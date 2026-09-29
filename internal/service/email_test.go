package service_test

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"testing"
	"time"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/entitlement/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/smtp"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

// templateMatcher is a custom gomock matcher for email templates
type templateMatcher struct {
	expectedTemplate *email.Template
}

func (m *templateMatcher) Matches(x interface{}) bool {
	template, ok := x.(*email.Template)
	if !ok {
		return false
	}

	if m.expectedTemplate.Path != template.Path {
		return false
	}

	expectedData := m.expectedTemplate.Data.Get()
	actualData := template.Data.Get()

	return m.compareStructs(expectedData, actualData)
}

func (m *templateMatcher) compareStructs(expected, actual interface{}) bool {
	expectedStr := fmt.Sprintf("%+v", expected)
	actualStr := fmt.Sprintf("%+v", actual)

	return expectedStr == actualStr
}

func (m *templateMatcher) String() string {
	return fmt.Sprintf("is template with path %s", m.expectedTemplate.Path)
}

func matchTemplate(expected *email.Template) gomock.Matcher {
	return &templateMatcher{expectedTemplate: expected}
}

func TestNewEmailService(t *testing.T) {
	type args struct {
		client       service.EmailSender
		templatesDir string
		smtpConf     *config.SMTPConfig
		opts         []service.Option
	}
	tests := []struct {
		name    string
		args    args
		want    service.EmailService
		wantErr error
	}{
		{
			name: "new email service",
			args: args{
				client: func() service.EmailSender {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mocksvc.NewMockEmailSender(ctrl)
				}(),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
				opts: []service.Option{
					service.WithLogger(mocklog.NewMockLogger(nil)),
					service.WithTracer(mocktrace.NewMockTracer(nil)),
				},
			},
			want: func() service.EmailService {
				svc, err := service.NewEmailService(
					func() service.EmailSender {
						ctrl := gomock.NewController(t)
						defer ctrl.Finish()
						return mocksvc.NewMockEmailSender(ctrl)
					}(),
					"/templates",
					new(config.SMTPConfig),
					service.WithLogger(mocklog.NewMockLogger(nil)),
					service.WithTracer(mocktrace.NewMockTracer(nil)),
				)
				if err != nil {
					panic(err)
				}
				return svc
			}(),
		},
		{
			name: "new email service with no email sender",
			args: args{
				client:       nil,
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
				opts: []service.Option{
					service.WithLogger(mocklog.NewMockLogger(nil)),
					service.WithTracer(mocktrace.NewMockTracer(nil)),
				},
			},
			wantErr: smtp.ErrNoSMTPClient,
		},
		{
			name: "new email service with invalid options",
			args: args{
				client: func() service.EmailSender {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mocksvc.NewMockEmailSender(ctrl)
				}(),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
				opts: []service.Option{
					service.WithLogger(nil),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new email service with no logger",
			args: args{
				client: func() service.EmailSender {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mocksvc.NewMockEmailSender(ctrl)
				}(),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
				opts: []service.Option{
					service.WithTracer(mocktrace.NewMockTracer(nil)),
				},
			},
			want: func() service.EmailService {
				svc, err := service.NewEmailService(
					func() service.EmailSender {
						ctrl := gomock.NewController(t)
						defer ctrl.Finish()
						return mocksvc.NewMockEmailSender(ctrl)
					}(),
					"/templates",
					new(config.SMTPConfig),
					service.WithLogger(log.DefaultLogger()),
					service.WithTracer(mocktrace.NewMockTracer(nil)),
				)
				if err != nil {
					panic(err)
				}
				return svc
			}(),
		},
		{
			name: "new email service with no tracer",
			args: args{
				client: func() service.EmailSender {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mocksvc.NewMockEmailSender(ctrl)
				}(),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
				opts: []service.Option{
					service.WithLogger(mocklog.NewMockLogger(nil)),
				},
			},
			want: func() service.EmailService {
				svc, err := service.NewEmailService(
					func() service.EmailSender {
						ctrl := gomock.NewController(t)
						defer ctrl.Finish()
						return mocksvc.NewMockEmailSender(ctrl)
					}(),
					"/templates",
					new(config.SMTPConfig),
					service.WithLogger(mocklog.NewMockLogger(nil)),
					service.WithTracer(tracing.NoopTracer()),
				)
				if err != nil {
					panic(err)
				}
				return svc
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := service.NewEmailService(tt.args.client, tt.args.templatesDir, tt.args.smtpConf, tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, got)
			}
		})
	}
}

func TestEmailService_SendAuthPasswordResetEmail(t *testing.T) {
	type fields struct {
		runtimeFn    func(ctrl *gomock.Controller, ctx context.Context) service.Runtime
		client       func(ctrl *gomock.Controller, ctx context.Context, templatesDir, token string, smtpConf *config.SMTPConfig, recipient email.Recipient) service.EmailSender
		templatesDir string
		smtpConf     *config.SMTPConfig
	}
	type args struct {
		ctx       context.Context
		recipient email.Recipient
		token     string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "send auth password reset email",
			fields: fields{
				runtimeFn: func(ctrl *gomock.Controller, ctx context.Context) service.Runtime {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.emailService/SendAuthPasswordResetEmail", gomock.Len(0)).Return(ctx, span)

					return service.NewRuntimeForTest(mocklog.NewMockLogger(ctrl), tracer)
				},
				client: func(ctrl *gomock.Controller, _ context.Context, templatesDir, token string, smtpConf *config.SMTPConfig, recipient email.Recipient) service.EmailSender {
					subject := "[Action Required] Reset your password"

					passwordResetURL := fmt.Sprintf("%s/reset-password?token=%s", smtpConf.ClientURL, token)
					template, err := email.NewTemplate(
						path.Join(templatesDir, service.AuthPasswordResetTemplate),
						&email.PasswordResetTemplateData{
							Subject:          subject,
							FirstName:        recipient.FirstName,
							LastName:         recipient.LastName,
							PasswordResetURL: passwordResetURL,
							SupportEmail:     smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := mocksvc.NewMockEmailSender(ctrl)
					client.EXPECT().SendEmail(gomock.Any(), subject, recipient.Email, matchTemplate(template)).Return(nil)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					ClientURL:      "https://example.com",
					SupportAddress: "support@example.com",
				},
			},
			args: args{
				ctx: context.Background(),
				recipient: email.Recipient{
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				},
				token: "test-token",
			},
		},
		{
			name: "send auth password reset email failed",
			fields: fields{
				runtimeFn: func(ctrl *gomock.Controller, ctx context.Context) service.Runtime {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.emailService/SendAuthPasswordResetEmail", gomock.Len(0)).Return(ctx, span)

					return service.NewRuntimeForTest(mocklog.NewMockLogger(ctrl), tracer)
				},
				client: func(ctrl *gomock.Controller, _ context.Context, templatesDir, token string, smtpConf *config.SMTPConfig, recipient email.Recipient) service.EmailSender {
					subject := "[Action Required] Reset your password"

					passwordResetURL := fmt.Sprintf("%s/reset-password?token=%s", smtpConf.ClientURL, token)
					template, err := email.NewTemplate(
						path.Join(templatesDir, service.AuthPasswordResetTemplate),
						&email.PasswordResetTemplateData{
							Subject:          subject,
							FirstName:        recipient.FirstName,
							LastName:         recipient.LastName,
							PasswordResetURL: passwordResetURL,
							SupportEmail:     smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := mocksvc.NewMockEmailSender(ctrl)
					client.EXPECT().SendEmail(gomock.Any(), subject, recipient.Email, matchTemplate(template)).Return(assert.AnError)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					ClientURL:      "https://example.com",
					SupportAddress: "support@example.com",
				},
			},
			args: args{
				ctx: context.Background(),
				recipient: email.Recipient{
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				},
				token: "test-token",
			},
			wantErr: service.ErrEmailSend,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := func() service.EmailService {
				rt := tt.fields.runtimeFn(ctrl, tt.args.ctx)
				svc, err := service.NewEmailService(
					tt.fields.client(ctrl, tt.args.ctx, tt.fields.templatesDir, tt.args.token, tt.fields.smtpConf, tt.args.recipient),
					tt.fields.templatesDir,
					tt.fields.smtpConf,
					service.WithLogger(service.RuntimeLogger(rt)),
					service.WithTracer(service.RuntimeTracer(rt)),
				)
				if err != nil {
					panic(err)
				}
				return svc
			}()
			assert.ErrorIs(t, s.SendAuthPasswordResetEmail(tt.args.ctx, tt.args.recipient, tt.args.token), tt.wantErr)
		})
	}
}

func TestEmailService_SendOrganizationInvitationEmail(t *testing.T) {
	type fields struct {
		runtimeFn    func(ctrl *gomock.Controller, ctx context.Context) service.Runtime
		client       func(ctrl *gomock.Controller, ctx context.Context, templatesDir, token string, smtpConf *config.SMTPConfig, organizationID model.ID, organizationName string, recipient email.Recipient) service.EmailSender
		templatesDir string
		smtpConf     *config.SMTPConfig
	}
	type args struct {
		ctx              context.Context
		invitationPath   string
		organizationID   model.ID
		organizationName string
		recipient        email.Recipient
		token            string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "send invitation email",
			fields: fields{
				runtimeFn: func(ctrl *gomock.Controller, ctx context.Context) service.Runtime {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.emailService/SendOrganizationInvitationEmail", gomock.Len(0)).Return(ctx, span)

					return service.NewRuntimeForTest(mocklog.NewMockLogger(ctrl), tracer)
				},
				client: func(ctrl *gomock.Controller, _ context.Context, templatesDir, token string, smtpConf *config.SMTPConfig, organizationID model.ID, organizationName string, recipient email.Recipient) service.EmailSender {
					subject := fmt.Sprintf("[Action Required] You have been invited to join %s", organizationName)

					invitationURL := fmt.Sprintf("%s/organizations/join?organization=%s&token=%s", smtpConf.ClientURL, organizationID.String(), token)
					template, err := email.NewTemplate(
						path.Join(templatesDir, service.OrganizationInviteTemplate),
						&email.OrganizationInviteTemplateData{
							Subject:          subject,
							OrganizationName: organizationName,
							InvitationURL:    invitationURL,
							SupportEmail:     smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := mocksvc.NewMockEmailSender(ctrl)
					client.EXPECT().SendEmail(gomock.Any(), subject, recipient.Email, matchTemplate(template)).Return(nil)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					ClientURL:      "https://example.com",
					SupportAddress: "support@example.com",
				},
			},
			args: args{
				ctx:              context.Background(),
				invitationPath:   "/invitation",
				organizationID:   model.MustNewID(model.ResourceTypeOrganization),
				organizationName: "test",
				recipient: email.Recipient{
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				},
			},
		},
		{
			name: "send invitation email failed",
			fields: fields{
				runtimeFn: func(ctrl *gomock.Controller, ctx context.Context) service.Runtime {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.emailService/SendOrganizationInvitationEmail", gomock.Len(0)).Return(ctx, span)

					return service.NewRuntimeForTest(mocklog.NewMockLogger(ctrl), tracer)
				},
				client: func(ctrl *gomock.Controller, _ context.Context, templatesDir, token string, smtpConf *config.SMTPConfig, organizationID model.ID, organizationName string, recipient email.Recipient) service.EmailSender {
					subject := fmt.Sprintf("[Action Required] You have been invited to join %s", organizationName)

					invitationURL := fmt.Sprintf("%s/organizations/join?organization=%s&token=%s", smtpConf.ClientURL, organizationID.String(), token)
					template, err := email.NewTemplate(
						path.Join(templatesDir, service.OrganizationInviteTemplate),
						&email.OrganizationInviteTemplateData{
							Subject:          subject,
							OrganizationName: organizationName,
							InvitationURL:    invitationURL,
							SupportEmail:     smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := mocksvc.NewMockEmailSender(ctrl)
					client.EXPECT().SendEmail(gomock.Any(), subject, recipient.Email, matchTemplate(template)).Return(assert.AnError)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					ClientURL:      "https://example.com",
					SupportAddress: "support@example.com",
				},
			},
			args: args{
				ctx:              context.Background(),
				invitationPath:   "/invitation",
				organizationID:   model.MustNewID(model.ResourceTypeOrganization),
				organizationName: "test",
				recipient: email.Recipient{
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				},
			},
			wantErr: service.ErrEmailSend,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := func() service.EmailService {
				rt := tt.fields.runtimeFn(ctrl, tt.args.ctx)
				svc, err := service.NewEmailService(
					tt.fields.client(ctrl, tt.args.ctx, tt.fields.templatesDir, tt.args.token, tt.fields.smtpConf, tt.args.organizationID, tt.args.organizationName, tt.args.recipient),
					tt.fields.templatesDir,
					tt.fields.smtpConf,
					service.WithLogger(service.RuntimeLogger(rt)),
					service.WithTracer(service.RuntimeTracer(rt)),
				)
				if err != nil {
					panic(err)
				}
				return svc
			}()
			assert.ErrorIs(t, s.SendOrganizationInvitationEmail(tt.args.ctx, tt.args.organizationID, tt.args.organizationName, tt.args.recipient, tt.args.token), tt.wantErr)
		})
	}
}

func TestEmailService_SendUserWelcomeEmail(t *testing.T) {
	type fields struct {
		runtimeFn    func(ctrl *gomock.Controller, ctx context.Context) service.Runtime
		client       func(ctrl *gomock.Controller, ctx context.Context, templatesDir string, smtpConf *config.SMTPConfig, recipient email.Recipient) service.EmailSender
		templatesDir string
		smtpConf     *config.SMTPConfig
	}
	type args struct {
		ctx       context.Context
		recipient email.Recipient
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "send welcome email",
			fields: fields{
				runtimeFn: func(ctrl *gomock.Controller, ctx context.Context) service.Runtime {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(gomock.Any(), "service.emailService/SendUserWelcomeEmail", gomock.Len(0)).Return(ctx, span)

					return service.NewRuntimeForTest(mocklog.NewMockLogger(ctrl), tracer)
				},
				client: func(ctrl *gomock.Controller, _ context.Context, templatesDir string, smtpConf *config.SMTPConfig, recipient email.Recipient) service.EmailSender {
					subject := "Welcome to Elemo"

					template, err := email.NewTemplate(
						path.Join(templatesDir, service.UserWelcomeTemplate),
						&email.UserWelcomeTemplateData{
							Subject:      subject,
							FirstName:    recipient.FirstName,
							LastName:     recipient.LastName,
							LoginURL:     fmt.Sprintf("%s/redirect?url=%s", smtpConf.ClientURL, url.QueryEscape(fmt.Sprintf("%s/auth/login", smtpConf.ClientURL))),
							SupportEmail: smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := mocksvc.NewMockEmailSender(ctrl)
					client.EXPECT().SendEmail(gomock.Any(), subject, recipient.Email, matchTemplate(template)).Return(assert.AnError)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					ClientURL:      "https://example.com",
					SupportAddress: "support@example.com",
				},
			},
			args: args{
				ctx: context.Background(),
				recipient: email.Recipient{
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				},
			},
			wantErr: service.ErrEmailSend,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := func() service.EmailService {
				rt := tt.fields.runtimeFn(ctrl, tt.args.ctx)
				svc, err := service.NewEmailService(
					tt.fields.client(ctrl, tt.args.ctx, tt.fields.templatesDir, tt.fields.smtpConf, tt.args.recipient),
					tt.fields.templatesDir,
					tt.fields.smtpConf,
					service.WithLogger(service.RuntimeLogger(rt)),
					service.WithTracer(service.RuntimeTracer(rt)),
				)
				if err != nil {
					panic(err)
				}
				return svc
			}()
			assert.ErrorIs(t, s.SendUserWelcomeEmail(tt.args.ctx, tt.args.recipient), tt.wantErr)
		})
	}
}

func TestEmailService_SendLicenseExpiryEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	now := time.Date(2026, time.September, 8, 9, 0, 0, 0, time.UTC)
	expiresAt := now.Add(7 * 24 * time.Hour)
	graceEndsAt := expiresAt.Add(30 * 24 * time.Hour)
	status := entitlement.AirGapStatus{
		State:         license.StateValid,
		LicenseID:     "license-id",
		Customer:      "ACME Inc.",
		SeatsLicensed: 25,
		ExpiresAt:     &expiresAt,
		GraceEndsAt:   &graceEndsAt,
	}
	smtpConf := &config.SMTPConfig{
		ClientURL:      "https://example.com/",
		SupportAddress: "support@example.com",
	}
	expectedTemplate, err := email.NewTemplate(
		path.Join("/templates", service.AirGapLicenseExpiryTemplate),
		&email.LicenseExpiryTemplateData{
			Subject:        "License expiration reminder",
			Customer:       "ACME Inc.",
			LicenseID:      "license-id",
			LicenseState:   "valid",
			LicenseExpires: expiresAt.Format(time.RFC1123),
			GraceEnds:      graceEndsAt.Format(time.RFC1123),
			SeatsLicensed:  25,
			SettingsURL:    "https://example.com/settings",
			SupportEmail:   "support@example.com",
		},
	)
	require.NoError(t, err)

	sender := mocksvc.NewMockEmailSender(ctrl)
	sender.EXPECT().SendEmail(
		gomock.Any(),
		"License expiration reminder",
		"billing@example.com",
		matchTemplate(expectedTemplate),
	).Return(nil)

	svc, err := service.NewEmailService(sender, "/templates", smtpConf)
	require.NoError(t, err)
	require.NoError(t, svc.SendLicenseExpiryEmail(ctx, "billing@example.com", status))
}

func TestEmailService_SendLicenseExpiryEmailInvalidStatus(t *testing.T) {
	t.Parallel()

	svc, err := service.NewEmailService(
		mocksvc.NewMockEmailSender(gomock.NewController(t)),
		"/templates",
		&config.SMTPConfig{},
	)
	require.NoError(t, err)

	err = svc.SendLicenseExpiryEmail(
		context.Background(),
		"billing@example.com",
		entitlement.AirGapStatus{},
	)
	require.ErrorIs(t, err, email.ErrInvalidLicenseExpiryTemplateData)
}

func TestEmailService_SendLicenseExpiryEmailSendFailure(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	graceEndsAt := expiresAt.Add(30 * 24 * time.Hour)
	sender := mocksvc.NewMockEmailSender(ctrl)
	sender.EXPECT().
		SendEmail(gomock.Any(), gomock.Any(), "billing@example.com", gomock.Any()).
		Return(assert.AnError)

	svc, err := service.NewEmailService(
		sender,
		"/templates",
		&config.SMTPConfig{
			ClientURL:      "https://example.com",
			SupportAddress: "support@example.com",
		},
	)
	require.NoError(t, err)

	err = svc.SendLicenseExpiryEmail(
		context.Background(),
		"billing@example.com",
		entitlement.AirGapStatus{
			State:         license.StateValid,
			LicenseID:     "license-id",
			Customer:      "ACME Inc.",
			SeatsLicensed: 10,
			ExpiresAt:     &expiresAt,
			GraceEndsAt:   &graceEndsAt,
		},
	)
	require.ErrorIs(t, err, service.ErrEmailSend)
}
