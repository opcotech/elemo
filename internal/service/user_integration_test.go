package service_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

type UserServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	userService service.UserService

	normalUser        *model.User
	normalUserContext context.Context

	systemOwner        *model.User
	systemOwnerContext context.Context
}

func (s *UserServiceIntegrationTestSuite) SetupSuite() {
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

	s.userService, err = service.NewUserService(
		service.WithUserRepository(s.UserRepo),
		service.WithUserTokenRepository(s.UserTokenRepository),
		service.WithPermissionService(permissionService),
		service.WithLicenseService(licenseService),
	)
	s.Require().NoError(err)
}

func (s *UserServiceIntegrationTestSuite) SetupTest() {
	s.systemOwner = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.systemOwner))
	s.systemOwnerContext = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.systemOwner.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.systemOwner.ID, s.Neo4jDB))

	s.normalUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.normalUser))
	s.normalUserContext = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.normalUser.ID)
}

func (s *UserServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *UserServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *UserServiceIntegrationTestSuite) TestCreateUser() {
	err := s.userService.Create(s.normalUserContext, testModel.NewUser())
	s.Assert().ErrorIs(err, service.ErrNoPermission)

	err = s.userService.Create(s.systemOwnerContext, testModel.NewUser())
	s.Assert().NoError(err)
}

