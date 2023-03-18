package config

import (
	"errors"
	"time"
)

var (
	ErrNoConfig = errors.New("no config provided") // no configuration provided
)

// LicenseConfig is the configuration for the license.
type LicenseConfig struct {
	File string `mapstructure:"file"`
}

// LogConfig is the configuration for the logger.
type LogConfig struct {
	Level string `mapstructure:"level"`
}

// DatabaseConfig is the configuration for the database.
type DatabaseConfig struct {
	URL                          string        `mapstructure:"url"`
	Username                     string        `mapstructure:"username"`
	Password                     string        `mapstructure:"password"`
	Name                         string        `mapstructure:"name"`
	MaxTransactionRetryTime      time.Duration `mapstructure:"max_transaction_retry_time"`
	MaxConnectionPoolSize        int           `mapstructure:"max_connection_pool_size"`
	MaxConnectionLifetime        time.Duration `mapstructure:"max_connection_lifetime"`
	ConnectionAcquisitionTimeout time.Duration `mapstructure:"connection_acquisition_timeout"`
	SocketConnectTimeout         time.Duration `mapstructure:"socket_connect_timeout"`
	SocketKeepalive              bool          `mapstructure:"socket_keepalive"`
	FetchSize                    int           `mapstructure:"fetch_size"`
}

// TLSConfig is the configuration for TLS.
type TLSConfig struct {
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// CORSConfig represents the CORS configuration.
type CORSConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// ServerConfig is the configuration for the HTTP server.
type ServerConfig struct {
	Address      string        `mapstructure:"address"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	CORS         CORSConfig    `mapstructure:"cors"`
}

// TracingConfig is the configuration for the tracing.
type TracingConfig struct {
	ServiceName       string  `mapstructure:"service_name"`
	CollectorEndpoint string  `mapstructure:"collector_endpoint"`
	TraceRatio        float64 `mapstructure:"trace_ratio"`
}

// Config is the combined configuration for the service.
type Config struct {
	Log           LogConfig      `mapstructure:"log"`
	License       LicenseConfig  `mapstructure:"license"`
	Server        ServerConfig   `mapstructure:"server"`
	MetricsServer ServerConfig   `mapstructure:"metrics_server"`
	Database      DatabaseConfig `mapstructure:"database"`
	TLS           TLSConfig      `mapstructure:"tls"`
	Tracing       TracingConfig  `mapstructure:"tracing"`
}
