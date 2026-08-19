package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

// NewUserService creates a new UserService for testing.
func NewUserService(t *testing.T, neo4jDBConf *config.GraphDatabaseConfig) service.UserService {
	neo4jDB, _ := testRepo.NewNeo4jDatabase(t, neo4jDBConf)

	permissionRepo, err := repository.NewNeo4jPermissionRepository(
		repository.WithNeo4jDatabase(neo4jDB),
	)
	require.NoError(t, err)

	userRepo, err := repository.NewNeo4jUserRepository(
		repository.WithNeo4jDatabase(neo4jDB),
	)
	require.NoError(t, err)

	licenseRepo, err := repository.NewNeo4jLicenseRepository(
		repository.WithNeo4jDatabase(neo4jDB),
	)
	require.NoError(t, err)

	permissionSvc, err := service.NewPermissionService(
		permissionRepo,
	)
	require.NoError(t, err)

	licenseSvc, err := service.NewLicenseService(
		testutil.ParseLicense(t),
		licenseRepo,
		service.WithPermissionService(permissionSvc),
	)
	require.NoError(t, err)

	s, err := service.NewUserService(
		service.WithUserRepository(userRepo),
		service.WithPermissionService(permissionSvc),
		service.WithLicenseService(licenseSvc),
	)
	require.NoError(t, err)

	return s
}

// NewResourceOwner creates a new user with organization.create on Installation.
func NewResourceOwner(t *testing.T, neo4jDBConf *config.GraphDatabaseConfig) *repository.User {
	neo4jDB, _ := testRepo.NewNeo4jDatabase(t, neo4jDBConf)

	userRepo, err := repository.NewNeo4jUserRepository(
		repository.WithNeo4jDatabase(neo4jDB),
	)
	require.NoError(t, err)

	owner, err := userRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	require.NoError(t, err)

	require.NoError(t, testRepo.MakeUserSystemOwner(owner.ID, neo4jDB))

	return owner
}
