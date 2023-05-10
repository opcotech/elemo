package model

const (
	HealthCheckComponentCacheDB      HealthCheckComponent = "cache_database"      // cache database
	HealthCheckComponentGraphDB      HealthCheckComponent = "graph_database"      // graph database
	HealthCheckComponentRelationalDB HealthCheckComponent = "relational_database" // relational database
	HealthCheckComponentLicense      HealthCheckComponent = "license"             // license
)

const (
	HealthStatusUnknown   HealthStatus = iota // component is in an unknown state
	HealthStatusHealthy                       // component is healthy
	HealthStatusUnhealthy                     // component is unhealthy
)

var (
	healthStatusKeys = map[string]HealthStatus{
		"unknown":   HealthStatusUnknown,
		"healthy":   HealthStatusHealthy,
		"unhealthy": HealthStatusUnhealthy,
	}
	healthStatusValues = map[HealthStatus]string{
		HealthStatusUnknown:   "unknown",
		HealthStatusHealthy:   "healthy",
		HealthStatusUnhealthy: "unhealthy",
	}
)

// HealthCheckComponent represents a component of the application.
type HealthCheckComponent string

// HealthStatus is the status of a component.
type HealthStatus uint8

// String returns the string representation of the health status.
func (s HealthStatus) String() string {
	return healthStatusValues[s]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (s HealthStatus) MarshalText() (text []byte, err error) {
	if s > HealthStatusUnhealthy {
		return nil, ErrInvalidHealthStatus
	}
	return []byte(s.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (s *HealthStatus) UnmarshalText(text []byte) error {
	if v, ok := healthStatusKeys[string(text)]; ok {
		*s = v
		return nil
	}
	return ErrInvalidHealthStatus
}

// VersionInfo represents the version information of the application.
type VersionInfo struct {
	Version   string `validate:"required,semver"`
	Commit    string `validate:"required,alphanum,len=7"`
	Date      string `validate:"required"`
	GoVersion string `validate:"required,semver"`
}
