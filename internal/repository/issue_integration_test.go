package repository_test

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func cacheKeysWithoutIssueListGeneration(keys []string) []string {
	filtered := make([]string, 0, len(keys))
	for _, key := range keys {
		if strings.HasPrefix(key, "issue:list:gen:") {
			continue
		}
		filtered = append(filtered, key)
	}
	return filtered
}

type IssueRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser      *repository.User
	testOrg       *repository.Organization
	testNamespace *repository.Namespace
	testProject   *repository.Project
	createOpts    repository.CreateIssueOpts
}

func (s *IssueRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *IssueRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.testNamespace, err = s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)
	s.testProject, err = s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(s.testNamespace.ID, s.testUser.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID)
}

func (s *IssueRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *IssueRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *IssueRepositoryIntegrationTestSuite) TestCreate() {
	issue, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeIssue), issue.ID)
	s.Assert().Equal(uint(1), issue.NumericID)
	s.Assert().Equal(model.FormatIssueKey(s.testProject.Key, 1), issue.Key)
	s.Assert().NotNil(issue.CreatedAt)
	s.Assert().Nil(issue.UpdatedAt)
	s.Require().NotNil(issue.Project)
	s.Assert().Equal(s.testProject.ID, issue.Project.ID)
	s.Require().NotNil(issue.ReportedBy)
	s.Assert().Equal(s.createOpts.ReportedBy, issue.ReportedBy.ID)
	s.Require().NotNil(issue.WatcherCount)
	s.Assert().Equal(int64(1), *issue.WatcherCount)
}

func (s *IssueRepositoryIntegrationTestSuite) TestCreateSequentialNumericIDs() {
	first, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().Equal(uint(1), first.NumericID)
	s.Assert().Equal(model.FormatIssueKey(s.testProject.Key, 1), first.Key)

	second, err := s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)
	s.Assert().Equal(uint(2), second.NumericID)
	s.Assert().Equal(model.FormatIssueKey(s.testProject.Key, 2), second.Key)

	third, err := s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)
	s.Assert().Equal(uint(3), third.NumericID)

	otherProject, err := s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(s.testNamespace.ID, s.testUser.ID))
	s.Require().NoError(err)

	other, err := s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(otherProject.ID, s.testUser.ID))
	s.Require().NoError(err)
	s.Assert().Equal(uint(1), other.NumericID)
	s.Assert().Equal(model.FormatIssueKey(otherProject.Key, 1), other.Key)
}

func (s *IssueRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	issue, err := s.IssueRepo.Get(context.Background(), created.ID, repository.IssueDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, issue.ID)
	s.Assert().Equal(uint(1), issue.NumericID)
	s.Assert().Equal(model.FormatIssueKey(s.testProject.Key, 1), issue.Key)
	s.Assert().Equal(s.createOpts.Kind, issue.Kind)
	s.Assert().Equal(s.createOpts.Title, issue.Title)
	s.Assert().Equal(s.createOpts.Description, issue.Description)
	s.Assert().Equal(s.createOpts.Status, issue.Status)
	s.Assert().Equal(s.createOpts.Priority, issue.Priority)
	s.Require().NotNil(issue.ReportedBy)
	s.Assert().Equal(s.createOpts.ReportedBy, issue.ReportedBy.ID)
	s.Assert().WithinDuration(*created.CreatedAt, *issue.CreatedAt, 100*time.Millisecond)
}

func (s *IssueRepositoryIntegrationTestSuite) TestGetByKey() {
	created, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	issue, err := s.IssueRepo.GetByKey(context.Background(), s.testNamespace.ID, created.Key, repository.IssueDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, issue.ID)
	s.Assert().Equal(created.Key, issue.Key)
	s.Assert().Equal(created.NumericID, issue.NumericID)
}

