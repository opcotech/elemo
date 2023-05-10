package redis_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/repository/redis"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type CachedCommentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser    *model.User
	testOrg     *model.Organization
	testDoc     *model.Document
	comment     *model.Comment
	commentRepo *redis.CachedCommentRepository
}

func (s *CachedCommentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.commentRepo, _ = redis.NewCachedCommentRepository(s.CommentRepo, redis.WithDatabase(s.RedisDB))
}

func (s *CachedCommentRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.testDoc = testModel.NewDocument(s.testUser.ID)
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.testDoc))

	s.comment = testModel.NewComment(s.testUser.ID)

	s.Require().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.commentRepo.Create(context.Background(), s.testDoc.ID, s.comment))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeComment), s.comment.ID)
	s.Assert().NotNil(s.comment.CreatedAt)
	s.Assert().Nil(s.comment.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.CommentRepo.Create(context.Background(), s.testDoc.ID, s.comment))

	original, err := s.CommentRepo.Get(context.Background(), s.comment.ID)
	s.Require().NoError(err)

	usingCache, err := s.commentRepo.Get(context.Background(), s.comment.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cached, err := s.commentRepo.Get(context.Background(), s.comment.ID)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache.ID, cached.ID)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TestGetAll() {
	s.Require().NoError(s.CommentRepo.Create(context.Background(), s.testDoc.ID, s.comment))
	s.Require().NoError(s.CommentRepo.Create(context.Background(), s.testDoc.ID, testModel.NewComment(s.testUser.ID)))

	originalComments, err := s.CommentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 0, 10)
	s.Require().NoError(err)

	usingCacheComments, err := s.commentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(originalComments, usingCacheComments)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedComments, err := s.commentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(len(usingCacheComments), len(cachedComments))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.CommentRepo.Create(context.Background(), s.testDoc.ID, s.comment))

	newContent := "new content"
	comment, err := s.commentRepo.Update(context.Background(), s.comment.ID, newContent)
	s.Require().NoError(err)

	s.Assert().Equal(s.comment.ID, comment.ID)
	s.Assert().Equal(s.comment.CreatedBy, comment.CreatedBy)
	s.Assert().Equal(newContent, comment.Content)
	s.Assert().WithinDuration(*s.comment.CreatedAt, *comment.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(comment.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.CommentRepo.Create(context.Background(), s.testDoc.ID, s.comment))

	_, err := s.commentRepo.Get(context.Background(), s.comment.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.commentRepo.Delete(context.Background(), s.comment.ID))

	_, err = s.commentRepo.Get(context.Background(), s.comment.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedCommentRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedCommentRepositoryIntegrationTestSuite))
}
