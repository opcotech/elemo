package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/service"
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

	s, err := service.NewUserService(
		service.WithUserRepository(userRepo),
		service.WithPermissionRepository(permissionRepo),
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
	CREATE
		(rt:` + model.ResourceTypeResourceType.String() + ` {id: $rt_label, system: true, created_at: datetime()}),
		(r:` + model.ResourceTypeRole.String() + ` {id: "Owner", system: true, created_at: datetime()}),
		(r)-[:` + neo4j.EdgeKindHasPermission.String() + ` {kind: $perm_kind, created_at: datetime()}]->(rt),
		(u)-[:` + neo4j.EdgeKindMemberOf.String() + `]->(r)
	`

	params := map[string]any{
		"id":        owner.ID.String(),
		"rt_label":  model.ResourceTypeUser.String(),
		"perm_kind": model.PermissionKindAll.String(),
	}

	_, err = neo4jDB.GetWriteSession(context.Background()).Run(context.Background(), cypher, params)
	require.NoError(t, err)

	return owner
}
