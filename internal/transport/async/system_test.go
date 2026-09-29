package async

import (
	"context"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/deployment"
	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/entitlement/license"
	"github.com/opcotech/elemo/internal/pkg/log"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/queue"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
)

func TestNewSystemHealthCheckTaskHandler(t *testing.T) {
	type args struct {
		opts []TaskHandlerOption
	}
	tests := []struct {
		name    string
		args    args
		want    *SystemHealthCheckTaskHandler
		wantErr error
	}{
		{
			name: "create new task handler",
			args: args{
				opts: []TaskHandlerOption{
					WithTaskLogger(mocklog.NewMockLogger(nil)),
					WithTaskTracer(mocktrace.NewMockTracer(nil)),
				},
			},
			want: &SystemHealthCheckTaskHandler{
				baseTaskHandler: &baseTaskHandler{
					logger: mocklog.NewMockLogger(nil),
					tracer: mocktrace.NewMockTracer(nil),
				},
			},
		},
		{
			name: "create new task handler with invalid option",
			args: args{
				opts: []TaskHandlerOption{
					WithTaskLogger(nil),
				},
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewSystemHealthCheckTaskHandler(tt.args.opts...)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSystemHealthCheckTaskHandler_ProcessTask(t *testing.T) {
	type fields struct {
		baseTaskHandler func(ctx context.Context, task *asynq.Task, ctrl *gomock.Controller) *baseTaskHandler
	}
	type args struct {
		ctx  context.Context
		task *asynq.Task
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "process task",
			fields: fields{
				baseTaskHandler: func(ctx context.Context, _ *asynq.Task, ctrl *gomock.Controller) *baseTaskHandler {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "transport.asynq.SystemHealthCheckTaskHandler/ProcessTask").Return(ctx, span)

					return &baseTaskHandler{
						logger: mocklog.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				task: func() *asynq.Task {
					task, _ := queue.NewSystemHealthCheckTask()
					return task
				}(),
			},
		},
		{
			name: "process task with invalid payload",
			fields: fields{
				baseTaskHandler: func(ctx context.Context, _ *asynq.Task, ctrl *gomock.Controller) *baseTaskHandler {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "transport.asynq.SystemHealthCheckTaskHandler/ProcessTask").Return(ctx, span)

					return &baseTaskHandler{
						logger: mocklog.NewMockLogger(nil),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				task: func() *asynq.Task {
					return asynq.NewTask(
						queue.TaskTypeSystemHealthCheck.String(),
						[]byte(`{"message"`),
						asynq.Timeout(queue.DefaultTaskTimeout),
						asynq.Retention(queue.DefaultTaskRetention),
					)
				}(),
			},
			wantErr: ErrTaskPayloadUnmarshal,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := &SystemHealthCheckTaskHandler{
				baseTaskHandler: tt.fields.baseTaskHandler(tt.args.ctx, tt.args.task, ctrl),
			}

			err := h.ProcessTask(tt.args.ctx, tt.args.task)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

type staticEntitlementReporter struct {
	status entitlement.Status
	err    error
}

func (r staticEntitlementReporter) Status(context.Context) (entitlement.Status, error) {
	return r.status, r.err
}

func TestNewSystemLicenseExpiryTaskHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	emailService := mocksvc.NewMockEmailService(ctrl)
	reporter := staticEntitlementReporter{}

	handler, err := NewSystemLicenseExpiryTaskHandler(
		WithTaskEmailService(emailService),
		WithTaskEntitlementReporter(reporter),
	)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	_, err = NewSystemLicenseExpiryTaskHandler(WithTaskEntitlementReporter(reporter))
	assert.ErrorIs(t, err, ErrNoEmailService)

	_, err = NewSystemLicenseExpiryTaskHandler(WithTaskEmailService(emailService))
	assert.ErrorIs(t, err, ErrNoEntitlementReporter)
}

func TestSystemLicenseExpiryTaskHandler_ProcessTask(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	nearExpiry := now.Add(6 * 24 * time.Hour)
	farExpiry := now.Add(8 * 24 * time.Hour)
	graceEnd := now.Add(30 * 24 * time.Hour)

	airGapStatus := func(state license.State, expiresAt *time.Time) entitlement.Status {
		return entitlement.Status{
			DeploymentMode: deployment.ModeAirGap,
			AirGap: &entitlement.AirGapStatus{
				State:         state,
				LicenseID:     "license-id",
				Customer:      "ACME Inc.",
				SeatsLicensed: 10,
				ExpiresAt:     expiresAt,
				GraceEndsAt:   &graceEnd,
			},
		}
	}

	tests := []struct {
		name         string
		billingEmail string
		status       entitlement.Status
		reporterErr  error
		send         bool
		emailErr     error
		wantErr      error
	}{
		{
			name:         "self-hosted",
			billingEmail: "billing@example.com",
			status:       entitlement.Status{DeploymentMode: deployment.ModeSelfHosted},
		},
		{
			name:   "empty billing email",
			status: airGapStatus(license.StateValid, &nearExpiry),
		},
		{
			name:         "invalid billing email",
			billingEmail: "not-an-email",
			status:       airGapStatus(license.StateValid, &nearExpiry),
		},
		{
			name:         "valid outside reminder window",
			billingEmail: "billing@example.com",
			status:       airGapStatus(license.StateValid, &farExpiry),
		},
		{
			name:         "valid within reminder window",
			billingEmail: "billing@example.com",
			status:       airGapStatus(license.StateValid, &nearExpiry),
			send:         true,
		},
		{
			name:         "grace",
			billingEmail: "billing@example.com",
			status:       airGapStatus(license.StateGrace, &nearExpiry),
			send:         true,
		},
		{
			name:         "expired",
			billingEmail: "billing@example.com",
			status:       airGapStatus(license.StateExpired, &nearExpiry),
			send:         true,
		},
		{
			name:         "email failure",
			billingEmail: "billing@example.com",
			status:       airGapStatus(license.StateGrace, &nearExpiry),
			send:         true,
			emailErr:     assert.AnError,
			wantErr:      assert.AnError,
		},
		{
			name:         "missing",
			billingEmail: "billing@example.com",
			status:       airGapStatus(license.StateMissing, nil),
		},
		{
			name:         "invalid",
			billingEmail: "billing@example.com",
			status:       airGapStatus(license.StateInvalid, nil),
		},
		{
			name:         "not yet valid",
			billingEmail: "billing@example.com",
			status:       airGapStatus(license.StateNotYetValid, &farExpiry),
		},
		{
			name:         "reporter failure",
			billingEmail: "billing@example.com",
			reporterErr:  assert.AnError,
			wantErr:      assert.AnError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			emailService := mocksvc.NewMockEmailService(ctrl)
			if tt.send {
				emailService.EXPECT().
					SendLicenseExpiryEmail(gomock.Any(), tt.billingEmail, *tt.status.AirGap).
					Return(tt.emailErr)
			}

			handler, err := NewSystemLicenseExpiryTaskHandler(
				WithTaskEmailService(emailService),
				WithTaskEntitlementReporter(staticEntitlementReporter{
					status: tt.status,
					err:    tt.reporterErr,
				}),
				WithTaskBillingEmail(tt.billingEmail),
			)
			assert.NoError(t, err)

			err = handler.ProcessTask(context.Background(), nil)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
