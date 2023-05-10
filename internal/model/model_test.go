package model

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestID_String(t *testing.T) {
	type fields struct {
		inner xid.ID
		label ResourceType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "empty",
			fields: fields{
				inner: xid.NilID(),
				label: ResourceType(0),
			},
			want: xid.NilID().String(),
		},
		{
			name: "with ID",
			fields: fields{
				inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc},
				label: ResourceType(0),
			},
			want: "041061050o3gg28a1c60",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id := ID{
				Inner: tt.fields.inner,
				Type:  tt.fields.label,
			}
			assert.Equal(t, tt.want, id.String())
		})
	}
}

func TestID_Label(t *testing.T) {
	type fields struct {
		inner xid.ID
		label ResourceType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "empty",
			fields: fields{
				inner: xid.NilID(),
				label: ResourceType(0),
			},
			want: "",
		},
		{
			name: "with Type",
			fields: fields{
				inner: xid.NilID(),
				label: ResourceTypeAssignment,
			},
			want: ResourceTypeAssignment.String(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id := ID{
				Inner: tt.fields.inner,
				Type:  tt.fields.label,
			}
			assert.Equal(t, tt.want, id.Label())
		})
	}
}

func TestNewID(t *testing.T) {
	type args struct {
		typ ResourceType
	}
	tests := []struct {
		name     string
		args     args
		wantType ResourceType
		wantErr  error
	}{
		{
			name: "empty",
			args: args{
				typ: ResourceType(0),
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "too long",
			args: args{
				typ: ResourceType(100),
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "valid",
			args: args{
				typ: ResourceTypeAssignment,
			},
			wantType: ResourceTypeAssignment,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewID(tt.args.typ)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.NotEqual(t, xid.NilID(), got.Inner)
				assert.Equal(t, tt.wantType, got.Type)
			}
		})
	}
}

func TestMustNewID(t *testing.T) {
	type args struct {
		typ ResourceType
	}
	tests := []struct {
		name     string
		args     args
		wantType ResourceType
		panics   bool
	}{
		{
			name: "empty",
			args: args{
				typ: ResourceType(0),
			},
			panics: true,
		},
		{
			name: "too long",
			args: args{
				typ: ResourceType(100),
			},
			panics: true,
		},
		{
			name: "valid",
			args: args{
				typ: ResourceTypeAssignment,
			},
			wantType: ResourceTypeAssignment,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.panics {
				assert.Panics(t, func() {
					MustNewID(tt.args.typ)
				})
			} else {
				got := MustNewID(tt.args.typ)
				assert.Equal(t, tt.wantType, got.Type)
				assert.NotNil(t, got.Inner)
			}
		})
	}
}

func TestNewNilID(t *testing.T) {
	type args struct {
		typ ResourceType
	}
	tests := []struct {
		name    string
		args    args
		want    ID
		wantErr error
	}{
		{
			name: "empty",
			args: args{
				typ: ResourceType(0),
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "too long",
			args: args{
				typ: ResourceType(100),
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "valid",
			args: args{
				typ: ResourceTypeAssignment,
			},
			want: ID{
				Inner: xid.NilID(),
				Type:  ResourceTypeAssignment,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewNilID(tt.args.typ)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMustNewNilID(t *testing.T) {
	type args struct {
		typ ResourceType
	}
	tests := []struct {
		name   string
		args   args
		want   ID
		panics bool
	}{
		{
			name: "empty",
			args: args{
				typ: ResourceType(0),
			},
			panics: true,
		},
		{
			name: "too long",
			args: args{
				typ: ResourceType(100),
			},
			panics: true,
		},
		{
			name: "valid",
			args: args{
				typ: ResourceTypeAssignment,
			},
			want: ID{
				Inner: xid.NilID(),
				Type:  ResourceTypeAssignment,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.panics {
				assert.Panics(t, func() {
					MustNewNilID(tt.args.typ)
				})
			} else {
				got := MustNewNilID(tt.args.typ)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNewIDFromString(t *testing.T) {
	type args struct {
		id  string
		typ ResourceType
	}
	tests := []struct {
		name    string
		args    args
		want    ID
		wantErr error
	}{
		{
			name: "empty",
			args: args{
				id:  "",
				typ: ResourceType(0),
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "id too long",
			args: args{
				id:  "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
				typ: ResourceTypeAssignment,
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "type too long",
			args: args{
				id:  "041061050o3gg28a1c60",
				typ: ResourceType(100),
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "valid",
			args: args{
				id:  "041061050o3gg28a1c60",
				typ: ResourceTypeAssignment,
			},
			want: ID{
				Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc},
				Type:  ResourceTypeAssignment,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewIDFromString(tt.args.id, tt.args.typ.String())
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestID_IsNil(t *testing.T) {
	type fields struct {
		inner xid.ID
		typ   ResourceType
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "nil",
			fields: fields{
				inner: xid.NilID(),
				typ:   ResourceTypeAssignment,
			},
			want: true,
		},
		{
			name: "not nil",
			fields: fields{
				inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc},
				typ:   ResourceTypeAssignment,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id := ID{
				Inner: tt.fields.inner,
				Type:  tt.fields.typ,
			}
			assert.Equal(t, tt.want, id.IsNil())
		})
	}
}
