//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/testutil"
)

func prepareOrganization(t *testing.T) *model.Organization {
	organization, err := model.NewOrganization(testutil.GenerateRandomString(10), testutil.GenerateEmail(10))
	require.NoError(t, err)

	organization.Logo = "https://www.gravatar.com/avatar"
	organization.Website = "https://example.com/"

	return organization
}

func TestOrganizationRepository_Create(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer cleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, owner.ID, organization)
	require.NoError(t, err)
}

func TestOrganizationRepository_Get(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer cleanupNeo4jStore(t, ctx, db)

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

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, owner.ID, organization)
	require.NoError(t, err)

	role, err := model.NewRole("admin")
	require.NoError(t, err)

	err = roleRepo.Create(ctx, owner.ID, organization.ID, role)
	require.NoError(t, err)

	got, err := orgRepo.Get(ctx, organization.ID)
	require.NoError(t, err)

	assert.Equal(t, organization.ID, got.ID)
	assert.Equal(t, organization.Name, got.Name)
	assert.Equal(t, organization.Email, got.Email)
	assert.Equal(t, organization.Logo, got.Logo)
	assert.Equal(t, organization.Website, got.Website)
	assert.Len(t, got.Members, 1)
	assert.Len(t, got.Namespaces, 0)
	assert.Len(t, got.Teams, 1)
	assert.Nil(t, got.UpdatedAt)
}

func TestOrganizationRepository_GetAll(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer cleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	err = orgRepo.Create(ctx, owner.ID, prepareOrganization(t))
	require.NoError(t, err)

	err = orgRepo.Create(ctx, owner.ID, prepareOrganization(t))
	require.NoError(t, err)

	err = orgRepo.Create(ctx, owner.ID, prepareOrganization(t))
	require.NoError(t, err)

	got, err := orgRepo.GetAll(ctx, 0, 10)
	require.NoError(t, err)
	assert.Len(t, got, 3)

	got, err = orgRepo.GetAll(ctx, 0, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = orgRepo.GetAll(ctx, 1, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = orgRepo.GetAll(ctx, 2, 2)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = orgRepo.GetAll(ctx, 3, 2)
	require.NoError(t, err)
	assert.Len(t, got, 0)
}

func TestOrganizationRepository_Update(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer cleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, owner.ID, organization)
	require.NoError(t, err)

	patch := map[string]any{
		"name":  "new name",
		"email": "newemail@example.com",
	}

	updated, err := orgRepo.Update(ctx, organization.ID, patch)
	require.NoError(t, err)

	assert.Equal(t, organization.ID, updated.ID)
	assert.Equal(t, patch["name"], updated.Name)
	assert.Equal(t, patch["email"], updated.Email)
	assert.Equal(t, organization.Logo, updated.Logo)
	assert.Equal(t, organization.Website, updated.Website)
	assert.Len(t, updated.Members, 1)
	assert.Len(t, updated.Namespaces, 0)
	assert.Len(t, updated.Teams, 0)
	assert.NotNil(t, updated.UpdatedAt)
}

func TestOrganizationRepository_AddMember(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer cleanupNeo4jStore(t, ctx, db)

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

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, owner.ID, organization)
	require.NoError(t, err)

	role, err := model.NewRole("admin")
	require.NoError(t, err)

	err = roleRepo.Create(ctx, owner.ID, organization.ID, role)
	require.NoError(t, err)

	got, err := orgRepo.Get(ctx, organization.ID)
	require.NoError(t, err)
	assert.Len(t, got.Members, 1)

	member := prepareUser(t)
	err = userRepo.Create(ctx, member)
	require.NoError(t, err)

	err = orgRepo.AddMember(ctx, organization.ID, member.ID)
	require.NoError(t, err)

	got, err = orgRepo.Get(ctx, organization.ID)
	require.NoError(t, err)
	assert.Len(t, got.Members, 2)
}

func TestOrganizationRepository_RemoveMember(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer cleanupNeo4jStore(t, ctx, db)

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

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, owner.ID, organization)
	require.NoError(t, err)

	role, err := model.NewRole("admin")
	require.NoError(t, err)

	err = roleRepo.Create(ctx, owner.ID, organization.ID, role)
	require.NoError(t, err)

	got, err := orgRepo.Get(ctx, organization.ID)
	require.NoError(t, err)
	assert.Len(t, got.Members, 1)

	err = orgRepo.RemoveMember(ctx, organization.ID, owner.ID)
	require.NoError(t, err)

	got, err = orgRepo.Get(ctx, organization.ID)
	require.NoError(t, err)
	assert.Len(t, got.Members, 0)
}

func TestOrganizationRepository_Delete(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer cleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, owner.ID, organization)
	require.NoError(t, err)

	err = orgRepo.Delete(ctx, organization.ID)
	require.NoError(t, err)

	_, err = orgRepo.Get(ctx, organization.ID)
	require.Error(t, err)
}
