package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreateCommentOpts creates repository.CreateCommentOpts for tests.
func NewCreateCommentOpts(belongsTo, createdBy model.ID) repository.CreateCommentOpts {
	return repository.CreateCommentOpts{
		BelongsTo: belongsTo,
		Content:   pkg.GenerateRandomString(10),
		CreatedBy: createdBy,
	}
}

// NewRepositoryComment creates a repository.Comment for mock returns.
func NewRepositoryComment(createdBy model.ID) *repository.Comment {
	return &repository.Comment{
		ID:        model.MustNewID(model.ResourceTypeComment),
		Content:   pkg.GenerateRandomString(10),
		CreatedBy: createdBy,
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}
}

// NewComment creates a new comment with the given user ID and text. The
// comment is not created in the database.
//
// Deprecated: prefer NewCreateCommentOpts / NewRepositoryComment.
func NewComment(createdBy model.ID) *model.Comment {
	comment, err := model.NewComment(pkg.GenerateRandomString(10), createdBy)
	if err != nil {
		panic(err)
	}
	return comment
}
