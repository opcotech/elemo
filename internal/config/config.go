package config

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrNoConfig = errors.New("no config provided") // no configuration provided
)

// TemplateConfig is the configuration for template files.
type TemplateConfig struct {
	Directory string `mapstructure:"directory"`
}

// SMTPConfig is the configuration for the SMTP server used for sending
// notification emails.
type SMTPConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	FromAddress string `mapstructure:"from_address"`
}

// LicenseConfig is the configuration for the license.
type LicenseConfig struct {
	File string `mapstructure:"file"`
}

// LogConfig is the configuration for the logger.
type LogConfig struct {
	Level string `mapstructure:"level"`
}

// RedisConfig is the configuration for the Redis database.
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Username     string        `mapstructure:"username"`
	Password     string        `mapstructure:"password"`
	Database     int           `mapstructure:"database"`
	IsSecure     bool          `mapstructure:"is_secure"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolSize     int           `mapstructure:"pool_size"`
}

// Address returns the address for the Redis database.
func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// CacheDatabaseConfig is the configuration for the cache database.
type CacheDatabaseConfig struct {
	RedisConfig           `mapstructure:",squash"`
	MaxIdleConnections    int           `mapstructure:"max_idle_connections"`
	MinIdleConnections    int           `mapstructure:"min_idle_connections"`
	ConnectionMaxIdleTime time.Duration `mapstructure:"connection_max_idle_time"`
	ConnectionMaxLifetime time.Duration `mapstructure:"connection_max_lifetime"`
}

// ConnectionURL returns the connection URL for the cache database.
func (c *CacheDatabaseConfig) ConnectionURL() string {
	if c.IsSecure {
		return fmt.Sprintf("redis://%s:%s@%s/%d?sslmode=require", c.Username, c.Password, c.Address(), c.Database)
	}

	return fmt.Sprintf("redis://%s:%s@%s:%d/%d?sslmode=disable", c.Username, c.Password, c.Host, c.Port, c.Database)
}

// GraphDatabaseConfig is the configuration for the graph database.
type GraphDatabaseConfig struct {
	Host                         string        `mapstructure:"host"`
	Port                         int           `mapstructure:"port"`
	Username                     string        `mapstructure:"username"`
	Password                     string        `mapstructure:"password"`
	Database                     string        `mapstructure:"database"`
	IsSecure                     bool          `mapstructure:"is_secure"`
	MaxTransactionRetryTime      time.Duration `mapstructure:"max_transaction_retry_time"`
	MaxConnectionPoolSize        int           `mapstructure:"max_connection_pool_size"`
	MaxConnectionLifetime        time.Duration `mapstructure:"max_connection_lifetime"`
	ConnectionAcquisitionTimeout time.Duration `mapstructure:"connection_acquisition_timeout"`
	SocketConnectTimeout         time.Duration `mapstructure:"socket_connect_timeout"`
	SocketKeepalive              bool          `mapstructure:"socket_keepalive"`
	FetchSize                    int           `mapstructure:"fetch_size"`
}

// ConnectionURL returns the connection URL for the graph database.
func (c *GraphDatabaseConfig) ConnectionURL() string {
	if c.IsSecure {
		return fmt.Sprintf("neo4j+s://%s:%d", c.Host, c.Port)
	}

	return fmt.Sprintf("neo4j://%s:%d", c.Host, c.Port)
}

// RelationalDatabaseConfig is the configuration for the relational database.
type RelationalDatabaseConfig struct {
	Host                  string        `mapstructure:"host"`
	Port                  int           `mapstructure:"port"`
	Username              string        `mapstructure:"username"`
	Password              string        `mapstructure:"password"`
	Database              string        `mapstructure:"database"`
	IsSecure              bool          `mapstructure:"is_secure"`
	MaxConnections        int           `mapstructure:"max_connections"`
	MaxConnectionLifetime time.Duration `mapstructure:"max_connection_lifetime"`
	MaxConnectionIdleTime time.Duration `mapstructure:"max_connection_idle_time"`
	MinConnections        int           `mapstructure:"min_connections"`
}

// ConnectionURL returns the connection URL for the relational database.
func (c *RelationalDatabaseConfig) ConnectionURL() string {
	if c.IsSecure {
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=require", c.Username, c.Password, c.Host, c.Port, c.Database)
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.Username, c.Password, c.Host, c.Port, c.Database)
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

// SessionConfig is the configuration for the session.
type SessionConfig struct {
	CookieName string `mapstructure:"cookie_name"`
	MaxAge     int    `mapstructure:"max_age"`
	Secure     bool   `mapstructure:"secure"`
}

// ServerConfig is the configuration for the HTTP server.
type ServerConfig struct {
	Address                string        `mapstructure:"address"`
	ReadTimeout            time.Duration `mapstructure:"read_timeout"`
	WriteTimeout           time.Duration `mapstructure:"write_timeout"`
	RequestThrottleLimit   int           `mapstructure:"request_throttle_limit"`
	RequestThrottleBacklog int           `mapstructure:"request_throttle_backlog"`
	RequestThrottleTimeout time.Duration `mapstructure:"request_throttle_timeout"`
	CORS                   CORSConfig    `mapstructure:"cors"`
	Session                SessionConfig `mapstructure:"session"`
	TLS                    TLSConfig     `mapstructure:"tls"`
}

// WorkerConfig is the configuration for the async worker.
type WorkerConfig struct {
	Concurrency              int           `mapstructure:"concurrency"`
	StrictPriority           bool          `mapstructure:"strict_priority"`
	ShutdownTimeout          time.Duration `mapstructure:"shutdown_timeout"`
	HealthCheckInterval      time.Duration `mapstructure:"health_check_interval"`
	DelayedTaskCheckInterval time.Duration `mapstructure:"delayed_task_check_interval"`
	GroupGracePeriod         time.Duration `mapstructure:"group_grace_period"`
	GroupMaxDelay            time.Duration `mapstructure:"group_max_delay"`
	GroupMaxSize             int           `mapstructure:"group_max_size"`
	LogLevel                 string        `mapstructure:"log_level"`
	RateLimit                float64       `mapstructure:"rate_limit"`
	RateLimitBurst           int           `mapstructure:"rate_limit_burst"`
	Broker                   RedisConfig   `mapstructure:"broker"`
}

// TracingConfig is the configuration for the tracing.
type TracingConfig struct {
	ServiceName       string  `mapstructure:"service_name"`
	CollectorEndpoint string  `mapstructure:"collector_endpoint"`
	TraceRatio        float64 `mapstructure:"trace_ratio"`
}

// Config is the combined configuration for the service.
type Config struct {
	Log                 LogConfig                `mapstructure:"log"`
	License             LicenseConfig            `mapstructure:"license"`
	Server              ServerConfig             `mapstructure:"server"`
	MetricsServer       ServerConfig             `mapstructure:"metrics_server"`
	Worker              WorkerConfig             `mapstructure:"worker"`
	WorkerMetricsServer ServerConfig             `mapstructure:"worker_metrics_server"`
	GraphDatabase       GraphDatabaseConfig      `mapstructure:"graph_database"`
	RelationalDatabase  RelationalDatabaseConfig `mapstructure:"relational_database"`
	CacheDatabase       CacheDatabaseConfig      `mapstructure:"cache_database"`
	Tracing             TracingConfig            `mapstructure:"tracing"`
	SMTP                SMTPConfig               `mapstructure:"smtp"`
	Template            TemplateConfig           `mapstructure:"template"`
}
