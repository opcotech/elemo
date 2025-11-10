package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestWithLogger(t *testing.T) {
	type args struct {
		logger log.Logger
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    log.Logger
	}{
		{
			name: "WithLogger sets the logger for the baseService",
			args: args{
				logger: mock.NewMockLogger(nil),
			},
			want: mock.NewMockLogger(nil),
		},
		{
			name: "WithLogger returns an error if no logger is provided",
			args: args{
				logger: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s baseService

			err := WithLogger(tt.args.logger)(&s)
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}

			if !tt.wantErr {
				assert.Equal(t, tt.want, s.logger)
			}
		})
	}
}

func TestWithTracer(t *testing.T) {
	type args struct {
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    tracing.Tracer
	}{
		{
			name: "WithTracer sets the tracer for the baseService",
			args: args{
				tracer: mock.NewMockTracer(nil),
			},
			want: mock.NewMockTracer(nil),
		},
		{
			name: "WithTracer returns an error if no tracer is provided",
			args: args{
				tracer: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s baseService

			err := WithTracer(tt.args.tracer)(&s)
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}

			if !tt.wantErr {
				assert.Equal(t, tt.want, s.tracer)
			}
		})
	}
}

func TestWithLicenseService(t *testing.T) {
	type args struct {
		licenseService LicenseService
	}
	tests := []struct {
		name    string
		argsFn  func(ctrl *gomock.Controller) args
		want    func(ctrl *gomock.Controller) LicenseService
		wantErr bool
	}{
		{
			name: "set the license service for the baseService",
			argsFn: func(ctrl *gomock.Controller) args {
				return args{licenseService: mock.NewMockLicenseService(ctrl)}
			},
			want: func(ctrl *gomock.Controller) LicenseService { return mock.NewMockLicenseService(ctrl) },
		},
		{
			name:    "return an error if no license service is provided",
			argsFn:  func(_ *gomock.Controller) args { return args{licenseService: nil} },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var s baseService
			args := tt.argsFn(ctrl)
			err := WithLicenseService(args.licenseService)(&s)
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want(ctrl), s.licenseService)
			}
		})
	}
}

func TestWithPermissionService(t *testing.T) {
	type args struct {
		permissionService PermissionService
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    PermissionService
	}{
		{
			name: "set the permission service for the baseService",
			args: args{
				permissionService: mock.NewPermissionService(nil),
			},
			want: mock.NewPermissionService(nil),
		},
		{
			name: "return an error if no permission service is provided",
			args: args{
				permissionService: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s baseService

			err := WithPermissionService(tt.args.permissionService)(&s)
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}

			if !tt.wantErr {
				assert.Equal(t, tt.want, s.permissionService)
			}
		})
	}
}

func TestWithEmailService(t *testing.T) {
	type args struct {
		emailService EmailService
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    EmailService
	}{
		{
			name: "set the email service for the baseService",
			args: args{
				emailService: mock.NewEmailService(nil),
			},
			want: mock.NewEmailService(nil),
		},
		{
			name: "return an error if no email service is provided",
			args: args{
				emailService: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s baseService

			err := WithEmailService(tt.args.emailService)(&s)
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}

			if !tt.wantErr {
				assert.Equal(t, tt.want, s.emailService)
			}
		})
	}
}

func Test_newService(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *baseService
		wantErr error
	}{
		{
			name: "newService returns a baseService with the provided options",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
				},
			},
			want: &baseService{
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
		},
		{
			name: "newService returns default logger if no logger is provided",
			args: args{
				opts: []Option{
					WithTracer(mock.NewMockTracer(nil)),
				},
			},
			want: &baseService{
				logger: log.DefaultLogger(),
				tracer: mock.NewMockTracer(nil),
			},
		},
		{
			name: "newService returns default tracer if no tracer is provided",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
				},
			},
			want: &baseService{
				logger: mock.NewMockLogger(nil),
				tracer: tracing.NoopTracer(),
			},
		},
		{
			name: "newService returns error if nil logger is provided",
			args: args{
				opts: []Option{
					WithLogger(nil),
					WithTracer(mock.NewMockTracer(nil)),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "newService returns error if nil tracer is provided",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(nil),
				},
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := newService(tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
