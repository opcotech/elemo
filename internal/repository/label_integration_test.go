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

type LabelRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser *model.User
	testOrg  *model.Organization
	testDoc  *model.Document

	label *model.Label
}

func (s *LabelRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *LabelRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.testDoc = testModel.NewDocument(s.testUser.ID)
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.testDoc))

	s.label = testModel.NewLabel()
}

func (s *LabelRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *LabelRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *LabelRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.LabelRepo.Create(context.Background(), s.label))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeLabel), s.label.ID)
	s.Assert().NotNil(s.label.CreatedAt)
}

func (s *LabelRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.LabelRepo.Create(context.Background(), s.label))

	label, err := s.LabelRepo.Get(context.Background(), s.label.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.label.ID, label.ID)
	s.Assert().Equal(s.label.Name, label.Name)
	s.Assert().Equal(s.label.Description, label.Description)
	s.Assert().WithinDuration(*s.label.CreatedAt, *label.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(s.label.UpdatedAt, label.UpdatedAt)
}

func (s *LabelRepositoryIntegrationTestSuite) TestGetAll() {
	s.Require().NoError(s.LabelRepo.Create(context.Background(), s.label))
	s.Require().NoError(s.LabelRepo.Create(context.Background(), testModel.NewLabel()))
	s.Require().NoError(s.LabelRepo.Create(context.Background(), testModel.NewLabel()))

	labels, err := s.LabelRepo.GetAll(context.Background(), 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(labels, 3)

	labels, err = s.LabelRepo.GetAll(context.Background(), 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(labels, 2)

	labels, err = s.LabelRepo.GetAll(context.Background(), 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(labels, 1)

	labels, err = s.LabelRepo.GetAll(context.Background(), 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(labels, 0)
}

func (s *LabelRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.LabelRepo.Create(context.Background(), s.label))

	patch := map[string]any{
		"name":        "new name",
		"description": "new description",
	}

	label, err := s.LabelRepo.Update(context.Background(), s.label.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(s.label.ID, label.ID)
	s.Assert().Equal(patch["name"], label.Name)
	s.Assert().Equal(patch["description"], label.Description)
	s.Assert().WithinDuration(*s.label.CreatedAt, *label.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(label.UpdatedAt)
}

func (s *LabelRepositoryIntegrationTestSuite) TestAttachTo() {
	s.Require().NoError(s.LabelRepo.Create(context.Background(), s.label))

	s.Require().NoError(s.LabelRepo.AttachTo(context.Background(), s.label.ID, s.testDoc.ID))

	document, err := s.DocumentRepo.Get(context.Background(), s.testDoc.ID)
	s.Require().NoError(err)

	s.Assert().Len(document.Labels, 1)
	s.Assert().Equal(document.Labels[0], s.label.ID)
}

func (s *LabelRepositoryIntegrationTestSuite) TestDetachFrom() {
	s.Require().NoError(s.LabelRepo.Create(context.Background(), s.label))

	s.Require().NoError(s.LabelRepo.AttachTo(context.Background(), s.label.ID, s.testDoc.ID))
	s.Require().NoError(s.LabelRepo.DetachFrom(context.Background(), s.label.ID, s.testDoc.ID))

	document, err := s.DocumentRepo.Get(context.Background(), s.testDoc.ID)
	s.Require().NoError(err)

	s.Assert().Len(document.Labels, 0)
}

func (s *LabelRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.LabelRepo.Create(context.Background(), s.label))

	s.Require().NoError(s.LabelRepo.Delete(context.Background(), s.label.ID))

	_, err := s.LabelRepo.Get(context.Background(), s.label.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestLabelRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(LabelRepositoryIntegrationTestSuite))
}

type CachedLabelRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser  *model.User
	testOrg   *model.Organization
	testDoc   *model.Document
	label     *model.Label
	labelRepo *repository.RedisCachedLabelRepository
}

func (s *CachedLabelRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.labelRepo, _ = repository.NewCachedLabelRepository(s.LabelRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedLabelRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.testDoc = testModel.NewDocument(s.testUser.ID)
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.testDoc))

	s.label = testModel.NewLabel()

	s.Require().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.labelRepo.Create(context.Background(), s.label))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeLabel), s.label.ID)
	s.Assert().NotNil(s.label.CreatedAt)
	s.Assert().Nil(s.label.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.LabelRepo.Create(context.Background(), s.label))

	original, err := s.LabelRepo.Get(context.Background(), s.label.ID)
	s.Require().NoError(err)

	usingCache, err := s.labelRepo.Get(context.Background(), s.label.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cached, err := s.labelRepo.Get(context.Background(), s.label.ID)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache.ID, cached.ID)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestGetAll() {
	s.Require().NoError(s.LabelRepo.Create(context.Background(), s.label))
	s.Require().NoError(s.LabelRepo.Create(context.Background(), testModel.NewLabel()))

	originalLabels, err := s.LabelRepo.GetAll(context.Background(), 0, 10)
	s.Require().NoError(err)

	usingCacheLabels, err := s.labelRepo.GetAll(context.Background(), 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(originalLabels, usingCacheLabels)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedLabels, err := s.labelRepo.GetAll(context.Background(), 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(len(usingCacheLabels), len(cachedLabels))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestAttachTo() {
	s.Require().NoError(s.LabelRepo.Create(context.Background(), s.label))

	_, err := s.labelRepo.GetAll(context.Background(), 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.labelRepo.AttachTo(context.Background(), s.label.ID, s.testDoc.ID))

	document, err := s.DocumentRepo.Get(context.Background(), s.testDoc.ID)
	s.Require().NoError(err)
	s.Assert().Len(document.Labels, 1)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestDetachFrom() {
	s.Require().NoError(s.LabelRepo.Create(context.Background(), s.label))

	_, err := s.labelRepo.GetAll(context.Background(), 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.labelRepo.AttachTo(context.Background(), s.label.ID, s.testDoc.ID))
	s.Require().NoError(s.labelRepo.DetachFrom(context.Background(), s.label.ID, s.testDoc.ID))

	document, err := s.DocumentRepo.Get(context.Background(), s.testDoc.ID)
	s.Require().NoError(err)
	s.Assert().Len(document.Labels, 0)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.LabelRepo.Create(context.Background(), s.label))

	patch := map[string]any{
		"name":        "new name",
		"description": "new description",
	}

	label, err := s.labelRepo.Update(context.Background(), s.label.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(s.label.ID, label.ID)
	s.Assert().Equal(patch["name"], label.Name)
	s.Assert().Equal(patch["description"], label.Description)
	s.Assert().WithinDuration(*s.label.CreatedAt, *label.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(label.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.LabelRepo.Create(context.Background(), s.label))

	_, err := s.labelRepo.Get(context.Background(), s.label.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.labelRepo.Delete(context.Background(), s.label.ID))

	_, err = s.labelRepo.Get(context.Background(), s.label.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedLabelRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedLabelRepositoryIntegrationTestSuite))
}
