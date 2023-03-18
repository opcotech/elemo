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

type CommentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser *model.User
	testOrg  *model.Organization
	testDoc  *model.Document
	comment  *model.Comment
}

func (s *CommentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *CommentRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.testDoc = testModel.NewDocument(s.testUser.ID)
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.testDoc))

	s.comment = testModel.NewComment(s.testUser.ID)
}

func (s *CommentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *CommentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CommentRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.CommentRepo.Create(context.Background(), s.testDoc.ID, s.comment))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeComment), s.comment.ID)
	s.Assert().NotNil(s.comment.CreatedAt)
}

func (s *CommentRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.CommentRepo.Create(context.Background(), s.testDoc.ID, s.comment))

	comment, err := s.CommentRepo.Get(context.Background(), s.comment.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.comment.ID, comment.ID)
	s.Assert().Equal(s.comment.CreatedBy, comment.CreatedBy)
	s.Assert().Equal(s.comment.Content, comment.Content)
	s.Assert().WithinDuration(*s.comment.CreatedAt, *comment.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(comment.UpdatedAt)
}

func (s *CommentRepositoryIntegrationTestSuite) TestGetAllBelongsTo() {
	s.Require().NoError(s.CommentRepo.Create(context.Background(), s.testDoc.ID, s.comment))
	s.Require().NoError(s.CommentRepo.Create(context.Background(), s.testDoc.ID, testModel.NewComment(s.testUser.ID)))
	s.Require().NoError(s.CommentRepo.Create(context.Background(), s.testDoc.ID, testModel.NewComment(s.testUser.ID)))

	comments, err := s.CommentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(comments, 3)

	comments, err = s.CommentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(comments, 2)

	comments, err = s.CommentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(comments, 1)

	comments, err = s.CommentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(comments, 0)
}

func (s *CommentRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.CommentRepo.Create(context.Background(), s.testDoc.ID, s.comment))

	newContent := "new content"
	comment, err := s.CommentRepo.Update(context.Background(), s.comment.ID, newContent)
	s.Require().NoError(err)

	s.Assert().Equal(s.comment.ID, comment.ID)
	s.Assert().Equal(s.comment.CreatedBy, comment.CreatedBy)
	s.Assert().Equal(newContent, comment.Content)
	s.Assert().WithinDuration(*s.comment.CreatedAt, *comment.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(comment.UpdatedAt)
}

func (s *CommentRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.CommentRepo.Create(context.Background(), s.testDoc.ID, s.comment))

	s.Require().NoError(s.CommentRepo.Delete(context.Background(), s.comment.ID))

	_, err := s.CommentRepo.Get(context.Background(), s.comment.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestCommentRepositoryIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(CommentRepositoryIntegrationTestSuite))
}
