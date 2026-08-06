package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStatus_String(t *testing.T) {
	tests := []struct {
		name string
		s    UserStatus
		want string
	}{
		{"active", UserStatusActive, "active"},
		{"pending", UserStatusPending, "pending"},
		{"inactive", UserStatusInactive, "inactive"},
		{"deleted", UserStatusDeleted, "deleted"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}

func TestUserStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       UserStatus
		want    []byte
		wantErr bool
	}{
		{"active", UserStatusActive, []byte("active"), false},
		{"pending", UserStatusPending, []byte("pending"), false},
		{"inactive", UserStatusInactive, []byte("inactive"), false},
		{"deleted", UserStatusDeleted, []byte("deleted"), false},
		{"status high", UserStatus(255), nil, true},
		{"status low", UserStatus(0), nil, true},
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

func TestUserStatus_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       *UserStatus
		text    []byte
		want    UserStatus
		wantErr bool
	}{
		{"active", new(UserStatus), []byte("active"), UserStatusActive, false},
		{"pending", new(UserStatus), []byte("pending"), UserStatusPending, false},
		{"inactive", new(UserStatus), []byte("inactive"), UserStatusInactive, false},
		{"deleted", new(UserStatus), []byte("deleted"), UserStatusDeleted, false},
		{"status high", new(UserStatus), []byte("100"), UserStatus(0), true},
		{"status low", new(UserStatus), []byte("0"), UserStatus(0), true},
		{"status invalid", new(UserStatus), []byte("invalid"), UserStatus(0), true},
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