func (s *IssueRepositoryIntegrationTestSuite) TestGetByKeyScopedToNamespace() {
	created, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	otherNamespace, err := s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)
	otherProjectOpts := testModel.NewCreateProjectOpts(otherNamespace.ID, s.testUser.ID)
	otherProjectOpts.Key = s.testProject.Key
	otherProject, err := s.ProjectRepo.Create(context.Background(), otherProjectOpts)
	s.Require().NoError(err)
	otherIssue, err := s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(otherProject.ID, s.testUser.ID))
	s.Require().NoError(err)
	s.Require().Equal(created.Key, otherIssue.Key)

	got, err := s.IssueRepo.GetByKey(context.Background(), s.testNamespace.ID, created.Key, repository.IssueDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, got.ID)

	gotOther, err := s.IssueRepo.GetByKey(context.Background(), otherNamespace.ID, otherIssue.Key, repository.IssueDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(otherIssue.ID, gotOther.ID)

	_, err = s.IssueRepo.GetByKey(context.Background(), otherNamespace.ID, "ZZZZZZ-999", repository.IssueDetailProjection())
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func (s *IssueRepositoryIntegrationTestSuite) TestListForProject() {
	_, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)

	issues, err := s.IssueRepo.ListForProject(context.Background(), repository.IssueListQuery{ProjectID: s.testProject.ID, Page: repository.CursorPage{Size: 10}, Projection: repository.IssueListForProjectProjection()})
	s.Require().NoError(err)
	s.Assert().Len(issues.Items, 3)
	for _, issue := range issues.Items {
		s.Require().NotNil(issue.ReportedBy)
		s.Assert().Equal(s.testUser.ID, issue.ReportedBy.ID)
	}

	issues, err = s.IssueRepo.ListForProject(context.Background(), repository.IssueListQuery{ProjectID: s.testProject.ID, Page: repository.CursorPage{Size: 2}, Projection: repository.IssueListForProjectProjection()})
	s.Require().NoError(err)
	s.Assert().Len(issues.Items, 2)
}

func (s *IssueRepositoryIntegrationTestSuite) TestListForNamespace() {
	_, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	otherProject, err := s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(s.testNamespace.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(otherProject.ID, s.testUser.ID))
	s.Require().NoError(err)

	issues, err := s.IssueRepo.ListForNamespace(context.Background(), repository.IssueListForNamespaceQuery{
		NamespaceID: s.testNamespace.ID,
		Page:        repository.CursorPage{Size: 10},
		Projection:  repository.IssueListForNamespaceProjection(),
	})
	s.Require().NoError(err)
	s.Assert().Len(issues.Items, 2)

	keys := make([]string, 0, len(issues.Items))
	for _, issue := range issues.Items {
		s.Require().NotNil(issue.Project)
		keys = append(keys, issue.Key)
	}
	s.Assert().Contains(keys, model.FormatIssueKey(s.testProject.Key, 1))
	s.Assert().Contains(keys, model.FormatIssueKey(otherProject.Key, 1))
}

func (s *IssueRepositoryIntegrationTestSuite) TestListForUser() {
	assigned, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.AssignmentRepo.Create(context.Background(), testModel.NewCreateAssignmentOpts(
		s.testUser.ID,
		assigned.ID,
		model.AssignmentKindAssignee,
	))
	s.Require().NoError(err)

	_, err = s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)

	issues, err := s.IssueRepo.ListForUser(context.Background(), repository.IssueListForUserQuery{
		UserID:     s.testUser.ID,
		Page:       repository.CursorPage{Size: 10},
		Projection: repository.IssueListForUserProjection(),
	})
	s.Require().NoError(err)
	s.Require().Len(issues.Items, 1)
	s.Assert().Equal(assigned.ID, issues.Items[0].ID)
	s.Require().NotNil(issues.Items[0].Project)
}

