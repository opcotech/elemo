//go:build integration

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

	_, err = s.PermissionRepo.Create(context.Background(), repository.CreateGrantOpts{
		Principal: s.owner.ID,
		Scope:     s.organization.ID,
		Actions:   testModel.OrgAdminActions(),
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

	hasPermission, err := s.PermissionRepo.Has(context.Background(), s.owner.ID, ns.ID, model.ActionNamespaceRead)
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

	namespaces, err := s.namespaceService.List(s.ctx, s.organization.ID, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Assert().Len(namespaces.Items, 2)
}

func (s *NamespaceServiceIntegrationTestSuite) TestListAccessibleFromProjectViewer() {
	created, err := s.namespaceService.Create(s.ctx, s.organization.ID, service.CreateNamespaceOpts{
		Name: "viewer-namespace", Description: "viewer namespace description",
	})
	s.Require().NoError(err)

	sibling, err := s.namespaceService.Create(s.ctx, s.organization.ID, service.CreateNamespaceOpts{
		Name: "other-namespace", Description: "other namespace description",
	})
	s.Require().NoError(err)

	project, err := s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(created.ID, s.owner.ID))
	s.Require().NoError(err)

	viewer, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	_, err = s.PermissionRepo.Create(context.Background(), repository.CreateGrantOpts{
		Principal: viewer.ID,
		Scope:     project.ID,
		Actions:   testModel.ProjectViewerActions(),
	})
	s.Require().NoError(err)
	viewerCtx := context.WithValue(context.Background(), pkg.CtxKeyUserID, viewer.ID)

	accessible, err := s.namespaceService.ListAccessible(viewerCtx, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Require().Len(accessible.Items, 1)
	s.Assert().Equal(created.ID, accessible.Items[0].ID)
	s.Assert().Equal(s.organization.ID, accessible.Items[0].Organization.ID)
	s.Assert().Equal(s.organization.Name, accessible.Items[0].Organization.Name)

	listed, err := s.namespaceService.List(viewerCtx, s.organization.ID, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Require().Len(listed.Items, 1)
	s.Assert().Equal(created.ID, listed.Items[0].ID)

	_, err = s.namespaceService.Get(viewerCtx, created.ID)
	s.Assert().ErrorIs(err, service.ErrNoPermission)
	_, err = s.namespaceService.Get(viewerCtx, sibling.ID)
	s.Assert().ErrorIs(err, service.ErrNoPermission)
}

func (s *NamespaceServiceIntegrationTestSuite) TestOrgMemberDoesNotListUnreadableNamespaces() {
	_, err := s.namespaceService.Create(s.ctx, s.organization.ID, service.CreateNamespaceOpts{
		Name: "member-hidden", Description: "member hidden description",
	})
	s.Require().NoError(err)

	member, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), s.organization.ID, member.ID))
	_, err = s.PermissionRepo.Create(context.Background(), repository.CreateGrantOpts{
		Principal: s.organization.ID,
		Scope:     s.organization.ID,
		Actions:   []model.Action{model.ActionOrganizationRead},
	})
	s.Require().NoError(err)

	memberCtx := context.WithValue(context.Background(), pkg.CtxKeyUserID, member.ID)
	listed, err := s.namespaceService.List(memberCtx, s.organization.ID, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Assert().Empty(listed.Items)
}

func (s *NamespaceServiceIntegrationTestSuite) TestListAccessibleFromCrossOrgProjectViewer() {
	ns, err := s.namespaceService.Create(s.ctx, s.organization.ID, service.CreateNamespaceOpts{
		Name: "shared-namespace", Description: "shared namespace description",
	})
	s.Require().NoError(err)

	project, err := s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(ns.ID, s.owner.ID))
	s.Require().NoError(err)

	partnerOwner, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	partnerOrg, err := s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(partnerOwner.ID))
	s.Require().NoError(err)
	collaborator, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), partnerOrg.ID, collaborator.ID))

	_, err = s.PermissionRepo.Create(context.Background(), repository.CreateGrantOpts{
		Principal: partnerOrg.ID,
		Scope:     project.ID,
		Actions:   testModel.ProjectViewerActions(),
	})
	s.Require().NoError(err)

	collaboratorCtx := context.WithValue(context.Background(), pkg.CtxKeyUserID, collaborator.ID)

	hasOrgRead, err := s.PermissionRepo.Has(context.Background(), collaborator.ID, s.organization.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().False(hasOrgRead)

	accessible, err := s.namespaceService.ListAccessible(collaboratorCtx, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Require().Len(accessible.Items, 1)
	s.Assert().Equal(ns.ID, accessible.Items[0].ID)
	s.Assert().Equal(s.organization.ID, accessible.Items[0].Organization.ID)
	s.Assert().Equal(s.organization.Name, accessible.Items[0].Organization.Name)

	listed, err := s.namespaceService.List(collaboratorCtx, s.organization.ID, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Require().Len(listed.Items, 1)
	s.Assert().Equal(ns.ID, listed.Items[0].ID)

	_, err = s.namespaceService.Get(collaboratorCtx, ns.ID)
	s.Assert().ErrorIs(err, service.ErrNoPermission)
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
