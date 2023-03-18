package neo4j_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

type PermissionRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *model.User
	testOrg    *model.Organization
	permission *model.Permission
}

func (s *PermissionRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *PermissionRepositoryIntegrationTestSuite) SetupTest() {
	orgOwner := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), orgOwner))

	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), orgOwner.ID, s.testOrg))

	s.permission = testModel.NewPermission(s.testUser.ID, s.testOrg.ID, model.PermissionKindRead)
}

func (s *PermissionRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *PermissionRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *PermissionRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypePermission), s.permission.ID)
	s.Assert().NotNil(s.permission.CreatedAt)
	s.Assert().Nil(s.permission.UpdatedAt)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	permission, err := s.PermissionRepo.Get(context.Background(), s.permission.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.permission.ID, permission.ID)
	s.Assert().Equal(s.permission.Subject, permission.Subject)
	s.Assert().Equal(s.permission.Target, permission.Target)
	s.Assert().Equal(s.permission.Kind, permission.Kind)
	s.Assert().WithinDuration(*s.permission.CreatedAt, *permission.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(permission.UpdatedAt)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestHasPermission() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	hasPermission, err := s.PermissionRepo.HasPermission(
		context.Background(),
		s.permission.Subject,
		s.permission.Target,
		model.PermissionKindRead,
	)
	s.Require().NoError(err)
	s.Assert().True(hasPermission)

	hasPermission, err = s.PermissionRepo.HasPermission(
		context.Background(),
		s.permission.Subject,
		s.permission.Target,
		model.PermissionKindDelete,
	)
	s.Require().NoError(err)
	s.Assert().False(hasPermission)

	hasPermission, err = s.PermissionRepo.HasPermission(
		context.Background(),
		s.testUser.ID,
		model.MustNewNilID(model.ResourceTypeOrganization),
		model.PermissionKindCreate,
	)
	s.Require().NoError(err)
	s.Assert().False(hasPermission)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestGetBySubject() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	permissions, err := s.PermissionRepo.GetBySubject(context.Background(), s.permission.Subject)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 1)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestGetByTarget() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	permissions, err := s.PermissionRepo.GetByTarget(context.Background(), s.permission.Target)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 2) // the owner and the test user
}

func (s *PermissionRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	updatedKind := model.PermissionKindDelete
	permission, err := s.PermissionRepo.Update(context.Background(), s.permission.ID, updatedKind)
	s.Require().NoError(err)

	s.Assert().Equal(s.permission.ID, permission.ID)
	s.Assert().Equal(s.permission.Subject, permission.Subject)
	s.Assert().Equal(s.permission.Target, permission.Target)
	s.Assert().Equal(updatedKind, permission.Kind)
	s.Assert().WithinDuration(*s.permission.CreatedAt, *permission.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(permission.UpdatedAt)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestHasAnyRelation() {
	hasRelation, err := s.PermissionRepo.HasAnyRelation(context.Background(), s.testUser.ID, s.testOrg.ID)
	s.Require().NoError(err)
	s.Assert().False(hasRelation)

	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), s.testOrg.ID, s.testUser.ID))

	hasRelation, err = s.PermissionRepo.HasAnyRelation(context.Background(), s.testUser.ID, s.testOrg.ID)
	s.Require().NoError(err)
	s.Assert().True(hasRelation)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestHasSystemRole() {
	hasRole, err := s.PermissionRepo.HasSystemRole(
		context.Background(),
		s.testUser.ID,
		model.SystemRoleOwner,
		model.SystemRoleAdmin,
		model.SystemRoleSupport,
	)
	s.Require().NoError(err)
	s.Assert().False(hasRole)

	// Elevate user to system owner
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.testUser.ID, s.Neo4jDB))

	hasRole, err = s.PermissionRepo.HasSystemRole(
		context.Background(),
		s.testUser.ID,
		model.SystemRoleOwner,
		model.SystemRoleAdmin,
		model.SystemRoleSupport,
	)
	s.Require().NoError(err)
	s.Assert().True(hasRole)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), s.permission))

	s.Require().NoError(s.PermissionRepo.Delete(context.Background(), s.permission.ID))

	_, err := s.PermissionRepo.Get(context.Background(), s.permission.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestPermissionRepositoryIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(PermissionRepositoryIntegrationTestSuite))
}
