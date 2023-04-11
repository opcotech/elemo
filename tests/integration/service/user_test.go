//go:build integration

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
	testService "github.com/opcotech/elemo/internal/testutil/service"
)

func TestUserService_Create(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)
	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	s := testService.NewUserService(t, neo4jDBConf)
	owner := testService.NewResourceOwner(t, neo4jDBConf)

	user := testModel.NewUser()
	err := s.Create(context.WithValue(ctx, pkg.CtxKeyUserID, owner.ID), user)
	require.NoError(t, err)

	assert.NotNil(t, user.ID)
	assert.NotNil(t, user.CreatedAt)
	assert.Nil(t, user.UpdatedAt)
}

func TestUserService_Get(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)
	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	s := testService.NewUserService(t, neo4jDBConf)
	owner := testService.NewResourceOwner(t, neo4jDBConf)

	user := testModel.NewUser()
	err := s.Create(context.WithValue(ctx, pkg.CtxKeyUserID, owner.ID), user)
	require.NoError(t, err)

	got, err := s.Get(ctx, user.ID)
	require.NoError(t, err)

	assert.Equal(t, user.ID, got.ID)
}

func TestUserService_GetByEmail(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)
	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	s := testService.NewUserService(t, neo4jDBConf)
	owner := testService.NewResourceOwner(t, neo4jDBConf)

	user := testModel.NewUser()
	err := s.Create(context.WithValue(ctx, pkg.CtxKeyUserID, owner.ID), user)
	require.NoError(t, err)

	got, err := s.GetByEmail(ctx, user.Email)
	require.NoError(t, err)

	assert.Equal(t, user.ID, got.ID)
}

func TestUserService_GetAll(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)
	defer testRepo.CleanupNeo4jStore(t, ctx, db)
	testRepo.CleanupNeo4jStore(t, ctx, db)

	s := testService.NewUserService(t, neo4jDBConf)
	owner := testService.NewResourceOwner(t, neo4jDBConf)

	user := testModel.NewUser()
	err := s.Create(context.WithValue(ctx, pkg.CtxKeyUserID, owner.ID), user)
	require.NoError(t, err)

	err = s.Create(context.WithValue(ctx, pkg.CtxKeyUserID, owner.ID), testModel.NewUser())
	require.NoError(t, err)

	err = s.Create(context.WithValue(ctx, pkg.CtxKeyUserID, owner.ID), testModel.NewUser())
	require.NoError(t, err)

	got, err := s.GetAll(ctx, 0, 10)
	require.NoError(t, err)
	assert.Len(t, got, 3+1) // +1 for the owner

	got, err = s.GetAll(ctx, 0, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = s.GetAll(ctx, 1, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = s.GetAll(ctx, 2, 2)
	require.NoError(t, err)
	assert.Len(t, got, 1+1) // +1 for the owner

	got, err = s.GetAll(ctx, 3+1, 2) // +1 for the owner
	require.NoError(t, err)
	assert.Len(t, got, 0)
}

func TestUserService_Update(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)
	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	s := testService.NewUserService(t, neo4jDBConf)
	owner := testService.NewResourceOwner(t, neo4jDBConf)

	user := testModel.NewUser()
	err := s.Create(context.WithValue(ctx, pkg.CtxKeyUserID, owner.ID), user)
	require.NoError(t, err)

	patch := map[string]any{
		"email": "email@example.com",
	}

	updated, err := s.Update(context.WithValue(ctx, pkg.CtxKeyUserID, user.ID), user.ID, patch)
	require.NoError(t, err)

	assert.Equal(t, user.ID, updated.ID)
	assert.Equal(t, patch["email"], updated.Email)
}

func TestUserService_Delete(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)
	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	s := testService.NewUserService(t, neo4jDBConf)
	owner := testService.NewResourceOwner(t, neo4jDBConf)

	user := testModel.NewUser()
	err := s.Create(context.WithValue(ctx, pkg.CtxKeyUserID, owner.ID), user)
	require.NoError(t, err)

	target := testModel.NewUser()
	require.NoError(t, s.Create(context.WithValue(ctx, pkg.CtxKeyUserID, owner.ID), target))

	permRepo, err := neo4j.NewPermissionRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	permission, err := model.NewPermission(user.ID, target.ID, model.PermissionKindDelete)
	require.NoError(t, err)

	err = permRepo.Create(ctx, permission)
	require.NoError(t, err)

	// Soft delete
	err = s.Delete(context.WithValue(ctx, pkg.CtxKeyUserID, user.ID), target.ID, false)
	require.NoError(t, err)

	got, err := s.Get(ctx, target.ID)
	require.NoError(t, err)

	assert.Equal(t, target.ID, got.ID)
	assert.Equal(t, target.Email, got.Email)
	assert.Equal(t, model.UserStatusDeleted, got.Status)
	assert.Equal(t, password.UnusablePassword, got.Password)
	assert.NotNil(t, got.UpdatedAt)

	// Hard delete
	err = s.Delete(context.WithValue(ctx, pkg.CtxKeyUserID, user.ID), target.ID, true)
	require.NoError(t, err)

	_, err = s.Get(ctx, target.ID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}
