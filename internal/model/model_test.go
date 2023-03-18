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
		label string
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
				label: "",
			},
			want: xid.NilID().String(),
		},
		{
			name: "with ID",
			fields: fields{
				inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc},
				label: "",
			},
			want: "041061050o3gg28a1c60",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id := ID{
				inner: tt.fields.inner,
				label: tt.fields.label,
			}
			assert.Equal(t, tt.want, id.String())
		})
	}
}

func TestID_Label(t *testing.T) {
	type fields struct {
		inner xid.ID
		label string
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
				label: "",
			},
			want: "",
		},
		{
			name: "with label",
			fields: fields{
				inner: xid.NilID(),
				label: "test",
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id := ID{
				inner: tt.fields.inner,
				label: tt.fields.label,
			}
			assert.Equalf(t, tt.want, id.Label(), "Label()")
		})
	}
}

func TestNewID(t *testing.T) {
	type args struct {
		typ string
	}
	tests := []struct {
		name     string
		args     args
		wantType string
		wantErr  error
	}{
		{
			name: "empty",
			args: args{
				typ: "",
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "too short",
			args: args{
				typ: "abc",
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "too long",
			args: args{
				typ: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "valid",
			args: args{
				typ: "test",
			},
			wantType: "test",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewID(tt.args.typ)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.NotEqual(t, xid.NilID(), got.inner)
				assert.Equal(t, tt.wantType, got.label)
			}
		})
	}
}

func TestMustNewID(t *testing.T) {
	type args struct {
		typ string
	}
	tests := []struct {
		name     string
		args     args
		wantType string
		panics   bool
	}{
		{
			name: "empty",
			args: args{
				typ: "",
			},
			panics: true,
		},
		{
			name: "too short",
			args: args{
				typ: "abc",
			},
			panics: true,
		},
		{
			name: "too long",
			args: args{
				typ: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
			},
			panics: true,
		},
		{
			name: "valid",
			args: args{
				typ: "test",
			},
			wantType: "test",
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
				assert.Equal(t, tt.wantType, got.label)
				assert.NotNil(t, got.inner)
			}
		})
	}
}

func TestNewNilID(t *testing.T) {
	type args struct {
		typ string
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
				typ: "",
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "too short",
			args: args{
				typ: "abc",
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "too long",
			args: args{
				typ: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "valid",
			args: args{
				typ: "test",
			},
			want: ID{
				inner: xid.NilID(),
				label: "test",
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
		typ string
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
				typ: "",
			},
			panics: true,
		},
		{
			name: "too short",
			args: args{
				typ: "abc",
			},
			panics: true,
		},
		{
			name: "too long",
			args: args{
				typ: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
			},
			panics: true,
		},
		{
			name: "valid",
			args: args{
				typ: "test",
			},
			want: ID{
				inner: xid.NilID(),
				label: "test",
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
		typ string
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
				typ: "",
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "id too short",
			args: args{
				id:  "abc",
				typ: "abcd",
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "id too long",
			args: args{
				id:  "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
				typ: "abcd",
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "type too short",
			args: args{
				id:  "041061050o3gg28a1c60",
				typ: "ab",
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "type too long",
			args: args{
				id:  "041061050o3gg28a1c60",
				typ: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "valid",
			args: args{
				id:  "041061050o3gg28a1c60",
				typ: "test",
			},
			want: ID{
				inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc},
				label: "test",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewIDFromString(tt.args.id, tt.args.typ)
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
		Type  string
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
				Type:  "abcd",
			},
			want: true,
		},
		{
			name: "not nil",
			fields: fields{
				inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc},
				Type:  "abcd",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id := ID{
				inner: tt.fields.inner,
				label: tt.fields.Type,
			}
			assert.Equal(t, tt.want, id.IsNil())
		})
	}
}
