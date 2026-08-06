package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/auth"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
)

// NewCreateUserTokenOpts creates repository.CreateUserTokenOpts for tests.
func NewCreateUserTokenOpts(userID model.ID) (string, repository.CreateUserTokenOpts) {
	encoded, token, err := auth.GenerateToken(model.UserTokenContextConfirm.String(), map[string]any{
		"user_id": userID,
	})
	if err != nil {
		panic(err)
	}

	return encoded, repository.CreateUserTokenOpts{
		UserID:  userID,
		SentTo:  testutil.GenerateEmail(10),
		Token:   token,
		Context: model.UserTokenContextConfirm,
	}
}

// NewRepositoryUserToken creates a repository.UserToken for mock returns.
func NewRepositoryUserToken(userID model.ID) (string, *repository.UserToken) {
	encoded, opts := NewCreateUserTokenOpts(userID)
	return encoded, &repository.UserToken{
		ID:        model.MustNewID(model.ResourceTypeUserToken),
		UserID:    opts.UserID,
		SentTo:    opts.SentTo,
		Token:     opts.Token,
		Context:   opts.Context,
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}
}

// NewUserToken creates a user token. It does not create the db record.
// Prefer NewCreateUserTokenOpts / NewRepositoryUserToken for new tests.
func NewUserToken(userID model.ID) (string, *repository.UserToken) {
	return NewRepositoryUserToken(userID)
}
