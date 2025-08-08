package service

import (
	"context"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/smtp"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestNewEmailService(t *testing.T) {
	type args struct {
		client       EmailSender
		templatesDir string
		smtpConf     *config.SMTPConfig
		opts         []Option
	}
	tests := []struct {
		name    string
		args    args
		want    EmailService
		wantErr error
	}{
		{
			name: "new email service",
			args: args{
				client:       new(mock.SMTPClient),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
				},
			},
			want: &emailService{
				baseService: &baseService{
					logger: new(mock.Logger),
					tracer: new(mock.Tracer),
				},
				client:       new(mock.SMTPClient),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
			},
		},
		{
			name: "new email service with no email sender",
			args: args{
				client:       nil,
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
				opts: []Option{
					WithLogger(new(mock.Logger)),
					WithTracer(new(mock.Tracer)),
				},
			},
			wantErr: smtp.ErrNoSMTPClient,
		},
		{
			name: "new email service with invalid options",
			args: args{
				client:       new(mock.SMTPClient),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
				opts: []Option{
					WithLogger(nil),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new email service with no logger",
			args: args{
				client:       new(mock.SMTPClient),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
				opts: []Option{
					WithTracer(new(mock.Tracer)),
				},
			},
			want: &emailService{
				baseService: &baseService{
					logger: log.DefaultLogger(),
					tracer: new(mock.Tracer),
				},
				client:       new(mock.SMTPClient),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
			},
		},
		{
			name: "new email service with no tracer",
			args: args{
				client:       new(mock.SMTPClient),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
				opts: []Option{
					WithLogger(new(mock.Logger)),
				},
			},
			want: &emailService{
				baseService: &baseService{
					logger: new(mock.Logger),
					tracer: tracing.NoopTracer(),
				},
				client:       new(mock.SMTPClient),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewEmailService(tt.args.client, tt.args.templatesDir, tt.args.smtpConf, tt.args.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestEmailService_SendAuthPasswordResetEmail(t *testing.T) {
	type fields struct {
		baseService  func(ctx context.Context) *baseService
		client       func(ctx context.Context, templatesDir, resetPath string, smtpConf *config.SMTPConfig, user *model.User) EmailSender
		templatesDir string
		smtpConf     *config.SMTPConfig
	}
	type args struct {
		ctx       context.Context
		resetPath string
		user      *model.User
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.emailService/SendAuthPasswordResetEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				client: func(ctx context.Context, templatesDir, resetPath string, smtpConf *config.SMTPConfig, user *model.User) EmailSender {
					subject := "Reset your password"

					template, err := email.NewTemplate(
						path.Join(templatesDir, authPasswordResetTemplate),
						&email.PasswordResetTemplateData{
							Subject:          subject,
							Username:         user.Username,
							FirstName:        user.FirstName,
							PasswordResetURL: fmt.Sprintf("https://%s", path.Join(smtpConf.Hostname, resetPath)),
							SupportEmail:     smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := new(mock.SMTPClient)
					client.On("SendEmail", ctx, subject, user.Email, template).Return(nil)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					Hostname:       "example.com",
					SupportAddress: "support@example.com",
				},
			},
			args: args{
				ctx:       context.Background(),
				resetPath: "/reset",
				user: &model.User{
					Username:  "test",
					FirstName: "Test",
					Email:     "test@example.com",
				},
			},
		},
		{
			name: "send auth password reset email failed",
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.emailService/SendAuthPasswordResetEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				client: func(ctx context.Context, templatesDir, resetPath string, smtpConf *config.SMTPConfig, user *model.User) EmailSender {
					subject := "Reset your password"

					template, err := email.NewTemplate(
						path.Join(templatesDir, authPasswordResetTemplate),
						&email.PasswordResetTemplateData{
							Subject:          subject,
							Username:         user.Username,
							FirstName:        user.FirstName,
							PasswordResetURL: fmt.Sprintf("https://%s", path.Join(smtpConf.Hostname, resetPath)),
							SupportEmail:     smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := new(mock.SMTPClient)
					client.On("SendEmail", ctx, subject, user.Email, template).Return(assert.AnError)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					Hostname:       "example.com",
					SupportAddress: "support@example.com",
				},
			},
			args: args{
				ctx:       context.Background(),
				resetPath: "/reset",
				user: &model.User{
					Username:  "test",
					FirstName: "Test",
					Email:     "test@example.com",
				},
			},
			wantErr: ErrEmailSend,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &emailService{
				baseService:  tt.fields.baseService(tt.args.ctx),
				client:       tt.fields.client(tt.args.ctx, tt.fields.templatesDir, tt.args.resetPath, tt.fields.smtpConf, tt.args.user),
				templatesDir: tt.fields.templatesDir,
				smtpConf:     tt.fields.smtpConf,
			}
			assert.ErrorIs(t, s.SendAuthPasswordResetEmail(tt.args.ctx, tt.args.resetPath, tt.args.user), tt.wantErr)
		})
	}
}

func TestEmailService_SendOrganizationInvitationEmail(t *testing.T) {
	type fields struct {
		baseService  func(ctx context.Context) *baseService
		client       func(ctx context.Context, templatesDir, invitationPath string, smtpConf *config.SMTPConfig, organization *model.Organization, user *model.User) EmailSender
		templatesDir string
		smtpConf     *config.SMTPConfig
	}
	type args struct {
		ctx            context.Context
		invitationPath string
		organization   *model.Organization
		user           *model.User
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.emailService/SendOrganizationInvitationEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				client: func(ctx context.Context, templatesDir, invitationPath string, smtpConf *config.SMTPConfig, organization *model.Organization, user *model.User) EmailSender {
					subject := fmt.Sprintf("You have been invited to join %s", organization.Name)

					template, err := email.NewTemplate(
						path.Join(templatesDir, organizationInviteTemplate),
						&email.OrganizationInviteTemplateData{
							Subject:          subject,
							Username:         user.Username,
							FirstName:        user.FirstName,
							OrganizationName: organization.Name,
							InvitationURL:    fmt.Sprintf("https://%s", path.Join(smtpConf.Hostname, invitationPath)),
							SupportEmail:     smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := new(mock.SMTPClient)
					client.On("SendEmail", ctx, subject, user.Email, template).Return(nil)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					Hostname:       "example.com",
					SupportAddress: "support@example.com",
				},
			},
			args: args{
				ctx:            context.Background(),
				invitationPath: "/invitation",
				organization: &model.Organization{
					Name: "test",
				},
				user: &model.User{
					Username:  "test",
					FirstName: "Test",
					Email:     "test@example.com",
				},
			},
		},
		{
			name: "send invitation email failed",
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.emailService/SendOrganizationInvitationEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				client: func(ctx context.Context, templatesDir, invitationPath string, smtpConf *config.SMTPConfig, organization *model.Organization, user *model.User) EmailSender {
					subject := fmt.Sprintf("You have been invited to join %s", organization.Name)

					template, err := email.NewTemplate(
						path.Join(templatesDir, organizationInviteTemplate),
						&email.OrganizationInviteTemplateData{
							Subject:          subject,
							Username:         user.Username,
							FirstName:        user.FirstName,
							OrganizationName: organization.Name,
							InvitationURL:    fmt.Sprintf("https://%s", path.Join(smtpConf.Hostname, invitationPath)),
							SupportEmail:     smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := new(mock.SMTPClient)
					client.On("SendEmail", ctx, subject, user.Email, template).Return(assert.AnError)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					Hostname:       "example.com",
					SupportAddress: "support@example.com",
				},
			},
			args: args{
				ctx:            context.Background(),
				invitationPath: "/invitation",
				organization: &model.Organization{
					Name: "test",
				},
				user: &model.User{
					Username:  "test",
					FirstName: "Test",
					Email:     "test@example.com",
				},
			},
			wantErr: ErrEmailSend,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &emailService{
				baseService:  tt.fields.baseService(tt.args.ctx),
				client:       tt.fields.client(tt.args.ctx, tt.fields.templatesDir, tt.args.invitationPath, tt.fields.smtpConf, tt.args.organization, tt.args.user),
				templatesDir: tt.fields.templatesDir,
				smtpConf:     tt.fields.smtpConf,
			}
			assert.ErrorIs(t, s.SendOrganizationInvitationEmail(tt.args.ctx, tt.args.invitationPath, tt.args.organization, tt.args.user), tt.wantErr)
		})
	}
}

func TestEmailService_SendSystemLicenseExpiryEmail(t *testing.T) {
	type fields struct {
		baseService  func(ctx context.Context) *baseService
		client       func(ctx context.Context, templatesDir string, smtpConf *config.SMTPConfig, licenseID, licenseEmail, licenseOrganization string, licenseExpiresAt time.Time) EmailSender
		templatesDir string
		smtpConf     *config.SMTPConfig
	}
	type args struct {
		ctx                 context.Context
		licenseID           string
		licenseEmail        string
		licenseOrganization string
		licenseExpiresAt    time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "send license expiry email",
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.emailService/SendSystemLicenseExpiryEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				client: func(ctx context.Context, templatesDir string, smtpConf *config.SMTPConfig, licenseID, licenseEmail, licenseOrganization string, licenseExpiresAt time.Time) EmailSender {
					subject := fmt.Sprintf("Your license for %s is about to expire", licenseOrganization)

					template, err := email.NewTemplate(
						path.Join(templatesDir, systemLicenseExpiryTemplate),
						&email.LicenseExpiryTemplateData{
							Subject:             subject,
							LicenseID:           licenseID,
							LicenseEmail:        licenseEmail,
							LicenseOrganization: licenseOrganization,
							LicenseExpiresAt:    licenseExpiresAt.Format(time.RFC850),
							ServerURL:           fmt.Sprintf("https://%s", smtpConf.Hostname),
							RenewEmail:          renewEmailAddress,
							SupportEmail:        smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := new(mock.SMTPClient)
					client.On("SendEmail", ctx, subject, licenseEmail, template).Return(nil)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					Hostname:       "example.com",
					SupportAddress: "support@example.com",
				},
			},
			args: args{
				ctx:                 context.Background(),
				licenseID:           "123456789",
				licenseEmail:        "info@example.com",
				licenseOrganization: "ACME Inc.",
				licenseExpiresAt:    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "send auth password reset email failed",
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.emailService/SendSystemLicenseExpiryEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				client: func(ctx context.Context, templatesDir string, smtpConf *config.SMTPConfig, licenseID, licenseEmail, licenseOrganization string, licenseExpiresAt time.Time) EmailSender {
					subject := fmt.Sprintf("Your license for %s is about to expire", licenseOrganization)

					template, err := email.NewTemplate(
						path.Join(templatesDir, systemLicenseExpiryTemplate),
						&email.LicenseExpiryTemplateData{
							Subject:             subject,
							LicenseID:           licenseID,
							LicenseEmail:        licenseEmail,
							LicenseOrganization: licenseOrganization,
							LicenseExpiresAt:    licenseExpiresAt.Format(time.RFC850),
							ServerURL:           fmt.Sprintf("https://%s", smtpConf.Hostname),
							RenewEmail:          renewEmailAddress,
							SupportEmail:        smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := new(mock.SMTPClient)
					client.On("SendEmail", ctx, subject, licenseEmail, template).Return(assert.AnError)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					Hostname:       "example.com",
					SupportAddress: "support@example.com",
				},
			},
			args: args{
				ctx:                 context.Background(),
				licenseID:           "123456789",
				licenseEmail:        "info@example.com",
				licenseOrganization: "ACME Inc.",
				licenseExpiresAt:    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: ErrEmailSend,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &emailService{
				baseService: tt.fields.baseService(tt.args.ctx),
				client: tt.fields.client(
					tt.args.ctx,
					tt.fields.templatesDir,
					tt.fields.smtpConf,
					tt.args.licenseID,
					tt.args.licenseEmail,
					tt.args.licenseOrganization,
					tt.args.licenseExpiresAt,
				),
				templatesDir: tt.fields.templatesDir,
				smtpConf:     tt.fields.smtpConf,
			}
			err := s.SendSystemLicenseExpiryEmail(
				tt.args.ctx,
				tt.args.licenseID,
				tt.args.licenseEmail,
				tt.args.licenseOrganization,
				tt.args.licenseExpiresAt,
			)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestEmailService_SendUserWelcomeEmail(t *testing.T) {
	type fields struct {
		baseService  func(ctx context.Context) *baseService
		client       func(ctx context.Context, templatesDir string, smtpConf *config.SMTPConfig, user *model.User) EmailSender
		templatesDir string
		smtpConf     *config.SMTPConfig
	}
	type args struct {
		ctx  context.Context
		user *model.User
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
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.emailService/SendUserWelcomeEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				client: func(ctx context.Context, templatesDir string, smtpConf *config.SMTPConfig, user *model.User) EmailSender {
					subject := fmt.Sprintf("Welcome to %s", smtpConf.Hostname)

					template, err := email.NewTemplate(
						path.Join(templatesDir, userWelcomeTemplate),
						&email.UserWelcomeTemplateData{
							Subject:      subject,
							Username:     user.Username,
							FirstName:    user.FirstName,
							LoginURL:     fmt.Sprintf("https://%s/sign-in", smtpConf.Hostname),
							SupportEmail: smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := new(mock.SMTPClient)
					client.On("SendEmail", ctx, subject, user.Email, template).Return(nil)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					Hostname:       "example.com",
					SupportAddress: "support@example.com",
				},
			},
			args: args{
				ctx: context.Background(),
				user: &model.User{
					Username:  "test",
					FirstName: "Test",
					Email:     "test@example.com",
				},
			},
		},
		{
			name: "send auth password reset email failed",
			fields: fields{
				baseService: func(ctx context.Context) *baseService {
					span := new(mock.Span)
					span.On("End", []trace.SpanEndOption(nil)).Return()

					tracer := new(mock.Tracer)
					tracer.On("Start", ctx, "service.emailService/SendUserWelcomeEmail", []trace.SpanStartOption(nil)).Return(ctx, span)

					return &baseService{
						logger: new(mock.Logger),
						tracer: tracer,
					}
				},
				client: func(ctx context.Context, templatesDir string, smtpConf *config.SMTPConfig, user *model.User) EmailSender {
					subject := fmt.Sprintf("Welcome to %s", smtpConf.Hostname)

					template, err := email.NewTemplate(
						path.Join(templatesDir, userWelcomeTemplate),
						&email.UserWelcomeTemplateData{
							Subject:      subject,
							Username:     user.Username,
							FirstName:    user.FirstName,
							LoginURL:     fmt.Sprintf("https://%s/sign-in", smtpConf.Hostname),
							SupportEmail: smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := new(mock.SMTPClient)
					client.On("SendEmail", ctx, subject, user.Email, template).Return(assert.AnError)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					Hostname:       "example.com",
					SupportAddress: "support@example.com",
				},
			},
			args: args{
				ctx: context.Background(),
				user: &model.User{
					Username:  "test",
					FirstName: "Test",
					Email:     "test@example.com",
				},
			},
			wantErr: ErrEmailSend,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &emailService{
				baseService:  tt.fields.baseService(tt.args.ctx),
				client:       tt.fields.client(tt.args.ctx, tt.fields.templatesDir, tt.fields.smtpConf, tt.args.user),
				templatesDir: tt.fields.templatesDir,
				smtpConf:     tt.fields.smtpConf,
			}
			assert.ErrorIs(t, s.SendUserWelcomeEmail(tt.args.ctx, tt.args.user), tt.wantErr)
		})
	}
}
