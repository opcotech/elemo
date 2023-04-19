//go:build integration

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/password"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testService "github.com/opcotech/elemo/internal/testutil/service"
)

func TestUserService_CreateGabor(t *testing.T) {
	ctx := context.Background()

	owner := testService.NewResourceOwner(t, neo4jDBConf)
	s := testService.NewUserService(t, neo4jDBConf)

	user := testModel.NewUser()
	user.FirstName = "Gábor"
	user.LastName = "Boros"
	user.Email = "gabor@elemo.app"
	user.Password = password.HashPassword("AppleTree123")
	user.Picture = "https://github.com/gabor-boros.png"

	err := s.Create(context.WithValue(ctx, pkg.CtxKeyUserID, owner.ID), user)
	require.NoError(t, err)

	assert.NotNil(t, user.ID)
	assert.NotNil(t, user.CreatedAt)
	assert.Nil(t, user.UpdatedAt)
}