func (s *IssueRepositoryIntegrationTestSuite) TestListForIssue() {
	parent, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	childOpts := testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID)
	childOpts.Parent = &parent.ID
	_, err = s.IssueRepo.Create(context.Background(), childOpts)
	s.Require().NoError(err)

	related, err := s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.IssueRepo.AddRelation(context.Background(), repository.CreateIssueRelationOpts{
		Source: parent.ID,
		Target: related.ID,
		Kind:   model.IssueRelationKindBlocks,
	})
	s.Require().NoError(err)

	issues, err := s.IssueRepo.ListForIssue(context.Background(), repository.IssueListForIssueQuery{IssueID: parent.ID, Page: repository.CursorPage{Size: 10}, Projection: repository.IssueDetailProjection()})
	s.Require().NoError(err)
	s.Assert().Len(issues.Items, 2)
}

func (s *IssueRepositoryIntegrationTestSuite) TestAddWatcher() {
	created, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	watcher, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.IssueRepo.AddWatcher(context.Background(), created.ID, watcher.ID))
}

func (s *IssueRepositoryIntegrationTestSuite) TestGetWatchers() {
	created, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	watchers, err := s.IssueRepo.GetWatchers(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Len(watchers, 1)

	watcher, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.IssueRepo.AddWatcher(context.Background(), created.ID, watcher.ID))

	watchers, err = s.IssueRepo.GetWatchers(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Len(watchers, 2)
}

func (s *IssueRepositoryIntegrationTestSuite) TestRemoveWatcher() {
	created, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Require().NotNil(created.ReportedBy)
	s.Require().NoError(s.IssueRepo.RemoveWatcher(context.Background(), created.ID, created.ReportedBy.ID))
	watchers, err := s.IssueRepo.GetWatchers(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Empty(watchers)
}

func (s *IssueRepositoryIntegrationTestSuite) TestAddRelation() {
	created, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	related, err := s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.IssueRepo.AddRelation(context.Background(), repository.CreateIssueRelationOpts{
		Source: created.ID,
		Target: related.ID,
		Kind:   model.IssueRelationKindBlocks,
	})
	s.Require().NoError(err)
}

func (s *IssueRepositoryIntegrationTestSuite) TestGetRelations() {
	created, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	related, err := s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.IssueRepo.AddRelation(context.Background(), repository.CreateIssueRelationOpts{
		Source: created.ID,
		Target: related.ID,
		Kind:   model.IssueRelationKindBlocks,
	})
	s.Require().NoError(err)
	relations, err := s.IssueRepo.GetRelations(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Len(relations, 1)
}

func (s *IssueRepositoryIntegrationTestSuite) TestRemoveRelation() {
	created, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	related, err := s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.IssueRepo.AddRelation(context.Background(), repository.CreateIssueRelationOpts{
		Source: created.ID,
		Target: related.ID,
		Kind:   model.IssueRelationKindBlocks,
	})
	s.Require().NoError(err)
	s.Require().NoError(s.IssueRepo.RemoveRelation(context.Background(), created.ID, related.ID, model.IssueRelationKindBlocks))
}

func (s *IssueRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	issue, err := s.IssueRepo.Update(context.Background(), created.ID, repository.UpdateIssueOpts{
		Title:       optional.Some("new title"),
		Description: optional.Some("new description"),
		Status:      optional.Some(model.IssueStatusClosed),
	}, repository.IssueDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal("new title", issue.Title)
	s.Assert().Equal("new description", issue.Description)
	s.Assert().Equal(model.IssueStatusClosed, issue.Status)
	s.Assert().NotNil(issue.UpdatedAt)
}

func (s *IssueRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.IssueRepo.Delete(context.Background(), created.ID))
	_, err = s.IssueRepo.Get(context.Background(), created.ID, repository.IssueDetailProjection())
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestIssueRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IssueRepositoryIntegrationTestSuite))
}

type CachedIssueRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser      *repository.User
	testOrg       *repository.Organization
	testNamespace *repository.Namespace
	testProject   *repository.Project
	createOpts    repository.CreateIssueOpts
	issueRepo     *repository.RedisCachedIssueRepository
}

