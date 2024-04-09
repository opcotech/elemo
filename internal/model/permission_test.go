package model

import (
	"testing"
	"time"

	"github.com/rs/xid"
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

func TestNewPermission(t *testing.T) {
	type args struct {
		subject ID
		target  ID
		kind    PermissionKind
	}
	tests := []struct {
		name    string
		args    args
		want    Permission
		wantErr error
	}{
		{
			name: "create new permission",
			args: args{
				subject: ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
				target:  ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xc, 0xb, 0xa}, Type: ResourceTypeRole},
				kind:    PermissionKindCreate,
			},
			want: Permission{
				ID:      ID{Inner: xid.NilID(), Type: ResourceTypePermission},
				Subject: ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
				Target:  ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xc, 0xb, 0xa}, Type: ResourceTypeRole},
				Kind:    PermissionKindCreate,
			},
		},
		{
			name: "create new permission with nil subject",
			args: args{
				subject: ID{},
				target:  ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xc, 0xb, 0xa}, Type: ResourceTypeRole},
				kind:    PermissionKindCreate,
			},
			wantErr: ErrInvalidPermissionDetails,
		},
		{
			name: "create new permission with nil target",
			args: args{
				subject: ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
				target:  ID{},
				kind:    PermissionKindCreate,
			},
			wantErr: ErrInvalidPermissionDetails,
		},
		{
			name: "create new permission with invalid kind",
			args: args{
				subject: ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
				target:  ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xc, 0xb, 0xa}, Type: ResourceTypeRole},
				kind:    PermissionKind(0),
			},
			wantErr: ErrInvalidPermissionDetails,
		},
		{
			name: "create new permission with equal subject and target",
			args: args{
				subject: ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
				target:  ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
				kind:    PermissionKindCreate,
			},
			wantErr: ErrInvalidPermissionDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			permission, err := NewPermission(tt.args.subject, tt.args.target, tt.args.kind)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, *permission)
			}
		})
	}
}

func TestPermission_Validate(t *testing.T) {
	type fields struct {
		ID        ID
		Kind      PermissionKind
		Subject   ID
		Target    ID
		CreatedAt *time.Time
		UpdatedAt *time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid permission",
			fields: fields{
				ID:        ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypePermission},
				Kind:      PermissionKindCreate,
				Subject:   ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
				Target:    ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xc, 0xb, 0xa}, Type: ResourceTypeRole},
				CreatedAt: &time.Time{},
				UpdatedAt: &time.Time{},
			},
		},
		{
			name: "invalid permission id",
			fields: fields{
				ID:        ID{Inner: xid.NilID(), Type: ResourceType(0)},
				Kind:      PermissionKindCreate,
				Subject:   ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
				Target:    ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xc, 0xb, 0xa}, Type: ResourceTypeRole},
				CreatedAt: &time.Time{},
				UpdatedAt: &time.Time{},
			},
			wantErr: ErrInvalidPermissionDetails,
		},
		{
			name: "invalid permission kind",
			fields: fields{
				ID:        ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypePermission},
				Kind:      PermissionKind(0),
				Subject:   ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
				Target:    ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xc, 0xb, 0xa}, Type: ResourceTypeRole},
				CreatedAt: &time.Time{},
				UpdatedAt: &time.Time{},
			},
			wantErr: ErrInvalidPermissionDetails,
		},
		{
			name: "invalid permission subject",
			fields: fields{
				ID:        ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypePermission},
				Kind:      PermissionKindCreate,
				Subject:   ID{Inner: xid.NilID(), Type: ResourceType(0)},
				Target:    ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xc, 0xb, 0xa}, Type: ResourceTypeRole},
				CreatedAt: &time.Time{},
				UpdatedAt: &time.Time{},
			},
			wantErr: ErrInvalidPermissionDetails,
		},
		{
			name: "invalid permission target",
			fields: fields{
				ID:        ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypePermission},
				Kind:      PermissionKindCreate,
				Subject:   ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
				Target:    ID{Inner: xid.NilID(), Type: ResourceType(0)},
				CreatedAt: &time.Time{},
				UpdatedAt: &time.Time{},
			},
			wantErr: ErrInvalidPermissionDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &Permission{
				ID:        tt.fields.ID,
				Kind:      tt.fields.Kind,
				Subject:   tt.fields.Subject,
				Target:    tt.fields.Target,
				CreatedAt: tt.fields.CreatedAt,
				UpdatedAt: tt.fields.UpdatedAt,
			}
			err := p.Validate()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
