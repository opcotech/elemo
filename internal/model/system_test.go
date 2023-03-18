package model

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthStatus_String(t *testing.T) {
	tests := []struct {
		name string
		s    HealthStatus
		want string
	}{
		{"health status unknown", HealthStatusUnknown, "unknown"},
		{"health status healthy", HealthStatusHealthy, "healthy"},
		{"health status unhealthy", HealthStatusUnhealthy, "unhealthy"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}

func TestHealthStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       HealthStatus
		want    []byte
		wantErr bool
	}{
		{"health status unknown", HealthStatusUnknown, []byte("unknown"), false},
		{"health status healthy", HealthStatusHealthy, []byte("healthy"), false},
		{"health status unhealthy", HealthStatusUnhealthy, []byte("unhealthy"), false},
		{"status high", HealthStatus(255), nil, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotText, err := tt.s.MarshalText()
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want, gotText)
			}
		})
	}
}

func TestHealthStatus_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       *HealthStatus
		text    []byte
		want    HealthStatus
		wantErr bool
	}{
		{"health status unknown", new(HealthStatus), []byte("unknown"), HealthStatusUnknown, false},
		{"health status healthy", new(HealthStatus), []byte("healthy"), HealthStatusHealthy, false},
		{"health status unhealthy", new(HealthStatus), []byte("unhealthy"), HealthStatusUnhealthy, false},
		{"status high", new(HealthStatus), []byte("high"), HealthStatusUnknown, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.s.UnmarshalText(tt.text); (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateVersionInfo(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		v       *VersionInfo
		wantErr bool
	}{
		{
			name: "valid",
			v: &VersionInfo{
				Version:   "1.0.0",
				Commit:    "1234567",
				Date:      "2020-01-01T00:00:00Z",
				GoVersion: "1.13.5",
			},
			wantErr: false,
		},
		{
			name: "invalid version",
			v: &VersionInfo{
				Version:   "1.0",
				Commit:    "1234567",
				Date:      "2020-01-01T00:00:00Z",
				GoVersion: "1.13.5",
			},
			wantErr: true,
		},
		{
			name: "invalid commit",
			v: &VersionInfo{
				Version:   "1.0.0",
				Commit:    "12345678",
				Date:      "2020-01-01T00:00:00Z",
				GoVersion: "1.13.5",
			},
			wantErr: true,
		},
		{
			name: "invalid date",
			v: &VersionInfo{
				Version:   "1.0.0",
				Commit:    "1234567",
				Date:      "2020-01-01",
				GoVersion: "1.13.5",
			},
			wantErr: true,
		},
		{
			name: "invalid go version",
			v: &VersionInfo{
				Version:   "1.0.0",
				Commit:    "1234567",
				Date:      "2020-01-01T00:00:00Z",
				GoVersion: "1.13",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validate.Struct(tt.v)
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}
		})
	}
}
