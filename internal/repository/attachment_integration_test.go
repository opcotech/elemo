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

type AttachmentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *repository.User
	testOrg    *repository.Organization
	testDoc    *repository.Document
	createOpts repository.CreateAttachmentOpts
}

func (s *AttachmentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *AttachmentRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.testDoc, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateAttachmentOpts(s.testDoc.ID, s.testUser.ID)
}

func (s *AttachmentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *AttachmentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *AttachmentRepositoryIntegrationTestSuite) TestCreate() {
	attachment, err := s.AttachmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeAttachment), attachment.ID)
	s.Assert().NotNil(attachment.CreatedAt)
}

func (s *AttachmentRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.AttachmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	attachment, err := s.AttachmentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, attachment.ID)
	s.Assert().Equal(s.createOpts.Name, attachment.Name)
	s.Assert().Equal(s.createOpts.FileID, attachment.FileID)
	s.Assert().Equal(s.createOpts.CreatedBy, attachment.CreatedBy)
	s.Assert().WithinDuration(*created.CreatedAt, *attachment.CreatedAt, 100*time.Millisecond)
}

func (s *AttachmentRepositoryIntegrationTestSuite) TestGetAllBelongsTo() {
	_, err := s.AttachmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.AttachmentRepo.Create(context.Background(), testModel.NewCreateAttachmentOpts(s.testDoc.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.AttachmentRepo.Create(context.Background(), testModel.NewCreateAttachmentOpts(s.testDoc.ID, s.testUser.ID))
	s.Require().NoError(err)

	attachments, err := s.AttachmentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(attachments, 3)

	attachments, err = s.AttachmentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(attachments, 2)
}

func (s *AttachmentRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.AttachmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	attachment, err := s.AttachmentRepo.Update(context.Background(), created.ID, repository.UpdateAttachmentOpts{Name: "new name"})
	s.Require().NoError(err)
	s.Assert().Equal("new name", attachment.Name)
	s.Assert().NotNil(attachment.UpdatedAt)
}

func (s *AttachmentRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.AttachmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.AttachmentRepo.Delete(context.Background(), created.ID))
	_, err = s.AttachmentRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestAttachmentRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AttachmentRepositoryIntegrationTestSuite))
}

type CachedAttachmentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser       *repository.User
	testOrg        *repository.Organization
	testDoc        *repository.Document
	createOpts     repository.CreateAttachmentOpts
	attachmentRepo *repository.RedisCachedAttachmentRepository
}

func (s *CachedAttachmentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.attachmentRepo, _ = repository.NewCachedAttachmentRepository(s.AttachmentRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedAttachmentRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.testDoc, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateAttachmentOpts(s.testDoc.ID, s.testUser.ID)
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedAttachmentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedAttachmentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedAttachmentRepositoryIntegrationTestSuite) TestCreate() {
	attachment, err := s.attachmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotNil(attachment.CreatedAt)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedAttachmentRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.attachmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.AttachmentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	usingCache, err := s.attachmentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedAttachmentRepositoryIntegrationTestSuite) TestGetAllBelongsTo() {
	_, err := s.attachmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.AttachmentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 0, 10)
	s.Require().NoError(err)
	usingCache, err := s.attachmentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedAttachmentRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.attachmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	attachment, err := s.attachmentRepo.Update(context.Background(), created.ID, repository.UpdateAttachmentOpts{Name: "new name"})
	s.Require().NoError(err)
	s.Assert().Equal("new name", attachment.Name)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedAttachmentRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.attachmentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.attachmentRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
	s.Require().NoError(s.attachmentRepo.Delete(context.Background(), created.ID))
	_, err = s.attachmentRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedAttachmentRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedAttachmentRepositoryIntegrationTestSuite))
}
