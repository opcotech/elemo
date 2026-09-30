package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/smtp"
)

const (
	licenseExpiryTemplate      = "email/license-expiry-reminder.html"
	authPasswordResetTemplate  = "email/password-reset.html"
	organizationInviteTemplate = "email/organization-invite.html"
	userWelcomeTemplate        = "email/user-welcome.html"
)

// EmailSender defines the interface to send emails.
type EmailSender interface {
	// SendEmail sends an email to the given address using a template.
	SendEmail(ctx context.Context, subject, to string, template *email.Template) error
}

// EmailService defines the interface to send emails from templates.
//
//go:generate go tool mockgen -destination=mock/mock_email_gen.go -package=mocksvc . EmailService,EmailSender
type EmailService interface {
	// SendAuthPasswordResetEmail sends an email to the user with a link to
	// reset the password.
	SendAuthPasswordResetEmail(ctx context.Context, recipient email.Recipient, token string) error
	// SendOrganizationInvitationEmail sends an email to the invited user.
	SendOrganizationInvitationEmail(ctx context.Context, organizationID model.ID, organizationName string, recipient email.Recipient, token string) error
	// SendLicenseExpiryEmail sends a license expiration reminder.
	SendLicenseExpiryEmail(ctx context.Context, recipient string, status entitlement.AirGapStatus) error
	// SendUserWelcomeEmail sends an email to the user to welcome it to the
	// system.
	SendUserWelcomeEmail(ctx context.Context, recipient email.Recipient) error
}

// emailService is the concrete implementation of the EmailService interface.
type emailService struct {
	runtime
	client       EmailSender
	templatesDir string
	smtpConf     *config.SMTPConfig
}

func (s *emailService) sendEmail(ctx context.Context, subject string, template string, data email.TemplateData, to string) error {
	tmpl, err := email.NewTemplate(path.Join(s.templatesDir, template), data)
	if err != nil {
		return errors.Join(ErrEmailSend, err)
	}

	if err := s.client.SendEmail(ctx, subject, to, tmpl); err != nil {
		return errors.Join(ErrEmailSend, err)
	}

	return nil
}

func (s *emailService) SendAuthPasswordResetEmail(ctx context.Context, recipient email.Recipient, token string) error {
	ctx, span := s.tracer.Start(ctx, "service.emailService/SendAuthPasswordResetEmail")
	defer span.End()

	passwordResetURL := fmt.Sprintf("%s/reset-password?token=%s", s.smtpConf.ClientURL, token)

	data := &email.PasswordResetTemplateData{
		Subject:          "[Action Required] Reset your password",
		FirstName:        recipient.FirstName,
		LastName:         recipient.LastName,
		PasswordResetURL: passwordResetURL,
		SupportEmail:     s.smtpConf.SupportAddress,
	}

	return s.sendEmail(ctx, data.Subject, authPasswordResetTemplate, data, recipient.Email)
}

func (s *emailService) SendOrganizationInvitationEmail(ctx context.Context, organizationID model.ID, organizationName string, recipient email.Recipient, token string) error {
	ctx, span := s.tracer.Start(ctx, "service.emailService/SendOrganizationInvitationEmail")
	defer span.End()

	invitationURL := fmt.Sprintf("%s/organizations/join?organization=%s&token=%s", s.smtpConf.ClientURL, organizationID.String(), token)

	data := &email.OrganizationInviteTemplateData{
		Subject:          fmt.Sprintf("[Action Required] You have been invited to join %s", organizationName),
		OrganizationName: organizationName,
		InvitationURL:    invitationURL,
		SupportEmail:     s.smtpConf.SupportAddress,
	}

	return s.sendEmail(ctx, data.Subject, organizationInviteTemplate, data, recipient.Email)
}

func (s *emailService) SendLicenseExpiryEmail(ctx context.Context, recipient string, status entitlement.AirGapStatus) error {
	ctx, span := s.tracer.Start(ctx, "service.emailService/SendLicenseExpiryEmail")
	defer span.End()

	if status.ExpiresAt == nil || status.GraceEndsAt == nil {
		return errors.Join(ErrEmailSend, email.ErrInvalidLicenseExpiryTemplateData)
	}

	data := &email.LicenseExpiryTemplateData{
		Subject:        "License expiration reminder",
		Customer:       status.Customer,
		LicenseID:      status.LicenseID,
		LicenseState:   status.State.String(),
		LicenseExpires: status.ExpiresAt.UTC().Format(time.RFC1123),
		GraceEnds:      status.GraceEndsAt.UTC().Format(time.RFC1123),
		SeatsLicensed:  status.SeatsLicensed,
		SettingsURL:    strings.TrimRight(s.smtpConf.ClientURL, "/") + "/settings",
		SupportEmail:   s.smtpConf.SupportAddress,
	}

	return s.sendEmail(ctx, data.Subject, licenseExpiryTemplate, data, recipient)
}

func (s *emailService) SendUserWelcomeEmail(ctx context.Context, recipient email.Recipient) error {
	ctx, span := s.tracer.Start(ctx, "service.emailService/SendUserWelcomeEmail")
	defer span.End()

	loginURL := fmt.Sprintf("%s/auth/login", s.smtpConf.ClientURL)

	data := &email.UserWelcomeTemplateData{
		Subject:      "Welcome to Elemo",
		FirstName:    recipient.FirstName,
		LastName:     recipient.LastName,
		LoginURL:     fmt.Sprintf("%s/redirect?url=%s", s.smtpConf.ClientURL, url.QueryEscape(loginURL)),
		SupportEmail: s.smtpConf.SupportAddress,
	}

	return s.sendEmail(ctx, data.Subject, userWelcomeTemplate, data, recipient.Email)
}

// NewEmailService creates a new email service.
func NewEmailService(client EmailSender, templatesDir string, smtpConf *config.SMTPConfig, opts ...Option) (EmailService, error) {
	rt, err := newRuntime(opts...)
	if err != nil {
		return nil, err
	}

	svc := &emailService{
		runtime:      rt,
		client:       client,
		templatesDir: templatesDir,
		smtpConf:     smtpConf,
	}

	if svc.client == nil {
		return nil, smtp.ErrNoSMTPClient
	}

	return svc, nil
}
