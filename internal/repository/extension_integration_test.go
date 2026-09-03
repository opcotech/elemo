package repository_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type ExtensionRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser    *repository.User
	testOrg     *repository.Organization
	testNS      *repository.Namespace
	testProject *repository.Project
	testIssue   *repository.Issue
	pluginID    string
}

func (s *ExtensionRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *ExtensionRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.pluginID = "com.elemo.timetracking"
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.testNS, err = s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)
	s.testProject, err = s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(s.testNS.ID, s.testUser.ID))
	s.Require().NoError(err)
	s.testIssue, err = s.IssueRepo.Create(context.Background(), testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)
}

func (s *ExtensionRepositoryIntegrationTestSuite) TearDownTest() {
	s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *ExtensionRepositoryIntegrationTestSuite) TearDownSuite() {
	s.CleanupContainers()
}

func (s *ExtensionRepositoryIntegrationTestSuite) TestNodeCRUD() {
	ctx := context.Background()
	created, err := s.ExtensionRepo.Create(ctx, repository.CreateExtensionOpts{
		PluginID:   s.pluginID,
		Kind:       "TimeEntry",
		Parent:     s.testIssue.ID,
		Properties: map[string]any{"seconds": int64(12), "note": "first"},
	})
	s.Require().NoError(err)
	s.Equal(s.pluginID, created.PluginID)
	s.Equal("TimeEntry", created.Kind)
	s.Require().NotNil(created.Parent)
	s.Equal(s.testIssue.ID, *created.Parent)

	got, err := s.ExtensionRepo.Get(ctx, s.pluginID, created.ID)
	s.Require().NoError(err)
	s.Equal(created.ID, got.ID)
	s.Equal(int64(12), toInt64(got.Properties["seconds"]))

	updated, err := s.ExtensionRepo.Update(ctx, s.pluginID, created.ID, repository.UpdateExtensionOpts{
		Properties: map[string]any{"seconds": int64(90)},
	})
	s.Require().NoError(err)
	s.Equal(int64(90), toInt64(updated.Properties["seconds"]))

	listed, err := s.ExtensionRepo.List(ctx, repository.ListExtensionFilter{
		PluginID: s.pluginID,
		Kind:     "TimeEntry",
		Scope:    s.testIssue.ID,
		Page:     repository.CursorPage{Size: 10},
	})
	s.Require().NoError(err)
	s.Len(listed.Items, 1)

	otherIssue, err := s.IssueRepo.Create(ctx, testModel.NewCreateIssueOpts(s.testProject.ID, s.testUser.ID))
	s.Require().NoError(err)
	moved, err := s.ExtensionRepo.Move(ctx, repository.MoveExtensionOpts{
		PluginID: s.pluginID,
		ID:       created.ID,
		Parent:   otherIssue.ID,
	})
	s.Require().NoError(err)
	s.Require().NotNil(moved.Parent)
	s.Equal(otherIssue.ID, *moved.Parent)

	s.Require().NoError(s.ExtensionRepo.Delete(ctx, s.pluginID, created.ID))
	_, err = s.ExtensionRepo.Get(ctx, s.pluginID, created.ID)
	s.ErrorIs(err, repository.ErrNotFound)
}

func (s *ExtensionRepositoryIntegrationTestSuite) TestGetNotFound() {
	ctx := context.Background()
	_, err := s.ExtensionRepo.Get(ctx, s.pluginID, model.MustNewID(model.ResourceTypeExtension))
	s.ErrorIs(err, repository.ErrNotFound)
}

func (s *ExtensionRepositoryIntegrationTestSuite) TestRelationCRUD() {
	ctx := context.Background()
	ext, err := s.ExtensionRepo.Create(ctx, repository.CreateExtensionOpts{
		PluginID:   s.pluginID,
		Kind:       "TimeEntry",
		Parent:     s.testIssue.ID,
		Properties: map[string]any{"seconds": int64(30)},
	})
	s.Require().NoError(err)

	rel, err := s.ExtensionRepo.CreateRelation(ctx, repository.CreateExtensionRelationOpts{
		PluginID: s.pluginID,
		Kind:     "LOGGED_ON",
		From:     ext.ID,
		To:       s.testIssue.ID,
	})
	s.Require().NoError(err)
	s.NotEmpty(rel.ID)
	s.Equal("LOGGED_ON", rel.Kind)

	out, in, err := s.ExtensionRepo.CountRelations(ctx, s.pluginID, "LOGGED_ON", ext.ID, s.testIssue.ID)
	s.Require().NoError(err)
	s.Equal(int64(1), out)
	s.Equal(int64(1), in)

	listed, err := s.ExtensionRepo.ListRelations(
		ctx,
		s.pluginID,
		"LOGGED_ON",
		s.testIssue.ID,
		model.PluginGraphRelationDirectionIncoming,
		repository.CursorPage{Size: 10},
	)
	s.Require().NoError(err)
	s.Len(listed.Items, 1)
	s.Equal(rel.ID, listed.Items[0].ID)

	s.Require().NoError(s.ExtensionRepo.DeleteRelation(ctx, s.pluginID, rel.ID))
	out, in, err = s.ExtensionRepo.CountRelations(ctx, s.pluginID, "LOGGED_ON", ext.ID, s.testIssue.ID)
	s.Require().NoError(err)
	s.Equal(int64(0), out)
	s.Equal(int64(0), in)
}

func (s *ExtensionRepositoryIntegrationTestSuite) TestDeleteByPlugin() {
	ctx := context.Background()
	ext, err := s.ExtensionRepo.Create(ctx, repository.CreateExtensionOpts{
		PluginID:   s.pluginID,
		Kind:       "TimeEntry",
		Parent:     s.testIssue.ID,
		Properties: map[string]any{"seconds": int64(5)},
	})
	s.Require().NoError(err)
	_, err = s.ExtensionRepo.CreateRelation(ctx, repository.CreateExtensionRelationOpts{
		PluginID: s.pluginID,
		Kind:     "LOGGED_ON",
		From:     ext.ID,
		To:       s.testIssue.ID,
	})
	s.Require().NoError(err)

	s.Require().NoError(s.ExtensionRepo.DeleteByPlugin(ctx, s.pluginID))
	_, err = s.ExtensionRepo.Get(ctx, s.pluginID, ext.ID)
	s.ErrorIs(err, repository.ErrNotFound)
}

func (s *ExtensionRepositoryIntegrationTestSuite) TestCreateMissingParent() {
	ctx := context.Background()
	_, err := s.ExtensionRepo.Create(ctx, repository.CreateExtensionOpts{
		PluginID:   s.pluginID,
		Kind:       "TimeEntry",
		Parent:     model.MustNewID(model.ResourceTypeIssue),
		Properties: map[string]any{"seconds": int64(1)},
	})
	s.ErrorIs(err, repository.ErrExtensionParent)
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	default:
		return 0
	}
}

func TestExtensionRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ExtensionRepositoryIntegrationTestSuite))
}
