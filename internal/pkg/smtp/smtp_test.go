package smtp

import (
	"context"
	"go.uber.org/mock/gomock"
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/testutil"
	"github.com/opcotech/elemo/internal/testutil/mock"
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
				client: mock.NewNetSMTPClient(nil),
				config: new(config.SMTPConfig),
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
			want: &Client{
				client: mock.NewNetSMTPClient(nil),
				config: new(config.SMTPConfig),
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
		},
		{
			name: "create new client with nil net client",
			args: args{
				client: nil,
				config: new(config.SMTPConfig),
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
			wantErr: ErrNoSMTPClient,
		},
		{
			name: "create new client with nil config",
			args: args{
				client: mock.NewNetSMTPClient(nil),
				config: nil,
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
			wantErr: config.ErrNoConfig,
		},
		{
			name: "create new client with nil logger",
			args: args{
				client: mock.NewNetSMTPClient(nil),
				config: new(config.SMTPConfig),
				logger: nil,
				tracer: new(mock.Tracer),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new client with nil tracer",
			args: args{
				client: mock.NewNetSMTPClient(nil),
				config: new(config.SMTPConfig),
				logger: new(mock.Logger),
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
				client: mock.NewNetSMTPClient(nil),
			},
			want: mock.NewNetSMTPClient(nil),
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
				logger: new(mock.Logger),
			},
			want: new(mock.Logger),
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
				tracer: new(mock.Tracer),
			},
			want: new(mock.Tracer),
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

func TestClient_Authenticate(t *testing.T) {
	type fields struct {
		client func(ctrl *gomock.Controller, auth smtp.Auth) WrappedClient
		config *config.SMTPConfig
		logger log.Logger
		tracer func(ctx context.Context) tracing.Tracer
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "authenticate with success",
			fields: fields{
				client: func(ctrl *gomock.Controller, auth smtp.Auth) WrappedClient {
					client := mock.NewNetSMTPClient(ctrl)
					client.EXPECT().Auth(auth).Return(nil)
					return client
				},
				config: &config.SMTPConfig{
					Host:     "host",
					Username: "username",
					Password: "password",
				},
				logger: new(mock.Logger),
				tracer: func(ctx context.Context) tracing.Tracer {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "smtp.Client/Authenticate", []trace.SpanStartOption(nil)).Return(ctx, span)

					return tracer
				},
			},
			args: args{
				ctx: context.Background(),
			},
		},
		{
			name: "authenticate with error",
			fields: fields{
				client: func(ctrl *gomock.Controller, auth smtp.Auth) WrappedClient {
					client := mock.NewNetSMTPClient(ctrl)
					client.EXPECT().Auth(auth).Return(assert.AnError)
					return client
				},
				config: &config.SMTPConfig{
					Host:     "host",
					Username: "username",
					Password: "password",
				},
				logger: new(mock.Logger),
				tracer: func(ctx context.Context) tracing.Tracer {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "smtp.Client/Authenticate", []trace.SpanStartOption(nil)).Return(ctx, span)

					return tracer
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: ErrAuthFailed,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := &Client{
				client: tt.fields.client(
					ctrl,
					smtp.PlainAuth(
						"",
						tt.fields.config.Username,
						tt.fields.config.Password,
						tt.fields.config.Host,
					),
				),
				config: tt.fields.config,
				logger: tt.fields.logger,
				tracer: tt.fields.tracer(tt.args.ctx),
			}
			assert.ErrorIs(t, c.Authenticate(tt.args.ctx), tt.wantErr)
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
				client: func(ctrl *gomock.Controller, ctx context.Context, _, to string) *Client {
					smtpConf := &config.SMTPConfig{
						FromAddress: "no-reply@example.com",
					}

					buf := mock.NewWriteCloser(ctrl)
					buf.EXPECT().Write(gomock.Any()).Return(10, nil).AnyTimes()
					buf.EXPECT().Close().Return(nil)

					client := mock.NewNetSMTPClient(ctrl)
					client.EXPECT().Mail(smtpConf.FromAddress).Return(nil)
					client.EXPECT().Rcpt(to).Return(nil)
					client.EXPECT().Data().Return(buf, nil)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "smtp.Client/SendEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &Client{
						client: client,
						config: smtpConf,
						logger: new(mock.Logger),
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

					client := mock.NewNetSMTPClient(ctrl)
					client.EXPECT().Mail(smtpConf.FromAddress).Return(assert.AnError)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "smtp.Client/SendEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &Client{
						client: client,
						config: smtpConf,
						logger: new(mock.Logger),
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
			wantErr: ErrComposeEmail,
		},
		{
			name: "send email with setting rcpt error",
			fields: fields{
				client: func(ctrl *gomock.Controller, ctx context.Context, _, to string) *Client {
					smtpConf := &config.SMTPConfig{
						FromAddress: "no-reply@example.com",
					}

					client := mock.NewNetSMTPClient(ctrl)
					client.EXPECT().Mail(smtpConf.FromAddress).Return(nil)
					client.EXPECT().Rcpt(to).Return(assert.AnError)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "smtp.Client/SendEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &Client{
						client: client,
						config: smtpConf,
						logger: new(mock.Logger),
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
			wantErr: ErrComposeEmail,
		},
		{
			name: "send email with getting data error",
			fields: fields{
				client: func(ctrl *gomock.Controller, ctx context.Context, _, to string) *Client {
					smtpConf := &config.SMTPConfig{
						FromAddress: "no-reply@example.com",
					}

					client := mock.NewNetSMTPClient(ctrl)
					client.EXPECT().Mail(smtpConf.FromAddress).Return(nil)
					client.EXPECT().Rcpt(to).Return(nil)
					client.EXPECT().Data().Return(mock.NewWriteCloser(nil), assert.AnError)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "smtp.Client/SendEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &Client{
						client: client,
						config: smtpConf,
						logger: new(mock.Logger),
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
			wantErr: ErrComposeEmail,
		},
		{
			name: "send email with buffer write error",
			fields: fields{
				client: func(ctrl *gomock.Controller, ctx context.Context, _, to string) *Client {
					smtpConf := &config.SMTPConfig{
						FromAddress: "no-reply@example.com",
					}

					buf := mock.NewWriteCloser(ctrl)
					buf.EXPECT().Write(gomock.Any()).Return(0, assert.AnError).AnyTimes()

					client := mock.NewNetSMTPClient(ctrl)
					client.EXPECT().Mail(smtpConf.FromAddress).Return(nil)
					client.EXPECT().Rcpt(to).Return(nil)
					client.EXPECT().Data().Return(buf, nil)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "smtp.Client/SendEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &Client{
						client: client,
						config: smtpConf,
						logger: new(mock.Logger),
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
			wantErr: ErrComposeEmail,
		},
		{
			name: "send email with buffer close error",
			fields: fields{
				client: func(ctrl *gomock.Controller, ctx context.Context, _, to string) *Client {
					smtpConf := &config.SMTPConfig{
						FromAddress: "no-reply@example.com",
					}

					buf := mock.NewWriteCloser(ctrl)
					buf.EXPECT().Write(gomock.Any()).Return(10, nil).AnyTimes()
					buf.EXPECT().Close().Return(assert.AnError)

					client := mock.NewNetSMTPClient(ctrl)
					client.EXPECT().Mail(smtpConf.FromAddress).Return(nil)
					client.EXPECT().Rcpt(to).Return(nil)
					client.EXPECT().Data().Return(buf, nil)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "smtp.Client/SendEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &Client{
						client: client,
						config: smtpConf,
						logger: new(mock.Logger),
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
			wantErr: ErrComposeEmail,
		},
		{
			name: "send email with invalid template",
			fields: fields{
				client: func(_ *gomock.Controller, ctx context.Context, _, _ string) *Client {
					smtpConf := &config.SMTPConfig{
						FromAddress: "no-reply@example.com",
					}

					client := mock.NewNetSMTPClient(nil)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "smtp.Client/SendEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &Client{
						client: client,
						config: smtpConf,
						logger: new(mock.Logger),
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
					Data: nil,
				},
			},
			wantErr: ErrComposeEmail,
		},
		{
			name: "send email with template render error",
			fields: fields{
				client: func(_ *gomock.Controller, ctx context.Context, _, _ string) *Client {
					smtpConf := &config.SMTPConfig{
						FromAddress: "no-reply@example.com",
					}

					client := mock.NewNetSMTPClient(nil)

					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "smtp.Client/SendEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &Client{
						client: client,
						config: smtpConf,
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
			},
			args: args{
				ctx:     context.Background(),
				subject: "subject",
				to:      "test-user@example.com",
				template: &email.Template{
					Path: testutil.NewTempFile(t, "template", "{{ ?? }}"),
					Data: &testTemplateData{Field: "value"},
				},
			},
			wantErr: ErrComposeEmail,
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
