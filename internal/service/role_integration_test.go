package service_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

type RoleServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	roleService         service.RoleService
	organizationService service.OrganizationService
	emailService        service.EmailService

	owner        *model.User
	role         *model.Role
	organization *model.Organization

	ctx context.Context
}

func (s *RoleServiceIntegrationTestSuite) SetupSuite() {
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

	s.roleService, err = service.NewRoleService(
		service.WithRoleRepository(s.RoleRepo),
		service.WithUserRepository(s.UserRepo),
		service.WithPermissionService(permissionService),
		service.WithLicenseService(licenseService),
	)
	s.Require().NoError(err)

	// Create a mock email sender for integration tests
	ctrl := gomock.NewController(s.T())
	emailSender := mock.NewEmailSender(ctrl)

	// Create a real EmailService with mock sender
	smtpConf := &config.SMTPConfig{
		ClientURL:      "http://localhost:3000",
		SupportAddress: "support@example.com",
	}
	s.emailService, err = service.NewEmailService(emailSender, "templates", smtpConf)
	s.Require().NoError(err)

	s.organizationService, err = service.NewOrganizationService(
		service.WithUserRepository(s.UserRepo),
		service.WithOrganizationRepository(s.OrganizationRepo),
		service.WithPermissionService(permissionService),
		service.WithLicenseService(licenseService),
		service.WithUserTokenRepository(s.UserTokenRepository),
		service.WithEmailService(s.emailService),
	)
	s.Require().NoError(err)
}

func (s *RoleServiceIntegrationTestSuite) SetupTest() {
	s.owner = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.owner))

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.owner.ID, s.Neo4jDB))

	s.organization = testModel.NewOrganization()
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	s.role = testModel.NewRole()
}

func (s *RoleServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *RoleServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *RoleServiceIntegrationTestSuite) TestCreate() {
	err := s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role)
	s.Require().NoError(err)
	s.Require().NotEmpty(s.role.ID)
	s.Assert().NotNil(s.role.CreatedAt)
	s.Assert().Nil(s.role.UpdatedAt)
}

func (s *RoleServiceIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	role, err := s.roleService.Get(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Equal(s.role.ID, role.ID)
	s.Assert().Equal(s.role.Name, role.Name)
	s.Assert().Equal(s.role.Description, role.Description)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, role.Members)
	s.Assert().ElementsMatch(s.role.Permissions, role.Permissions)
	s.Assert().NotNil(s.role.CreatedAt)
	s.Assert().Nil(s.role.UpdatedAt)
}

func (s *RoleServiceIntegrationTestSuite) TestGetAllBelongsTo() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	roles, err := s.roleService.GetAllBelongsTo(s.ctx, s.organization.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(roles, 3)

	roles, err = s.roleService.GetAllBelongsTo(s.ctx, s.organization.ID, 0, 2)
	s.Require().NoError(err)
	s.Assert().Len(roles, 2)

	roles, err = s.roleService.GetAllBelongsTo(s.ctx, s.organization.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(roles, 2)

	roles, err = s.roleService.GetAllBelongsTo(s.ctx, s.organization.ID, 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(roles, 1)

	roles, err = s.roleService.GetAllBelongsTo(s.ctx, s.organization.ID, 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(roles, 0)
}

func (s *RoleServiceIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	patch := map[string]any{
		"name":        "new name",
		"description": "new description",
	}

	updated, err := s.roleService.Update(s.ctx, s.role.ID, s.organization.ID, patch)
	s.Require().NoError(err)
	s.Assert().Equal(s.role.ID, updated.ID)
	s.Assert().Equal(patch["name"], updated.Name)
	s.Assert().Equal(patch["description"], updated.Description)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, updated.Members)
	s.Assert().ElementsMatch(s.role.Permissions, updated.Permissions)
	s.Assert().NotNil(updated.CreatedAt)
	s.Assert().NotNil(updated.UpdatedAt)
}

func (s *RoleServiceIntegrationTestSuite) TestGetMembers() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	user := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), user))

	members, err := s.roleService.GetMembers(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)

	memberIDs := make([]model.ID, len(members))
	for i, member := range members {
		memberIDs[i] = member.ID
	}
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, memberIDs)

	err = s.roleService.AddMember(s.ctx, s.role.ID, user.ID, s.organization.ID)
	s.Require().NoError(err)

	members, err = s.roleService.GetMembers(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	memberIDs = make([]model.ID, len(members))
	for i, member := range members {
		memberIDs[i] = member.ID
	}
	s.Assert().ElementsMatch([]model.ID{s.owner.ID, user.ID}, memberIDs)
}

