package main

import (
	"context"
	"time"

	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
)

// discardEmailService satisfies service.EmailService without sending mail.
type discardEmailService struct{}

func (discardEmailService) SendAuthPasswordResetEmail(
	_ context.Context,
	_ email.Recipient,
	_ string,
) error {
	return nil
}

func (discardEmailService) SendOrganizationInvitationEmail(
	_ context.Context,
	_ model.ID,
	_ string,
	_ email.Recipient,
	_ string,
) error {
	return nil
}

func (discardEmailService) SendSystemLicenseExpiryEmail(
	_ context.Context,
	_, _, _ string,
	_ time.Time,
) error {
	return nil
}

func (discardEmailService) SendUserWelcomeEmail(_ context.Context, _ email.Recipient) error {
	return nil
}

// discardNotificationService satisfies service.NotificationService without writes.
type discardNotificationService struct{}

func (discardNotificationService) Create(_ context.Context, opts service.CreateNotificationOpts) (*service.Notification, error) {
	return &service.Notification{
		Title:       opts.Title,
		Description: opts.Description,
		Recipient:   opts.Recipient,
	}, nil
}

func (discardNotificationService) Get(_ context.Context, _, _ model.ID) (*service.Notification, error) {
	return nil, nil
}

func (discardNotificationService) ListByRecipient(_ context.Context, _ model.ID, _ service.CursorPage) (service.Page[*service.Notification], error) {
	return service.Page[*service.Notification]{}, nil
}

func (discardNotificationService) Update(_ context.Context, _, _ model.ID, _ service.UpdateNotificationOpts) (*service.Notification, error) {
	return nil, nil
}

func (discardNotificationService) Delete(_ context.Context, _, _ model.ID) error {
	return nil
}

// noopSearchService drops write-through indexing during seed. Reindex is
// called on a real SearchService after the graph is populated.
type noopSearchService struct{}

var _ service.SearchService = noopSearchService{}

func (noopSearchService) Index(_ context.Context, _ service.IndexInput) error {
	return nil
}

func (noopSearchService) EnqueueIndex(_ context.Context, _ model.ID) error {
	return nil
}

func (noopSearchService) IndexIDs(_ context.Context, _ *repository.Neo4jDatabase, _ ...model.ID) error {
	return nil
}

func (noopSearchService) Delete(_ context.Context, _ model.ID) error {
	return nil
}

func (noopSearchService) DeleteByScope(_ context.Context, _ model.ID) error {
	return nil
}

func (noopSearchService) DeleteAll(_ context.Context) error {
	return nil
}

func (noopSearchService) Search(_ context.Context, _ service.SearchQuery) (service.Page[*service.SearchResult], error) {
	return service.Page[*service.SearchResult]{}, nil
}

func (noopSearchService) Reindex(_ context.Context, _ service.SearchReindexSources, _ service.SearchReindexOptions) error {
	return nil
}
