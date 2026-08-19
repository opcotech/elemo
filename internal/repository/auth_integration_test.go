package repository_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/suite"
)

type UserTokenIntegrationTestSuite struct {
	testutil.ConfigurationTestSuite
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	createOpts repository.CreateUserTokenOpts
	userID     model.ID
}

func (s *UserTokenIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)
	s.SetupPg(&s.ContainerIntegrationTestSuite, container)
}

func (s *UserTokenIntegrationTestSuite) SetupTest() {
	testUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.userID = testUser.ID

	_, s.createOpts = testModel.NewCreateUserTokenOpts(testUser.ID)
}

func (s *UserTokenIntegrationTestSuite) TearDownTest() {
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *UserTokenIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *UserTokenIntegrationTestSuite) TestCreate() {
	token, err := s.UserTokenRepository.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeUserToken), token.ID)
	s.Assert().NotEmpty(token.CreatedAt)
}

func (s *UserTokenIntegrationTestSuite) TestGet() {
	created, err := s.UserTokenRepository.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	token, err := s.UserTokenRepository.Get(context.Background(), created.UserID, created.Context)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, token.ID)
}

func (s *UserTokenIntegrationTestSuite) TestDelete() {
	created, err := s.UserTokenRepository.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	s.Require().NoError(s.UserTokenRepository.Delete(context.Background(), created.UserID, created.Context))

	_, err = s.UserTokenRepository.Get(context.Background(), created.UserID, created.Context)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestUserTokenIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserTokenIntegrationTestSuite))
}
