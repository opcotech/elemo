package model

import "errors"

const (
	HealthStatusUnknown   HealthStatus = iota // component is in an unknown state
	HealthStatusHealthy                       // component is healthy
	HealthStatusUnhealthy                     // component is unhealthy
)

var (
	ErrInvalidHealthStatus = errors.New("invalid health status") // health status is invalid

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
