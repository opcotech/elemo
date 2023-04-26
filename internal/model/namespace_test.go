package model

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNamespace(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *Namespace
		wantErr error
	}{
		{
			name: "create namespace with valid details",
			args: args{
				name: "test",
			},
			want: &Namespace{
				ID:        ID{Inner: xid.NilID(), Type: ResourceTypeNamespace},
				Name:      "test",
				Projects:  make([]ID, 0),
				Documents: make([]ID, 0),
			},
		},
		{
			name: "create namespace with invalid name",
			args: args{
				name: "t",
			},
			wantErr: ErrInvalidNamespaceDetails,
		},
		{
			name: "create namespace with empty name",
			args: args{
				name: "",
			},
			wantErr: ErrInvalidNamespaceDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewNamespace(tt.args.name)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.want != nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNamespace_Validate(t *testing.T) {
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
			name: "validate namespace with valid details",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeNamespace},
				Name:        "test",
				Description: "test description",
			},
		},
		{
			name: "validate namespace with invalid ID",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceType(0)},
				Name:        "test",
				Description: "test description",
			},
			wantErr: ErrInvalidNamespaceDetails,
		},
		{
			name: "validate namespace with invalid name",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeNamespace},
				Name:        "t",
				Description: "test description",
			},
			wantErr: ErrInvalidNamespaceDetails,
		},
		{
			name: "validate namespace with empty name",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeNamespace},
				Name:        "",
				Description: "test description",
			},
			wantErr: ErrInvalidNamespaceDetails,
		},
		{
			name: "validate namespace with invalid description",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeNamespace},
				Name:        "test",
				Description: "t",
			},
			wantErr: ErrInvalidNamespaceDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := &Namespace{
				ID:          tt.fields.ID,
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
			}
			require.ErrorIs(t, n.Validate(), tt.wantErr)
		})
	}
}
