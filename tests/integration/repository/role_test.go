//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

func TestRoleRepository_Create(t *testing.T) {
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

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	roleRepo, err := neo4j.NewRoleRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	role, err := model.NewRole("test")
	require.NoError(t, err)

	role.Description = "Test role"

	err = roleRepo.Create(ctx, user.ID, organization.ID, role)
	require.NoError(t, err)
}

func TestRoleRepository_Get(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	permRepo, err := neo4j.NewPermissionRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	roleRepo, err := neo4j.NewRoleRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

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

	role, err := model.NewRole("test")
	require.NoError(t, err)
	role.Description = "Test role"
	require.NoError(t, roleRepo.Create(ctx, user.ID, organization.ID, role))

	perm, err := model.NewPermission(role.ID, organization.ID, model.PermissionKindWrite)
	require.NoError(t, err)
	require.NoError(t, permRepo.Create(ctx, perm))

	role, err = roleRepo.Get(ctx, role.ID)
	require.NoError(t, err)

	require.NotNil(t, role)
	require.Equal(t, "test", role.Name)
	require.Equal(t, "Test role", role.Description)
	require.Len(t, role.Members, 1)
	require.Equal(t, user.ID, role.Members[0])
	require.Len(t, role.Permissions, 1)
	require.Equal(t, perm.ID, role.Permissions[0])
}

func TestRoleRepository_GetAll(t *testing.T) {
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

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	roleRepo, err := neo4j.NewRoleRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user1role1 := prepareUser(t)
	require.NoError(t, userRepo.Create(ctx, user1role1))

	user1role2 := prepareUser(t)
	require.NoError(t, userRepo.Create(ctx, user1role2))

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user1role1.ID, organization)
	require.NoError(t, err)

	role, err := model.NewRole("test")
	require.NoError(t, err)
	require.NoError(t, roleRepo.Create(ctx, user1role1.ID, organization.ID, role))

	role2, err := model.NewRole("test 2")
	require.NoError(t, err)
	require.NoError(t, roleRepo.Create(ctx, user1role2.ID, organization.ID, role2))

	roles, err := roleRepo.GetAllBelongsTo(ctx, organization.ID, 0, 10)
	require.NoError(t, err)
	require.Len(t, roles, 2)

	roles, err = roleRepo.GetAllBelongsTo(ctx, organization.ID, 0, 1)
	require.NoError(t, err)
	require.Len(t, roles, 1)

	roles, err = roleRepo.GetAllBelongsTo(ctx, organization.ID, 1, 1)
	require.NoError(t, err)
	require.Len(t, roles, 1)

	roles, err = roleRepo.GetAllBelongsTo(ctx, organization.ID, 2, 1)
	require.NoError(t, err)
	require.Len(t, roles, 0)

	roles, err = roleRepo.GetAllBelongsTo(ctx, organization.ID, 0, 0)
	require.NoError(t, err)
	require.Len(t, roles, 0)
}

func TestRoleRepository_Update(t *testing.T) {
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

	roleRepo, err := neo4j.NewRoleRepository(
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

	role, err := model.NewRole("test")
	require.NoError(t, err)
	role.Description = "Test role"
	require.NoError(t, roleRepo.Create(ctx, user.ID, organization.ID, role))

	role, err = roleRepo.Get(ctx, role.ID)
	require.NoError(t, err)

	newDescription := "Updated description"
	patch := map[string]any{
		"description": newDescription,
	}

	updatedRole, err := roleRepo.Update(ctx, role.ID, patch)
	require.NoError(t, err)

	assert.Equal(t, role.ID, updatedRole.ID)
	assert.Equal(t, role.Name, updatedRole.Name)
	assert.Equal(t, newDescription, updatedRole.Description)
	assert.Len(t, role.Members, 1)
	assert.Equal(t, user.ID, role.Members[0])
}

func TestRoleRepository_AddMember(t *testing.T) {
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

	roleRepo, err := neo4j.NewRoleRepository(
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

	additionalUser := prepareUser(t)
	require.NoError(t, userRepo.Create(ctx, additionalUser))

	role, err := model.NewRole("test")
	require.NoError(t, err)
	require.NoError(t, roleRepo.Create(ctx, user.ID, organization.ID, role))

	role, err = roleRepo.Get(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, role.Members, 1)

	require.NoError(t, roleRepo.AddMember(ctx, role.ID, additionalUser.ID))

	role, err = roleRepo.Get(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, role.Members, 2)

	assert.Contains(t, role.Members, user.ID)
	assert.Contains(t, role.Members, additionalUser.ID)
}

func TestRoleRepository_RemoveMember(t *testing.T) {
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

	roleRepo, err := neo4j.NewRoleRepository(
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

	additionalUser := prepareUser(t)
	require.NoError(t, userRepo.Create(ctx, additionalUser))

	role, err := model.NewRole("test")
	require.NoError(t, err)
	require.NoError(t, roleRepo.Create(ctx, user.ID, organization.ID, role))

	role, err = roleRepo.Get(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, role.Members, 1)

	require.NoError(t, roleRepo.AddMember(ctx, role.ID, additionalUser.ID))

	role, err = roleRepo.Get(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, role.Members, 2)

	require.NoError(t, roleRepo.RemoveMember(ctx, role.ID, additionalUser.ID))

	role, err = roleRepo.Get(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, role.Members, 1)

	assert.Contains(t, role.Members, user.ID)
	assert.NotContains(t, role.Members, additionalUser.ID)
}

func TestRoleRepository_Delete(t *testing.T) {
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

	roleRepo, err := neo4j.NewRoleRepository(
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

	role, err := model.NewRole("test")
	require.NoError(t, err)
	role.Description = "Test role"
	require.NoError(t, roleRepo.Create(ctx, user.ID, organization.ID, role))

	err = roleRepo.Delete(ctx, role.ID)
	require.NoError(t, err)

	_, err = roleRepo.Get(ctx, role.ID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}
