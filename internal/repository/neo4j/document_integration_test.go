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

type DocumentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser *model.User
	testOrg  *model.Organization
	document *model.Document
}

func (s *DocumentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *DocumentRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.document = testModel.NewDocument(s.testUser.ID)
}

func (s *DocumentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *DocumentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *DocumentRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.document))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeDocument), s.document.ID)
	s.Assert().NotNil(s.document.CreatedAt)
	s.Assert().Nil(s.document.UpdatedAt)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.document))

	document, err := s.DocumentRepo.Get(context.Background(), s.document.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.document.ID, document.ID)
	s.Assert().Equal(s.document.Name, document.Name)
	s.Assert().Equal(s.document.Excerpt, document.Excerpt)
	s.Assert().Equal(s.document.FileID, document.FileID)
	s.Assert().Equal(s.document.CreatedBy, document.CreatedBy)
	s.Assert().Equal(s.document.Labels, document.Labels)
	s.Assert().Equal(s.document.Comments, document.Comments)
	s.Assert().Equal(s.document.Attachments, document.Attachments)
	s.Assert().WithinDuration(*s.document.CreatedAt, *document.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(s.document.UpdatedAt)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestGetByCreator() {
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.document))
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, testModel.NewDocument(s.testUser.ID)))
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, testModel.NewDocument(s.testUser.ID)))

	documents, err := s.DocumentRepo.GetByCreator(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(documents, 3)

	documents, err = s.DocumentRepo.GetByCreator(context.Background(), s.testUser.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(documents, 2)

	documents, err = s.DocumentRepo.GetByCreator(context.Background(), s.testUser.ID, 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(documents, 1)

	documents, err = s.DocumentRepo.GetByCreator(context.Background(), s.testUser.ID, 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(documents, 0)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestGetAllBelongsTo() {
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testOrg.ID, s.document))
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testOrg.ID, testModel.NewDocument(s.testUser.ID)))
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testOrg.ID, testModel.NewDocument(s.testUser.ID)))

	documents, err := s.DocumentRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(documents, 3)

	documents, err = s.DocumentRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(documents, 2)

	documents, err = s.DocumentRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(documents, 1)

	documents, err = s.DocumentRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(documents, 0)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.document))

	patch := map[string]any{
		"name":    "new name",
		"excerpt": "new excerpt",
	}

	document, err := s.DocumentRepo.Update(context.Background(), s.document.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(s.document.ID, document.ID)
	s.Assert().Equal(patch["name"], document.Name)
	s.Assert().Equal(patch["excerpt"], document.Excerpt)
	s.Assert().Equal(s.document.FileID, document.FileID)
	s.Assert().Equal(s.document.CreatedBy, document.CreatedBy)
	s.Assert().Equal(s.document.Labels, document.Labels)
	s.Assert().Equal(s.document.Comments, document.Comments)
	s.Assert().Equal(s.document.Attachments, document.Attachments)
	s.Assert().WithinDuration(*s.document.CreatedAt, *document.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(document.UpdatedAt)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.document))

	s.Require().NoError(s.DocumentRepo.Delete(context.Background(), s.document.ID))

	_, err := s.DocumentRepo.Get(context.Background(), s.document.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestDocumentRepositoryIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(DocumentRepositoryIntegrationTestSuite))
}
