package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreateNotificationOpts creates repository.CreateNotificationOpts for tests.
func NewCreateNotificationOpts(recipient model.ID) repository.CreateNotificationOpts {
	return repository.CreateNotificationOpts{
		Title:       pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(10),
		Recipient:   recipient,
	}
}

// NewRepositoryNotification creates a repository.Notification for mock returns.
func NewRepositoryNotification(recipient model.ID) *repository.Notification {
	opts := NewCreateNotificationOpts(recipient)
	return &repository.Notification{
		ID:          model.MustNewID(model.ResourceTypeNotification),
		Title:       opts.Title,
		Description: opts.Description,
		Recipient:   opts.Recipient,
		Read:        false,
		CreatedAt:   convert.ToPointer(time.Now().UTC()),
	}
}
