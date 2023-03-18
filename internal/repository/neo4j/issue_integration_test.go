package neo4j_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type IssueRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser      *model.User
	testOrg       *model.Organization
	testNamespace *model.Namespace
	testProject   *model.Project
	issue         *model.Issue
}

func (s *IssueRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *IssueRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.testNamespace = testModel.NewNamespace()
	s.Require().NoError(s.NamespaceRepo.Create(context.Background(), s.testOrg.ID, s.testNamespace))

	s.testProject = testModel.NewProject()
	s.Require().NoError(s.ProjectRepo.Create(context.Background(), s.testNamespace.ID, s.testProject))

	s.issue = testModel.NewIssue(s.testUser.ID)
}

func (s *IssueRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *IssueRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *IssueRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeIssue), s.issue.ID)
	s.Assert().NotNil(s.issue.CreatedAt)
	s.Assert().Nil(s.issue.UpdatedAt)
}

func (s *IssueRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	issue, err := s.IssueRepo.Get(context.Background(), s.issue.ID)
	s.Require().NoError(err)

	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeIssue), issue.ID)
	s.Assert().Equal(s.issue.NumericID, issue.NumericID)
	s.Assert().Equal(s.issue.Parent, issue.Parent)
	s.Assert().Equal(s.issue.Kind, issue.Kind)
	s.Assert().Equal(s.issue.Title, issue.Title)
	s.Assert().Equal(s.issue.Description, issue.Description)
	s.Assert().Equal(s.issue.Status, issue.Status)
	s.Assert().Equal(s.issue.Priority, issue.Priority)
	s.Assert().Equal(s.issue.Resolution, issue.Resolution)
	s.Assert().Equal(s.issue.ReportedBy, issue.ReportedBy)
	s.Assert().Equal(s.issue.Assignees, issue.Assignees)
	s.Assert().Equal(s.issue.Labels, issue.Labels)
	s.Assert().Equal(s.issue.Comments, issue.Comments)
	s.Assert().Equal(s.issue.Attachments, issue.Attachments)
	s.Assert().ElementsMatch([]model.ID{issue.ReportedBy}, issue.Watchers)
	s.Assert().Equal(s.issue.Relations, issue.Relations)
	s.Assert().Equal(s.issue.Links, issue.Links)
	s.Assert().WithinDuration(*s.issue.DueDate, *issue.DueDate, 100*time.Millisecond)
	s.Assert().WithinDuration(*s.issue.CreatedAt, *issue.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(issue.UpdatedAt)
}

func (s *IssueRepositoryIntegrationTestSuite) TestAddWatcher() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	watcher := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), watcher))

	s.Require().NoError(s.IssueRepo.AddWatcher(context.Background(), s.issue.ID, watcher.ID))
}

func (s *IssueRepositoryIntegrationTestSuite) TestGetWatchers() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	watchers, err := s.IssueRepo.GetWatchers(context.Background(), s.issue.ID)
	s.Require().NoError(err)
	s.Assert().Len(watchers, 1)

	watcher := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), watcher))
	s.Require().NoError(s.IssueRepo.AddWatcher(context.Background(), s.issue.ID, watcher.ID))

	watchers, err = s.IssueRepo.GetWatchers(context.Background(), s.issue.ID)
	s.Require().NoError(err)
	s.Assert().Len(watchers, 2)
}

func (s *IssueRepositoryIntegrationTestSuite) TestRemoveWatcher() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	s.Require().NoError(s.IssueRepo.RemoveWatcher(context.Background(), s.issue.ID, s.issue.ReportedBy))

	watchers, err := s.IssueRepo.GetWatchers(context.Background(), s.issue.ID)
	s.Require().NoError(err)
	s.Assert().Empty(watchers)
}

func (s *IssueRepositoryIntegrationTestSuite) TestAddRelation() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	relatedIssue := testModel.NewIssue(s.testUser.ID)
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, relatedIssue))

	relation, err := model.NewIssueRelation(s.issue.ID, relatedIssue.ID, model.IssueRelationKindBlocks)
	s.Require().NoError(err)
	s.Require().NoError(s.IssueRepo.AddRelation(context.Background(), relation))
}

func (s *IssueRepositoryIntegrationTestSuite) TestGetRelations() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	relatedIssue := testModel.NewIssue(s.testUser.ID)
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, relatedIssue))

	relation, err := model.NewIssueRelation(s.issue.ID, relatedIssue.ID, model.IssueRelationKindBlocks)
	s.Require().NoError(err)
	s.Require().NoError(s.IssueRepo.AddRelation(context.Background(), relation))

	relations, err := s.IssueRepo.GetRelations(context.Background(), s.issue.ID)
	s.Require().NoError(err)
	s.Assert().Len(relations, 1)
}

func (s *IssueRepositoryIntegrationTestSuite) TestRemoveRelation() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	relatedIssue := testModel.NewIssue(s.testUser.ID)
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, relatedIssue))

	relation, err := model.NewIssueRelation(s.issue.ID, relatedIssue.ID, model.IssueRelationKindBlocks)
	s.Require().NoError(err)
	s.Require().NoError(s.IssueRepo.AddRelation(context.Background(), relation))

	s.Require().NoError(s.IssueRepo.RemoveRelation(context.Background(), s.issue.ID, relatedIssue.ID, model.IssueRelationKindBlocks))
}

func (s *IssueRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	dueDate := time.Now().Add(1 * time.Hour)
	patch := map[string]any{
		"title":       "New title",
		"description": "New description",
		"kind":        model.IssueKindBug.String(),
		"status":      model.IssueStatusClosed.String(),
		"priority":    model.IssuePriorityHigh.String(),
		"resolution":  model.IssueResolutionFixed.String(),
		"due_date":    dueDate.Format(time.RFC3339Nano),
	}

	issue, err := s.IssueRepo.Update(context.Background(), s.issue.ID, patch)
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
}

func (s *IssueRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.IssueRepo.Create(context.Background(), s.testProject.ID, s.issue))

	s.Require().NoError(s.IssueRepo.Delete(context.Background(), s.issue.ID))

	_, err := s.IssueRepo.Get(context.Background(), s.issue.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestIssueRepositoryIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(IssueRepositoryIntegrationTestSuite))
}
