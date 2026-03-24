package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"go.uber.org/zap"
	grpcpkg "google.golang.org/grpc"

	"github.com/JoX23/go-without-magic/internal/config"
	grpcservice "github.com/JoX23/go-without-magic/internal/grpc"
	"github.com/JoX23/go-without-magic/internal/grpc/pb"
	httphandler "github.com/JoX23/go-without-magic/internal/handler/http"
	"github.com/JoX23/go-without-magic/internal/middleware"
	"github.com/JoX23/go-without-magic/internal/observability"
	"github.com/JoX23/go-without-magic/internal/repository/memory"
	"github.com/JoX23/go-without-magic/internal/service"
	"github.com/JoX23/go-without-magic/pkg/health"
	"github.com/JoX23/go-without-magic/pkg/shutdown"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: %v\n", err)
		os.Exit(1)
	}
}

// run separa la lógica de main para poder retornar errores
// y testear el arranque sin llamar a os.Exit directamente.
func run() error {
	// ── 1. Configuración ───────────────────────────────────────────────
	cfg, err := config.Load("config.yaml")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// ── 2. Logger ─────────────────────────────────────────────────────
	logger, err := observability.NewLogger(
		cfg.Observability.LogLevel,
		cfg.Service.Environment,
	)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer logger.Sync() //nolint:errcheck

	// ── 3. Observability (Tracing & Metrics) ──────────────────────────
	var tracerProvider *observability.TracerProvider
	var metrics *observability.Metrics
	var spanProcessor *observability.SpanProcessor

	if cfg.Observability.Tracing.Enabled {
		tracerProvider, err = observability.NewTracerProvider(
			cfg.Observability.Tracing.ServiceName,
			cfg.Observability.Tracing.ServiceVersion,
		)
		if err != nil {
			return fmt.Errorf("creating tracer provider: %w", err)
		}
		defer tracerProvider.Shutdown(context.Background()) //nolint:errcheck

		spanProcessor = observability.NewSpanProcessor()
		logger.Info("tracing enabled", zap.String("service", cfg.Observability.Tracing.ServiceName))
	}

	if cfg.Observability.Metrics.Enabled {
		metrics = observability.NewMetrics()
		logger.Info("metrics enabled", zap.String("path", cfg.Observability.Metrics.Path))
	}

	logger.Info("starting service",
		zap.String("name", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
		zap.String("environment", cfg.Service.Environment),
	)

	// ── 4. Repositorio ─────────────────────────────────────────────────
	// En local usamos memoria; para producción cambia por postgres.New(cfg.Database)
	repo := memory.NewUserRepository()
	productRepo := memory.NewProductRepository()

	// ── 5. Capa de servicio ────────────────────────────────────────────
	userSvc := service.NewUserService(repo, logger)
	productSvc := service.NewProductService(productRepo, logger)

	// ── 6. HTTP Handler ────────────────────────────────────────────────
	userHandler := httphandler.NewUserHandler(userSvc, logger)
	productHandler := httphandler.NewProductHandler(productSvc, logger)

	mux := http.NewServeMux()

	// Rutas de negocio
	userHandler.RegisterRoutes(mux)
	productHandler.RegisterRoutes(mux)

	// Rutas de infraestructura
	// Sin checkers reales en modo memoria — agregar repo cuando uses postgres
	mux.Handle("/healthz", health.NewHandler())

	// Rutas de observabilidad
	if cfg.Observability.Metrics.Enabled {
		mux.Handle(cfg.Observability.Metrics.Path, metrics.Handler())
	}

	// ── 7. Middlewares ─────────────────────────────────────────────────
	var handler http.Handler = mux

	// Aplicar middlewares en orden (de afuera hacia adentro)
	middlewares := []middleware.Middleware{
		middleware.RecoveryPanic(logger),
		middleware.RequestID(),
		middleware.Logging(logger),
	}

	// Agregar tracing si está habilitado
	if cfg.Observability.Tracing.Enabled {
		tracer := tracerProvider.Tracer("http")
		middlewares = append(middlewares, middleware.Tracing(tracer, spanProcessor, metrics))
	}

	// Agregar métricas de negocio si están habilitadas
	if cfg.Observability.Metrics.Enabled {
		middlewares = append(middlewares, middleware.BusinessMetrics(metrics, spanProcessor))
	}

	handler = middleware.Chain(middlewares...)(handler)

	// ── 8. Servidor HTTP ───────────────────────────────────────────────
	addr := fmt.Sprintf(":%d", cfg.Server.HTTPPort)

	// Chequear que el puerto esté disponible ANTES de crear el servidor
	// Esto nos da error temprano en lugar de una goroutine silenciosa
	httpLis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("cannot bind to %s: %w", addr, err)
	}

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      handler, // Usar handler con middlewares aplicados
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Canal para reportar errores del servidor
	serverErrors := make(chan error, 2)

	// Arrancar servidor HTTP en goroutine
	go func() {
		logger.Info("HTTP server listening", zap.String("addr", addr))
		serverErrors <- httpServer.Serve(httpLis)
	}()

	var grpcServer *grpcpkg.Server
	if cfg.Server.GRPCPort > 0 {
		grpcAddr := fmt.Sprintf(":%d", cfg.Server.GRPCPort)
		grpcLis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			return fmt.Errorf("cannot bind to %s: %w", grpcAddr, err)
		}

		grpcServer = grpcpkg.NewServer(grpcpkg.UnaryInterceptor(grpcservice.UnaryServerInterceptor(logger)))
		pb.RegisterUserServiceServer(grpcServer, grpcservice.NewUserServiceServer(userSvc, logger))

		go func() {
			logger.Info("gRPC server listening", zap.String("addr", grpcAddr))
			serverErrors <- grpcServer.Serve(grpcLis)
		}()
	}

	// ── 9. Graceful Shutdown ───────────────────────────────────────────
	shutdownMgr := shutdown.NewManager(cfg.Server.ShutdownTimeout, logger).
		Register("http", httpServer)
	if grpcServer != nil {
		shutdownMgr.Register("grpc", &grpcservice.GRPCServerAdapter{Server: grpcServer})
	}

	// Iniciar el signal handler en goroutine (no bloquea)
	go shutdownMgr.Wait()

	// Esperar: error del servidor O finalización del shutdown
	// Si el servidor falla al startup, retornamos el error inmediatamente
	err = <-serverErrors
	if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, grpcpkg.ErrServerStopped) {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
