package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config contiene toda la configuración del servicio.
// Todos los campos son tipados — nunca map[string]interface{}.
type Config struct {
	Service       ServiceConfig       `mapstructure:"service"`
	Server        ServerConfig        `mapstructure:"server"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
}

type ServiceConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

type ServerConfig struct {
	HTTPPort        int           `mapstructure:"http_port"`
	GRPCPort        int           `mapstructure:"grpc_port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type ObservabilityConfig struct {
	LogLevel string        `mapstructure:"log_level"`
	Tracing  TracingConfig `mapstructure:"tracing"`
	Metrics  MetricsConfig `mapstructure:"metrics"`
}

type TracingConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	ServiceName    string `mapstructure:"service_name"`
	ServiceVersion string `mapstructure:"service_version"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"` // e.g., "/metrics"
}

// KafkaConfig configura el transporte Kafka (opt-in).
// Si Brokers está vacío, el consumer/producer no se inicializan.
type KafkaConfig struct {
	Brokers          []string      `mapstructure:"brokers"`
	ConsumerGroup    string        `mapstructure:"consumer_group"`
	Topics           []string      `mapstructure:"topics"`
	DLTSuffix        string        `mapstructure:"dlt_suffix"`
	MaxRetries       int           `mapstructure:"max_retries"`
	SessionTimeout   time.Duration `mapstructure:"session_timeout"`
	RebalanceTimeout time.Duration `mapstructure:"rebalance_timeout"`
	CircuitBreaker   CBConfig      `mapstructure:"circuit_breaker"`
}

type CBConfig struct {
	MaxFails int           `mapstructure:"max_fails"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

// Load carga configuración desde archivo YAML + variables de entorno.
//
// Las variables de entorno sobreescriben el archivo YAML.
// Formato: APP_SERVER_HTTP_PORT=9090 → server.http_port
func Load(path string) (*Config, error) {
	v := viper.New()

	// Valores por defecto seguros
	setDefaults(v)

	// Archivo de configuración
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Variables de entorno: APP_SERVER_HTTP_PORT → server.http_port
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// No es error fatal si el archivo no existe (usamos defaults + env)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file %q: %w", path, err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate verifica que los campos críticos estén presentes.
func (c *Config) validate() error {
	if c.Service.Name == "" {
		return fmt.Errorf("service.name is required")
	}
	if c.Server.HTTPPort == 0 {
		return fmt.Errorf("server.http_port is required")
	}
	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("service.version", "dev")
	v.SetDefault("service.environment", "local")
	v.SetDefault("server.http_port", 8080)
	v.SetDefault("server.grpc_port", 50051)
	v.SetDefault("server.shutdown_timeout", 30*time.Second)
	v.SetDefault("server.read_timeout", 10*time.Second)
	v.SetDefault("server.write_timeout", 10*time.Second)
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("observability.log_level", "info")

	// Kafka: brokers vacío = transporte desactivado (opt-in)
	v.SetDefault("kafka.consumer_group", "go-without-magic")
	v.SetDefault("kafka.dlt_suffix", ".dlt")
	v.SetDefault("kafka.max_retries", 3)
	v.SetDefault("kafka.session_timeout", 10*time.Second)
	v.SetDefault("kafka.rebalance_timeout", 60*time.Second)
	v.SetDefault("kafka.circuit_breaker.max_fails", 5)
	v.SetDefault("kafka.circuit_breaker.timeout", 30*time.Second)
}