func (s *RoleServiceIntegrationTestSuite) TestAddMember() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	user := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), user))

	err := s.roleService.AddMember(s.ctx, s.role.ID, user.ID, s.organization.ID)
	s.Require().NoError(err)

	role, err := s.roleService.Get(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID, user.ID}, role.Members)
}

func (s *RoleServiceIntegrationTestSuite) TestRemoveMember() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	user := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), user))

	err := s.roleService.AddMember(s.ctx, s.role.ID, user.ID, s.organization.ID)
	s.Require().NoError(err)

	err = s.roleService.RemoveMember(s.ctx, s.role.ID, user.ID, s.organization.ID)
	s.Require().NoError(err)

	role, err := s.roleService.Get(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, role.Members)
}

func (s *RoleServiceIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	err := s.roleService.Delete(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)

	_, err = s.roleService.Get(s.ctx, s.role.ID, s.organization.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func (s *RoleServiceIntegrationTestSuite) TestAddPermission() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	// Create a document to use as target
	document := testModel.NewDocument(s.owner.ID)
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.organization.ID, document))

	// Add permission to role
	err := s.roleService.AddPermission(s.ctx, s.role.ID, s.organization.ID, document.ID, model.PermissionKindRead)
	s.Require().NoError(err)

	// Verify permission was added
	permissions, err := s.roleService.GetPermissions(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 1)
	s.Assert().Equal(s.role.ID, permissions[0].Subject)
	s.Assert().Equal(document.ID, permissions[0].Target)
	s.Assert().Equal(model.PermissionKindRead, permissions[0].Kind)
}

func (s *RoleServiceIntegrationTestSuite) TestAddPermissionMultipleKinds() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	// Create a document to use as target
	document := testModel.NewDocument(s.owner.ID)
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.organization.ID, document))

	// Add multiple permissions with different kinds
	err := s.roleService.AddPermission(s.ctx, s.role.ID, s.organization.ID, document.ID, model.PermissionKindRead)
	s.Require().NoError(err)

	err = s.roleService.AddPermission(s.ctx, s.role.ID, s.organization.ID, document.ID, model.PermissionKindWrite)
	s.Require().NoError(err)

	err = s.roleService.AddPermission(s.ctx, s.role.ID, s.organization.ID, document.ID, model.PermissionKindDelete)
	s.Require().NoError(err)

	// Verify all permissions were added
	permissions, err := s.roleService.GetPermissions(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 3)

	kinds := make([]model.PermissionKind, len(permissions))
	for i, p := range permissions {
		kinds[i] = p.Kind
	}
	s.Assert().Contains(kinds, model.PermissionKindRead)
	s.Assert().Contains(kinds, model.PermissionKindWrite)
	s.Assert().Contains(kinds, model.PermissionKindDelete)
}

func (s *RoleServiceIntegrationTestSuite) TestRemovePermission() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	// Create a document to use as target
	document := testModel.NewDocument(s.owner.ID)
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.organization.ID, document))

	// Add permission to role
	err := s.roleService.AddPermission(s.ctx, s.role.ID, s.organization.ID, document.ID, model.PermissionKindRead)
	s.Require().NoError(err)

	// Get the permission ID
	permissions, err := s.roleService.GetPermissions(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 1)
	permissionID := permissions[0].ID

	// Remove permission
	err = s.roleService.RemovePermission(s.ctx, s.role.ID, s.organization.ID, permissionID)
	s.Require().NoError(err)

	// Verify permission was removed
	permissions, err = s.roleService.GetPermissions(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 0)
}

func (s *RoleServiceIntegrationTestSuite) TestGetPermissions() {
	s.Require().NoError(s.roleService.Create(s.ctx, s.owner.ID, s.organization.ID, s.role))

	// Initially no permissions
	permissions, err := s.roleService.GetPermissions(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 0)

	// Create multiple documents
	doc1 := testModel.NewDocument(s.owner.ID)
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.organization.ID, doc1))
	doc2 := testModel.NewDocument(s.owner.ID)
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.organization.ID, doc2))

	// Add permissions on different targets
	err = s.roleService.AddPermission(s.ctx, s.role.ID, s.organization.ID, doc1.ID, model.PermissionKindRead)
	s.Require().NoError(err)

	err = s.roleService.AddPermission(s.ctx, s.role.ID, s.organization.ID, doc2.ID, model.PermissionKindWrite)
	s.Require().NoError(err)

	// Verify all permissions are returned
	permissions, err = s.roleService.GetPermissions(s.ctx, s.role.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 2)

	// Verify permissions have correct subjects and targets
	for _, p := range permissions {
		s.Assert().Equal(s.role.ID, p.Subject)
		s.Assert().True(p.Target == doc1.ID || p.Target == doc2.ID)
	}
}

func TestRoleServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(RoleServiceIntegrationTestSuite))
}
