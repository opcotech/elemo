package queue

import (
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/opcotech/elemo/internal/license"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
)

func TestNewSystemHealthCheckTask(t *testing.T) {
	tests := []struct {
		name    string
		want    *asynq.Task
		wantErr error
	}{
		{
			name: "create new task",
			want: asynq.NewTask(TaskTypeSystemHealthCheck.String(),
				[]byte(`{"message":"healthy"}`),
				asynq.Timeout(DefaultTaskTimeout),
				asynq.Retention(DefaultTaskRetention)),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewSystemHealthCheckTask()
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewSystemLicenseExpiryTask(t *testing.T) {
	licenseID, _ := xid.FromString("bvn6c05roa2mnak37ms0")

	type args struct {
		license *license.License
	}
	tests := []struct {
		name    string
		args    args
		want    *asynq.Task
		wantErr error
	}{
		{
			name: "create new task",
			args: args{
				license: &license.License{
					ID:           licenseID,
					Email:        "info@example.com",
					Organization: "ACME Inc.",
					ExpiresAt:    time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC),
				},
			},
			want: asynq.NewTask(TaskTypeSystemLicenseExpiry.String(),
				[]byte(`{"LicenseID":"`+licenseID.String()+`","LicenseEmail":"info@example.com","LicenseOrganization":"ACME Inc.","LicenseExpiresAt":"2099-12-31T00:00:00Z"}`),
				asynq.Timeout(DefaultTaskTimeout),
				asynq.Queue(MessageQueueHighPriority),
			),
		},
		{
			name: "create new task with no license",
			args: args{
				license: nil,
			},
			wantErr: license.ErrNoLicense,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewSystemLicenseExpiryTask(tt.args.license)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
