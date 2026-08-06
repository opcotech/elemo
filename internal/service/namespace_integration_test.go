package service_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
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

	owner        *repository.User
	organization *repository.Organization

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
	var err error
	s.owner, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.owner.ID, s.Neo4jDB))

	s.organization, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.owner.ID))
	s.Require().NoError(err)

	_, err = s.PermissionRepo.Create(context.Background(), repository.CreatePermissionOpts{
		Subject: s.owner.ID,
		Target:  s.organization.ID,
		Kind:    model.PermissionKindWrite,
	})
	s.Require().NoError(err)
}

func (s *NamespaceServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *NamespaceServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *NamespaceServiceIntegrationTestSuite) TestCreate() {
	ns, err := s.namespaceService.Create(s.ctx, s.organization.ID, service.CreateNamespaceOpts{
		Name:        "test-namespace",
		Description: "test namespace description",
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(ns.ID)
	s.Assert().NotNil(ns.CreatedAt)

	hasPermission, err := s.PermissionRepo.HasPermission(
		context.Background(),
		s.owner.ID,
		ns.ID,
		model.PermissionKindAll,
	)
	s.Require().NoError(err)
	s.Assert().True(hasPermission)
}

func (s *NamespaceServiceIntegrationTestSuite) TestCreateWithoutPermission() {
	otherUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	otherCtx := context.WithValue(context.Background(), pkg.CtxKeyUserID, otherUser.ID)

	_, err = s.namespaceService.Create(otherCtx, s.organization.ID, service.CreateNamespaceOpts{
		Name: "unauthorized-ns", Description: "should fail description",
	})
	s.Assert().ErrorIs(err, service.ErrNoPermission)
}

func (s *NamespaceServiceIntegrationTestSuite) TestGet() {
	created, err := s.namespaceService.Create(s.ctx, s.organization.ID, service.CreateNamespaceOpts{
		Name: "get-namespace", Description: "get namespace description",
	})
	s.Require().NoError(err)

	ns, err := s.namespaceService.Get(s.ctx, created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, ns.ID)
	s.Assert().Equal(created.Name, ns.Name)
}

func (s *NamespaceServiceIntegrationTestSuite) TestGetAll() {
	_, err := s.namespaceService.Create(s.ctx, s.organization.ID, service.CreateNamespaceOpts{
		Name: "ns-one", Description: "ns one description",
	})
	s.Require().NoError(err)
	_, err = s.namespaceService.Create(s.ctx, s.organization.ID, service.CreateNamespaceOpts{
		Name: "ns-two", Description: "ns two description",
	})
	s.Require().NoError(err)

	namespaces, err := s.namespaceService.GetAll(s.ctx, s.organization.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(namespaces, 2)
}

func (s *NamespaceServiceIntegrationTestSuite) TestUpdate() {
	created, err := s.namespaceService.Create(s.ctx, s.organization.ID, service.CreateNamespaceOpts{
		Name: "upd-namespace", Description: "upd namespace description",
	})
	s.Require().NoError(err)

	ns, err := s.namespaceService.Update(s.ctx, created.ID, service.UpdateNamespaceOpts{
		Name: optional.Some("updated-name"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("updated-name", ns.Name)
	s.Assert().NotNil(ns.UpdatedAt)
}

func (s *NamespaceServiceIntegrationTestSuite) TestDelete() {
	created, err := s.namespaceService.Create(s.ctx, s.organization.ID, service.CreateNamespaceOpts{
		Name: "del-namespace", Description: "del namespace description",
	})
	s.Require().NoError(err)

	s.Require().NoError(s.namespaceService.Delete(s.ctx, created.ID))
	_, err = s.namespaceService.Get(s.ctx, created.ID)
	s.Assert().Error(err)
}

func TestNamespaceServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NamespaceServiceIntegrationTestSuite))
}
