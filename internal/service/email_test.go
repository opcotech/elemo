package service

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/smtp"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/testutil/mock"
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
				client: func() EmailSender {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mock.NewEmailSender(ctrl)
				}(),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
				},
			},
			want: &emailService{
				baseService: &baseService{
					logger: mock.NewMockLogger(nil),
					tracer: mock.NewMockTracer(nil),
				},
				client: func() EmailSender {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mock.NewEmailSender(ctrl)
				}(),
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
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
				},
			},
			wantErr: smtp.ErrNoSMTPClient,
		},
		{
			name: "new email service with invalid options",
			args: args{
				client: func() EmailSender {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mock.NewEmailSender(ctrl)
				}(),
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
				client: func() EmailSender {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mock.NewEmailSender(ctrl)
				}(),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
				opts: []Option{
					WithTracer(mock.NewMockTracer(nil)),
				},
			},
			want: &emailService{
				baseService: &baseService{
					logger: log.DefaultLogger(),
					tracer: mock.NewMockTracer(nil),
				},
				client: func() EmailSender {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mock.NewEmailSender(ctrl)
				}(),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
			},
		},
		{
			name: "new email service with no tracer",
			args: args{
				client: func() EmailSender {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mock.NewEmailSender(ctrl)
				}(),
				templatesDir: "/templates",
				smtpConf:     new(config.SMTPConfig),
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
				},
			},
			want: &emailService{
				baseService: &baseService{
					logger: mock.NewMockLogger(nil),
					tracer: tracing.NoopTracer(),
				},
				client: func() EmailSender {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mock.NewEmailSender(ctrl)
				}(),
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
		baseService  func(ctrl *gomock.Controller, ctx context.Context) *baseService
		client       func(ctrl *gomock.Controller, ctx context.Context, templatesDir, token string, smtpConf *config.SMTPConfig, user *model.User) EmailSender
		templatesDir string
		smtpConf     *config.SMTPConfig
	}
	type args struct {
		ctx   context.Context
		user  *model.User
		token string
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.emailService/SendAuthPasswordResetEmail", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				client: func(ctrl *gomock.Controller, ctx context.Context, templatesDir, token string, smtpConf *config.SMTPConfig, user *model.User) EmailSender {
					subject := "[Action Required] Reset your password"

					passwordResetURL := fmt.Sprintf("%s/reset-password?token=%s", smtpConf.ClientURL, token)
					template, err := email.NewTemplate(
						path.Join(templatesDir, authPasswordResetTemplate),
						&email.PasswordResetTemplateData{
							Subject:          subject,
							FirstName:        user.FirstName,
							LastName:         user.LastName,
							PasswordResetURL: passwordResetURL,
							SupportEmail:     smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := mock.NewEmailSender(ctrl)
					client.EXPECT().SendEmail(ctx, subject, user.Email, matchTemplate(template)).Return(nil)

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
				user: &model.User{
					Username:  "test",
					FirstName: "Test",
					LastName:  "User",
					Email:     "test@example.com",
				},
				token: "test-token",
			},
		},
		{
			name: "send auth password reset email failed",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.emailService/SendAuthPasswordResetEmail", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				client: func(ctrl *gomock.Controller, ctx context.Context, templatesDir, token string, smtpConf *config.SMTPConfig, user *model.User) EmailSender {
					subject := "[Action Required] Reset your password"

					passwordResetURL := fmt.Sprintf("%s/reset-password?token=%s", smtpConf.ClientURL, token)
					template, err := email.NewTemplate(
						path.Join(templatesDir, authPasswordResetTemplate),
						&email.PasswordResetTemplateData{
							Subject:          subject,
							FirstName:        user.FirstName,
							LastName:         user.LastName,
							PasswordResetURL: passwordResetURL,
							SupportEmail:     smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := mock.NewEmailSender(ctrl)
					client.EXPECT().SendEmail(ctx, subject, user.Email, matchTemplate(template)).Return(assert.AnError)

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
				user: &model.User{
					Username:  "test",
					FirstName: "Test",
					LastName:  "User",
					Email:     "test@example.com",
				},
				token: "test-token",
			},
			wantErr: ErrEmailSend,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := &emailService{
				baseService:  tt.fields.baseService(ctrl, tt.args.ctx),
				client:       tt.fields.client(ctrl, tt.args.ctx, tt.fields.templatesDir, tt.args.token, tt.fields.smtpConf, tt.args.user),
				templatesDir: tt.fields.templatesDir,
				smtpConf:     tt.fields.smtpConf,
			}
			assert.ErrorIs(t, s.SendAuthPasswordResetEmail(tt.args.ctx, tt.args.user, tt.args.token), tt.wantErr)
		})
	}
}

