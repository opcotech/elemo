package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoPriority_String(t *testing.T) {
	tests := []struct {
		name string
		p    TodoPriority
		want string
	}{
		{"normal", TodoPriorityNormal, "normal"},
		{"important", TodoPriorityImportant, "important"},
		{"urgent", TodoPriorityUrgent, "urgent"},
		{"critical", TodoPriorityCritical, "critical"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.p.String())
		})
	}
}

func TestTodoPriority_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		p        TodoPriority
		wantText []byte
		wantErr  bool
	}{
		{"normal", TodoPriorityNormal, []byte("normal"), false},
		{"important", TodoPriorityImportant, []byte("important"), false},
		{"urgent", TodoPriorityUrgent, []byte("urgent"), false},
		{"critical", TodoPriorityCritical, []byte("critical"), false},
		{"status high", TodoPriority(100), []byte("100"), true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotText, err := tt.p.MarshalText()
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}
			if !tt.wantErr {
				assert.Equal(t, tt.wantText, gotText)
			}
		})
	}
}

func TestTodoPriority_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		text    []byte
		want    TodoPriority
		wantErr error
	}{
		{"normal", []byte("normal"), TodoPriorityNormal, nil},
		{"important", []byte("important"), TodoPriorityImportant, nil},
		{"urgent", []byte("urgent"), TodoPriorityUrgent, nil},
		{"critical", []byte("critical"), TodoPriorityCritical, nil},
		{"invalid", []byte("invalid"), TodoPriority(0), ErrInvalidTodoPriority},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var p TodoPriority
			err := p.UnmarshalText(tt.text)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, tt.want, p)
			}
		})
	}
}
