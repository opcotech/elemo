//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/testutil"
)

func TestNamespaceRepository_Create(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, owner.ID, organization)
	require.NoError(t, err)

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)
}

func TestNamespaceRepository_Get(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	cleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, owner.ID, organization)
	require.NoError(t, err)

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	namespace.Description = testutil.GenerateRandomString(10)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	ns, err := namespaceRepo.Get(ctx, namespace.ID)
	require.NoError(t, err)

	require.Equal(t, namespace.ID, ns.ID)
	require.Equal(t, namespace.Name, ns.Name)
	require.Equal(t, namespace.Description, ns.Description)
	require.Equal(t, namespace.Projects, ns.Projects)
	require.Equal(t, namespace.Documents, ns.Documents)
	require.WithinDuration(t, *namespace.CreatedAt, *ns.CreatedAt, 1*time.Second)
	require.Nil(t, ns.UpdatedAt)
}

func TestNamespaceRepository_GetAll(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	cleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, owner.ID, organization)
	require.NoError(t, err)

	namespace1, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	namespace2, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	namespace3, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	require.NoError(t, namespaceRepo.Create(ctx, organization.ID, namespace1))
	require.NoError(t, namespaceRepo.Create(ctx, organization.ID, namespace2))
	require.NoError(t, namespaceRepo.Create(ctx, organization.ID, namespace3))

	namespaces, err := namespaceRepo.GetAll(ctx, organization.ID, 0, 10)
	require.NoError(t, err)
	require.Len(t, namespaces, 3)

	namespaces, err = namespaceRepo.GetAll(ctx, organization.ID, 1, 10)
	require.NoError(t, err)
	require.Len(t, namespaces, 2)

	namespaces, err = namespaceRepo.GetAll(ctx, organization.ID, 2, 10)
	require.NoError(t, err)
	require.Len(t, namespaces, 1)

	namespaces, err = namespaceRepo.GetAll(ctx, organization.ID, 3, 10)
	require.NoError(t, err)
	require.Len(t, namespaces, 0)
}

func TestNamespaceRepository_Update(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	cleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, owner.ID, organization)
	require.NoError(t, err)

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	namespace.Description = testutil.GenerateRandomString(10)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	ns, err := namespaceRepo.Get(ctx, namespace.ID)
	require.NoError(t, err)

	require.Equal(t, namespace.ID, ns.ID)

	patch := map[string]any{
		"name":        "new name",
		"description": "new description",
	}

	updated, err := namespaceRepo.Update(ctx, ns.ID, patch)
	require.NoError(t, err)

	require.Equal(t, updated.ID, ns.ID)
	require.Equal(t, patch["name"], updated.Name)
	require.Equal(t, patch["description"], updated.Description)
	require.Equal(t, ns.Projects, updated.Projects)
	require.Equal(t, ns.Documents, updated.Documents)
	require.WithinDuration(t, *ns.CreatedAt, *updated.CreatedAt, 1*time.Second)
	require.NotNil(t, updated.UpdatedAt)
}

func TestNamespaceRepository_Delete(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	cleanupNeo4jStore(t, ctx, db)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	owner := prepareUser(t)
	err = userRepo.Create(ctx, owner)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, owner.ID, organization)
	require.NoError(t, err)

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	namespace.Description = testutil.GenerateRandomString(10)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	ns, err := namespaceRepo.Get(ctx, namespace.ID)
	require.NoError(t, err)

	require.Equal(t, namespace.ID, ns.ID)

	err = namespaceRepo.Delete(ctx, namespace.ID)
	require.NoError(t, err)

	ns, err = namespaceRepo.Get(ctx, namespace.ID)
	require.Error(t, err)
}
