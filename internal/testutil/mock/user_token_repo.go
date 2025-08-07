package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type UserTokenRepository struct {
	mock.Mock
}

// Create mocks the Create method.
func (r *UserTokenRepository) Create(ctx context.Context, token *model.UserToken) error {
	args := r.Called(ctx, token)
	return args.Error(0)
}

// Get mocks the Get method.
func (r *UserTokenRepository) Get(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) (*model.UserToken, error) {
	args := r.Called(ctx, userID, tokenCtx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserToken), args.Error(1)
}

// Delete mocks the Delete method.
func (r *UserTokenRepository) Delete(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) error {
	args := r.Called(ctx, userID, tokenCtx)
	return args.Error(0)
}
