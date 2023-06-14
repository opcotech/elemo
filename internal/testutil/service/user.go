package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

// NewUserService creates a new UserService for testing.
func NewUserService(t *testing.T, neo4jDBConf *config.GraphDatabaseConfig) service.UserService {
	neo4jDB, _ := testRepo.NewNeo4jDatabase(t, neo4jDBConf)

	permissionRepo, err := neo4j.NewPermissionRepository(
		neo4j.WithDatabase(neo4jDB),
	)
	require.NoError(t, err)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(neo4jDB),
	)
	require.NoError(t, err)

	licenseRepo, err := neo4j.NewLicenseRepository(
		neo4j.WithDatabase(neo4jDB),
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

// NewResourceOwner creates a new user with the Owner role.
func NewResourceOwner(t *testing.T, neo4jDBConf *config.GraphDatabaseConfig) *model.User {
	neo4jDB, _ := testRepo.NewNeo4jDatabase(t, neo4jDBConf)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(neo4jDB),
	)
	require.NoError(t, err)

	owner := testModel.NewUser()
	err = userRepo.Create(context.Background(), owner)
	require.NoError(t, err)

	cypher := `
	MATCH (u:` + owner.ID.Label() + ` {id: $id})
	MATCH (r:` + model.ResourceTypeRole.String() + ` {id: $role_label, system: true})
	CREATE (u)-[:` + neo4j.EdgeKindMemberOf.String() + `]->(r)`

	params := map[string]any{
		"id":         owner.ID.String(),
		"role_label": model.SystemRoleOwner.String(),
		"perm_kind":  model.PermissionKindAll.String(),
	}

	_, err = neo4jDB.GetWriteSession(context.Background()).Run(context.Background(), cypher, params)
	require.NoError(t, err)

	return owner
}
