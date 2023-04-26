package model

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLabel(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *Label
		wantErr error
	}{
		{
			name: "create Type with valid details",
			args: args{
				name: "test",
			},
			want: &Label{
				ID:   ID{Inner: xid.NilID(), Type: ResourceTypeLabel},
				Name: "test",
			},
		},
		{
			name: "create Type with invalid name",
			args: args{
				name: "t",
			},
			wantErr: ErrInvalidLabelDetails,
		},
		{
			name: "create Type with empty name",
			args: args{
				name: "",
			},
			wantErr: ErrInvalidLabelDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewLabel(tt.args.name)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestLabel_Validate(t *testing.T) {
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
			name: "validate Type with valid details",
			fields: fields{
				ID:   ID{Inner: xid.NilID(), Type: ResourceTypeLabel},
				Name: "test",
			},
		},
		{
			name: "validate Type with invalid ID",
			fields: fields{
				ID:   ID{},
				Name: "test",
			},
			wantErr: ErrInvalidLabelDetails,
		},
		{
			name: "validate Type with invalid name",
			fields: fields{
				ID:   ID{Inner: xid.NilID(), Type: ResourceTypeLabel},
				Name: "t",
			},
			wantErr: ErrInvalidLabelDetails,
		},
		{
			name: "validate Type with invalid description",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeLabel},
				Name:        "test",
				Description: "t",
			},
			wantErr: ErrInvalidLabelDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := &Label{
				ID:          tt.fields.ID,
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
			}
			err := l.Validate()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
