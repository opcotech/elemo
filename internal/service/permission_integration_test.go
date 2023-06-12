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

type PermissionServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	organizationService service.OrganizationService
	permissionService   service.PermissionService

	owner        *model.User
	guest        *model.User
	organization *model.Organization
	permission   *model.Permission

	ctx context.Context
}

func (s *PermissionServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)

	var err error
	s.permissionService, err = service.NewPermissionService(s.PermissionRepo)
	s.Require().NoError(err)

	licenseService, err := service.NewLicenseService(
		testutil.ParseLicense(s.T()),
		s.LicenseRepo,
		service.WithPermissionService(s.permissionService),
	)
	s.Require().NoError(err)

	s.organizationService, err = service.NewOrganizationService(
		service.WithUserRepository(s.UserRepo),
		service.WithOrganizationRepository(s.OrganizationRepo),
		service.WithPermissionService(s.permissionService),
		service.WithLicenseService(licenseService),
	)
	s.Require().NoError(err)
}

func (s *PermissionServiceIntegrationTestSuite) SetupTest() {
	s.owner = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.owner))

	s.guest = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.guest))

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.owner.ID, s.Neo4jDB))

	s.organization = testModel.NewOrganization()
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	s.permission = testModel.NewPermission(s.guest.ID, s.organization.ID, model.PermissionKindRead)
}

func (s *PermissionServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *PermissionServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *PermissionServiceIntegrationTestSuite) TestCreate() {
	err := s.permissionService.Create(s.ctx, s.permission)
	s.Require().NoError(err)
	s.Require().NotEmpty(s.permission.ID)
	s.Assert().NotNil(s.permission.CreatedAt)
	s.Assert().Nil(s.permission.UpdatedAt)
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserCreate() {
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, s.guest.ID)
	err := s.permissionService.CtxUserCreate(ctx, s.permission)
	s.Require().ErrorIs(err, service.ErrNoPermission)

	ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	err = s.permissionService.CtxUserCreate(ctx, s.permission)
	s.Require().NoError(err)
}

func (s *PermissionServiceIntegrationTestSuite) TestGetBySubject() {
	s.Require().NoError(s.permissionService.Create(s.ctx, s.permission))

	permissions, err := s.permissionService.GetBySubject(s.ctx, s.guest.ID)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 1)
	s.Assert().Equal(s.permission.ID, permissions[0].ID)
}

func (s *PermissionServiceIntegrationTestSuite) TestGetByTarget() {
	s.Require().NoError(s.permissionService.Create(s.ctx, s.permission))

	permissions, err := s.permissionService.GetByTarget(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 2) // +1 for organization owner permission

	userIDs := make([]model.ID, 0, len(permissions))
	for _, permission := range permissions {
		userIDs = append(userIDs, permission.Subject)
	}

	s.Assert().ElementsMatch([]model.ID{s.owner.ID, s.guest.ID}, userIDs)
}

