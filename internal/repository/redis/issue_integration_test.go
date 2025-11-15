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

type CachedIssueRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser      *model.User
	testOrg       *model.Organization
	testNamespace *model.Namespace
	testProject   *model.Project
	issue         *model.Issue
	issueRepo     *redis.CachedIssueRepository
}

func (s *CachedIssueRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.issueRepo, _ = redis.NewCachedIssueRepository(s.IssueRepo, redis.WithDatabase(s.RedisDB))
}

func (s *CachedIssueRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.testNamespace = testModel.NewNamespace()
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testUser.ID, s.testOrg.ID, s.testNamespace))

	s.testProject = testModel.NewProject()
	s.Require().NoError(s.ProjectRepo.Create(context.Background(), s.testNamespace.ID, s.testProject))

	s.issue = testModel.NewIssue(s.testUser.ID)

	s.Require().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.issueRepo.Create(context.Background(), s.testProject.ID, s.issue))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeIssue), s.issue.ID)
	s.Assert().NotNil(s.issue.CreatedAt)
	s.Assert().Nil(s.issue.UpdatedAt)

	s.issue.Parent = nil
	s.Require().NoError(s.issueRepo.Create(context.Background(), s.testProject.ID, s.issue))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeIssue), s.issue.ID)
	s.Assert().NotNil(s.issue.CreatedAt)
	s.Assert().Nil(s.issue.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	original, err := s.IssueRepo.Get(context.Background(), s.issue.ID)
	s.Require().NoError(err)

	usingCache, err := s.issueRepo.Get(context.Background(), s.issue.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cached, err := s.issueRepo.Get(context.Background(), s.issue.ID)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache.ID, cached.ID)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestGetAllForProject() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, testModel.NewIssue(s.testUser.ID)))

	originalIssues, err := s.IssueRepo.GetAllForProject(context.Background(), s.testProject.ID, 0, 10)
	s.Require().NoError(err)

	usingCacheIssues, err := s.issueRepo.GetAllForProject(context.Background(), s.testProject.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(originalIssues, usingCacheIssues)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedIssues, err := s.issueRepo.GetAllForProject(context.Background(), s.testProject.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(len(usingCacheIssues), len(cachedIssues))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestGetAllForIssue() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	relatedIssue1 := testModel.NewIssue(s.testUser.ID)
	relatedIssue1.Parent = &s.issue.ID
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, relatedIssue1))

	relatedIssue2 := testModel.NewIssue(s.testUser.ID)
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, relatedIssue2))

	relation, err := model.NewIssueRelation(s.issue.ID, relatedIssue2.ID, model.IssueRelationKindBlocks)
	s.Require().NoError(err)

	s.Require().NoError(s.IssueRepo.AddRelation(context.Background(), relation))

	originalIssues, err := s.IssueRepo.GetAllForIssue(context.Background(), s.issue.ID, 0, 10)
	s.Require().NoError(err)

	usingCacheIssues, err := s.issueRepo.GetAllForIssue(context.Background(), s.issue.ID, 0, 10)
	s.Require().NoError(err)

	s.Assert().Equal(originalIssues, usingCacheIssues)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedIssues, err := s.issueRepo.GetAllForIssue(context.Background(), s.issue.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(len(usingCacheIssues), len(cachedIssues))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestAddWatcher() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	watcher := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), watcher))

	s.Require().NoError(s.issueRepo.AddWatcher(context.Background(), s.issue.ID, watcher.ID))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestGetWatchers() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	watcher := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), watcher))

	s.Require().NoError(s.issueRepo.AddWatcher(context.Background(), s.issue.ID, watcher.ID))
	watchers, err := s.issueRepo.GetWatchers(context.Background(), s.issue.ID)
	s.Require().NoError(err)
	s.Assert().Len(watchers, 2)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestRemoveWatcher() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	watcher := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), watcher))

	s.Require().NoError(s.issueRepo.AddWatcher(context.Background(), s.issue.ID, watcher.ID))
	s.Require().NoError(s.issueRepo.RemoveWatcher(context.Background(), s.issue.ID, watcher.ID))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestAddRelation() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	relatedIssue := testModel.NewIssue(s.testUser.ID)
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, relatedIssue))

	relation, err := model.NewIssueRelation(s.issue.ID, relatedIssue.ID, model.IssueRelationKindBlocks)
	s.Require().NoError(err)

	s.Require().NoError(s.issueRepo.AddRelation(context.Background(), relation))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestGetRelations() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	relatedIssue := testModel.NewIssue(s.testUser.ID)
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, relatedIssue))

	relation, err := model.NewIssueRelation(s.issue.ID, relatedIssue.ID, model.IssueRelationKindBlocks)
	s.Require().NoError(err)

	s.Require().NoError(s.issueRepo.AddRelation(context.Background(), relation))

	relations, err := s.issueRepo.GetRelations(context.Background(), s.issue.ID)
	s.Require().NoError(err)
	s.Assert().Len(relations, 1)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestRemoveRelation() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	relatedIssue := testModel.NewIssue(s.testUser.ID)
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, relatedIssue))

	relation, err := model.NewIssueRelation(s.issue.ID, relatedIssue.ID, model.IssueRelationKindBlocks)
	s.Require().NoError(err)

	s.Require().NoError(s.issueRepo.AddRelation(context.Background(), relation))
	s.Require().NoError(s.issueRepo.RemoveRelation(context.Background(), s.issue.ID, relatedIssue.ID, model.IssueRelationKindBlocks))

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	dueDate := time.Now().UTC().Add(1 * time.Hour)
	patch := map[string]any{
		"title":       "New title",
		"description": "New description",
		"kind":        model.IssueKindBug.String(),
		"status":      model.IssueStatusClosed.String(),
		"priority":    model.IssuePriorityHigh.String(),
		"resolution":  model.IssueResolutionFixed.String(),
		"due_date":    dueDate.Format(time.RFC3339Nano),
	}

	issue, err := s.issueRepo.Update(context.Background(), s.issue.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(s.issue.ID, issue.ID)
	s.Assert().Equal(patch["kind"], issue.Kind.String())
	s.Assert().Equal(patch["title"], issue.Title)
	s.Assert().Equal(patch["description"], issue.Description)
	s.Assert().Equal(patch["status"], issue.Status.String())
	s.Assert().Equal(patch["priority"], issue.Priority.String())
	s.Assert().Equal(patch["resolution"], issue.Resolution.String())
	s.Assert().Equal(s.issue.ReportedBy, issue.ReportedBy)
	s.Assert().Equal(s.issue.Assignees, issue.Assignees)
	s.Assert().Equal(s.issue.Labels, issue.Labels)
	s.Assert().Equal(s.issue.Comments, issue.Comments)
	s.Assert().Equal(s.issue.Attachments, issue.Attachments)
	s.Assert().ElementsMatch([]model.ID{issue.ReportedBy}, issue.Watchers)
	s.Assert().Equal(s.issue.Relations, issue.Relations)
	s.Assert().Equal(s.issue.Links, issue.Links)
	s.Assert().WithinDuration(dueDate, *issue.DueDate, 100*time.Millisecond)
	s.Assert().WithinDuration(*s.issue.CreatedAt, *issue.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(issue.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	_, err := s.issueRepo.Get(context.Background(), s.issue.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.issueRepo.Delete(context.Background(), s.issue.ID))

	_, err = s.issueRepo.Get(context.Background(), s.issue.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedIssueRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedIssueRepositoryIntegrationTestSuite))
}
