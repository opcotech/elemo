package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreateAttachmentOpts creates repository.CreateAttachmentOpts for tests.
func NewCreateAttachmentOpts(belongsTo, createdBy model.ID) repository.CreateAttachmentOpts {
	return repository.CreateAttachmentOpts{
		BelongsTo: belongsTo,
		Name:      pkg.GenerateRandomString(10),
		FileID:    pkg.GenerateRandomString(10),
		CreatedBy: createdBy,
	}
}

// NewRepositoryAttachment creates a repository.Attachment for mock returns.
func NewRepositoryAttachment(createdBy model.ID) *repository.Attachment {
	return &repository.Attachment{
		ID:        model.MustNewID(model.ResourceTypeAttachment),
		Name:      pkg.GenerateRandomString(10),
		FileID:    pkg.GenerateRandomString(10),
		CreatedBy: createdBy,
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}
}

// NewAttachment creates a new attachment with random values. It does not
// create the db record.
//
// Deprecated: prefer NewCreateAttachmentOpts / NewRepositoryAttachment.
func NewAttachment(createdBy model.ID) *model.Attachment {
	attachment, err := model.NewAttachment(
		pkg.GenerateRandomString(10),
		pkg.GenerateRandomString(10),
		createdBy,
	)
	if err != nil {
		panic(err)
	}
	return attachment
}
