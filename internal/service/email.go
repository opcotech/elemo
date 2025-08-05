package service

import (
	"context"
	"errors"
	"fmt"
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
//go:generate mockgen -source=email.go -destination=../testutil/mock/email_sender_gen.go -package=mock
type EmailSender interface {
	// SendEmail sends an email to the given address using a template.
	SendEmail(ctx context.Context, subject, to string, template *email.Template) error
}

// EmailService defines the interface to send emails from templates.
type EmailService interface {
	// SendEmail sends an email from a template to the list of active users.
	// SendEmail(ctx context.Context, subject string, template *email.Template, data any, users []*model.User) error

	// SendAuthPasswordResetEmail sends an email to the user with a link to
	// reset the password.
	SendAuthPasswordResetEmail(ctx context.Context, resetPath string, user *model.User) error
	// SendOrganizationInvitationEmail sends an email to the invited user.
	SendOrganizationInvitationEmail(ctx context.Context, invitationPath string, organization *model.Organization, user *model.User) error
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

func (s *emailService) SendAuthPasswordResetEmail(ctx context.Context, resetPath string, user *model.User) error {
	ctx, span := s.tracer.Start(ctx, "service.emailService/SendAuthPasswordResetEmail")
	defer span.End()

	data := &email.PasswordResetTemplateData{
		Subject:          "Reset your password",
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		PasswordResetURL: fmt.Sprintf("https://%s", path.Join(s.smtpConf.Hostname, resetPath)),
		SupportEmail:     s.smtpConf.SupportAddress,
	}

	return s.sendEmail(ctx, data.Subject, authPasswordResetTemplate, data, user)
}

func (s *emailService) SendOrganizationInvitationEmail(ctx context.Context, invitationPath string, organization *model.Organization, user *model.User) error {
	ctx, span := s.tracer.Start(ctx, "service.emailService/SendOrganizationInvitationEmail")
	defer span.End()

	data := &email.OrganizationInviteTemplateData{
		Subject:          fmt.Sprintf("You have been invited to join %s", organization.Name),
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		OrganizationName: organization.Name,
		InvitationURL:    fmt.Sprintf("https://%s", path.Join(s.smtpConf.Hostname, invitationPath)),
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
		ServerURL:           fmt.Sprintf("https://%s", s.smtpConf.Hostname),
		RenewEmail:          renewEmailAddress,
		SupportEmail:        s.smtpConf.SupportAddress,
	}

	return s.sendEmail(ctx, data.Subject, systemLicenseExpiryTemplate, data, &model.User{Email: licenseEmail})
}

func (s *emailService) SendUserWelcomeEmail(ctx context.Context, user *model.User) error {
	ctx, span := s.tracer.Start(ctx, "service.emailService/SendUserWelcomeEmail")
	defer span.End()

	data := &email.UserWelcomeTemplateData{
		Subject:      fmt.Sprintf("Welcome to %s", s.smtpConf.Hostname),
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		LoginURL:     fmt.Sprintf("https://%s/sign-in", s.smtpConf.Hostname),
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
