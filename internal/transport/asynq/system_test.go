package asynq

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/testutil/mock"
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

func TestNewSystemHealthCheckTaskHandler(t *testing.T) {
	type args struct {
		opts []TaskOption
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
				opts: []TaskOption{
					WithTaskLogger(new(mock.Logger)),
					WithTaskTracer(new(mock.Tracer)),
				},
			},
			want: &SystemHealthCheckTaskHandler{
				baseTaskHandler: &baseTaskHandler{
					logger: new(mock.Logger),
					tracer: new(mock.Tracer),
				},
			},
		},
		{
			name: "create new task handler with invalid option",
			args: args{
				opts: []TaskOption{
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
		baseTaskHandler func(ctx context.Context, task *asynq.Task) *baseTaskHandler
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
				baseTaskHandler: func(ctx context.Context, task *asynq.Task) *baseTaskHandler {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "transport.asynq.SystemHealthCheckTaskHandler/ProcessTask", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseTaskHandler{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				task: func() *asynq.Task {
					task, _ := NewSystemHealthCheckTask()
					return task
				}(),
			},
		},
		{
			name: "process task with invalid payload",
			fields: fields{
				baseTaskHandler: func(ctx context.Context, task *asynq.Task) *baseTaskHandler {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "transport.asynq.SystemHealthCheckTaskHandler/ProcessTask", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseTaskHandler{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				task: func() *asynq.Task {
					return asynq.NewTask(
						TaskTypeSystemHealthCheck.String(),
						[]byte(`{"message"`),
						asynq.Timeout(DefaultTaskTimeout),
						asynq.Retention(DefaultTaskRetention),
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

			h := &SystemHealthCheckTaskHandler{
				baseTaskHandler: tt.fields.baseTaskHandler(tt.args.ctx, tt.args.task),
			}

			err := h.ProcessTask(tt.args.ctx, tt.args.task)
			assert.ErrorIs(t, err, tt.wantErr)
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

func TestNewSystemLicenseExpiryTaskHandler(t *testing.T) {
	type args struct {
		opts []TaskOption
	}
	tests := []struct {
		name    string
		args    args
		want    *SystemLicenseExpiryTaskHandler
		wantErr error
	}{
		{
			name: "create new task handler",
			args: args{
				opts: []TaskOption{
					WithTaskEmailService(new(mock.EmailService)),
					WithTaskLogger(new(mock.Logger)),
					WithTaskTracer(new(mock.Tracer)),
				},
			},
			want: &SystemLicenseExpiryTaskHandler{
				baseTaskHandler: &baseTaskHandler{
					logger:       new(mock.Logger),
					tracer:       new(mock.Tracer),
					emailService: new(mock.EmailService),
				},
			},
		},
		{
			name: "create new task handler with invalid option",
			args: args{
				opts: []TaskOption{
					WithTaskLogger(nil),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new task handler with no email service",
			args: args{
				opts: []TaskOption{
					WithTaskLogger(new(mock.Logger)),
					WithTaskTracer(new(mock.Tracer)),
				},
			},
			wantErr: ErrNoEmailService,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewSystemLicenseExpiryTaskHandler(tt.args.opts...)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSystemLicenseExpiryTaskHandler_ProcessTask(t *testing.T) {
	type fields struct {
		baseTaskHandler func(ctx context.Context, task *asynq.Task) *baseTaskHandler
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
				baseTaskHandler: func(ctx context.Context, task *asynq.Task) *baseTaskHandler {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "transport.asynq.SystemLicenseExpiryTaskHandler/ProcessTask", []trace.SpanStartOption(nil)).Return(ctx, span)

					var payload LicenseExpiryTaskPayload
					_ = json.Unmarshal(task.Payload(), &payload)

					emailService := new(mock.EmailService)
					emailService.On("SendSystemLicenseExpiryEmail", ctx,
						payload.LicenseID,
						payload.LicenseEmail,
						payload.LicenseOrganization,
						payload.LicenseExpiresAt,
					).Return(nil)

					return &baseTaskHandler{
						logger:       new(mock.Logger),
						tracer:       tracer,
						emailService: emailService,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				task: func() *asynq.Task {
					task, _ := NewSystemLicenseExpiryTask(&license.License{
						ID:           xid.New(),
						Email:        "info@exameple.com",
						Organization: "ACME Inc.",
						ExpiresAt:    time.Now().Add(24 * time.Hour),
					})
					return task
				}(),
			},
		},
		{
			name: "process task skip email sending",
			fields: fields{
				baseTaskHandler: func(ctx context.Context, task *asynq.Task) *baseTaskHandler {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "transport.asynq.SystemLicenseExpiryTaskHandler/ProcessTask", []trace.SpanStartOption(nil)).Return(ctx, span)

					var payload LicenseExpiryTaskPayload
					_ = json.Unmarshal(task.Payload(), &payload)

					return &baseTaskHandler{
						logger:       new(mock.Logger),
						tracer:       tracer,
						emailService: new(mock.EmailService),
					}
				},
			},
			args: args{
				ctx: context.Background(),
				task: func() *asynq.Task {
					task, _ := NewSystemLicenseExpiryTask(&license.License{
						ID:           xid.New(),
						Email:        "info@exameple.com",
						Organization: "ACME Inc.",
						ExpiresAt:    time.Now().Add(240 * time.Hour),
					})
					return task
				}(),
			},
		},
		{
			name: "process task with invalid payload",
			fields: fields{
				baseTaskHandler: func(ctx context.Context, task *asynq.Task) *baseTaskHandler {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "transport.asynq.SystemLicenseExpiryTaskHandler/ProcessTask", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseTaskHandler{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				task: func() *asynq.Task {
					return asynq.NewTask(
						TaskTypeSystemLicenseExpiry.String(),
						[]byte(`{"LicenseID"`),
						asynq.Timeout(DefaultTaskTimeout),
						asynq.Queue(MessageQueueHighPriority),
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

			h := &SystemLicenseExpiryTaskHandler{
				baseTaskHandler: tt.fields.baseTaskHandler(tt.args.ctx, tt.args.task),
			}

			err := h.ProcessTask(tt.args.ctx, tt.args.task)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
