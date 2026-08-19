//go:build integration

package service_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

func serviceCreateUserOpts() service.CreateUserOpts {
	o := testModel.NewCreateUserOpts()
	return service.CreateUserOpts{
		Username:  o.Username,
		Email:     o.Email,
		Password:  o.Password,
		Status:    o.Status,
		FirstName: o.FirstName,
		LastName:  o.LastName,
		Picture:   o.Picture,
		Title:     o.Title,
		Bio:       o.Bio,
		Phone:     o.Phone,
		Address:   o.Address,
		Links:     o.Links,
		Languages: o.Languages,
	}
}

type UserServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	userService service.UserService

	normalUser        *repository.User
	normalUserContext context.Context

	systemOwner        *repository.User
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
	var err error
	s.systemOwner, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.systemOwnerContext = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.systemOwner.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.systemOwner.ID, s.Neo4jDB))

	s.normalUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.normalUserContext = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.normalUser.ID)
}

func (s *UserServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *UserServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *UserServiceIntegrationTestSuite) TestCreateUser() {
	_, err := s.userService.Create(s.normalUserContext, serviceCreateUserOpts())
	s.Assert().ErrorIs(err, service.ErrNoPermission)

	_, err = s.userService.Create(s.systemOwnerContext, serviceCreateUserOpts())
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
	s.Assert().WithinDuration(*s.systemOwner.CreatedAt, *got.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(got.UpdatedAt)
}

func (s *UserServiceIntegrationTestSuite) TestGetByEmail() {
	got, err := s.userService.GetByEmail(s.normalUserContext, s.systemOwner.Email)
	s.Assert().NoError(err)
	s.Assert().Equal(s.systemOwner.Email, got.Email)
	s.Assert().Equal(s.systemOwner.Username, got.Username)
}

func (s *UserServiceIntegrationTestSuite) TestGetAll() {
	users, err := s.userService.List(s.normalUserContext, service.CursorPage{Size: 10})
	s.Assert().NoError(err)
	s.Assert().Len(users.Items, 2)

	users, err = s.userService.List(s.normalUserContext, service.CursorPage{Size: 1})
	s.Assert().NoError(err)
	s.Assert().Len(users.Items, 1)
	s.Assert().True(users.PageInfo.HasMore)

	users, err = s.userService.List(s.normalUserContext, service.CursorPage{Size: 1, Token: users.PageInfo.NextPageToken})
	s.Assert().NoError(err)
	s.Assert().Len(users.Items, 1)
}

func (s *UserServiceIntegrationTestSuite) TestUpdate() {
	created, err := s.userService.Create(s.systemOwnerContext, serviceCreateUserOpts())
	s.Assert().NoError(err)

	updateOpts := service.UpdateUserOpts{
		Username: optional.Some("new_username"),
	}

	_, err = s.userService.Update(s.normalUserContext, created.ID, updateOpts)
	s.Assert().ErrorIs(err, service.ErrNoPermission)

	got, err := s.userService.Update(s.systemOwnerContext, created.ID, updateOpts)
	s.Assert().NoError(err)

	s.Assert().Equal("new_username", got.Username)
	s.Assert().Equal(created.Email, got.Email)
	s.Assert().NotNil(got.UpdatedAt)
}

func (s *UserServiceIntegrationTestSuite) TestDelete() {
	created, err := s.userService.Create(s.systemOwnerContext, serviceCreateUserOpts())
	s.Assert().NoError(err)

	err = s.userService.Delete(s.normalUserContext, created.ID, false)
	s.Assert().ErrorIs(err, service.ErrNoPermission)

	err = s.userService.Delete(s.systemOwnerContext, s.systemOwner.ID, false)
	s.Assert().ErrorIs(err, service.ErrNoPermission)

	err = s.userService.Delete(s.systemOwnerContext, created.ID, false)
	s.Assert().NoError(err)

	got, err := s.userService.Get(s.systemOwnerContext, created.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(created.Email, got.Email)
	s.Assert().Equal(password.UnusablePassword, got.Password)
	s.Assert().Equal(model.UserStatusDeleted, got.Status)
	s.Assert().NotNil(got.UpdatedAt)

	err = s.userService.Delete(s.systemOwnerContext, created.ID, true)
	s.Assert().NoError(err)

	_, err = s.userService.Get(s.systemOwnerContext, created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestUserServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceIntegrationTestSuite))
}
