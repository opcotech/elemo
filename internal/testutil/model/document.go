package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreateDocumentOpts creates repository.CreateDocumentOpts for tests.
func NewCreateDocumentOpts(library, createdBy model.ID) repository.CreateDocumentOpts {
	return repository.CreateDocumentOpts{
		Library:   library,
		Title:     pkg.GenerateRandomString(10),
		Excerpt:   pkg.GenerateRandomString(10),
		FileID:    pkg.GenerateRandomString(10),
		CreatedBy: createdBy,
	}
}

// NewRepositoryDocument creates a repository.Document for mock returns.
func NewRepositoryDocument(createdBy model.ID) *repository.Document {
	libraryID := model.MustNewID(model.ResourceTypeNamespace)
	return &repository.Document{
		ID:        model.MustNewID(model.ResourceTypeDocument),
		Title:     pkg.GenerateRandomString(10),
		Excerpt:   pkg.GenerateRandomString(10),
		FileID:    pkg.GenerateRandomString(10),
		CreatedBy: repository.PartialUser{ID: createdBy},
		Library: repository.DocumentLibrary{
			ID:   libraryID,
			Type: model.ResourceTypeNamespace,
			Name: "Library",
		},
		Relations:       make([]repository.DocumentRelation, 0),
		Labels:          make([]repository.PartialLabel, 0),
		CommentCount:    convert.ToPointer(int64(0)),
		AttachmentCount: convert.ToPointer(int64(0)),
		CreatedAt:       convert.ToPointer(time.Now().UTC()),
	}
}

// NewCreateFolderOpts creates repository.CreateFolderOpts for tests.
func NewCreateFolderOpts(library, createdBy model.ID) repository.CreateFolderOpts {
	return repository.CreateFolderOpts{
		Library:   library,
		Name:      pkg.GenerateRandomString(10),
		CreatedBy: createdBy,
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
