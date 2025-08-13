package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// AssignmentRepository is a repository for managing resource assignments.
//
//go:generate mockgen -source=assignment.go -destination=../testutil/mock/assignment_repo_gen.go -package=mock -mock_names "AssignmentRepository=AssignmentRepository"
type AssignmentRepository interface {
	Create(ctx context.Context, assignment *model.Assignment) error
	Get(ctx context.Context, id model.ID) (*model.Assignment, error)
	GetByUser(ctx context.Context, userID model.ID, offset, limit int) ([]*model.Assignment, error)
	GetByResource(ctx context.Context, resourceID model.ID, offset, limit int) ([]*model.Assignment, error)
	Delete(ctx context.Context, id model.ID) error
}
