package model

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRole(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *Role
		wantErr error
	}{
		{
			name: "create role",
			args: args{
				name: "admin",
			},
			want: &Role{
				ID:          ID{inner: xid.NilID(), label: RoleIDType},
				Name:        "admin",
				Members:     make([]ID, 0),
				Permissions: make([]ID, 0),
			},
		},
		{
			name: "create role with empty name",
			args: args{
				name: "",
			},
			wantErr: ErrInvalidRoleDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewRole(tt.args.name)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestRole_Validate(t *testing.T) {
	type fields struct {
		ID          ID
		Name        string
		Description string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "validate role with valid details",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: RoleIDType},
				Name:        "test",
				Description: "test description",
			},
		},
		{
			name: "validate role with invalid ID",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ""},
				Name:        "test",
				Description: "test description",
			},
			wantErr: ErrInvalidRoleDetails,
		},
		{
			name: "validate role with invalid name",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: RoleIDType},
				Name:        "t",
				Description: "test description",
			},
			wantErr: ErrInvalidRoleDetails,
		},
		{
			name: "validate role with empty name",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: RoleIDType},
				Name:        "",
				Description: "test description",
			},
			wantErr: ErrInvalidRoleDetails,
		},
		{
			name: "validate role with invalid description",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: RoleIDType},
				Name:        "test",
				Description: "t",
			},
			wantErr: ErrInvalidRoleDetails,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Role{
				ID:          tt.fields.ID,
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
			}
			require.ErrorIs(t, r.Validate(), tt.wantErr)
		})
	}
}
