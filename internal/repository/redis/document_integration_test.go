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

type CachedDocumentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser     *model.User
	testOrg      *model.Organization
	document     *model.Document
	documentRepo *redis.CachedDocumentRepository
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.documentRepo, _ = redis.NewCachedDocumentRepository(s.DocumentRepo, redis.WithDatabase(s.RedisDB))
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.document = testModel.NewDocument(s.testUser.ID)

	s.Require().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.documentRepo.Create(context.Background(), s.testUser.ID, s.document))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeDocument), s.document.ID)
	s.Assert().NotNil(s.document.CreatedAt)
	s.Assert().Nil(s.document.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.document))

	original, err := s.DocumentRepo.Get(context.Background(), s.document.ID)
	s.Require().NoError(err)

	usingCache, err := s.documentRepo.Get(context.Background(), s.document.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cached, err := s.documentRepo.Get(context.Background(), s.document.ID)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache.ID, cached.ID)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestGetByCreator() {
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.document))
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, testModel.NewDocument(s.testUser.ID)))

	originalDocuments, err := s.DocumentRepo.GetByCreator(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)

	usingCacheDocuments, err := s.documentRepo.GetByCreator(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(originalDocuments, usingCacheDocuments)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedDocuments, err := s.documentRepo.GetByCreator(context.Background(), s.testUser.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(len(usingCacheDocuments), len(cachedDocuments))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestGetAllBelongsTo() {
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testOrg.ID, s.document))
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testOrg.ID, testModel.NewDocument(s.testUser.ID)))

	originalDocuments, err := s.DocumentRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)

	usingCacheDocuments, err := s.documentRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(originalDocuments, usingCacheDocuments)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedDocuments, err := s.documentRepo.GetAllBelongsTo(context.Background(), s.testOrg.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(len(usingCacheDocuments), len(cachedDocuments))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.document))

	patch := map[string]any{
		"name":    "new name",
		"excerpt": "new excerpt",
	}

	document, err := s.documentRepo.Update(context.Background(), s.document.ID, patch)
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

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testOrg.ID, s.document))

	_, err := s.documentRepo.Get(context.Background(), s.document.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.documentRepo.Delete(context.Background(), s.document.ID))

	_, err = s.documentRepo.Get(context.Background(), s.document.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedDocumentRepositoryIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(CachedDocumentRepositoryIntegrationTestSuite))
}
