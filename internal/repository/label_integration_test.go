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

type LabelRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *repository.User
	testOrg    *repository.Organization
	testDoc    *repository.Document
	createOpts repository.CreateLabelOpts
}

func (s *LabelRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *LabelRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)

	s.testDoc, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)

	s.createOpts = testModel.NewCreateLabelOpts()
}

func (s *LabelRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *LabelRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *LabelRepositoryIntegrationTestSuite) TestCreate() {
	label, err := s.LabelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeLabel), label.ID)
	s.Assert().NotNil(label.CreatedAt)
}

func (s *LabelRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.LabelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	label, err := s.LabelRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	s.Assert().Equal(created.ID, label.ID)
	s.Assert().Equal(s.createOpts.Name, label.Name)
	s.Assert().Equal(s.createOpts.Description, label.Description)
	s.Assert().WithinDuration(*created.CreatedAt, *label.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(label.UpdatedAt)
}

func (s *LabelRepositoryIntegrationTestSuite) TestGetAll() {
	_, err := s.LabelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.LabelRepo.Create(context.Background(), testModel.NewCreateLabelOpts())
	s.Require().NoError(err)
	_, err = s.LabelRepo.Create(context.Background(), testModel.NewCreateLabelOpts())
	s.Require().NoError(err)

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
	created, err := s.LabelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	updateOpts := repository.UpdateLabelOpts{
		Name:        optional.Some("new name"),
		Description: optional.Some("new description"),
	}

	label, err := s.LabelRepo.Update(context.Background(), created.ID, updateOpts)
	s.Require().NoError(err)

	s.Assert().Equal(created.ID, label.ID)
	s.Assert().Equal("new name", label.Name)
	s.Assert().Equal("new description", label.Description)
	s.Assert().WithinDuration(*created.CreatedAt, *label.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(label.UpdatedAt)
}

func (s *LabelRepositoryIntegrationTestSuite) TestAttachTo() {
	created, err := s.LabelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	s.Require().NoError(s.LabelRepo.AttachTo(context.Background(), created.ID, s.testDoc.ID))

	document, err := s.DocumentRepo.Get(context.Background(), s.testDoc.ID)
	s.Require().NoError(err)

	s.Assert().Len(document.Labels, 1)
	s.Assert().Equal(document.Labels[0], created.ID)
}

func (s *LabelRepositoryIntegrationTestSuite) TestDetachFrom() {
	created, err := s.LabelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	s.Require().NoError(s.LabelRepo.AttachTo(context.Background(), created.ID, s.testDoc.ID))
	s.Require().NoError(s.LabelRepo.DetachFrom(context.Background(), created.ID, s.testDoc.ID))

	document, err := s.DocumentRepo.Get(context.Background(), s.testDoc.ID)
	s.Require().NoError(err)

	s.Assert().Len(document.Labels, 0)
}

func (s *LabelRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.LabelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	s.Require().NoError(s.LabelRepo.Delete(context.Background(), created.ID))

	_, err = s.LabelRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestLabelRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(LabelRepositoryIntegrationTestSuite))
}

type CachedLabelRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser   *repository.User
	testOrg    *repository.Organization
	testDoc    *repository.Document
	createOpts repository.CreateLabelOpts
	labelRepo  *repository.RedisCachedLabelRepository
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
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)

	s.testDoc, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)

	s.createOpts = testModel.NewCreateLabelOpts()
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestCreate() {
	label, err := s.labelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeLabel), label.ID)
	s.Assert().NotNil(label.CreatedAt)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.labelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	original, err := s.LabelRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	usingCache, err := s.labelRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cached, err := s.labelRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(usingCache.ID, cached.ID)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestGetAll() {
	_, err := s.labelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.labelRepo.Create(context.Background(), testModel.NewCreateLabelOpts())
	s.Require().NoError(err)

	original, err := s.LabelRepo.GetAll(context.Background(), 0, 10)
	s.Require().NoError(err)

	usingCache, err := s.labelRepo.GetAll(context.Background(), 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.labelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	updateOpts := repository.UpdateLabelOpts{
		Name:        optional.Some("new name"),
		Description: optional.Some("new description"),
	}

	label, err := s.labelRepo.Update(context.Background(), created.ID, updateOpts)
	s.Require().NoError(err)
	s.Assert().Equal("new name", label.Name)
	s.Assert().Equal("new description", label.Description)
	s.Assert().NotNil(label.UpdatedAt)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestAttachTo() {
	created, err := s.labelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	s.Require().NoError(s.labelRepo.AttachTo(context.Background(), created.ID, s.testDoc.ID))

	document, err := s.DocumentRepo.Get(context.Background(), s.testDoc.ID)
	s.Require().NoError(err)
	s.Assert().Len(document.Labels, 1)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestDetachFrom() {
	created, err := s.labelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	s.Require().NoError(s.labelRepo.AttachTo(context.Background(), created.ID, s.testDoc.ID))
	s.Require().NoError(s.labelRepo.DetachFrom(context.Background(), created.ID, s.testDoc.ID))

	document, err := s.DocumentRepo.Get(context.Background(), s.testDoc.ID)
	s.Require().NoError(err)
	s.Assert().Len(document.Labels, 0)
}

func (s *CachedLabelRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.labelRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	_, err = s.labelRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.labelRepo.Delete(context.Background(), created.ID))

	_, err = s.labelRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedLabelRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedLabelRepositoryIntegrationTestSuite))
}
