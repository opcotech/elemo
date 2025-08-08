package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// UserTokenRepository is a repository for managing user tokens.
type UserTokenRepository interface {
	Create(ctx context.Context, token *model.UserToken) error
	Get(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) (*model.UserToken, error)
	Delete(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) error
}
