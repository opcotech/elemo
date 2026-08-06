package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreateDocumentOpts creates repository.CreateDocumentOpts for tests.
func NewCreateDocumentOpts(belongsTo, createdBy model.ID) repository.CreateDocumentOpts {
	return repository.CreateDocumentOpts{
		BelongsTo: belongsTo,
		Name:      pkg.GenerateRandomString(10),
		Excerpt:   pkg.GenerateRandomString(10),
		FileID:    pkg.GenerateRandomString(10),
		CreatedBy: createdBy,
	}
}

// NewRepositoryDocument creates a repository.Document for mock returns.
func NewRepositoryDocument(createdBy model.ID) *repository.Document {
	return &repository.Document{
		ID:          model.MustNewID(model.ResourceTypeDocument),
		Name:        pkg.GenerateRandomString(10),
		Excerpt:     pkg.GenerateRandomString(10),
		FileID:      pkg.GenerateRandomString(10),
		CreatedBy:   createdBy,
		Labels:      make([]model.ID, 0),
		Comments:    make([]model.ID, 0),
		Attachments: make([]model.ID, 0),
		CreatedAt:   convert.ToPointer(time.Now().UTC()),
	}
}

// NewDocument creates a new document with random values. It does not create
// the db record.
//
// Deprecated: prefer NewCreateDocumentOpts / NewRepositoryDocument.
func NewDocument(createdBy model.ID) *model.Document {
	doc, err := model.NewDocument(
		pkg.GenerateRandomString(10),
		pkg.GenerateRandomString(10),
		createdBy,
	)
	if err != nil {
		panic(err)
	}

	doc.Excerpt = pkg.GenerateRandomString(10)

	return doc
}
