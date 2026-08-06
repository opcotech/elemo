package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionKind_String(t *testing.T) {
	tests := []struct {
		name string
		s    PermissionKind
		want string
	}{
		{"*", PermissionKindAll, "*"},
		{"create", PermissionKindCreate, "create"},
		{"read", PermissionKindRead, "read"},
		{"write", PermissionKindWrite, "write"},
		{"delete", PermissionKindDelete, "delete"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}

func TestPermissionKind_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       PermissionKind
		want    []byte
		wantErr bool
	}{
		{"*", PermissionKindAll, []byte("*"), false},
		{"create", PermissionKindCreate, []byte("create"), false},
		{"read", PermissionKindRead, []byte("read"), false},
		{"write", PermissionKindWrite, []byte("write"), false},
		{"delete", PermissionKindDelete, []byte("delete"), false},
		{"kind low", PermissionKind(0), nil, true},
		{"kind high", PermissionKind(100), nil, true},
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

func TestPermissionKind_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       *PermissionKind
		text    []byte
		want    PermissionKind
		wantErr bool
	}{
		{"*", new(PermissionKind), []byte("*"), PermissionKindAll, false},
		{"create", new(PermissionKind), []byte("create"), PermissionKindCreate, false},
		{"read", new(PermissionKind), []byte("read"), PermissionKindRead, false},
		{"write", new(PermissionKind), []byte("write"), PermissionKindWrite, false},
		{"delete", new(PermissionKind), []byte("delete"), PermissionKindDelete, false},
		{"kind low", new(PermissionKind), []byte("0"), PermissionKind(10), true},
		{"kind high", new(PermissionKind), []byte("255"), PermissionKind(10), true},
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
