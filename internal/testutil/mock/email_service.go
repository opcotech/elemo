package mock

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type EmailService struct {
	mock.Mock
}

func (e *EmailService) SendAuthPasswordResetEmail(ctx context.Context, resetPath string, user *model.User) error {
	args := e.Called(ctx, resetPath, user)
	return args.Error(0)
}

func (e *EmailService) SendOrganizationInvitationEmail(ctx context.Context, invitationPath string, organization *model.Organization, user *model.User) error {
	args := e.Called(ctx, invitationPath, organization, user)
	return args.Error(0)
}

func (e *EmailService) SendSystemLicenseExpiryEmail(ctx context.Context, licenseID, licenseEmail, licenseOrganization string, licenseExpiresAt time.Time) error {
	args := e.Called(ctx, licenseID, licenseEmail, licenseOrganization, licenseExpiresAt)
	return args.Error(0)
}

func (e *EmailService) SendUserWelcomeEmail(ctx context.Context, user *model.User) error {
	args := e.Called(ctx, user)
	return args.Error(0)
}
