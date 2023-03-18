package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefault(t *testing.T) {
	type args struct {
		value    int
		fallback int
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "should return the value if it is not the zero value of the type",
			args: args{
				value:    1,
				fallback: 0,
			},
			want: 1,
		},
		{
			name: "should return the fallback if the value is the zero value of the type",
			args: args{
				value:    0,
				fallback: 1,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, GetDefault(tt.args.value, tt.args.fallback))
		})
	}
}

func TestGetDefaultPtr(t *testing.T) {
	type args struct {
		value    *int
		fallback int
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "should return the value if it is not nil",
			args: args{
				value: func() *int {
					v := 1
					return &v
				}(),
				fallback: 0,
			},
			want: 1,
		},
		{
			name: "should return the fallback if the value is nil",
			args: args{
				value:    nil,
				fallback: 1,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, GetDefaultPtr(tt.args.value, tt.args.fallback))
		})
	}
}
