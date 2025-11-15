package service_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

type NamespaceServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	namespaceService service.NamespaceService

	owner        *model.User
	organization *model.Organization
	namespace    *model.Namespace

	ctx context.Context
}

func (s *NamespaceServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)
	s.SetupPg(&s.ContainerIntegrationTestSuite, container)

	permissionService, err := service.NewPermissionService(s.PermissionRepo)
	s.Require().NoError(err)

	licenseService, err := service.NewLicenseService(
		testutil.ParseLicense(s.T()),
		s.LicenseRepo,
		service.WithPermissionService(permissionService),
	)
	s.Require().NoError(err)

	s.namespaceService, err = service.NewNamespaceService(
		service.WithNamespaceRepository(s.NamespaceRepo),
		service.WithPermissionService(permissionService),
		service.WithLicenseService(licenseService),
	)
	s.Require().NoError(err)
}

func (s *NamespaceServiceIntegrationTestSuite) SetupTest() {
	s.owner = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.owner))

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.owner.ID, s.Neo4jDB))

	s.organization = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.owner.ID, s.organization))

	// Grant write permission on organization to owner
	perm, err := model.NewPermission(s.owner.ID, s.organization.ID, model.PermissionKindWrite)
	s.Require().NoError(err)
	s.Require().NoError(s.PermissionRepo.Create(context.Background(), perm))

	s.namespace = testModel.NewNamespace()
}

func (s *NamespaceServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *NamespaceServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *NamespaceServiceIntegrationTestSuite) TestCreate() {
	err := s.namespaceService.Create(s.ctx, s.organization.ID, s.namespace)
	s.Require().NoError(err)
	s.Require().NotEmpty(s.namespace.ID)
	s.Assert().NotNil(s.namespace.CreatedAt)
	s.Assert().Nil(s.namespace.UpdatedAt)

	// Verify that the creator has * permission on the namespace
	hasPermission, err := s.PermissionRepo.HasPermission(
		context.Background(),
		s.owner.ID,
		s.namespace.ID,
		model.PermissionKindAll,
	)
	s.Require().NoError(err)
	s.Assert().True(hasPermission)
}

func (s *NamespaceServiceIntegrationTestSuite) TestCreateWithoutPermission() {
	// Create a user without permission
	user := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), user))
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, user.ID)

	err := s.namespaceService.Create(ctx, s.organization.ID, s.namespace)
	s.Require().Error(err)
	s.Assert().ErrorIs(err, service.ErrNoPermission)
}

func (s *NamespaceServiceIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.namespaceService.Create(s.ctx, s.organization.ID, s.namespace))

	namespace, err := s.namespaceService.Get(s.ctx, s.namespace.ID)
	s.Require().NoError(err)
	s.Assert().Equal(s.namespace.ID, namespace.ID)
	s.Assert().Equal(s.namespace.Name, namespace.Name)
	s.Assert().Equal(s.namespace.Description, namespace.Description)
	s.Assert().ElementsMatch(s.namespace.Projects, namespace.Projects)
	s.Assert().ElementsMatch(s.namespace.Documents, namespace.Documents)
	s.Assert().Equal(s.namespace.CreatedAt, namespace.CreatedAt)
	s.Assert().Equal(s.namespace.UpdatedAt, namespace.UpdatedAt)
}

func (s *NamespaceServiceIntegrationTestSuite) TestGetWithoutPermission() {
	s.Require().NoError(s.namespaceService.Create(s.ctx, s.organization.ID, s.namespace))

	// Create a user without permission
	user := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), user))
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, user.ID)

	_, err := s.namespaceService.Get(ctx, s.namespace.ID)
	s.Require().Error(err)
	s.Assert().ErrorIs(err, service.ErrNoPermission)
}

