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
	_, err = s.PermissionRepo.Create(context.Background(), testModel.NewCreateGrantOpts(
		s.testUser.ID,
		s.testOrg.ID,
		testModel.OrgAdminActions()...,
	))
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
	doc, err := s.DocumentRepo.Get(context.Background(), created.ID, repository.DocumentDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, doc.ID)
	s.Assert().Equal(s.createOpts.Title, doc.Title)
	s.Assert().Equal(s.createOpts.Excerpt, doc.Excerpt)
	s.Assert().Equal(s.createOpts.FileID, doc.FileID)
	s.Assert().Equal(s.createOpts.CreatedBy, doc.CreatedBy.ID)
	s.Require().NotNil(doc.CommentCount)
	s.Assert().Equal(int64(0), *doc.CommentCount)
	s.Require().NotNil(doc.AttachmentCount)
	s.Assert().Equal(int64(0), *doc.AttachmentCount)
	s.Assert().WithinDuration(*created.CreatedAt, *doc.CreatedAt, 100*time.Millisecond)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestGetByCreator() {
	_, err := s.DocumentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)

	docs, err := s.DocumentRepo.ListByCreator(context.Background(), s.testUser.ID, s.testUser.ID, repository.CursorPage{Size: 10}, repository.DocumentListProjection())
	s.Require().NoError(err)
	s.Assert().Len(docs.Items, 2)

	docs, err = s.DocumentRepo.ListByCreator(context.Background(), s.testUser.ID, s.testUser.ID, repository.CursorPage{Size: 1}, repository.DocumentListProjection())
	s.Require().NoError(err)
	s.Assert().Len(docs.Items, 1)
	s.Assert().True(docs.PageInfo.HasMore)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestListLibrary() {
	_, err := s.DocumentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)

	docs, err := s.DocumentRepo.ListLibrary(context.Background(), s.testOrg.ID, s.testUser.ID, nil, repository.LibraryListFilter{All: true}, repository.CursorPage{Size: 10}, repository.DocumentListProjection())
	s.Require().NoError(err)
	s.Assert().Len(docs.Items, 3)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestMoveToFolderAndRelate() {
	folder, err := s.FolderRepo.Create(context.Background(), testModel.NewCreateFolderOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)

	created, err := s.DocumentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().Nil(created.Folder)

	moved, err := s.DocumentRepo.MoveToFolder(context.Background(), created.ID, &folder.ID)
	s.Require().NoError(err)
	s.Require().NotNil(moved.Folder)
	s.Assert().Equal(folder.ID, moved.Folder.ID)

	inFolder, err := s.DocumentRepo.ListLibrary(context.Background(), s.testOrg.ID, s.testUser.ID, nil, repository.LibraryListFilter{FolderID: &folder.ID}, repository.CursorPage{Size: 10}, repository.DocumentListProjection())
	s.Require().NoError(err)
	s.Assert().Len(inFolder.Items, 1)

	atRoot, err := s.DocumentRepo.ListLibrary(context.Background(), s.testOrg.ID, s.testUser.ID, nil, repository.LibraryListFilter{}, repository.CursorPage{Size: 10}, repository.DocumentListProjection())
	s.Require().NoError(err)
	s.Assert().Empty(atRoot.Items)

	cleared, err := s.DocumentRepo.MoveToFolder(context.Background(), created.ID, nil)
	s.Require().NoError(err)
	s.Assert().Nil(cleared.Folder)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.DocumentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	doc, err := s.DocumentRepo.Update(context.Background(), created.ID, repository.UpdateDocumentOpts{
		Title:   optional.Some("new title"),
		Excerpt: optional.Some("new excerpt"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new title", doc.Title)
	s.Assert().Equal("new excerpt", doc.Excerpt)
	s.Assert().NotNil(doc.UpdatedAt)
}

func (s *DocumentRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.DocumentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.DocumentRepo.Delete(context.Background(), created.ID))
	_, err = s.DocumentRepo.Get(context.Background(), created.ID, repository.DocumentDetailProjection())
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
	_, err = s.PermissionRepo.Create(context.Background(), testModel.NewCreateGrantOpts(
		s.testUser.ID,
		s.testOrg.ID,
		testModel.OrgAdminActions()...,
	))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID)
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
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
	original, err := s.DocumentRepo.Get(context.Background(), created.ID, repository.DocumentDetailProjection())
	s.Require().NoError(err)
	usingCache, err := s.documentRepo.Get(context.Background(), created.ID, repository.DocumentDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestGetByCreator() {
	_, err := s.documentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.DocumentRepo.ListByCreator(context.Background(), s.testUser.ID, s.testUser.ID, repository.CursorPage{Size: 10}, repository.DocumentListProjection())
	s.Require().NoError(err)
	usingCache, err := s.documentRepo.ListByCreator(context.Background(), s.testUser.ID, s.testUser.ID, repository.CursorPage{Size: 10}, repository.DocumentListProjection())
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestListLibrary() {
	_, err := s.documentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.DocumentRepo.ListLibrary(context.Background(), s.testOrg.ID, s.testUser.ID, nil, repository.LibraryListFilter{All: true}, repository.CursorPage{Size: 10}, repository.DocumentListProjection())
	s.Require().NoError(err)
	usingCache, err := s.documentRepo.ListLibrary(context.Background(), s.testOrg.ID, s.testUser.ID, nil, repository.LibraryListFilter{All: true}, repository.CursorPage{Size: 10}, repository.DocumentListProjection())
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.documentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	doc, err := s.documentRepo.Update(context.Background(), created.ID, repository.UpdateDocumentOpts{
		Title: optional.Some("new title"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new title", doc.Title)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedDocumentRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.documentRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.documentRepo.Get(context.Background(), created.ID, repository.DocumentDetailProjection())
	s.Require().NoError(err)
	s.Require().NoError(s.documentRepo.Delete(context.Background(), created.ID))
	_, err = s.documentRepo.Get(context.Background(), created.ID, repository.DocumentDetailProjection())
	s.Assert().ErrorIs(err, repository.ErrNotFound)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedDocumentRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedDocumentRepositoryIntegrationTestSuite))
}
