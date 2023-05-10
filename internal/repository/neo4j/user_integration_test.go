package neo4j_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type UserRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	user *model.User
}

func (s *UserRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *UserRepositoryIntegrationTestSuite) SetupTest() {
	s.user = testModel.NewUser()
}

func (s *UserRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *UserRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *UserRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.user))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeUser), s.user.ID)
	s.Assert().NotNil(s.user.CreatedAt)
	s.Assert().Nil(s.user.UpdatedAt)
}

func (s *UserRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.user))

	user, err := s.UserRepo.Get(context.Background(), s.user.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.user.Username, user.Username)
	s.Assert().Equal(s.user.Email, user.Email)
	s.Assert().Equal(s.user.Password, user.Password)
	s.Assert().Equal(s.user.Status, user.Status)
	s.Assert().Equal(s.user.FirstName, user.FirstName)
	s.Assert().Equal(s.user.LastName, user.LastName)
	s.Assert().Equal(s.user.Picture, user.Picture)
	s.Assert().Equal(s.user.Title, user.Title)
	s.Assert().Equal(s.user.Bio, user.Bio)
	s.Assert().Equal(s.user.Phone, user.Phone)
	s.Assert().Equal(s.user.Address, user.Address)
	s.Assert().Equal(s.user.Links, user.Links)
	s.Assert().Equal(s.user.Languages, user.Languages)
	s.Assert().Equal(s.user.Documents, user.Documents)
	s.Assert().Equal(s.user.Permissions, user.Permissions)
	s.Assert().WithinDuration(*s.user.CreatedAt, *user.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(s.user.UpdatedAt)
}

func (s *UserRepositoryIntegrationTestSuite) TestGetByEmail() {
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.user))

	user, err := s.UserRepo.GetByEmail(context.Background(), s.user.Email)
	s.Require().NoError(err)

	s.Assert().Equal(s.user.Username, user.Username)
	s.Assert().Equal(s.user.Email, user.Email)
	s.Assert().Equal(s.user.Password, user.Password)
	s.Assert().Equal(s.user.Status, user.Status)
	s.Assert().Equal(s.user.FirstName, user.FirstName)
	s.Assert().Equal(s.user.LastName, user.LastName)
	s.Assert().Equal(s.user.Picture, user.Picture)
	s.Assert().Equal(s.user.Title, user.Title)
	s.Assert().Equal(s.user.Bio, user.Bio)
	s.Assert().Equal(s.user.Phone, user.Phone)
	s.Assert().Equal(s.user.Address, user.Address)
	s.Assert().Equal(s.user.Links, user.Links)
	s.Assert().Equal(s.user.Languages, user.Languages)
	s.Assert().Equal(s.user.Documents, user.Documents)
	s.Assert().Equal(s.user.Permissions, user.Permissions)
	s.Assert().WithinDuration(*s.user.CreatedAt, *user.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(s.user.UpdatedAt)
}

func (s *UserRepositoryIntegrationTestSuite) TestGetAll() {
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.user))
	s.Require().NoError(s.UserRepo.Create(context.Background(), testModel.NewUser()))
	s.Require().NoError(s.UserRepo.Create(context.Background(), testModel.NewUser()))

	users, err := s.UserRepo.GetAll(context.Background(), 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(users, 3)

	users, err = s.UserRepo.GetAll(context.Background(), 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(users, 2)

	users, err = s.UserRepo.GetAll(context.Background(), 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(users, 1)

	users, err = s.UserRepo.GetAll(context.Background(), 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(users, 0)
}

func (s *UserRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.user))

	patch := map[string]any{
		"username": "new username",
		"email":    testutil.GenerateEmail(10),
		"languages": []string{
			model.LanguageEN.String(),
		},
	}

	user, err := s.UserRepo.Update(context.Background(), s.user.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(patch["username"], user.Username)
	s.Assert().Equal(patch["email"], user.Email)
	s.Assert().Equal(s.user.Password, user.Password)
	s.Assert().Equal(s.user.Status, user.Status)
	s.Assert().Equal(s.user.FirstName, user.FirstName)
	s.Assert().Equal(s.user.LastName, user.LastName)
	s.Assert().Equal(s.user.Picture, user.Picture)
	s.Assert().Equal(s.user.Title, user.Title)
	s.Assert().Equal(s.user.Bio, user.Bio)
	s.Assert().Equal(s.user.Phone, user.Phone)
	s.Assert().Equal(s.user.Address, user.Address)
	s.Assert().Equal(s.user.Links, user.Links)
	s.Assert().ElementsMatch([]model.Language{model.LanguageEN}, user.Languages)
	s.Assert().Equal(s.user.Documents, user.Documents)
	s.Assert().Equal(s.user.Permissions, user.Permissions)
	s.Assert().WithinDuration(*s.user.CreatedAt, *user.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(s.user.UpdatedAt)
}

func (s *UserRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.user))

	err := s.UserRepo.Delete(context.Background(), s.user.ID)
	s.Require().NoError(err)

	_, err = s.UserRepo.Get(context.Background(), s.user.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestUserRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryIntegrationTestSuite))
}
