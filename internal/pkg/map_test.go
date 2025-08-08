package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name string
		args []map[string]any
		want map[string]any
	}{
		{
			name: "merge maps",
			args: []map[string]any{
				{"a": 1, "b": 2},
				{"b": 3, "c": 4},
			},
			want: map[string]any{
				"a": 1,
				"b": 3,
				"c": 4,
			},
		},
		{
			name: "merge maps with nil",
			args: []map[string]any{
				{"a": 1, "b": 2},
				{"b": 3, "c": 4},
				nil,
			},
			want: map[string]any{
				"a": 1,
				"b": 3,
				"c": 4,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, MergeMaps(tt.args...))
		})
	}
}