func (s *CachedIssueRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.issueRepo, _ = repository.NewCachedIssueRepository(s.IssueRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedIssueRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.testNamespace, err = s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)
	s.testProject, err = s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(s.testNamespace.ID, s.testUser.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID)
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestCreate() {
	issue, err := s.issueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotNil(issue.CreatedAt)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 0)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.issueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.IssueRepo.Get(context.Background(), created.ID, repository.IssueDetailProjection())
	s.Require().NoError(err)
	usingCache, err := s.issueRepo.Get(context.Background(), created.ID, repository.IssueDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestGetByKey() {
	created, err := s.issueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.IssueRepo.GetByKey(context.Background(), s.testNamespace.ID, created.Key, repository.IssueDetailProjection())
	s.Require().NoError(err)
	usingCache, err := s.issueRepo.GetByKey(context.Background(), s.testNamespace.ID, created.Key, repository.IssueDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestListForProject() {
	_, err := s.issueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.IssueRepo.ListForProject(context.Background(), repository.IssueListQuery{ProjectID: s.testProject.ID, Page: repository.CursorPage{Size: 10}, Projection: repository.IssueListForProjectProjection()})
	s.Require().NoError(err)
	usingCache, err := s.issueRepo.ListForProject(context.Background(), repository.IssueListQuery{ProjectID: s.testProject.ID, Page: repository.CursorPage{Size: 10}, Projection: repository.IssueListForProjectProjection()})
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.issueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	issue, err := s.issueRepo.Update(context.Background(), created.ID, repository.UpdateIssueOpts{
		Title: optional.Some("new title"),
	}, repository.IssueDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal("new title", issue.Title)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestUpdateInvalidatesProjectList() {
	created, err := s.issueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	query := repository.IssueListQuery{
		ProjectID:  s.testProject.ID,
		Page:       repository.CursorPage{Size: 10},
		Projection: repository.IssueListForProjectProjection(),
	}
	_, err = s.issueRepo.ListForProject(context.Background(), query)
	s.Require().NoError(err)

	_, err = s.issueRepo.Update(context.Background(), created.ID, repository.UpdateIssueOpts{
		Status: optional.Some(model.IssueStatusDone),
	}, repository.IssueDetailProjection())
	s.Require().NoError(err)

	page, err := s.issueRepo.ListForProject(context.Background(), query)
	s.Require().NoError(err)
	s.Require().Len(page.Items, 1)
	s.Assert().Equal(model.IssueStatusDone, page.Items[0].Status)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestListForNamespace() {
	_, err := s.issueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.IssueRepo.ListForNamespace(context.Background(), repository.IssueListForNamespaceQuery{
		NamespaceID: s.testNamespace.ID,
		Page:        repository.CursorPage{Size: 10},
		Projection:  repository.IssueListForNamespaceProjection(),
	})
	s.Require().NoError(err)
	usingCache, err := s.issueRepo.ListForNamespace(context.Background(), repository.IssueListForNamespaceQuery{
		NamespaceID: s.testNamespace.ID,
		Page:        repository.CursorPage{Size: 10},
		Projection:  repository.IssueListForNamespaceProjection(),
	})
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestListForUser() {
	created, err := s.issueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.AssignmentRepo.Create(context.Background(), testModel.NewCreateAssignmentOpts(
		s.testUser.ID,
		created.ID,
		model.AssignmentKindAssignee,
	))
	s.Require().NoError(err)

	query := repository.IssueListForUserQuery{
		UserID:     s.testUser.ID,
		Page:       repository.CursorPage{Size: 10},
		Projection: repository.IssueListForUserProjection(),
	}
	original, err := s.IssueRepo.ListForUser(context.Background(), query)
	s.Require().NoError(err)
	usingCache, err := s.issueRepo.ListForUser(context.Background(), query)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.issueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.issueRepo.Get(context.Background(), created.ID, repository.IssueDetailProjection())
	s.Require().NoError(err)
	s.Require().NoError(s.issueRepo.Delete(context.Background(), created.ID))
	_, err = s.issueRepo.Get(context.Background(), created.ID, repository.IssueDetailProjection())
	s.Assert().ErrorIs(err, repository.ErrNotFound)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 0)
}

func TestCachedIssueRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedIssueRepositoryIntegrationTestSuite))
}
