package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// UserTokenRepository is a repository for managing user tokens.
//
//go:generate mockgen -source=auth.go -destination=../testutil/mock/user_token_repo_gen.go -package=mock -mock_names "UserTokenRepository=UserTokenRepository"
type UserTokenRepository interface {
	Create(ctx context.Context, token *model.UserToken) error
	Get(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) (*model.UserToken, error)
	Delete(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) error
}
