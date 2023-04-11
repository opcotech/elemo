//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/testutil"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

func TestRelationRepository_HasAnyRelation(t *testing.T) {
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

	relationRepo, err := neo4j.NewRelationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	hasRelation, err := relationRepo.HasAnyRelation(ctx, user.ID, project.ID)
	require.NoError(t, err)

	require.True(t, hasRelation)
}
