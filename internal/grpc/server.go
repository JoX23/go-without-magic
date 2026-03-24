package grpc

import (
	"context"
	"net"

	"go.uber.org/zap"
	grpcpkg "google.golang.org/grpc"
)

// GRPCServerAdapter hace que grpc.Server implemente shutdown.Server.
type GRPCServerAdapter struct {
	Server *grpcpkg.Server
}

func (g *GRPCServerAdapter) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		g.Server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		// timeout: forzar stop
		g.Server.Stop()
		return ctx.Err()
	}
}

// StartGRPCServer arranca gRPC en el puerto dado y retorna el server.
func StartGRPCServer(address string, logger *zap.Logger, register func(*grpcpkg.Server)) (*grpcpkg.Server, error) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	grpcServer := grpcpkg.NewServer(grpcpkg.UnaryInterceptor(UnaryServerInterceptor(logger)))
	register(grpcServer)

	go func() {
		logger.Info("gRPC server listening", zap.String("addr", address))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	return grpcServer, nil
}
