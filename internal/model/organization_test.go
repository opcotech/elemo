package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationStatus_String(t *testing.T) {
	tests := []struct {
		name string
		s    OrganizationStatus
		want string
	}{
		{
			name: "organization status active",
			s:    OrganizationStatusActive,
			want: "active",
		},
		{
			name: "organization status deleted",
			s:    OrganizationStatusDeleted,
			want: "deleted",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}

func TestOrganizationStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       OrganizationStatus
		want    []byte
		wantErr bool
	}{
		{"active", OrganizationStatusActive, []byte("active"), false},
		{"deleted", OrganizationStatusDeleted, []byte("deleted"), false},
		{"status high", OrganizationStatus(255), nil, true},
		{"status low", OrganizationStatus(0), nil, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.s.MarshalText()
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOrganizationStatus_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       *OrganizationStatus
		text    []byte
		want    OrganizationStatus
		wantErr bool
	}{
		{"active", new(OrganizationStatus), []byte("active"), OrganizationStatusActive, false},
		{"deleted", new(OrganizationStatus), []byte("deleted"), OrganizationStatusDeleted, false},
		{"status high", new(OrganizationStatus), []byte("100"), OrganizationStatus(0), true},
		{"status low", new(OrganizationStatus), []byte("0"), OrganizationStatus(0), true},
		{"status invalid", new(OrganizationStatus), []byte("invalid"), OrganizationStatus(0), true},
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