func (s *NamespaceServiceIntegrationTestSuite) TestGetAll() {
	namespace1 := testModel.NewNamespace()
	s.Require().NoError(s.namespaceService.Create(s.ctx, s.organization.ID, namespace1))

	namespace2 := testModel.NewNamespace()
	s.Require().NoError(s.namespaceService.Create(s.ctx, s.organization.ID, namespace2))

	namespace3 := testModel.NewNamespace()
	s.Require().NoError(s.namespaceService.Create(s.ctx, s.organization.ID, namespace3))

	namespaces, err := s.namespaceService.GetAll(s.ctx, s.organization.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().GreaterOrEqual(len(namespaces), 3)

	namespaceIDs := make([]model.ID, len(namespaces))
	for i, ns := range namespaces {
		namespaceIDs[i] = ns.ID
	}

	s.Assert().Contains(namespaceIDs, namespace1.ID)
	s.Assert().Contains(namespaceIDs, namespace2.ID)
	s.Assert().Contains(namespaceIDs, namespace3.ID)
}

func (s *NamespaceServiceIntegrationTestSuite) TestGetAllWithPagination() {
	// Create multiple namespaces
	for i := 0; i < 5; i++ {
		ns := testModel.NewNamespace()
		s.Require().NoError(s.namespaceService.Create(s.ctx, s.organization.ID, ns))
	}

	// Get first page
	namespaces, err := s.namespaceService.GetAll(s.ctx, s.organization.ID, 0, 2)
	s.Require().NoError(err)
	s.Assert().Len(namespaces, 2)

	// Get second page
	namespaces, err = s.namespaceService.GetAll(s.ctx, s.organization.ID, 2, 2)
	s.Require().NoError(err)
	s.Assert().LessOrEqual(len(namespaces), 2)
}

func (s *NamespaceServiceIntegrationTestSuite) TestGetAllWithoutPermission() {
	// Create a user without permission
	user := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), user))
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, user.ID)

	_, err := s.namespaceService.GetAll(ctx, s.organization.ID, 0, 10)
	s.Require().Error(err)
	s.Assert().ErrorIs(err, service.ErrNoPermission)
}

func (s *NamespaceServiceIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.namespaceService.Create(s.ctx, s.organization.ID, s.namespace))

	patch := map[string]any{
		"name":        "Updated Namespace Name",
		"description": "Updated description",
	}

	updatedNamespace, err := s.namespaceService.Update(s.ctx, s.namespace.ID, patch)
	s.Require().NoError(err)
	s.Assert().Equal("Updated Namespace Name", updatedNamespace.Name)
	s.Assert().Equal("Updated description", updatedNamespace.Description)
	s.Assert().NotNil(updatedNamespace.UpdatedAt)
}

func (s *NamespaceServiceIntegrationTestSuite) TestUpdateWithoutPermission() {
	s.Require().NoError(s.namespaceService.Create(s.ctx, s.organization.ID, s.namespace))

	// Create a user without permission
	user := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), user))
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, user.ID)

	patch := map[string]any{"name": "Updated Name"}
	_, err := s.namespaceService.Update(ctx, s.namespace.ID, patch)
	s.Require().Error(err)
	s.Assert().ErrorIs(err, service.ErrNoPermission)
}

func (s *NamespaceServiceIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.namespaceService.Create(s.ctx, s.organization.ID, s.namespace))

	err := s.namespaceService.Delete(s.ctx, s.namespace.ID)
	s.Require().NoError(err)

	_, err = s.namespaceService.Get(s.ctx, s.namespace.ID)
	s.Require().Error(err)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func (s *NamespaceServiceIntegrationTestSuite) TestDeleteWithoutPermission() {
	s.Require().NoError(s.namespaceService.Create(s.ctx, s.organization.ID, s.namespace))

	// Create a user without permission
	user := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), user))
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, user.ID)

	err := s.namespaceService.Delete(ctx, s.namespace.ID)
	s.Require().Error(err)
	s.Assert().ErrorIs(err, service.ErrNoPermission)

	// Verify namespace still exists
	_, err = s.namespaceService.Get(s.ctx, s.namespace.ID)
	s.Require().NoError(err)
}

func TestNamespaceServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NamespaceServiceIntegrationTestSuite))
}
