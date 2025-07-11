package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderValidator(t *testing.T) {
	assert.NotNil(t, RenderValidator())
}

func TestStruct(t *testing.T) {
	type args struct {
		s any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "validate struct",
			args: args{
				s: struct {
					Name string `validate:"required"`
				}{
					Name: "test",
				},
			},
		},
		{
			name: "validate struct with error",
			args: args{
				s: struct {
					Name string `validate:"required"`
				}{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantErr, Struct(tt.args.s) != nil)
		})
	}
}

func TestVar(t *testing.T) {
	type args struct {
		field any
		tag   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "validate var",
			args: args{
				field: "test",
				tag:   "required",
			},
		},
		{
			name: "validate var with error",
			args: args{
				field: "",
				tag:   "required",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantErr, Var(tt.args.field, tt.args.tag) != nil)
		})
	}
}
