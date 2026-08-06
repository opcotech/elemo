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

type PermissionServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	organizationService service.OrganizationService
	permissionService   service.PermissionService
	emailService        service.EmailService

	owner        *repository.User
	guest        *repository.User
	organization *service.Organization
	permission   *service.Permission
	createOpts   service.CreatePermissionOpts

	ctx context.Context
}

func (s *PermissionServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)
	s.SetupPg(&s.ContainerIntegrationTestSuite, container)

	var err error
	s.permissionService, err = service.NewPermissionService(s.PermissionRepo)
	s.Require().NoError(err)

	licenseService, err := service.NewLicenseService(
		testutil.ParseLicense(s.T()),
		s.LicenseRepo,
		service.WithPermissionService(s.permissionService),
	)
	s.Require().NoError(err)

	ctrl := gomock.NewController(s.T())
	emailSender := mock.NewEmailSender(ctrl)

	smtpConf := &config.SMTPConfig{
		ClientURL:      "http://localhost:3000",
		SupportAddress: "support@example.com",
	}
	s.emailService, err = service.NewEmailService(emailSender, "templates", smtpConf)
	s.Require().NoError(err)

	s.organizationService, err = service.NewOrganizationService(
		service.WithUserRepository(s.UserRepo),
		service.WithOrganizationRepository(s.OrganizationRepo),
		service.WithPermissionService(s.permissionService),
		service.WithLicenseService(licenseService),
		service.WithUserTokenRepository(s.UserTokenRepository),
		service.WithEmailService(s.emailService),
	)
	s.Require().NoError(err)
}

func (s *PermissionServiceIntegrationTestSuite) SetupTest() {
	var err error
	s.owner, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.guest, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.owner.ID, s.Neo4jDB))

	repoOrgOpts := testModel.NewCreateOrganizationOpts(s.owner.ID)
	s.organization, err = s.organizationService.Create(s.ctx, s.owner.ID, service.CreateOrganizationOpts{
		Name:    repoOrgOpts.Name,
		Email:   repoOrgOpts.Email,
		Logo:    repoOrgOpts.Logo,
		Website: repoOrgOpts.Website,
		Status:  repoOrgOpts.Status,
	})
	s.Require().NoError(err)

	s.createOpts = service.CreatePermissionOpts{
		Subject: s.guest.ID,
		Target:  s.organization.ID,
		Kind:    model.PermissionKindRead,
	}
	s.permission = nil
}

