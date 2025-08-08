package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/smtp"
)

const (
	renewEmailAddress = "renew@elemo.app"

	authPasswordResetTemplate   = "email/password-reset.html"
	organizationInviteTemplate  = "email/organization-invite.html"
	systemLicenseExpiryTemplate = "email/license-expiry-reminder.html"
	userWelcomeTemplate         = "email/user-welcome.html"
)

// EmailSender defines the interface to send emails.
//
//go:generate mockgen -source=email.go -destination=../testutil/mock/email_sender_gen.go -package=mock -mock_names EmailSender=EmailSender
type EmailSender interface {
	// SendEmail sends an email to the given address using a template.
	SendEmail(ctx context.Context, subject, to string, template *email.Template) error
}

// EmailService defines the interface to send emails from templates.
//
//go:generate mockgen -source=email.go -destination=../testutil/mock/email_service_gen.go -package=mock -mock_names EmailService=EmailService
type EmailService interface {
	// SendEmail sends an email from a template to the list of active users.
	// SendEmail(ctx context.Context, subject string, template *email.Template, data any, users []*model.User) error

	// SendAuthPasswordResetEmail sends an email to the user with a link to
	// reset the password.
	SendAuthPasswordResetEmail(ctx context.Context, user *model.User, token string) error
	// SendOrganizationInvitationEmail sends an email to the invited user.
	SendOrganizationInvitationEmail(ctx context.Context, organization *model.Organization, user *model.User, token string) error
	// SendSystemLicenseExpiryEmail sends an email to the license owner when the license is about to expire.
	SendSystemLicenseExpiryEmail(ctx context.Context, licenseID, licenseEmail, licenseOrganization string, licenseExpiresAt time.Time) error
	// SendUserWelcomeEmail sends an email to the user to welcome it to the
	// system.
	SendUserWelcomeEmail(ctx context.Context, user *model.User) error
}

// emailService is the concrete implementation of the EmailService interface.
type emailService struct {
	*baseService
	client       EmailSender
	templatesDir string
	smtpConf     *config.SMTPConfig
}

func (s *emailService) sendEmail(ctx context.Context, subject string, template string, data email.TemplateData, user *model.User) error {
	tmpl, err := email.NewTemplate(path.Join(s.templatesDir, template), data)
	if err != nil {
		return errors.Join(ErrEmailSend, err)
	}

	if err := s.client.SendEmail(ctx, subject, user.Email, tmpl); err != nil {
		return errors.Join(ErrEmailSend, err)
	}

	return nil
}

func (s *emailService) SendAuthPasswordResetEmail(ctx context.Context, user *model.User, token string) error {
	ctx, span := s.tracer.Start(ctx, "service.emailService/SendAuthPasswordResetEmail")
	defer span.End()

	passwordResetURL := fmt.Sprintf("%s/reset-password?token=%s", s.smtpConf.ClientURL, token)

	data := &email.PasswordResetTemplateData{
		Subject:          "[Action Required] Reset your password",
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		PasswordResetURL: passwordResetURL,
		SupportEmail:     s.smtpConf.SupportAddress,
	}

	return s.sendEmail(ctx, data.Subject, authPasswordResetTemplate, data, user)
}

func (s *emailService) SendOrganizationInvitationEmail(ctx context.Context, organization *model.Organization, user *model.User, token string) error {
	ctx, span := s.tracer.Start(ctx, "service.emailService/SendOrganizationInvitationEmail")
	defer span.End()

	invitationURL := fmt.Sprintf("%s/organizations/join?workspace=%s&token=%s", s.smtpConf.ClientURL, organization.ID.String(), token)

	data := &email.OrganizationInviteTemplateData{
		Subject:          fmt.Sprintf("[Action Required] You have been invited to join %s", organization.Name),
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		OrganizationName: organization.Name,
		InvitationURL:    fmt.Sprintf("%s/redirect?url=%s", s.smtpConf.ClientURL, url.QueryEscape(invitationURL)),
		SupportEmail:     s.smtpConf.SupportAddress,
	}

	return s.sendEmail(ctx, data.Subject, organizationInviteTemplate, data, user)
}

func (s *emailService) SendSystemLicenseExpiryEmail(ctx context.Context, licenseID, licenseEmail, licenseOrganization string, licenseExpiresAt time.Time) error {
	ctx, span := s.tracer.Start(ctx, "service.emailService/SendSystemLicenseExpiryEmail")
	defer span.End()

	data := &email.LicenseExpiryTemplateData{
		Subject:             fmt.Sprintf("Your license for %s is about to expire", licenseOrganization),
		LicenseID:           licenseID,
		LicenseEmail:        licenseEmail,
		LicenseOrganization: licenseOrganization,
		LicenseExpiresAt:    licenseExpiresAt.Format(time.RFC850),
		ServerURL:           fmt.Sprintf("https://%s", s.smtpConf.ClientURL),
		RenewEmail:          renewEmailAddress,
		SupportEmail:        s.smtpConf.SupportAddress,
	}

	return s.sendEmail(ctx, data.Subject, systemLicenseExpiryTemplate, data, &model.User{Email: licenseEmail})
}

func (s *emailService) SendUserWelcomeEmail(ctx context.Context, user *model.User) error {
	ctx, span := s.tracer.Start(ctx, "service.emailService/SendUserWelcomeEmail")
	defer span.End()

	loginURL := fmt.Sprintf("%s/auth/login", s.smtpConf.ClientURL)

	data := &email.UserWelcomeTemplateData{
		Subject:      "Welcome to Elemo",
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		LoginURL:     fmt.Sprintf("%s/redirect?url=%s", s.smtpConf.ClientURL, url.QueryEscape(loginURL)),
		SupportEmail: s.smtpConf.SupportAddress,
	}

	return s.sendEmail(ctx, data.Subject, userWelcomeTemplate, data, user)
}

// NewEmailService creates a new email service.
func NewEmailService(client EmailSender, templatesDir string, smtpConf *config.SMTPConfig, opts ...Option) (EmailService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &emailService{
		baseService:  s,
		client:       client,
		templatesDir: templatesDir,
		smtpConf:     smtpConf,
	}

	if svc.client == nil {
		return nil, smtp.ErrNoSMTPClient
	}

	return svc, nil
}