func (s *PermissionServiceIntegrationTestSuite) TestGetBySubjectAndTarget() {
	s.Require().NoError(s.permissionService.Create(s.ctx, s.permission))

	// Create an organization for the guest user
	s.Require().NoError(s.organizationService.Create(s.ctx, s.guest.ID, testModel.NewOrganization()))

	// Check for specific subject and target permissions
	permissions, err := s.permissionService.GetBySubjectAndTarget(s.ctx, s.guest.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(permissions, 1)
	s.Assert().Equal(s.permission.ID, permissions[0].ID)
}

func (s *PermissionServiceIntegrationTestSuite) TestHasAnyRelation() {
	hasRelation, err := s.permissionService.HasAnyRelation(s.ctx, s.guest.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().False(hasRelation)

	s.Require().NoError(s.permissionService.Create(s.ctx, s.permission))

	hasRelation, err = s.permissionService.HasAnyRelation(s.ctx, s.guest.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().True(hasRelation)
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserHasAnyRelation() {
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, s.guest.ID)

	hasRelation := s.permissionService.CtxUserHasAnyRelation(ctx, s.organization.ID)
	s.Assert().False(hasRelation)

	s.Require().NoError(s.permissionService.Create(ctx, s.permission))

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
	s.Require().NoError(s.permissionService.Create(s.ctx, s.permission))

	tests := []struct {
		userID  model.ID
		kind    model.PermissionKind
		want    bool
		wantErr bool
	}{
		{s.guest.ID, model.PermissionKindAll, false, false},
		{s.guest.ID, model.PermissionKindCreate, false, false},
		{s.guest.ID, model.PermissionKindRead, true, false},
		{s.guest.ID, model.PermissionKindWrite, false, false},
		{s.guest.ID, model.PermissionKindDelete, false, false},
		{s.owner.ID, model.PermissionKindAll, true, false},
		{s.owner.ID, model.PermissionKindCreate, true, false},
		{s.owner.ID, model.PermissionKindRead, true, false},
		{s.owner.ID, model.PermissionKindWrite, true, false},
		{s.owner.ID, model.PermissionKindDelete, true, false},
	}

	for _, tt := range tests {
		hasPermission, err := s.permissionService.HasPermission(s.ctx, tt.userID, s.organization.ID, tt.kind)
		s.Assert().NoError(err)
		s.Assert().Equal(tt.want, hasPermission)
	}
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserHasPermission() {
	s.Require().NoError(s.permissionService.Create(s.ctx, s.permission))

	tests := []struct {
		userID  model.ID
		kind    model.PermissionKind
		want    bool
		wantErr bool
	}{
		{s.guest.ID, model.PermissionKindAll, false, false},
		{s.guest.ID, model.PermissionKindCreate, false, false},
		{s.guest.ID, model.PermissionKindRead, true, false},
		{s.guest.ID, model.PermissionKindWrite, false, false},
		{s.guest.ID, model.PermissionKindDelete, false, false},
		{s.owner.ID, model.PermissionKindAll, true, false},
		{s.owner.ID, model.PermissionKindCreate, true, false},
		{s.owner.ID, model.PermissionKindRead, true, false},
		{s.owner.ID, model.PermissionKindWrite, true, false},
		{s.owner.ID, model.PermissionKindDelete, true, false},
	}

	for _, tt := range tests {
		ctx := context.WithValue(s.ctx, pkg.CtxKeyUserID, tt.userID)
		hasPermission := s.permissionService.CtxUserHasPermission(ctx, s.organization.ID, tt.kind)
		s.Assert().Equal(tt.want, hasPermission)
	}
}

func (s *PermissionServiceIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.permissionService.Create(s.ctx, s.permission))

	updated, err := s.permissionService.Update(s.ctx, s.permission.ID, model.PermissionKindWrite)
	s.Require().NoError(err)
	s.Require().Equal(model.PermissionKindWrite, updated.Kind)
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserUpdate() {
	s.Require().NoError(s.permissionService.Create(s.ctx, s.permission))

	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, s.guest.ID)
	_, err := s.permissionService.CtxUserUpdate(ctx, s.permission.ID, model.PermissionKindWrite)
	s.Require().ErrorIs(err, service.ErrNoPermission)

	ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	updated, err := s.permissionService.CtxUserUpdate(ctx, s.permission.ID, model.PermissionKindWrite)
	s.Require().NoError(err)
	s.Require().Equal(model.PermissionKindWrite, updated.Kind)
}

func (s *PermissionServiceIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.permissionService.Create(s.ctx, s.permission))

	_, err := s.permissionService.Get(s.ctx, s.permission.ID)
	s.Require().NoError(err)

	err = s.permissionService.Delete(s.ctx, s.permission.ID)
	s.Require().NoError(err)

	_, err = s.permissionService.Get(s.ctx, s.permission.ID)
	s.Require().ErrorIs(err, repository.ErrNotFound)
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserDelete() {
	s.Require().NoError(s.permissionService.Create(s.ctx, s.permission))

	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, s.guest.ID)
	err := s.permissionService.CtxUserDelete(ctx, s.permission.ID)
	s.Require().ErrorIs(err, service.ErrNoPermission)

	ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	err = s.permissionService.CtxUserDelete(ctx, s.permission.ID)
	s.Require().NoError(err)
}

func TestPermissionServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionServiceIntegrationTestSuite))
}
