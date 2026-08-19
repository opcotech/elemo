//go:build integration

package repository_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/suite"
)

type CommentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *repository.User
	testOrg    *repository.Organization
	testDoc    *repository.Document
	createOpts repository.CreateCommentOpts
}

func (s *CommentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *CommentRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)

	s.testDoc, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)

	s.createOpts = testModel.NewCreateCommentOpts(s.testDoc.ID, s.testUser.ID)
}

func (s *CommentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *CommentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CommentRepositoryIntegrationTestSuite) TestCreate() {
	comment, err := s.CommentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeComment), comment.ID)
	s.Assert().NotNil(comment.CreatedAt)
}

func (s *CommentRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.CommentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	comment, err := s.CommentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	s.Assert().Equal(created.ID, comment.ID)
	s.Assert().Equal(s.createOpts.CreatedBy, comment.CreatedBy)
	s.Assert().Equal(s.createOpts.Content, comment.Content)
	s.Assert().WithinDuration(*created.CreatedAt, *comment.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(comment.UpdatedAt)
}

func (s *CommentRepositoryIntegrationTestSuite) TestGetAllBelongsTo() {
	_, err := s.CommentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.CommentRepo.Create(context.Background(), testModel.NewCreateCommentOpts(s.testDoc.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.CommentRepo.Create(context.Background(), testModel.NewCreateCommentOpts(s.testDoc.ID, s.testUser.ID))
	s.Require().NoError(err)

	comments, err := s.CommentRepo.ListBelongsTo(context.Background(), s.testDoc.ID, repository.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Assert().Len(comments.Items, 3)

	comments, err = s.CommentRepo.ListBelongsTo(context.Background(), s.testDoc.ID, repository.CursorPage{Size: 2})
	s.Require().NoError(err)
	s.Assert().Len(comments.Items, 2)
	s.Assert().True(comments.PageInfo.HasMore)

	comments, err = s.CommentRepo.ListBelongsTo(context.Background(), s.testDoc.ID, repository.CursorPage{Size: 2, Token: comments.PageInfo.NextPageToken})
	s.Require().NoError(err)
	s.Assert().Len(comments.Items, 1)
	s.Assert().False(comments.PageInfo.HasMore)
}

func (s *CommentRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.CommentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	comment, err := s.CommentRepo.Update(context.Background(), created.ID, repository.UpdateCommentOpts{Content: "new content"})
	s.Require().NoError(err)

	s.Assert().Equal(created.ID, comment.ID)
	s.Assert().Equal("new content", comment.Content)
	s.Assert().Equal(created.CreatedBy, comment.CreatedBy)
	s.Assert().WithinDuration(*created.CreatedAt, *comment.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(comment.UpdatedAt)
}

func (s *CommentRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.CommentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	s.Require().NoError(s.CommentRepo.Delete(context.Background(), created.ID))

	_, err = s.CommentRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestCommentRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CommentRepositoryIntegrationTestSuite))
}

type CachedCommentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser    *repository.User
	testOrg     *repository.Organization
	testDoc     *repository.Document
	createOpts  repository.CreateCommentOpts
	commentRepo *repository.RedisCachedCommentRepository
}

func (s *CachedCommentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.commentRepo, _ = repository.NewCachedCommentRepository(s.CommentRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedCommentRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)

	s.testDoc, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)

	s.createOpts = testModel.NewCreateCommentOpts(s.testDoc.ID, s.testUser.ID)
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TestCreate() {
	comment, err := s.commentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeComment), comment.ID)
	s.Assert().NotNil(comment.CreatedAt)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.commentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	original, err := s.CommentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	usingCache, err := s.commentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TestGetAllBelongsTo() {
	_, err := s.commentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.commentRepo.Create(context.Background(), testModel.NewCreateCommentOpts(s.testDoc.ID, s.testUser.ID))
	s.Require().NoError(err)

	original, err := s.CommentRepo.ListBelongsTo(context.Background(), s.testDoc.ID, repository.CursorPage{Size: 10})
	s.Require().NoError(err)

	usingCache, err := s.commentRepo.ListBelongsTo(context.Background(), s.testDoc.ID, repository.CursorPage{Size: 10})
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.commentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	comment, err := s.commentRepo.Update(context.Background(), created.ID, repository.UpdateCommentOpts{Content: "new content"})
	s.Require().NoError(err)
	s.Assert().Equal("new content", comment.Content)
	s.Assert().NotNil(comment.UpdatedAt)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedCommentRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.commentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	_, err = s.commentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.commentRepo.Delete(context.Background(), created.ID))

	_, err = s.commentRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedCommentRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedCommentRepositoryIntegrationTestSuite))
}
