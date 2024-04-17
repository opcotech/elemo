package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskType_String(t *testing.T) {
	tests := []struct {
		name string
		t    TaskType
		want string
	}{
		{"health check task", TaskTypeSystemHealthCheck, "system:health_check"},
		{"license expiry task", TaskTypeSystemLicenseExpiry, "system:license_expiry"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.t.String())
		})
	}
}
