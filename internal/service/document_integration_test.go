//go:build integration

package service_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

type DocumentServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.LocalStackContainerIntegrationTestSuite

	documentService   service.DocumentService
	staticFileService service.StaticFileService

	owner        *repository.User
	organization *repository.Organization

	ctx context.Context
}

func (s *DocumentServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)
	s.SetupLocalStack(&s.ContainerIntegrationTestSuite, container)

	permissionService, err := service.NewPermissionService(s.PermissionRepo)
	s.Require().NoError(err)

	licenseService, err := service.NewLicenseService(
		testutil.ParseLicense(s.T()),
		s.LicenseRepo,
		service.WithPermissionService(permissionService),
	)
	s.Require().NoError(err)

	s.staticFileService, err = service.NewStaticFileService(
		s.StaticFileRepository,
		service.WithLicenseService(licenseService),
	)
	s.Require().NoError(err)

	s.documentService, err = service.NewDocumentService(
		service.WithDocumentRepository(s.DocumentRepo),
		service.WithPermissionService(permissionService),
		service.WithLicenseService(licenseService),
		service.WithStaticFileService(s.staticFileService),
	)
	s.Require().NoError(err)
}

func (s *DocumentServiceIntegrationTestSuite) SetupTest() {
	var err error
	s.owner, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.owner.ID, s.Neo4jDB))

	s.organization, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.owner.ID))
	s.Require().NoError(err)

	_, err = s.PermissionRepo.Create(context.Background(), repository.CreateGrantOpts{
		Principal: s.owner.ID,
		Scope:     s.organization.ID,
		Actions:   testModel.OrgAdminActions(),
	})
	s.Require().NoError(err)
}

func (s *DocumentServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupLocalStack(&s.ContainerIntegrationTestSuite)
}

func (s *DocumentServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *DocumentServiceIntegrationTestSuite) createDocument(name string, content []byte) *service.Document {
	s.T().Helper()

	doc, err := s.documentService.Create(s.ctx, s.organization.ID, service.CreateDocumentOpts{
		Title:   name,
		Excerpt: "integration test document excerpt",
		Content: content,
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(doc.ID)
	return doc
}

func (s *DocumentServiceIntegrationTestSuite) TestCreate() {
	content := []byte("create document body")
	doc := s.createDocument("create-document", content)

	s.Assert().Equal("create-document", doc.Title)
	s.Assert().Equal(content, doc.Content)
	s.Assert().NotEmpty(doc.FileID)
	s.Assert().NotNil(doc.CreatedAt)

	stored, err := s.staticFileService.Get(context.Background(), doc.FileID)
	s.Require().NoError(err)
	s.Assert().Equal(content, stored)
}

func (s *DocumentServiceIntegrationTestSuite) TestCreateWithoutPermission() {
	otherUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	otherCtx := context.WithValue(context.Background(), pkg.CtxKeyUserID, otherUser.ID)

	_, err = s.documentService.Create(otherCtx, s.organization.ID, service.CreateDocumentOpts{
		Title:   "unauthorized-doc",
		Excerpt: "should fail excerpt",
		Content: []byte("unauthorized body"),
	})
	s.Assert().ErrorIs(err, service.ErrNoPermission)
}

func (s *DocumentServiceIntegrationTestSuite) TestGet() {
	content := []byte("get document body")
	created := s.createDocument("get-document", content)

	doc, err := s.documentService.Get(s.ctx, created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, doc.ID)
	s.Assert().Equal(created.Title, doc.Title)
	s.Assert().Equal(content, doc.Content)
	s.Assert().Equal(created.FileID, doc.FileID)
}

func (s *DocumentServiceIntegrationTestSuite) TestGetNotFound() {
	_, err := s.documentService.Get(s.ctx, model.MustNewID(model.ResourceTypeDocument))
	s.Assert().Error(err)
	s.Assert().ErrorIs(err, service.ErrDocumentGet)
}

func (s *DocumentServiceIntegrationTestSuite) TestGetWithoutPermission() {
	created := s.createDocument("perm-get-document", []byte("perm get body"))

	otherUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	otherCtx := context.WithValue(context.Background(), pkg.CtxKeyUserID, otherUser.ID)

	_, err = s.documentService.Get(otherCtx, created.ID)
	s.Assert().ErrorIs(err, service.ErrNoPermission)
}

func (s *DocumentServiceIntegrationTestSuite) TestUpdate() {
	created := s.createDocument("upd-document", []byte("original body"))

	doc, err := s.documentService.Update(s.ctx, created.ID, service.UpdateDocumentOpts{
		Title: optional.Some("updated-title"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("updated-title", doc.Title)
	s.Assert().Equal([]byte("original body"), doc.Content)
	s.Assert().NotNil(doc.UpdatedAt)

	updatedContent := []byte("updated body")
	doc, err = s.documentService.Update(s.ctx, created.ID, service.UpdateDocumentOpts{
		Content: optional.Some(updatedContent),
	})
	s.Require().NoError(err)
	s.Assert().Equal("updated-title", doc.Title)
	s.Assert().Equal(updatedContent, doc.Content)

	stored, err := s.staticFileService.Get(context.Background(), doc.FileID)
	s.Require().NoError(err)
	s.Assert().Equal(updatedContent, stored)
}

func (s *DocumentServiceIntegrationTestSuite) TestUpdateWithoutPermission() {
	created := s.createDocument("perm-upd-document", []byte("perm update body"))

	otherUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	otherCtx := context.WithValue(context.Background(), pkg.CtxKeyUserID, otherUser.ID)

	_, err = s.documentService.Update(otherCtx, created.ID, service.UpdateDocumentOpts{
		Title: optional.Some("should-fail"),
	})
	s.Assert().ErrorIs(err, service.ErrNoPermission)
}

func (s *DocumentServiceIntegrationTestSuite) TestDelete() {
	created := s.createDocument("del-document", []byte("delete body"))
	fileID := created.FileID

	s.Require().NoError(s.documentService.Delete(s.ctx, created.ID))

	_, err := s.documentService.Get(s.ctx, created.ID)
	s.Assert().Error(err)

	_, err = s.staticFileService.Get(context.Background(), fileID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func (s *DocumentServiceIntegrationTestSuite) TestDeleteWithoutPermission() {
	created := s.createDocument("perm-del-document", []byte("perm delete body"))

	otherUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	otherCtx := context.WithValue(context.Background(), pkg.CtxKeyUserID, otherUser.ID)

	err = s.documentService.Delete(otherCtx, created.ID)
	s.Assert().ErrorIs(err, service.ErrNoPermission)
}

func (s *DocumentServiceIntegrationTestSuite) TestDeleteNotFound() {
	err := s.documentService.Delete(s.ctx, model.MustNewID(model.ResourceTypeDocument))
	s.Assert().Error(err)
	s.Assert().ErrorIs(err, service.ErrDocumentDelete)
}

func (s *DocumentServiceIntegrationTestSuite) TestListLibrary() {
	s.createDocument("lib-doc-1", []byte("library one"))
	s.createDocument("lib-doc-2", []byte("library two"))

	page, err := s.documentService.ListLibrary(s.ctx, s.organization.ID, service.LibraryListFilter{All: true}, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Assert().GreaterOrEqual(len(page.Items), 2)
}

func TestDocumentServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(DocumentServiceIntegrationTestSuite))
}
