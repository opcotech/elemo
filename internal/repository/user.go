package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// UserRepository is a repository for managing users.
//
//go:generate mockgen -source=user.go -destination=../testutil/mock/user_repo_gen.go -package=mock -mock_names "UserRepository=UserRepository"
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Get(ctx context.Context, id model.ID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetAll(ctx context.Context, offset, limit int) ([]*model.User, error)
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.User, error)
	Delete(ctx context.Context, id model.ID) error
}
