//go:build integration

package repository_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/suite"
)

type FolderRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser *repository.User
	testOrg  *repository.Organization
}

func (s *FolderRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *FolderRepositoryIntegrationTestSuite) SetupTest() {
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
}

func (s *FolderRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *FolderRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *FolderRepositoryIntegrationTestSuite) TestCreateNestedAndList() {
	root, err := s.FolderRepo.Create(context.Background(), testModel.NewCreateFolderOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)

	childOpts := testModel.NewCreateFolderOpts(s.testOrg.ID, s.testUser.ID)
	childOpts.ParentID = &root.ID
	child, err := s.FolderRepo.Create(context.Background(), childOpts)
	s.Require().NoError(err)
	s.Require().NotNil(child.Parent)
	s.Assert().Equal(root.ID, child.Parent.ID)

	rootPage, err := s.FolderRepo.List(context.Background(), s.testOrg.ID, nil, s.testUser.ID, repository.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Assert().Len(rootPage.Items, 1)
	s.Assert().Equal(root.ID, rootPage.Items[0].ID)

	childPage, err := s.FolderRepo.List(context.Background(), s.testOrg.ID, &root.ID, s.testUser.ID, repository.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Assert().Len(childPage.Items, 1)
	s.Assert().Equal(child.ID, childPage.Items[0].ID)
}

func (s *FolderRepositoryIntegrationTestSuite) TestSiblingNameConflict() {
	opts := testModel.NewCreateFolderOpts(s.testOrg.ID, s.testUser.ID)
	_, err := s.FolderRepo.Create(context.Background(), opts)
	s.Require().NoError(err)

	dup := opts
	_, err = s.FolderRepo.Create(context.Background(), dup)
	s.Assert().ErrorIs(err, repository.ErrFolderNameConflict)
}

func (s *FolderRepositoryIntegrationTestSuite) TestCycleRejected() {
	parent, err := s.FolderRepo.Create(context.Background(), testModel.NewCreateFolderOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)

	childOpts := testModel.NewCreateFolderOpts(s.testOrg.ID, s.testUser.ID)
	childOpts.ParentID = &parent.ID
	child, err := s.FolderRepo.Create(context.Background(), childOpts)
	s.Require().NoError(err)

	_, err = s.FolderRepo.Update(context.Background(), parent.ID, repository.UpdateFolderOpts{
		ParentID: optional.Some(child.ID),
	})
	s.Assert().ErrorIs(err, repository.ErrFolderCycle)
}

func (s *FolderRepositoryIntegrationTestSuite) TestDeleteReparents() {
	parent, err := s.FolderRepo.Create(context.Background(), testModel.NewCreateFolderOpts(s.testOrg.ID, s.testUser.ID))
	s.Require().NoError(err)

	childOpts := testModel.NewCreateFolderOpts(s.testOrg.ID, s.testUser.ID)
	childOpts.ParentID = &parent.ID
	child, err := s.FolderRepo.Create(context.Background(), childOpts)
	s.Require().NoError(err)

	docOpts := testModel.NewCreateDocumentOpts(s.testOrg.ID, s.testUser.ID)
	docOpts.FolderID = &parent.ID
	doc, err := s.DocumentRepo.Create(context.Background(), docOpts)
	s.Require().NoError(err)

	s.Require().NoError(s.FolderRepo.Delete(context.Background(), parent.ID))

	reparented, err := s.FolderRepo.Get(context.Background(), child.ID)
	s.Require().NoError(err)
	s.Assert().Nil(reparented.Parent)

	movedDoc, err := s.DocumentRepo.Get(context.Background(), doc.ID, repository.DocumentDetailProjection())
	s.Require().NoError(err)
	s.Assert().Nil(movedDoc.Folder)
}

func TestFolderRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(FolderRepositoryIntegrationTestSuite))
}
