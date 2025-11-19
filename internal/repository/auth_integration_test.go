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

	token *model.UserToken
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
	testUser := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), testUser))

	_, s.token = testModel.NewUserToken(testUser.ID)

}

func (s *UserTokenIntegrationTestSuite) TearDownTest() {
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *UserTokenIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *UserTokenIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.UserTokenRepository.Create(context.Background(), s.token))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeUserToken), s.token.ID)
	s.Assert().NotEmpty(s.token.CreatedAt)
}

func (s *UserTokenIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.UserTokenRepository.Create(context.Background(), s.token))

	token, err := s.UserTokenRepository.Get(context.Background(), s.token.UserID, s.token.Context)
	s.Require().NoError(err)
	s.Assert().Equal(s.token.ID, token.ID)
}

func (s *UserTokenIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.UserTokenRepository.Create(context.Background(), s.token))

	s.Require().NoError(s.UserTokenRepository.Delete(context.Background(), s.token.UserID, s.token.Context))

	_, err := s.UserTokenRepository.Get(context.Background(), s.token.UserID, s.token.Context)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestUserTokenIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserTokenIntegrationTestSuite))
}
