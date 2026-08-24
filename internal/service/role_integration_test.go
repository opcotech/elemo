package service_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type RoleServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	roleService service.RoleService

	owner        *repository.User
	organization *repository.Organization
	ctx          context.Context
}

func (s *RoleServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)
	s.SetupPg(&s.ContainerIntegrationTestSuite, container)

	permissionService, err := service.NewPermissionService(s.PermissionRepo, s.RoleRepo)
	s.Require().NoError(err)

	licenseService, err := service.NewLicenseService(
		testutil.ParseLicense(s.T()),
		s.LicenseRepo,
		permissionService,
	)
	s.Require().NoError(err)

	notificationService, err := service.NewNotificationService(s.NotificationRepo)
	s.Require().NoError(err)

	s.roleService, err = service.NewRoleService(
		s.RoleRepo,
		permissionService,
		licenseService,
		s.OrganizationRepo,
		notificationService,
	)
	s.Require().NoError(err)
}

func (s *RoleServiceIntegrationTestSuite) SetupTest() {
	var err error
	s.owner, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)

	s.organization, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.owner.ID))
	s.Require().NoError(err)

	_, err = s.PermissionRepo.Create(context.Background(), repository.CreateGrantOpts{
		Principal: s.owner.ID,
		Scope:     s.organization.ID,
		Actions:   testModel.OrgAdminActions(),
	})
	s.Require().NoError(err)
}

func (s *RoleServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *RoleServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *RoleServiceIntegrationTestSuite) TestCreate() {
	role, err := s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, service.CreateRoleOpts{
		Name:        "test-role",
		Description: "test role description",
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(role.ID)
	s.Assert().NotNil(role.CreatedAt)
}

func (s *RoleServiceIntegrationTestSuite) TestGet() {
	created, err := s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, service.CreateRoleOpts{
		Name: "get-role", Description: "get role description",
	})
	s.Require().NoError(err)

	role, err := s.roleService.Get(s.ctx, created.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, role.ID)
	s.Assert().Equal(created.Name, role.Name)
}

func (s *RoleServiceIntegrationTestSuite) TestListBelongsTo() {
	_, err := s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, service.CreateRoleOpts{
		Name: "role-one", Description: "role one description",
	})
	s.Require().NoError(err)
	_, err = s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, service.CreateRoleOpts{
		Name: "role-two", Description: "role two description",
	})
	s.Require().NoError(err)

	roles, err := s.roleService.ListBelongsTo(s.ctx, s.organization.ID, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Assert().GreaterOrEqual(len(roles.Items), 2)
}

func (s *RoleServiceIntegrationTestSuite) TestUpdate() {
	created, err := s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, service.CreateRoleOpts{
		Name: "upd-role", Description: "upd role description",
	})
	s.Require().NoError(err)

	role, err := s.roleService.Update(s.ctx, created.ID, s.organization.ID, service.UpdateRoleOpts{
		Name: optional.Some("updated-role"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("updated-role", role.Name)
	s.Assert().NotNil(role.UpdatedAt)
}

func (s *RoleServiceIntegrationTestSuite) TestAddMember() {
	created, err := s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, service.CreateRoleOpts{
		Name: "member-role", Description: "member role description",
	})
	s.Require().NoError(err)

	member, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), s.organization.ID, member.ID))

	s.Require().NoError(s.roleService.AddMember(s.ctx, created.ID, member.ID, s.organization.ID))

	members, err := s.roleService.ListMembers(s.ctx, created.ID, s.organization.ID, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Assert().GreaterOrEqual(len(members.Items), 1)
}

func (s *RoleServiceIntegrationTestSuite) TestRemoveMember() {
	created, err := s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, service.CreateRoleOpts{
		Name: "rm-role", Description: "rm role description",
	})
	s.Require().NoError(err)

	member, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), s.organization.ID, member.ID))
	s.Require().NoError(s.roleService.AddMember(s.ctx, created.ID, member.ID, s.organization.ID))
	s.Require().NoError(s.roleService.RemoveMember(s.ctx, created.ID, member.ID, s.organization.ID))
}

func (s *RoleServiceIntegrationTestSuite) TestDelete() {
	created, err := s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, service.CreateRoleOpts{
		Name: "del-role", Description: "del role description",
	})
	s.Require().NoError(err)
	s.Require().NoError(s.roleService.Delete(s.ctx, created.ID, s.organization.ID))
}

func (s *RoleServiceIntegrationTestSuite) TestAddPermission() {
	s.T().Skip("role permissions are granted via CreateRoleOpts.Actions")
}

func TestRoleServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(RoleServiceIntegrationTestSuite))
}
