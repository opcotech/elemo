package service_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type SearchServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.SearchContainerIntegrationTestSuite

	permissionService service.PermissionService
	searchService     service.SearchService
	ctx               context.Context
}

func (s *SearchServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	name := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, name)
	s.SetupSearch(&s.ContainerIntegrationTestSuite, name)

	var err error
	s.permissionService, err = service.NewPermissionService(
		s.PermissionRepo,
		service.WithRoleRepository(s.RoleRepo),
	)
	s.Require().NoError(err)

	s.searchService, err = service.NewSearchService(
		s.SearchRepo,
		service.WithPermissionService(s.permissionService),
	)
	s.Require().NoError(err)
}

func (s *SearchServiceIntegrationTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *SearchServiceIntegrationTestSuite) TearDownTest() {
	s.CleanupSearch(&s.ContainerIntegrationTestSuite)
	s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *SearchServiceIntegrationTestSuite) TearDownSuite() {
	s.CleanupContainers()
}

func (s *SearchServiceIntegrationTestSuite) createUser() *repository.User {
	user, err := s.UserRepo.Create(s.ctx, testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	return user
}

func (s *SearchServiceIntegrationTestSuite) userCtx(userID model.ID) context.Context {
	return context.WithValue(s.ctx, pkg.CtxKeyUserID, userID)
}

func (s *SearchServiceIntegrationTestSuite) TestIndexSearchAndAuthz() {
	owner := s.createUser()
	viewer := s.createUser()

	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)
	issue, err := s.IssueRepo.Create(s.ctx, testModel.NewCreateIssueOpts(project.ID, owner.ID))
	s.Require().NoError(err)

	_, err = s.PermissionRepo.Create(s.ctx, repository.CreateGrantOpts{
		Principal: owner.ID,
		Scope:     org.ID,
		Actions: []model.Action{
			model.ActionOrganizationRead,
			model.ActionNamespaceRead,
			model.ActionProjectRead,
			model.ActionIssueRead,
		},
	})
	s.Require().NoError(err)
	_, err = s.PermissionRepo.Create(s.ctx, repository.CreateGrantOpts{
		Principal: viewer.ID,
		Scope:     project.ID,
		Actions:   []model.Action{model.ActionProjectRead, model.ActionIssueRead},
	})
	s.Require().NoError(err)

	s.Require().NoError(s.searchService.Index(s.ctx, service.IndexInput{
		ID:    org.ID,
		Title: org.Name,
	}))
	s.Require().NoError(s.searchService.Index(s.ctx, service.IndexInput{
		ID:      ns.ID,
		Title:   ns.Name,
		Content: ns.Description,
	}))
	s.Require().NoError(s.searchService.Index(s.ctx, service.IndexInput{
		ID:      project.ID,
		Title:   project.Name,
		Content: project.Description,
		Key:     project.Key,
	}))
	s.Require().NoError(s.searchService.Index(s.ctx, service.IndexInput{
		ID:      issue.ID,
		Title:   issue.Title,
		Content: issue.Description,
		Key:     issue.Key,
	}))

	ownerPage, err := s.searchService.Search(s.userCtx(owner.ID), service.SearchQuery{PageSize: 20})
	s.Require().NoError(err)
	s.Require().Nil(ownerPage.PageInfo.TotalCount)
	ids := make([]model.ID, 0, len(ownerPage.Items))
	for _, item := range ownerPage.Items {
		ids = append(ids, item.ID)
	}
	s.Assert().Contains(ids, org.ID)
	s.Assert().Contains(ids, project.ID)
	s.Assert().Contains(ids, issue.ID)

	viewerPage, err := s.searchService.Search(s.userCtx(viewer.ID), service.SearchQuery{PageSize: 20})
	s.Require().NoError(err)
	s.Require().Nil(viewerPage.PageInfo.TotalCount)
	viewerIDs := make([]model.ID, 0, len(viewerPage.Items))
	for _, item := range viewerPage.Items {
		viewerIDs = append(viewerIDs, item.ID)
	}
	s.Assert().NotContains(viewerIDs, org.ID)
	s.Assert().Contains(viewerIDs, project.ID)
	s.Assert().Contains(viewerIDs, issue.ID)
}

func (s *SearchServiceIntegrationTestSuite) TestDeleteByScopeRemovesDescendants() {
	owner := s.createUser()
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)

	_, err = s.PermissionRepo.Create(s.ctx, repository.CreateGrantOpts{
		Principal: owner.ID,
		Scope:     org.ID,
		Actions:   []model.Action{model.ActionOrganizationRead, model.ActionNamespaceRead, model.ActionProjectRead},
	})
	s.Require().NoError(err)

	s.Require().NoError(s.searchService.Index(s.ctx, service.IndexInput{ID: org.ID, Title: org.Name}))
	s.Require().NoError(s.searchService.Index(s.ctx, service.IndexInput{ID: ns.ID, Title: ns.Name}))
	s.Require().NoError(s.searchService.Index(s.ctx, service.IndexInput{ID: project.ID, Title: project.Name, Key: project.Key}))

	s.Require().NoError(s.searchService.DeleteByScope(s.ctx, project.ID))

	page, err := s.searchService.Search(s.userCtx(owner.ID), service.SearchQuery{PageSize: 20})
	s.Require().NoError(err)
	for _, item := range page.Items {
		s.Assert().NotEqual(project.ID, item.ID)
	}
}

func (s *SearchServiceIntegrationTestSuite) TestReindex() {
	owner := s.createUser()
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)
	issue, err := s.IssueRepo.Create(s.ctx, testModel.NewCreateIssueOpts(project.ID, owner.ID))
	s.Require().NoError(err)

	_, err = s.PermissionRepo.Create(s.ctx, repository.CreateGrantOpts{
		Principal: owner.ID,
		Scope:     org.ID,
		Actions: []model.Action{
			model.ActionOrganizationRead,
			model.ActionNamespaceRead,
			model.ActionProjectRead,
			model.ActionIssueRead,
		},
	})
	s.Require().NoError(err)

	s.Require().NoError(s.searchService.Reindex(s.ctx, service.SearchReindexSources{
		DB: s.Neo4jDB,
	}, service.SearchReindexOptions{}))

	page, err := s.searchService.Search(s.userCtx(owner.ID), service.SearchQuery{PageSize: 20})
	s.Require().NoError(err)
	ids := make([]model.ID, 0, len(page.Items))
	for _, item := range page.Items {
		ids = append(ids, item.ID)
	}
	s.Assert().Contains(ids, org.ID)
	s.Assert().Contains(ids, ns.ID)
	s.Assert().Contains(ids, project.ID)
	s.Assert().Contains(ids, issue.ID)
}

func TestSearchServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(SearchServiceIntegrationTestSuite))
}
