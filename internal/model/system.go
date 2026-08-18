package model

const (
	HealthCheckComponentCacheDB      HealthCheckComponent = "cache_database"      // cache database
	HealthCheckComponentGraphDB      HealthCheckComponent = "graph_database"      // graph database
	HealthCheckComponentRelationalDB HealthCheckComponent = "relational_database" // relational database
	HealthCheckComponentLicense      HealthCheckComponent = "license"             // license
	HealthCheckComponentMessageQueue HealthCheckComponent = "message_queue"       // message_queue
	HealthCheckComponentS3Storage    HealthCheckComponent = "s3_storage"          // s3 storage
)

const (
	HealthStatusUnknown   HealthStatus = iota // unknown
	HealthStatusHealthy                       // healthy
	HealthStatusUnhealthy                     // unhealthy
)

// HealthCheckComponent represents a component of the application.
type HealthCheckComponent string

// HealthStatus is the status of a component.
//
//go:generate go tool enumer -type=HealthStatus -text -transform=noop -linecomment -output=system_health_status_gen.go
type HealthStatus uint8

// VersionInfo represents the version information of the application.
type VersionInfo struct {
	Version   string `validate:"required,semver"`
	Commit    string `validate:"required,alphanum,len=7"`
	Date      string `validate:"required"`
	GoVersion string `validate:"required,semver"`
}
