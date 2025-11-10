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
)

type NamespaceRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser  *model.User
	testOrg   *model.Organization
	namespace *model.Namespace
}

func (s *NamespaceRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *NamespaceRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.namespace = testModel.NewNamespace()
}

func (s *NamespaceRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.namespace))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeNamespace), s.namespace.ID)
	s.Assert().NotNil(s.namespace.CreatedAt)
	s.Assert().Nil(s.namespace.UpdatedAt)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.namespace))

	namespace, err := s.NamespaceRepo.Get(context.Background(), s.namespace.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.namespace.ID, namespace.ID)
	s.Assert().Equal(s.namespace.Name, namespace.Name)
	s.Assert().Equal(s.namespace.Description, namespace.Description)
	s.Assert().Equal(s.namespace.Projects, namespace.Projects)
	s.Assert().Equal(s.namespace.Documents, namespace.Documents)
	s.Assert().WithinDuration(*s.namespace.CreatedAt, *namespace.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(s.namespace.UpdatedAt)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestGetAll() {
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.namespace))
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, testModel.NewNamespace()))
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, testModel.NewNamespace()))

	namespaces, err := s.NamespaceRepo.GetAll(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(namespaces, 3)

	namespaces, err = s.NamespaceRepo.GetAll(context.Background(), s.testOrg.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(namespaces, 2)

	namespaces, err = s.NamespaceRepo.GetAll(context.Background(), s.testOrg.ID, 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(namespaces, 1)

	namespaces, err = s.NamespaceRepo.GetAll(context.Background(), s.testOrg.ID, 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(namespaces, 0)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.namespace))

	patch := map[string]any{
		"name":        "new name",
		"description": "new description",
	}

	namespace, err := s.NamespaceRepo.Update(context.Background(), s.namespace.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(s.namespace.ID, namespace.ID)
	s.Assert().Equal(patch["name"], namespace.Name)
	s.Assert().Equal(patch["description"], namespace.Description)
	s.Assert().Equal(s.namespace.Projects, namespace.Projects)
	s.Assert().Equal(s.namespace.Documents, namespace.Documents)
	s.Assert().WithinDuration(*s.namespace.CreatedAt, *namespace.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(namespace.UpdatedAt)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.namespace))

	s.Require().NoError(s.NamespaceRepo.Delete(context.Background(), s.namespace.ID))

	_, err := s.NamespaceRepo.Get(context.Background(), s.namespace.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestNamespaceRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NamespaceRepositoryIntegrationTestSuite))
}
