package smtp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/pkg/log"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocksmtp "github.com/opcotech/elemo/internal/pkg/smtp/mock"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/testutil"
)

type testTemplateData struct {
	Field string
}

func (t *testTemplateData) Get() any {
	return t
}

func TestNewDatabase(t *testing.T) {
	type args struct {
		client WrappedClient
		config *config.SMTPConfig
		logger log.Logger
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr error
	}{
		{
			name: "create new client",
			args: args{
				client: new(mocksmtp.MockWrappedClient),
				config: new(config.SMTPConfig),
				logger: mocklog.NewMockLogger(nil),
				tracer: mocktrace.NewMockTracer(nil),
			},
			want: &Client{
				client: new(mocksmtp.MockWrappedClient),
				config: new(config.SMTPConfig),
				logger: mocklog.NewMockLogger(nil),
				tracer: mocktrace.NewMockTracer(nil),
			},
		},
		{
			name: "create new client with nil net client",
			args: args{
				client: nil,
				config: new(config.SMTPConfig),
				logger: mocklog.NewMockLogger(nil),
				tracer: mocktrace.NewMockTracer(nil),
			},
			wantErr: ErrNoSMTPClient,
		},
		{
			name: "create new client with nil config",
			args: args{
				client: new(mocksmtp.MockWrappedClient),
				config: nil,
				logger: mocklog.NewMockLogger(nil),
				tracer: mocktrace.NewMockTracer(nil),
			},
			wantErr: config.ErrNoConfig,
		},
		{
			name: "create new client with nil logger",
			args: args{
				client: new(mocksmtp.MockWrappedClient),
				config: new(config.SMTPConfig),
				logger: nil,
				tracer: mocktrace.NewMockTracer(nil),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new client with nil tracer",
			args: args{
				client: new(mocksmtp.MockWrappedClient),
				config: new(config.SMTPConfig),
				logger: mocklog.NewMockLogger(nil),
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, err := NewClient(
				WithWrappedClient(tt.args.client),
				WithConfig(tt.args.config),
				WithLogger(tt.args.logger),
				WithTracer(tt.args.tracer),
			)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db)
		})
	}
}

func TestWithConfig(t *testing.T) {
	type args struct {
		config *config.SMTPConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *config.SMTPConfig
		wantErr error
	}{
		{
			name: "create new option with config",
			args: args{
				config: new(config.SMTPConfig),
			},
			want: new(config.SMTPConfig),
		},
		{
			name: "create new option with nil config",
			args: args{
				config: nil,
			},
			wantErr: config.ErrNoConfig,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := new(Client)
			err := WithConfig(tt.args.config)(client)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, client.config)
		})
	}
}

func TestWithWrappedClient(t *testing.T) {
	type args struct {
		client WrappedClient
	}
	tests := []struct {
		name    string
		args    args
		want    WrappedClient
		wantErr error
	}{
		{
			name: "create new option with client",
			args: args{
				client: new(mocksmtp.MockWrappedClient),
			},
			want: new(mocksmtp.MockWrappedClient),
		},
		{
			name: "create new option with nil client",
			args: args{
				client: nil,
			},
			wantErr: ErrNoSMTPClient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := new(Client)
			err := WithWrappedClient(tt.args.client)(client)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, client.client)
		})
	}
}

func TestWithLogger(t *testing.T) {
	type args struct {
		logger log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    log.Logger
		wantErr error
	}{
		{
			name: "create new option with logger",
			args: args{
				logger: mocklog.NewMockLogger(nil),
			},
			want: mocklog.NewMockLogger(nil),
		},
		{
			name: "create new option with nil logger",
			args: args{
				logger: nil,
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := new(Client)
			err := WithLogger(tt.args.logger)(client)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, client.logger)
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
		want    tracing.Tracer
		wantErr error
	}{
		{
			name: "create new option with tracer",
			args: args{
				tracer: mocktrace.NewMockTracer(nil),
			},
			want: mocktrace.NewMockTracer(nil),
		},
		{
			name: "create new option with nil tracer",
			args: args{
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := new(Client)
			err := WithTracer(tt.args.tracer)(client)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, client.tracer)
		})
	}
}

func TestClient_SendEmail(t *testing.T) {
	type fields struct {
		client func(ctrl *gomock.Controller, ctx context.Context, subject, to string) *Client
	}
	type args struct {
		ctx      context.Context
		subject  string
		to       string
		template *email.Template
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "send email with success",
			fields: fields{
				client: func(ctrl *gomock.Controller, ctx context.Context, _, _ string) *Client {
					smtpConf := &config.SMTPConfig{
						FromAddress: "no-reply@example.com",
					}

					client := mocksmtp.NewMockWrappedClient(ctrl)
					client.EXPECT().DialAndSend(gomock.Any()).Return(nil)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Info(gomock.Any(), "email sent", gomock.Any())

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "smtp.Client/SendEmail").Return(ctx, span)

					return &Client{
						client: client,
						config: smtpConf,
						logger: logger,
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: "subject",
				to:      "test-user@example.com",
				template: &email.Template{
					Path: testutil.NewTempFile(t, "template", "{{ .Field }}"),
					Data: &testTemplateData{Field: "value"},
				},
			},
		},
		{
			name: "send email with setting mail error",
			fields: fields{
				client: func(ctrl *gomock.Controller, ctx context.Context, _, _ string) *Client {
					smtpConf := &config.SMTPConfig{
						FromAddress: "no-reply@example.com",
					}

					client := mocksmtp.NewMockWrappedClient(ctrl)
					client.EXPECT().DialAndSend(gomock.Any()).Return(assert.AnError)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End().Return()

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Error(gomock.Any(), ErrSendEmail.Error(), gomock.Any())

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "smtp.Client/SendEmail").Return(ctx, span)

					return &Client{
						client: client,
						config: smtpConf,
						logger: logger,
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: "subject",
				to:      "test-user@example.com",
				template: &email.Template{
					Path: testutil.NewTempFile(t, "template", "{{ .Field }}"),
					Data: &testTemplateData{Field: "value"},
				},
			},
			wantErr: ErrSendEmail,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := tt.fields.client(ctrl, tt.args.ctx, tt.args.subject, tt.args.to)
			err := c.SendEmail(tt.args.ctx, tt.args.subject, tt.args.to, tt.args.template)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