func (s *PermissionServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *PermissionServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *PermissionServiceIntegrationTestSuite) TestCreate() {
	permission, err := s.permissionService.Create(s.ctx, s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	s.Require().NotEmpty(permission.ID)
	s.Assert().NotNil(permission.CreatedAt)
	s.Assert().Nil(permission.UpdatedAt)
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserCreate() {
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, s.guest.ID)
	_, err := s.permissionService.CtxUserCreate(ctx, s.createOpts)
	s.Require().ErrorIs(err, service.ErrNoPermission)

	ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	permission, err := s.permissionService.CtxUserCreate(ctx, s.createOpts)
	s.Require().NoError(err)
	s.permission = permission
}

func (s *PermissionServiceIntegrationTestSuite) TestGetBySubject() {
	permission, err := s.permissionService.Create(s.ctx, s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	permissions, err := s.permissionService.GetBySubject(s.ctx, s.guest.ID)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 1)
	s.Assert().Equal(permission.ID, permissions[0].ID)
}

func (s *PermissionServiceIntegrationTestSuite) TestGetByTarget() {
	permission, err := s.permissionService.Create(s.ctx, s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	permissions, err := s.permissionService.GetByTarget(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 2) // +1 for organization owner permission

	userIDs := make([]model.ID, 0, len(permissions))
	for _, p := range permissions {
		userIDs = append(userIDs, p.Subject)
	}

	s.Assert().ElementsMatch([]model.ID{s.owner.ID, s.guest.ID}, userIDs)
}

func (s *PermissionServiceIntegrationTestSuite) TestGetBySubjectAndTarget() {
	permission, err := s.permissionService.Create(s.ctx, s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	repoOrgOpts := testModel.NewCreateOrganizationOpts(s.guest.ID)
	_, err = s.organizationService.Create(s.ctx, s.guest.ID, service.CreateOrganizationOpts{
		Name:    repoOrgOpts.Name,
		Email:   repoOrgOpts.Email,
		Logo:    repoOrgOpts.Logo,
		Website: repoOrgOpts.Website,
		Status:  repoOrgOpts.Status,
	})
	s.Require().NoError(err)

	permissions, err := s.permissionService.GetBySubjectAndTarget(s.ctx, s.guest.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 1)
	s.Assert().Equal(permission.ID, permissions[0].ID)
}

func (s *PermissionServiceIntegrationTestSuite) TestHasAnyRelation() {
	hasRelation, err := s.permissionService.HasAnyRelation(s.ctx, s.guest.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().False(hasRelation)

	permission, err := s.permissionService.Create(s.ctx, s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	hasRelation, err = s.permissionService.HasAnyRelation(s.ctx, s.guest.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().True(hasRelation)
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserHasAnyRelation() {
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, s.guest.ID)

	hasRelation := s.permissionService.CtxUserHasAnyRelation(ctx, s.organization.ID)
	s.Assert().False(hasRelation)

	permission, err := s.permissionService.Create(ctx, s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	hasRelation = s.permissionService.CtxUserHasAnyRelation(ctx, s.organization.ID)
	s.Assert().True(hasRelation)
}

func (s *PermissionServiceIntegrationTestSuite) TestHasSystemRole() {
	hasSystemRole, err := s.permissionService.HasSystemRole(s.ctx, s.guest.ID, model.SystemRoleOwner)
	s.Require().NoError(err)
	s.Assert().False(hasSystemRole)

	s.Require().NoError(testRepo.MakeUserSystemOwner(s.guest.ID, s.Neo4jDB))

	hasSystemRole, err = s.permissionService.HasSystemRole(s.ctx, s.guest.ID, model.SystemRoleOwner)
	s.Require().NoError(err)
	s.Assert().True(hasSystemRole)
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserHasSystemRole() {
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, s.guest.ID)

	hasSystemRole := s.permissionService.CtxUserHasSystemRole(ctx, model.SystemRoleOwner)
	s.Assert().False(hasSystemRole)

	s.Require().NoError(testRepo.MakeUserSystemOwner(s.guest.ID, s.Neo4jDB))

	hasSystemRole = s.permissionService.CtxUserHasSystemRole(ctx, model.SystemRoleOwner)
	s.Assert().True(hasSystemRole)
}

func (s *PermissionServiceIntegrationTestSuite) TestHasPermission() {
	permission, err := s.permissionService.Create(s.ctx, s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	tests := []struct {
		userID model.ID
		kind   model.PermissionKind
		want   bool
	}{
		{s.guest.ID, model.PermissionKindAll, false},
		{s.guest.ID, model.PermissionKindCreate, false},
		{s.guest.ID, model.PermissionKindRead, true},
		{s.guest.ID, model.PermissionKindWrite, false},
		{s.guest.ID, model.PermissionKindDelete, false},
		{s.owner.ID, model.PermissionKindAll, true},
		{s.owner.ID, model.PermissionKindCreate, true},
		{s.owner.ID, model.PermissionKindRead, true},
		{s.owner.ID, model.PermissionKindWrite, true},
		{s.owner.ID, model.PermissionKindDelete, true},
	}

	for _, tt := range tests {
		hasPermission, err := s.permissionService.HasPermission(s.ctx, tt.userID, s.organization.ID, tt.kind)
		s.Assert().NoError(err)
		s.Assert().Equal(tt.want, hasPermission)
	}
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserHasPermission() {
	permission, err := s.permissionService.Create(s.ctx, s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	tests := []struct {
		userID model.ID
		kind   model.PermissionKind
		want   bool
	}{
		{s.guest.ID, model.PermissionKindAll, false},
		{s.guest.ID, model.PermissionKindCreate, false},
		{s.guest.ID, model.PermissionKindRead, true},
		{s.guest.ID, model.PermissionKindWrite, false},
		{s.guest.ID, model.PermissionKindDelete, false},
		{s.owner.ID, model.PermissionKindAll, true},
		{s.owner.ID, model.PermissionKindCreate, true},
		{s.owner.ID, model.PermissionKindRead, true},
		{s.owner.ID, model.PermissionKindWrite, true},
		{s.owner.ID, model.PermissionKindDelete, true},
	}

	for _, tt := range tests {
		ctx := context.WithValue(s.ctx, pkg.CtxKeyUserID, tt.userID)
		hasPermission := s.permissionService.CtxUserHasPermission(ctx, s.organization.ID, tt.kind)
		s.Assert().Equal(tt.want, hasPermission)
	}
}

func (s *PermissionServiceIntegrationTestSuite) TestUpdate() {
	permission, err := s.permissionService.Create(s.ctx, s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	updated, err := s.permissionService.Update(s.ctx, permission.ID, model.PermissionKindWrite)
	s.Require().NoError(err)
	s.Require().Equal(model.PermissionKindWrite, updated.Kind)
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserUpdate() {
	permission, err := s.permissionService.Create(s.ctx, s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, s.guest.ID)
	_, err = s.permissionService.CtxUserUpdate(ctx, permission.ID, model.PermissionKindWrite)
	s.Require().ErrorIs(err, service.ErrNoPermission)

	ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	updated, err := s.permissionService.CtxUserUpdate(ctx, permission.ID, model.PermissionKindWrite)
	s.Require().NoError(err)
	s.Require().Equal(model.PermissionKindWrite, updated.Kind)
}

func (s *PermissionServiceIntegrationTestSuite) TestDelete() {
	permission, err := s.permissionService.Create(s.ctx, s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	_, err = s.permissionService.Get(s.ctx, permission.ID)
	s.Require().NoError(err)

	err = s.permissionService.Delete(s.ctx, permission.ID)
	s.Require().NoError(err)

	_, err = s.permissionService.Get(s.ctx, permission.ID)
	s.Require().ErrorIs(err, repository.ErrNotFound)
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserDelete() {
	permission, err := s.permissionService.Create(s.ctx, s.createOpts)
	s.Require().NoError(err)
	s.permission = permission

	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, s.guest.ID)
	err = s.permissionService.CtxUserDelete(ctx, permission.ID)
	s.Require().ErrorIs(err, service.ErrNoPermission)

	ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	err = s.permissionService.CtxUserDelete(ctx, permission.ID)
	s.Require().NoError(err)
}

func TestPermissionServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionServiceIntegrationTestSuite))
}
