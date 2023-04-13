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

func TestLicenseRepository_ActiveUserCount(t *testing.T) {
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

	licenseRepo, err := neo4j.NewLicenseRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	require.NoError(t, userRepo.Create(ctx, user))

	userPending := prepareUser(t)
	userPending.Status = model.UserStatusPending
	require.NoError(t, userRepo.Create(ctx, userPending))

	userDeleted := prepareUser(t)
	userDeleted.Status = model.UserStatusDeleted
	require.NoError(t, userRepo.Create(ctx, userDeleted))

	userInactive := prepareUser(t)
	userInactive.Status = model.UserStatusInactive
	require.NoError(t, userRepo.Create(ctx, userInactive))

	count, err := licenseRepo.ActiveUserCount(ctx)
	require.NoError(t, err)
	require.Equal(t, uint(2), count)
}

func TestLicenseRepository_ActiveOrganizationCount(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	licenseRepo, err := neo4j.NewLicenseRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	org := prepareOrganization(t)
	require.NoError(t, orgRepo.Create(ctx, user.ID, org))

	orgDeleted := prepareOrganization(t)
	orgDeleted.Status = model.OrganizationStatusDeleted
	require.NoError(t, orgRepo.Create(ctx, user.ID, orgDeleted))

	count, err := licenseRepo.ActiveOrganizationCount(ctx)
	require.NoError(t, err)
	require.Equal(t, uint(1), count)
}

func TestLicenseRepository_DocumentCount(t *testing.T) {
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

	licenseRepo, err := neo4j.NewLicenseRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	documentRepo, err := neo4j.NewDocumentRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	require.NoError(t, documentRepo.Create(ctx, user.ID, prepareDocument(t, user.ID)))

	count, err := licenseRepo.DocumentCount(ctx)
	require.NoError(t, err)

	require.Equal(t, uint(1), count)
}

func TestLicenseRepository_NamespaceCount(t *testing.T) {
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

	licenseRepo, err := neo4j.NewLicenseRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	namespaceRepo, err := neo4j.NewNamespaceRepository(
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

	org := prepareOrganization(t)
	require.NoError(t, orgRepo.Create(ctx, user.ID, org))

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, namespaceRepo.Create(ctx, org.ID, namespace))

	count, err := licenseRepo.NamespaceCount(ctx)
	require.NoError(t, err)

	require.Equal(t, uint(1), count)
}

func TestLicenseRepository_ProjectCount(t *testing.T) {
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

	licenseRepo, err := neo4j.NewLicenseRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	org := prepareOrganization(t)
	require.NoError(t, orgRepo.Create(ctx, user.ID, org))

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, namespaceRepo.Create(ctx, org.ID, namespace))

	project := prepareProject(t)
	require.NoError(t, projectRepo.Create(ctx, namespace.ID, project))

	count, err := licenseRepo.ProjectCount(ctx)
	require.NoError(t, err)

	require.Equal(t, uint(1), count)
}

func TestLicenseRepository_RoleCount(t *testing.T) {
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

	licenseRepo, err := neo4j.NewLicenseRepository(
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

	err = roleRepo.Create(ctx, user.ID, organization.ID, role)
	require.NoError(t, err)

	count, err := licenseRepo.RoleCount(ctx)
	require.NoError(t, err)

	require.Equal(t, uint(1), count)
}
