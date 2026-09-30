package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

// NewUserService creates a new UserService for testing.
func NewUserService(t *testing.T, neo4jDBConf *config.GraphDatabaseConfig) service.UserService {
	neo4jDB, _ := testRepo.NewNeo4jDatabase(t, neo4jDBConf)

	userRepo, err := repository.NewNeo4jUserRepository(
		repository.WithNeo4jDatabase(neo4jDB),
	)
	require.NoError(t, err)

	s, err := service.NewUserService(
		userRepo,
		mockrepo.NewMockUserTokenRepository(nil),
		entitlement.Unrestricted(),
	)
	require.NoError(t, err)

	return s
}