func TestEmailService_SendOrganizationInvitationEmail(t *testing.T) {
	type fields struct {
		baseService  func(ctrl *gomock.Controller, ctx context.Context) *baseService
		client       func(ctrl *gomock.Controller, ctx context.Context, templatesDir, token string, smtpConf *config.SMTPConfig, organization *model.Organization, user *model.User) EmailSender
		templatesDir string
		smtpConf     *config.SMTPConfig
	}
	type args struct {
		ctx            context.Context
		invitationPath string
		organization   *model.Organization
		user           *model.User
		token          string
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.emailService/SendOrganizationInvitationEmail", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				client: func(ctrl *gomock.Controller, ctx context.Context, templatesDir, token string, smtpConf *config.SMTPConfig, organization *model.Organization, user *model.User) EmailSender {
					subject := fmt.Sprintf("[Action Required] You have been invited to join %s", organization.Name)

					invitationURL := fmt.Sprintf("%s/organizations/join?organization=%s&token=%s", smtpConf.ClientURL, organization.ID.String(), token)
					template, err := email.NewTemplate(
						path.Join(templatesDir, organizationInviteTemplate),
						&email.OrganizationInviteTemplateData{
							Subject:          subject,
							OrganizationName: organization.Name,
							InvitationURL:    invitationURL,
							SupportEmail:     smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := mock.NewEmailSender(ctrl)
					client.EXPECT().SendEmail(ctx, subject, user.Email, matchTemplate(template)).Return(nil)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					ClientURL:      "https://example.com",
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
					LastName:  "User",
					Email:     "test@example.com",
				},
			},
		},
		{
			name: "send invitation email failed",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.emailService/SendOrganizationInvitationEmail", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				client: func(ctrl *gomock.Controller, ctx context.Context, templatesDir, token string, smtpConf *config.SMTPConfig, organization *model.Organization, user *model.User) EmailSender {
					subject := fmt.Sprintf("[Action Required] You have been invited to join %s", organization.Name)

					invitationURL := fmt.Sprintf("%s/organizations/join?organization=%s&token=%s", smtpConf.ClientURL, organization.ID.String(), token)
					template, err := email.NewTemplate(
						path.Join(templatesDir, organizationInviteTemplate),
						&email.OrganizationInviteTemplateData{
							Subject:          subject,
							OrganizationName: organization.Name,
							InvitationURL:    invitationURL,
							SupportEmail:     smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := mock.NewEmailSender(ctrl)
					client.EXPECT().SendEmail(ctx, subject, user.Email, matchTemplate(template)).Return(assert.AnError)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					ClientURL:      "https://example.com",
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
					LastName:  "User",
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

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := &emailService{
				baseService:  tt.fields.baseService(ctrl, tt.args.ctx),
				client:       tt.fields.client(ctrl, tt.args.ctx, tt.fields.templatesDir, tt.args.token, tt.fields.smtpConf, tt.args.organization, tt.args.user),
				templatesDir: tt.fields.templatesDir,
				smtpConf:     tt.fields.smtpConf,
			}
			assert.ErrorIs(t, s.SendOrganizationInvitationEmail(tt.args.ctx, tt.args.organization, tt.args.user, tt.args.token), tt.wantErr)
		})
	}
}

