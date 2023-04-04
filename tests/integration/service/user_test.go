//go:build integration

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/password"
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

	user := testModel.NewUser()
	err := s.Create(context.Background(), user)
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

	user := testModel.NewUser()
	err := s.Create(context.Background(), user)
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

	user := testModel.NewUser()
	err := s.Create(context.Background(), user)
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

	s := testService.NewUserService(t, neo4jDBConf)

	user1 := testModel.NewUser()
	err := s.Create(context.Background(), user1)
	require.NoError(t, err)

	user2 := testModel.NewUser()
	err = s.Create(context.Background(), user2)
	require.NoError(t, err)

	user3 := testModel.NewUser()
	err = s.Create(context.Background(), user3)
	require.NoError(t, err)

	got, err := s.GetAll(ctx, 0, 10)
	require.NoError(t, err)
	assert.Len(t, got, 3)

	got, err = s.GetAll(ctx, 0, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = s.GetAll(ctx, 1, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = s.GetAll(ctx, 2, 2)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = s.GetAll(ctx, 3, 2)
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

	user := testModel.NewUser()
	err := s.Create(context.Background(), user)
	require.NoError(t, err)

	patch := map[string]any{
		"email": "email@example.com",
	}

	updated, err := s.Update(ctx, user.ID, patch)
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

	user := testModel.NewUser()
	err := s.Create(context.Background(), user)
	require.NoError(t, err)

	// Soft delete
	err = s.Delete(ctx, user.ID, false)
	require.NoError(t, err)

	got, err := s.Get(ctx, user.ID)
	require.NoError(t, err)

	assert.Equal(t, user.ID, got.ID)
	assert.Equal(t, user.Email, got.Email)
	assert.Equal(t, model.UserStatusDeleted, got.Status)
	assert.Equal(t, password.UnusablePassword, got.Password)
	assert.NotNil(t, got.UpdatedAt)

	// Hard delete
	err = s.Delete(ctx, user.ID, true)
	require.NoError(t, err)

	_, err = s.Get(ctx, user.ID)
	assert.Error(t, err)
}
