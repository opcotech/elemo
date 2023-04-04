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
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

func prepareProject(t *testing.T) *model.Project {
	project, err := model.NewProject(testutil.GenerateRandomStringAlpha(6), testutil.GenerateRandomString(10))
	require.NoError(t, err)

	project.Description = testutil.GenerateRandomString(10)
	project.Logo = "https://www.gravatar.com/avatar"

	return project
}

func TestProjectRepository_Create(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
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

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)
}

func TestProjectRepository_Get(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
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

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	got, err := projectRepo.Get(ctx, project.ID)
	require.NoError(t, err)

	assert.Equal(t, project.ID, got.ID)
	assert.Equal(t, project.Name, got.Name)
	assert.Equal(t, project.Description, got.Description)
	assert.Equal(t, project.Logo, got.Logo)
	assert.Len(t, got.Teams, 0)
	assert.Len(t, got.Documents, 0)
	assert.WithinDuration(t, *project.CreatedAt, *got.CreatedAt, 1*time.Second)
	assert.Nil(t, got.UpdatedAt)
}

func TestProjectRepository_GetByKey(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
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

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	got, err := projectRepo.GetByKey(ctx, project.Key)
	require.NoError(t, err)

	assert.Equal(t, project.ID, got.ID)
	assert.Equal(t, project.Name, got.Name)
	assert.Equal(t, project.Description, got.Description)
	assert.Equal(t, project.Logo, got.Logo)
	assert.Len(t, got.Teams, 0)
	assert.Len(t, got.Documents, 0)
	assert.WithinDuration(t, *project.CreatedAt, *got.CreatedAt, 1*time.Second)
	assert.Nil(t, got.UpdatedAt)
}

func TestProjectRepository_GetAll(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
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

	err = projectRepo.Create(ctx, namespace.ID, prepareProject(t))
	require.NoError(t, err)

	err = projectRepo.Create(ctx, namespace.ID, prepareProject(t))
	require.NoError(t, err)

	err = projectRepo.Create(ctx, namespace.ID, prepareProject(t))
	require.NoError(t, err)

	got, err := projectRepo.GetAll(ctx, namespace.ID, 0, 10)
	require.NoError(t, err)
	assert.Len(t, got, 3)

	got, err = projectRepo.GetAll(ctx, namespace.ID, 0, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = projectRepo.GetAll(ctx, namespace.ID, 1, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = projectRepo.GetAll(ctx, namespace.ID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = projectRepo.GetAll(ctx, namespace.ID, 3, 2)
	require.NoError(t, err)
	assert.Len(t, got, 0)
}

func TestProjectRepository_Update(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
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

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	patch := map[string]any{
		"name":        testutil.GenerateRandomString(10),
		"description": testutil.GenerateRandomString(10),
	}

	updated, err := projectRepo.Update(ctx, project.ID, patch)
	require.NoError(t, err)

	assert.Equal(t, project.ID, updated.ID)
	assert.Equal(t, patch["name"], updated.Name)
	assert.Equal(t, patch["description"], updated.Description)
	assert.Equal(t, project.Logo, updated.Logo)
	assert.Len(t, updated.Teams, 0)
	assert.Len(t, updated.Documents, 0)
	assert.WithinDuration(t, *project.CreatedAt, *updated.CreatedAt, 1*time.Second)
	assert.NotNil(t, updated.UpdatedAt)
}

func TestProjectRepository_Delete(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
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

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	_, err = projectRepo.Get(ctx, project.ID)
	require.NoError(t, err)

	err = projectRepo.Delete(ctx, project.ID)
	require.NoError(t, err)

	_, err = projectRepo.Get(ctx, project.ID)
	require.Error(t, err)
}