func TestEmailService_SendSystemLicenseExpiryEmail(t *testing.T) {
	type fields struct {
		baseService  func(ctrl *gomock.Controller, ctx context.Context) *baseService
		client       func(ctrl *gomock.Controller, ctx context.Context, templatesDir string, smtpConf *config.SMTPConfig, licenseID, licenseEmail, licenseOrganization string, licenseExpiresAt time.Time) EmailSender
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.emailService/SendSystemLicenseExpiryEmail", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				client: func(ctrl *gomock.Controller, ctx context.Context, templatesDir string, smtpConf *config.SMTPConfig, licenseID, licenseEmail, licenseOrganization string, licenseExpiresAt time.Time) EmailSender {
					subject := fmt.Sprintf("Your license for %s is about to expire", licenseOrganization)

					template, err := email.NewTemplate(
						path.Join(templatesDir, systemLicenseExpiryTemplate),
						&email.LicenseExpiryTemplateData{
							Subject:             subject,
							LicenseID:           licenseID,
							LicenseEmail:        licenseEmail,
							LicenseOrganization: licenseOrganization,
							LicenseExpiresAt:    licenseExpiresAt.Format(time.RFC850),
							ServerURL:           fmt.Sprintf("https://%s", smtpConf.ClientURL),
							RenewEmail:          renewEmailAddress,
							SupportEmail:        smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := mock.NewEmailSender(ctrl)
					client.EXPECT().SendEmail(ctx, subject, licenseEmail, matchTemplate(template)).Return(nil)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					ClientURL:      "https://example.com",
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.emailService/SendSystemLicenseExpiryEmail", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				client: func(ctrl *gomock.Controller, ctx context.Context, templatesDir string, smtpConf *config.SMTPConfig, licenseID, licenseEmail, licenseOrganization string, licenseExpiresAt time.Time) EmailSender {
					subject := fmt.Sprintf("Your license for %s is about to expire", licenseOrganization)

					template, err := email.NewTemplate(
						path.Join(templatesDir, systemLicenseExpiryTemplate),
						&email.LicenseExpiryTemplateData{
							Subject:             subject,
							LicenseID:           licenseID,
							LicenseEmail:        licenseEmail,
							LicenseOrganization: licenseOrganization,
							LicenseExpiresAt:    licenseExpiresAt.Format(time.RFC850),
							ServerURL:           fmt.Sprintf("https://%s", smtpConf.ClientURL),
							RenewEmail:          renewEmailAddress,
							SupportEmail:        smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := mock.NewEmailSender(ctrl)
					client.EXPECT().SendEmail(ctx, subject, licenseEmail, matchTemplate(template)).Return(assert.AnError)

					return client
				},
				templatesDir: "/templates",
				smtpConf: &config.SMTPConfig{
					ClientURL:      "https://example.com",
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

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := &emailService{
				baseService: tt.fields.baseService(ctrl, tt.args.ctx),
				client: tt.fields.client(
					ctrl,
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
		baseService  func(ctrl *gomock.Controller, ctx context.Context) *baseService
		client       func(ctrl *gomock.Controller, ctx context.Context, templatesDir string, smtpConf *config.SMTPConfig, user *model.User) EmailSender
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
				baseService: func(ctrl *gomock.Controller, ctx context.Context) *baseService {
					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.emailService/SendUserWelcomeEmail", gomock.Len(0)).Return(ctx, span)

					return &baseService{
						logger: mock.NewMockLogger(ctrl),
						tracer: tracer,
					}
				},
				client: func(ctrl *gomock.Controller, ctx context.Context, templatesDir string, smtpConf *config.SMTPConfig, user *model.User) EmailSender {
					subject := "Welcome to Elemo"

					template, err := email.NewTemplate(
						path.Join(templatesDir, userWelcomeTemplate),
						&email.UserWelcomeTemplateData{
							Subject:      subject,
							FirstName:    user.FirstName,
							LastName:     user.LastName,
							LoginURL:     fmt.Sprintf("%s/redirect?url=%s", smtpConf.ClientURL, url.QueryEscape(fmt.Sprintf("%s/auth/login", smtpConf.ClientURL))),
							SupportEmail: smtpConf.SupportAddress,
						},
					)
					require.NoError(t, err)

					client := mock.NewEmailSender(ctrl)
					client.EXPECT().SendEmail(ctx, subject, user.Email, matchTemplate(template)).Return(assert.AnError)

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
				user: &model.User{
					Username:  "test",
					FirstName: "Test",
					LastName:  "User",
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

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := &emailService{
				baseService:  tt.fields.baseService(ctrl, tt.args.ctx),
				client:       tt.fields.client(ctrl, tt.args.ctx, tt.fields.templatesDir, tt.fields.smtpConf, tt.args.user),
				templatesDir: tt.fields.templatesDir,
				smtpConf:     tt.fields.smtpConf,
			}
			assert.ErrorIs(t, s.SendUserWelcomeEmail(tt.args.ctx, tt.args.user), tt.wantErr)
		})
	}
}
