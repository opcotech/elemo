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
	s.Assert().NotNil(issue.CreatedAt)
	s.Assert().Nil(issue.UpdatedAt)
}

func (s *IssueRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	issue, err := s.IssueRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, issue.ID)
	s.Assert().Equal(s.createOpts.NumericID, issue.NumericID)
	s.Assert().Equal(s.createOpts.Kind, issue.Kind)
	s.Assert().Equal(s.createOpts.Title, issue.Title)
	s.Assert().Equal(s.createOpts.Description, issue.Description)
	s.Assert().Equal(s.createOpts.Status, issue.Status)
	s.Assert().Equal(s.createOpts.Priority, issue.Priority)
	s.Assert().Equal(s.createOpts.ReportedBy, issue.ReportedBy)
	s.Assert().WithinDuration(*created.CreatedAt, *issue.CreatedAt, 100*time.Millisecond)
}

func (s *IssueRepositoryIntegrationTestSuite) TestGetAllForProject() {
	_, err := s.IssueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)

	issues, err := s.IssueRepo.GetAllForProject(context.Background(), s.testProject.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(issues, 3)

	issues, err = s.IssueRepo.GetAllForProject(context.Background(), s.testProject.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(issues, 2)
}

func (s *IssueRepositoryIntegrationTestSuite) TestGetAllForIssue() {
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

	issues, err := s.IssueRepo.GetAllForIssue(context.Background(), parent.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(issues, 2)
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
	s.Require().NoError(s.IssueRepo.RemoveWatcher(context.Background(), created.ID, created.ReportedBy))
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
	})
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
	_, err = s.IssueRepo.Get(context.Background(), created.ID)
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
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.issueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.IssueRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	usingCache, err := s.issueRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestGetAllForProject() {
	_, err := s.issueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.IssueRepo.GetAllForProject(context.Background(), s.testProject.ID, 0, 10)
	s.Require().NoError(err)
	usingCache, err := s.issueRepo.GetAllForProject(context.Background(), s.testProject.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.issueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	issue, err := s.issueRepo.Update(context.Background(), created.ID, repository.UpdateIssueOpts{
		Title: optional.Some("new title"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new title", issue.Title)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedIssueRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.issueRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.issueRepo.Get(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Require().NoError(s.issueRepo.Delete(context.Background(), created.ID))
	_, err = s.issueRepo.Get(context.Background(), created.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedIssueRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedIssueRepositoryIntegrationTestSuite))
}
