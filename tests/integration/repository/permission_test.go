//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/testutil"
)

func TestPermissionRepository_Create(t *testing.T) {
	ctx := context.Background()

	db, closer := testutil.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testutil.CleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	require.NoError(t, userRepo.Create(ctx, user))

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	repo, err := neo4j.NewPermissionRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	permission, err := model.NewPermission(user.ID, organization.ID, model.PermissionKindRead)
	require.NoError(t, err)

	// Create a new permission
	err = repo.Create(ctx, permission)
	require.NoError(t, err)
}

func TestPermissionRepository_Get(t *testing.T) {
	ctx := context.Background()

	db, closer := testutil.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testutil.CleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	require.NoError(t, userRepo.Create(ctx, user))

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	repo, err := neo4j.NewPermissionRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	permission, err := model.NewPermission(user.ID, organization.ID, model.PermissionKindRead)
	require.NoError(t, err)

	// Create a new permission
	err = repo.Create(ctx, permission)
	require.NoError(t, err)

	// Get the permission
	got, err := repo.Get(ctx, permission.ID)
	require.NoError(t, err)

	assert.Equal(t, permission.ID, got.ID)
	assert.Equal(t, permission.Subject, got.Subject)
	assert.Equal(t, permission.Target, got.Target)
	assert.Equal(t, permission.Kind, got.Kind)
	assert.WithinDuration(t, *permission.CreatedAt, *got.CreatedAt, 0)
	assert.Nil(t, got.UpdatedAt)
}

func TestPermissionRepository_GetBySubject(t *testing.T) {
	ctx := context.Background()

	db, closer := testutil.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testutil.CleanupNeo4jStore(t, ctx, db)
	testutil.CleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	require.NoError(t, userRepo.Create(ctx, user))

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	repo, err := neo4j.NewPermissionRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	readPerm, err := model.NewPermission(user.ID, organization.ID, model.PermissionKindRead)
	require.NoError(t, err)

	writePerm, err := model.NewPermission(user.ID, organization.ID, model.PermissionKindWrite)
	require.NoError(t, err)

	// Create a new permission
	err = repo.Create(ctx, readPerm)
	require.NoError(t, err)

	err = repo.Create(ctx, writePerm)
	require.NoError(t, err)

	// Get permissions
	got, err := repo.GetBySubject(ctx, user.ID)
	require.NoError(t, err)

	assert.Len(t, got, 2)
	for _, perm := range got {
		assert.Equal(t, user.ID, perm.Subject)
		assert.Equal(t, organization.ID, perm.Target)
		assert.WithinDuration(t, *readPerm.CreatedAt, *perm.CreatedAt, 1*time.Second)
		assert.Nil(t, perm.UpdatedAt)
	}
}

func TestPermissionRepository_GetByTarget(t *testing.T) {
	ctx := context.Background()

	db, closer := testutil.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testutil.CleanupNeo4jStore(t, ctx, db)
	testutil.CleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	require.NoError(t, userRepo.Create(ctx, user))

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	repo, err := neo4j.NewPermissionRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	readPerm, err := model.NewPermission(user.ID, organization.ID, model.PermissionKindRead)
	require.NoError(t, err)

	writePerm, err := model.NewPermission(user.ID, organization.ID, model.PermissionKindWrite)
	require.NoError(t, err)

	// Create a new permission
	err = repo.Create(ctx, readPerm)
	require.NoError(t, err)

	err = repo.Create(ctx, writePerm)
	require.NoError(t, err)

	// Get permissions
	got, err := repo.GetByTarget(ctx, organization.ID)
	require.NoError(t, err)

	assert.Len(t, got, 2)
	for _, perm := range got {
		assert.Equal(t, user.ID, perm.Subject)
		assert.Equal(t, organization.ID, perm.Target)
		assert.WithinDuration(t, *readPerm.CreatedAt, *perm.CreatedAt, 1*time.Second)
		assert.Nil(t, perm.UpdatedAt)
	}
}

func TestPermissionRepository_Update(t *testing.T) {
	ctx := context.Background()

	db, closer := testutil.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testutil.CleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	require.NoError(t, userRepo.Create(ctx, user))

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	repo, err := neo4j.NewPermissionRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	permission, err := model.NewPermission(user.ID, organization.ID, model.PermissionKindRead)
	require.NoError(t, err)

	// Create a new permission
	err = repo.Create(ctx, permission)
	require.NoError(t, err)

	// Get the permission
	got, err := repo.Get(ctx, permission.ID)
	require.NoError(t, err)

	assert.Equal(t, permission.ID, got.ID)
	assert.Equal(t, permission.Subject, got.Subject)
	assert.Equal(t, permission.Target, got.Target)
	assert.Equal(t, permission.Kind, got.Kind)
	assert.WithinDuration(t, *permission.CreatedAt, *got.CreatedAt, 0)
	assert.Nil(t, got.UpdatedAt)

	// Update the permission
	got, err = repo.Update(ctx, permission.ID, model.PermissionKindWrite)
	require.NoError(t, err)

	assert.Equal(t, permission.ID, got.ID)
	assert.Equal(t, permission.Subject, got.Subject)
	assert.Equal(t, permission.Target, got.Target)
	assert.Equal(t, model.PermissionKindWrite, got.Kind)
	assert.WithinDuration(t, *permission.CreatedAt, *got.CreatedAt, 0)
	assert.NotNil(t, got.UpdatedAt)
}

func TestPermissionRepository_Delete(t *testing.T) {
	ctx := context.Background()

	db, closer := testutil.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testutil.CleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	require.NoError(t, userRepo.Create(ctx, user))

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	repo, err := neo4j.NewPermissionRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	permission, err := model.NewPermission(user.ID, organization.ID, model.PermissionKindRead)
	require.NoError(t, err)

	// Create a new permission
	err = repo.Create(ctx, permission)
	require.NoError(t, err)

	// Get the permission
	got, err := repo.Get(ctx, permission.ID)
	require.NoError(t, err)
	assert.Equal(t, permission.ID, got.ID)

	// Delete the permission
	err = repo.Delete(ctx, permission.ID)
	require.NoError(t, err)

	// Get the permission again
	_, err = repo.Get(ctx, permission.ID)
	require.Error(t, err)
}
