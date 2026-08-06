package repository_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	createOpts repository.CreateUserOpts
}

func (s *UserRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *UserRepositoryIntegrationTestSuite) SetupTest() {
	s.createOpts = testModel.NewCreateUserOpts()
}

func (s *UserRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *UserRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *UserRepositoryIntegrationTestSuite) TestCreate() {
	user, err := s.UserRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeUser), user.ID)
	s.Assert().NotNil(user.CreatedAt)
	s.Assert().Nil(user.UpdatedAt)
}

func (s *UserRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.UserRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	user, err := s.UserRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.createOpts.Username, user.Username)
	s.Assert().Equal(s.createOpts.Email, user.Email)
	s.Assert().Equal(s.createOpts.Password, user.Password)
	s.Assert().Equal(s.createOpts.Status, user.Status)
	s.Assert().Equal(s.createOpts.FirstName, user.FirstName)
	s.Assert().Equal(s.createOpts.LastName, user.LastName)
	s.Assert().Equal(s.createOpts.Picture, user.Picture)
	s.Assert().Equal(s.createOpts.Title, user.Title)
	s.Assert().Equal(s.createOpts.Bio, user.Bio)
	s.Assert().Equal(s.createOpts.Phone, user.Phone)
	s.Assert().Equal(s.createOpts.Address, user.Address)
	s.Assert().Equal(s.createOpts.Links, user.Links)
	s.Assert().Equal(s.createOpts.Languages, user.Languages)
	s.Assert().WithinDuration(*created.CreatedAt, *user.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(user.UpdatedAt)
}

func (s *UserRepositoryIntegrationTestSuite) TestGetByEmail() {
	created, err := s.UserRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	user, err := s.UserRepo.GetByEmail(context.Background(), s.createOpts.Email)
	s.Require().NoError(err)

	s.Assert().Equal(created.ID, user.ID)
	s.Assert().Equal(s.createOpts.Email, user.Email)
}

func (s *UserRepositoryIntegrationTestSuite) TestGetAll() {
	_, err := s.UserRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	_, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

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
	created, err := s.UserRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	newEmail := testutil.GenerateEmail(10)
	updateOpts := repository.UpdateUserOpts{
		Username:  optional.Some("newusername"),
		Email:     optional.Some(newEmail),
		Languages: optional.Some([]model.Language{model.LanguageEN}),
	}

	user, err := s.UserRepo.Update(context.Background(), created.ID, updateOpts)
	s.Require().NoError(err)

	s.Assert().Equal("newusername", user.Username)
	s.Assert().Equal(newEmail, user.Email)
	s.Assert().Equal(created.Password, user.Password)
	s.Assert().Equal(created.Status, user.Status)
	s.Assert().Equal(created.FirstName, user.FirstName)
	s.Assert().Equal(created.LastName, user.LastName)
	s.Assert().ElementsMatch([]model.Language{model.LanguageEN}, user.Languages)
	s.Assert().WithinDuration(*created.CreatedAt, *user.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(user.UpdatedAt)
}

func (s *UserRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.UserRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	err = s.UserRepo.Delete(context.Background(), created.ID)
	s.Require().NoError(err)

	_, err = s.UserRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestUserRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryIntegrationTestSuite))
}

type CachedUserRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	createOpts repository.CreateUserOpts
	userRepo   *repository.RedisCachedUserRepository
}

func (s *CachedUserRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.userRepo, _ = repository.NewCachedUserRepository(s.UserRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedUserRepositoryIntegrationTestSuite) SetupTest() {
	s.createOpts = testModel.NewCreateUserOpts()
	s.Require().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedUserRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedUserRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedUserRepositoryIntegrationTestSuite) TestCreate() {
	user, err := s.userRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeUser), user.ID)
	s.Assert().NotNil(user.CreatedAt)
	s.Assert().Nil(user.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedUserRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.userRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	original, err := s.UserRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	usingCache, err := s.userRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cached, err := s.userRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache.ID, cached.ID)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedUserRepositoryIntegrationTestSuite) TestGetByEmail() {
	created, err := s.userRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	original, err := s.UserRepo.GetByEmail(context.Background(), created.Email)
	s.Require().NoError(err)

	usingCache, err := s.userRepo.GetByEmail(context.Background(), created.Email)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cached, err := s.userRepo.GetByEmail(context.Background(), created.Email)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache.ID, cached.ID)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedUserRepositoryIntegrationTestSuite) TestGetAll() {
	_, err := s.userRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.userRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	originalUsers, err := s.UserRepo.GetAll(context.Background(), 0, 10)
	s.Require().NoError(err)

	usingCacheUsers, err := s.userRepo.GetAll(context.Background(), 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(originalUsers, usingCacheUsers)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedUsers, err := s.userRepo.GetAll(context.Background(), 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(len(usingCacheUsers), len(cachedUsers))
}

func (s *CachedUserRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.userRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	newEmail := testutil.GenerateEmail(10)
	updateOpts := repository.UpdateUserOpts{
		Username:  optional.Some("newusername"),
		Email:     optional.Some(newEmail),
		Languages: optional.Some([]model.Language{model.LanguageEN}),
	}

	user, err := s.userRepo.Update(context.Background(), created.ID, updateOpts)
	s.Require().NoError(err)

	s.Assert().Equal("newusername", user.Username)
	s.Assert().Equal(newEmail, user.Email)
	s.Assert().ElementsMatch([]model.Language{model.LanguageEN}, user.Languages)
	s.Assert().NotNil(user.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedUserRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.userRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	_, err = s.userRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.userRepo.Delete(context.Background(), created.ID))

	_, err = s.userRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedUserRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedUserRepositoryIntegrationTestSuite))
}
