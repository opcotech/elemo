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

	actor        *repository.User
	actorContext context.Context
	other        *repository.User
}

func (s *UserServiceIntegrationTestSuite) SetupSuite() {
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

	s.userService, err = service.NewUserService(
		s.UserRepo,
		s.UserTokenRepository,
		licenseService,
	)
	s.Require().NoError(err)
}

func (s *UserServiceIntegrationTestSuite) SetupTest() {
	var err error
	s.actor, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.actorContext = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.actor.ID)

	s.other, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
}

func (s *UserServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *UserServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *UserServiceIntegrationTestSuite) TestCreateUser() {
	_, err := s.userService.Create(s.actorContext, serviceCreateUserOpts())
	s.Assert().NoError(err)
}

func (s *UserServiceIntegrationTestSuite) TestGet() {
	got, err := s.userService.Get(s.actorContext, s.other.ID)
	s.Assert().NoError(err)

	s.Assert().Equal(s.other.Username, got.Username)
	s.Assert().Equal(s.other.Email, got.Email)
	s.Assert().Equal(s.other.Password, got.Password)
	s.Assert().Equal(s.other.Status, got.Status)
	s.Assert().Equal(s.other.FirstName, got.FirstName)
	s.Assert().Equal(s.other.LastName, got.LastName)
	s.Assert().WithinDuration(*s.other.CreatedAt, *got.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(got.UpdatedAt)
}

func (s *UserServiceIntegrationTestSuite) TestGetByEmail() {
	got, err := s.userService.GetByEmail(s.actorContext, s.other.Email)
	s.Assert().NoError(err)
	s.Assert().Equal(s.other.Email, got.Email)
	s.Assert().Equal(s.other.Username, got.Username)
}

func (s *UserServiceIntegrationTestSuite) TestList() {
	users, err := s.userService.List(s.actorContext, service.CursorPage{Size: 10})
	s.Assert().NoError(err)
	s.Assert().Len(users.Items, 2)

	users, err = s.userService.List(s.actorContext, service.CursorPage{Size: 1})
	s.Assert().NoError(err)
	s.Assert().Len(users.Items, 1)
	s.Assert().True(users.PageInfo.HasMore)

	users, err = s.userService.List(s.actorContext, service.CursorPage{Size: 1, Token: users.PageInfo.NextPageToken})
	s.Assert().NoError(err)
	s.Assert().Len(users.Items, 1)
}

func (s *UserServiceIntegrationTestSuite) TestUpdate() {
	updateOpts := service.UpdateUserOpts{
		Username: optional.Some("new_username"),
	}

	_, err := s.userService.Update(s.actorContext, s.other.ID, updateOpts)
	s.Assert().ErrorIs(err, service.ErrNoPermission)

	got, err := s.userService.Update(s.actorContext, s.actor.ID, updateOpts)
	s.Assert().NoError(err)

	s.Assert().Equal("new_username", got.Username)
	s.Assert().Equal(s.actor.Email, got.Email)
	s.Assert().NotNil(got.UpdatedAt)
}

func (s *UserServiceIntegrationTestSuite) TestDelete() {
	created, err := s.userService.Create(s.actorContext, serviceCreateUserOpts())
	s.Assert().NoError(err)
	selfCtx := context.WithValue(context.Background(), pkg.CtxKeyUserID, created.ID)

	err = s.userService.Delete(s.actorContext, created.ID, false)
	s.Assert().ErrorIs(err, service.ErrNoPermission)

	err = s.userService.Delete(selfCtx, created.ID, false)
	s.Assert().NoError(err)

	got, err := s.userService.Get(selfCtx, created.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(created.Email, got.Email)
	s.Assert().Equal(password.UnusablePassword, got.Password)
	s.Assert().Equal(model.UserStatusDeleted, got.Status)
	s.Assert().NotNil(got.UpdatedAt)

	err = s.userService.Delete(selfCtx, created.ID, true)
	s.Assert().NoError(err)

	_, err = s.userService.Get(selfCtx, created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestUserServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceIntegrationTestSuite))
}
