//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

func TestTodoRepository_Create(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	todoRepo, err := neo4j.NewTodoRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	creator := prepareUser(t)
	err = userRepo.Create(ctx, creator)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	todo, err := model.NewTodo("test", owner.ID, creator.ID)
	require.NoError(t, err)

	todo.Description = "test description"
	todo.DueDate = convert.ToPointer(time.Now().Add(24 * time.Hour))

	err = todoRepo.Create(ctx, todo)
	require.NoError(t, err)
}

func TestTodoRepository_Get(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	todoRepo, err := neo4j.NewTodoRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	creator := prepareUser(t)
	err = userRepo.Create(ctx, creator)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	todo, err := model.NewTodo("test", owner.ID, creator.ID)
	require.NoError(t, err)

	todo.Description = "test description"
	todo.DueDate = convert.ToPointer(time.Now().Add(24 * time.Hour))

	err = todoRepo.Create(ctx, todo)
	require.NoError(t, err)

	got, err := todoRepo.Get(ctx, todo.ID)
	require.NoError(t, err)

	assert.Equal(t, todo.ID, got.ID)
	assert.Equal(t, todo.Title, got.Title)
	assert.Equal(t, todo.Description, got.Description)
	assert.Equal(t, todo.Completed, got.Completed)
	assert.Equal(t, todo.OwnedBy, got.OwnedBy)
	assert.Equal(t, todo.CreatedBy, got.CreatedBy)
	assert.WithinDuration(t, *todo.DueDate, *got.DueDate, 1*time.Second)
	assert.WithinDuration(t, *todo.CreatedAt, *got.CreatedAt, 1*time.Second)
	assert.Nil(t, got.UpdatedAt)
}

func TestTodoRepository_GetByOwner(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)
	testRepo.CleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	todoRepo, err := neo4j.NewTodoRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	creator := prepareUser(t)
	err = userRepo.Create(ctx, creator)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		todo, err := model.NewTodo("test", owner.ID, creator.ID)
		require.NoError(t, err)

		todo.Description = "test description"
		todo.Completed = i%2 == 0
		todo.DueDate = convert.ToPointer(time.Now().Add(24 * time.Hour))

		err = todoRepo.Create(ctx, todo)
		require.NoError(t, err)
	}

	got, err := todoRepo.GetByOwner(ctx, owner.ID, nil)
	require.NoError(t, err)
	assert.Len(t, got, 10)

	got, err = todoRepo.GetByOwner(ctx, owner.ID, convert.ToPointer(false))
	require.NoError(t, err)
	assert.Len(t, got, 5)

	got, err = todoRepo.GetByOwner(ctx, owner.ID, convert.ToPointer(true))
	require.NoError(t, err)
	assert.Len(t, got, 5)
}

func TestTodoRepository_Update(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	todoRepo, err := neo4j.NewTodoRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	creator := prepareUser(t)
	err = userRepo.Create(ctx, creator)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	todo, err := model.NewTodo("test", owner.ID, creator.ID)
	require.NoError(t, err)

	todo.Description = "test description"
	todo.DueDate = convert.ToPointer(time.Now().Add(24 * time.Hour))

	err = todoRepo.Create(ctx, todo)
	require.NoError(t, err)

	patch := map[string]any{
		"title":       "updated title",
		"description": "updated description",
		"completed":   true,
	}

	updated, err := todoRepo.Update(ctx, todo.ID, patch)
	require.NoError(t, err)

	assert.Equal(t, todo.ID, updated.ID)
	assert.Equal(t, patch["title"], updated.Title)
	assert.Equal(t, patch["description"], updated.Description)
	assert.Equal(t, patch["completed"], updated.Completed)
	assert.NotNil(t, updated.UpdatedAt)
}

func TestTodoRepository_Delete(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	todoRepo, err := neo4j.NewTodoRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	creator := prepareUser(t)
	err = userRepo.Create(ctx, creator)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	todo, err := model.NewTodo("test", owner.ID, creator.ID)
	require.NoError(t, err)

	todo.Description = "test description"
	todo.DueDate = convert.ToPointer(time.Now().Add(24 * time.Hour))

	err = todoRepo.Create(ctx, todo)
	require.NoError(t, err)

	_, err = todoRepo.Get(ctx, todo.ID)
	require.NoError(t, err)

	err = todoRepo.Delete(ctx, todo.ID)
	require.NoError(t, err)

	_, err = todoRepo.Get(ctx, todo.ID)
	require.Error(t, err)
}
