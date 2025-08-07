package model

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/auth"
	"github.com/opcotech/elemo/internal/testutil"
)

// NewUserToken creates a new user token. It does not create the db record.
func NewUserToken(userID model.ID) (string, *model.UserToken) {
	encoded, token, err := auth.GenerateToken(model.UserTokenContextConfirm.String(), map[string]any{
		"user_id": userID,
	})
	if err != nil {
		panic(err)
	}

	userToken, err := model.NewUserToken(userID, testutil.GenerateEmail(10), token, model.UserTokenContextConfirm)
	if err != nil {
		panic(err)
	}

	return encoded, userToken
}