func (s *UserServiceIntegrationTestSuite) TestGet() {
	got, err := s.userService.Get(s.normalUserContext, s.systemOwner.ID)
	s.Assert().NoError(err)

	s.Assert().Equal(s.systemOwner.Username, got.Username)
	s.Assert().Equal(s.systemOwner.Email, got.Email)
	s.Assert().Equal(s.systemOwner.Password, got.Password)
	s.Assert().Equal(s.systemOwner.Status, got.Status)
	s.Assert().Equal(s.systemOwner.FirstName, got.FirstName)
	s.Assert().Equal(s.systemOwner.LastName, got.LastName)
	s.Assert().Equal(s.systemOwner.Picture, got.Picture)
	s.Assert().Equal(s.systemOwner.Title, got.Title)
	s.Assert().Equal(s.systemOwner.Bio, got.Bio)
	s.Assert().Equal(s.systemOwner.Phone, got.Phone)
	s.Assert().Equal(s.systemOwner.Address, got.Address)
	s.Assert().Equal(s.systemOwner.Links, got.Links)
	s.Assert().Equal(s.systemOwner.Languages, got.Languages)
	s.Assert().Equal(s.systemOwner.Documents, got.Documents)
	s.Assert().Equal(s.systemOwner.Permissions, got.Permissions)
	s.Assert().WithinDuration(*s.systemOwner.CreatedAt, *got.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(got.UpdatedAt)
}

func (s *UserServiceIntegrationTestSuite) TestGetByEmail() {
	got, err := s.userService.GetByEmail(s.normalUserContext, s.systemOwner.Email)
	s.Assert().NoError(err)

	s.Assert().Equal(s.systemOwner.Username, got.Username)
	s.Assert().Equal(s.systemOwner.Email, got.Email)
	s.Assert().Equal(s.systemOwner.Password, got.Password)
	s.Assert().Equal(s.systemOwner.Status, got.Status)
	s.Assert().Equal(s.systemOwner.FirstName, got.FirstName)
	s.Assert().Equal(s.systemOwner.LastName, got.LastName)
	s.Assert().Equal(s.systemOwner.Picture, got.Picture)
	s.Assert().Equal(s.systemOwner.Title, got.Title)
	s.Assert().Equal(s.systemOwner.Bio, got.Bio)
	s.Assert().Equal(s.systemOwner.Phone, got.Phone)
	s.Assert().Equal(s.systemOwner.Address, got.Address)
	s.Assert().Equal(s.systemOwner.Links, got.Links)
	s.Assert().Equal(s.systemOwner.Languages, got.Languages)
	s.Assert().Equal(s.systemOwner.Documents, got.Documents)
	s.Assert().Equal(s.systemOwner.Permissions, got.Permissions)
	s.Assert().WithinDuration(*s.systemOwner.CreatedAt, *got.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(got.UpdatedAt)
}

func (s *UserServiceIntegrationTestSuite) TestGetAll() {
	users, err := s.userService.GetAll(s.normalUserContext, 0, 10)
	s.Assert().NoError(err)
	s.Assert().Len(users, 2)

	users, err = s.userService.GetAll(s.normalUserContext, 0, 1)
	s.Assert().NoError(err)
	s.Assert().Len(users, 1)

	users, err = s.userService.GetAll(s.normalUserContext, 1, 1)
	s.Assert().NoError(err)
	s.Assert().Len(users, 1)

	users, err = s.userService.GetAll(s.normalUserContext, 2, 1)
	s.Assert().NoError(err)
	s.Assert().Len(users, 0)
}

func (s *UserServiceIntegrationTestSuite) TestUpdate() {
	targetUser := testModel.NewUser()
	s.Assert().NoError(s.userService.Create(s.systemOwnerContext, targetUser))

	patch := map[string]any{
		"username": "new_username",
	}

	_, err := s.userService.Update(s.normalUserContext, targetUser.ID, patch)
	s.Assert().ErrorIs(err, service.ErrNoPermission)

	got, err := s.userService.Update(s.systemOwnerContext, targetUser.ID, patch)
	s.Assert().NoError(err)

	s.Assert().Equal(patch["username"], got.Username)
	s.Assert().Equal(targetUser.Email, got.Email)
	s.Assert().Equal(targetUser.Password, got.Password)
	s.Assert().Equal(targetUser.Status, got.Status)
	s.Assert().Equal(targetUser.FirstName, got.FirstName)
	s.Assert().Equal(targetUser.LastName, got.LastName)
	s.Assert().Equal(targetUser.Picture, got.Picture)
	s.Assert().Equal(targetUser.Title, got.Title)
	s.Assert().Equal(targetUser.Bio, got.Bio)
	s.Assert().Equal(targetUser.Phone, got.Phone)
	s.Assert().Equal(targetUser.Address, got.Address)
	s.Assert().Equal(targetUser.Links, got.Links)
	s.Assert().Equal(targetUser.Languages, got.Languages)
	s.Assert().Equal(targetUser.Documents, got.Documents)
	s.Assert().Equal(targetUser.Permissions, got.Permissions)
	s.Assert().WithinDuration(*targetUser.CreatedAt, *got.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(got.UpdatedAt)
}

func (s *UserServiceIntegrationTestSuite) TestDelete() {
	targetUser := testModel.NewUser()
	s.Assert().NoError(s.userService.Create(s.systemOwnerContext, targetUser))

	// Only users with the necessary permissions can delete users
	err := s.userService.Delete(s.normalUserContext, targetUser.ID, false)
	s.Assert().ErrorIs(err, service.ErrNoPermission)

	// Nobody can delete itself
	err = s.userService.Delete(s.systemOwnerContext, s.systemOwner.ID, false)
	s.Assert().ErrorIs(err, service.ErrNoPermission)

	// System owner can delete any user
	err = s.userService.Delete(s.systemOwnerContext, targetUser.ID, false)
	s.Assert().NoError(err)

	// Soft delete does not delete the user permanently, but marks it as deleted
	// and updates its password to be unusable
	got, err := s.userService.Get(s.systemOwnerContext, targetUser.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(targetUser.Email, got.Email)
	s.Assert().Equal(password.UnusablePassword, got.Password)
	s.Assert().Equal(model.UserStatusDeleted, got.Status)
	s.Assert().NotNil(got.UpdatedAt)

	// Force delete deletes the user permanently
	err = s.userService.Delete(s.systemOwnerContext, targetUser.ID, true)
	s.Assert().NoError(err)

	// User is deleted
	_, err = s.userService.Get(s.systemOwnerContext, targetUser.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestUserServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceIntegrationTestSuite))
}
