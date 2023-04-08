package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	CommentIDType = "Comment"
)

// Comment represents a comment on a resource.
type Comment struct {
	ID        ID         `json:"id" validate:"required,dive"`
	Content   string     `json:"content" validate:"required,min=5,max=2000"`
	CreatedBy ID         `json:"created_by" validate:"required,dive"`
	CreatedAt *time.Time `json:"created_at" validate:"omitempty"`
	UpdatedAt *time.Time `json:"updated_at" validate:"omitempty"`
}

func (c *Comment) Validate() error {
	if err := validate.Struct(c); err != nil {
		return errors.Join(ErrInvalidCommentDetails, err)
	}
	if err := c.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidCommentDetails, err)
	}
	if err := c.CreatedBy.Validate(); err != nil {
		return errors.Join(ErrInvalidCommentDetails, err)
	}
	return nil
}

// NewComment creates a new Comment.
func NewComment(content string, createdBy ID) (*Comment, error) {
	comment := &Comment{
		ID:        MustNewNilID(CommentIDType),
		Content:   content,
		CreatedBy: createdBy,
	}

	if err := comment.Validate(); err != nil {
		return nil, err
	}

	return comment, nil
}
