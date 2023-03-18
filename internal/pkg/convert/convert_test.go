package convert

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToPointer(t *testing.T) {
	tests := []struct {
		name string
		src  any
		want any
	}{
		{
			name: "string to pointer",
			src:  "test",
			want: func() *string {
				v := "test"
				return &v
			}(),
		},
		{
			name: "int to pointer",
			src:  1,
			want: func() *int {
				v := 1
				return &v
			}(),
		},
		{
			name: "struct to pointer",
			src:  struct{ Name string }{Name: "test"},
			want: func() *struct{ Name string } {
				v := struct{ Name string }{Name: "test"}
				return &v
			}(),
		},
		{
			name: "slice to pointer",
			src:  []string{"test"},
			want: func() *[]string {
				v := []string{"test"}
				return &v
			}(),
		},
		{
			name: "map to pointer",
			src:  map[string]string{"test": "test"},
			want: func() *map[string]string {
				v := map[string]string{"test": "test"}
				return &v
			}(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ToPointer(tt.src)
			assert.Equal(t, reflect.ValueOf(tt.want).Elem().Interface(), reflect.ValueOf(got).Elem().Interface())
		})
	}
}

func TestConvertAnyToAny(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		output  any
		want    any
		wantErr bool
	}{
		{
			name:    "convert string to int",
			input:   "1",
			output:  ToPointer(1),
			wantErr: true,
		},
		{
			name:    "convert int to string",
			input:   1,
			output:  ToPointer("1"),
			wantErr: true,
		},
		{
			name: "convert map to struct",
			input: map[string]any{
				"foo": "bar",
			},
			output: &struct {
				Foo string `json:"foo"`
			}{
				Foo: "bar",
			},
			want: &struct {
				Foo string `json:"foo"`
			}{
				Foo: "bar",
			},
		},
		{
			name: "convert struct to map",
			input: &struct {
				Foo string `json:"foo"`
			}{
				Foo: "bar",
			},
			output: ToPointer(map[string]any{
				"foo": "bar",
			}),
			want: map[string]any{
				"foo": "bar",
			},
		},
		{
			name: "convert to non-pointer",
			input: &struct {
				Foo string `json:"foo"`
			}{
				Foo: "bar",
			},
			output: map[string]any{
				"foo": "bar",
			},
			wantErr: true,
		},
		{
			name: "convert to nil",
			input: map[string]string{
				"foo": "bar",
			},
			output: nil,
			want:   nil,
		},
		{
			name:    "convert invalid input",
			input:   func() {},
			output:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := AnyToAny(tt.input, tt.output)
			if (err != nil) != tt.wantErr {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMustConvertAnyToAny(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		output  any
		want    any
		wantErr bool
	}{
		{
			name:    "convert string to int",
			input:   "1",
			output:  ToPointer(1),
			wantErr: true,
		},
		{
			name:    "convert int to string",
			input:   1,
			output:  ToPointer("1"),
			wantErr: true,
		},
		{
			name: "convert map to struct",
			input: map[string]any{
				"foo": "bar",
			},
			output: &struct {
				Foo string `json:"foo"`
			}{
				Foo: "bar",
			},
			want: &struct {
				Foo string `json:"foo"`
			}{
				Foo: "bar",
			},
		},
		{
			name: "convert struct to map",
			input: &struct {
				Foo string `json:"foo"`
			}{
				Foo: "bar",
			},
			output: ToPointer(map[string]any{
				"foo": "bar",
			}),
			want: map[string]any{
				"foo": "bar",
			},
		},
		{
			name: "convert to non-pointer",
			input: &struct {
				Foo string `json:"foo"`
			}{
				Foo: "bar",
			},
			output: map[string]any{
				"foo": "bar",
			},
			wantErr: true,
		},
		{
			name: "convert to nil",
			input: map[string]string{
				"foo": "bar",
			},
			output: nil,
			want:   nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantErr {
				require.Panics(t, func() {
					MustAnyToAny(tt.input, tt.output)
				})
				return
			}

			MustAnyToAny(tt.input, tt.output)
		})
	}
}
