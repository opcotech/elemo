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

type DocumentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *repository.User
	testOrg    *repository.Organization
	createOpts repository.CreateDocumentOpts
}

func (s *DocumentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *DocumentRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID)
}

func (s *DocumentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *DocumentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *DocumentRepositoryIntegrationTestSuite) TestCreate() {
	doc, err := s.DocumentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeDocument), doc.ID)
	s.Assert().NotNil(doc.CreatedAt)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.DocumentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	doc, err := s.DocumentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, doc.ID)
	s.Assert().Equal(s.createOpts.Name, doc.Name)
	s.Assert().Equal(s.createOpts.Excerpt, doc.Excerpt)
	s.Assert().Equal(s.createOpts.FileID, doc.FileID)
	s.Assert().Equal(s.createOpts.CreatedBy, doc.CreatedBy)
	s.Assert().WithinDuration(*created.CreatedAt, *doc.CreatedAt, 100*time.Millisecond)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestGetByCreator() {
	_, err := s.DocumentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)

	docs, err := s.DocumentRepo.GetByCreator(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(docs, 2)

	docs, err = s.DocumentRepo.GetByCreator(context.Background(), s.testUser.ID, 0, 1)
	s.Require().NoError(err)
	s.Assert().Len(docs, 1)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestGetAllBelongsTo() {
	_, err := s.DocumentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)

	docs, err := s.DocumentRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(docs, 3)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.DocumentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	doc, err := s.DocumentRepo.Update(context.Background(), created.ID, repository.UpdateDocumentOpts{
		Name:    optional.Some("new name"),
		Excerpt: optional.Some("new excerpt"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", doc.Name)
	s.Assert().Equal("new excerpt", doc.Excerpt)
	s.Assert().NotNil(doc.UpdatedAt)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.DocumentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.DocumentRepo.Delete(context.Background(), created.ID))
	_, err = s.DocumentRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestDocumentRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(DocumentRepositoryIntegrationTestSuite))
}

type CachedDocumentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser     *repository.User
	testOrg      *repository.Organization
	createOpts   repository.CreateDocumentOpts
	documentRepo *repository.RedisCachedDocumentRepository
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.documentRepo, _ = repository.NewCachedDocumentRepository(s.DocumentRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID)
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestCreate() {
	doc, err := s.documentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotNil(doc.CreatedAt)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.documentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.DocumentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	usingCache, err := s.documentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestGetByCreator() {
	_, err := s.documentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.DocumentRepo.GetByCreator(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)
	usingCache, err := s.documentRepo.GetByCreator(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestGetAllBelongsTo() {
	_, err := s.documentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.DocumentRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	usingCache, err := s.documentRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.documentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	doc, err := s.documentRepo.Update(context.Background(), created.ID, repository.UpdateDocumentOpts{
		Name: optional.Some("new name"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", doc.Name)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.documentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.documentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Require().NoError(s.documentRepo.Delete(context.Background(), created.ID))
	_, err = s.documentRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedDocumentRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedDocumentRepositoryIntegrationTestSuite))
}
