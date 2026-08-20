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

type SearchIndexIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	ctx context.Context
}

func (s *SearchIndexIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *SearchIndexIntegrationTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *SearchIndexIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *SearchIndexIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *SearchIndexIntegrationTestSuite) TestListSearchableRecords() {
	owner, err := s.UserRepo.Create(s.ctx, testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)
	issue, err := s.IssueRepo.Create(s.ctx, testModel.NewCreateIssueOpts(project.ID, owner.ID))
	s.Require().NoError(err)

	records, err := repository.ListSearchableRecords(
		s.ctx,
		s.Neo4jDB,
		model.ResourceTypeIssue,
		"",
		10,
	)
	s.Require().NoError(err)
	s.Require().Len(records, 1)
	s.Assert().Equal(issue.ID, records[0].ID)
	s.Assert().Equal(issue.Title, records[0].Title)
	s.Assert().Equal(issue.Description, records[0].Content)
	s.Assert().Equal(issue.Key, records[0].Key)
	s.Assert().Contains(records[0].Ancestry, issue.ID)
	s.Assert().Contains(records[0].Ancestry, project.ID)
	s.Assert().Contains(records[0].Ancestry, ns.ID)
	s.Assert().Contains(records[0].Ancestry, org.ID)

	byIDs, err := repository.ListSearchableRecordsByIDs(
		s.ctx,
		s.Neo4jDB,
		model.ResourceTypeIssue,
		[]model.ID{issue.ID},
	)
	s.Require().NoError(err)
	s.Require().Len(byIDs, 1)
	s.Assert().Equal(records[0].ID, byIDs[0].ID)
	s.Assert().Equal(records[0].Key, byIDs[0].Key)
}

func (s *SearchIndexIntegrationTestSuite) TestListSearchableRecordsPagination() {
	owner, err := s.UserRepo.Create(s.ctx, testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	first, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	second, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)

	page, err := repository.ListSearchableRecords(
		s.ctx,
		s.Neo4jDB,
		model.ResourceTypeOrganization,
		"",
		1,
	)
	s.Require().NoError(err)
	s.Require().Len(page, 1)

	rest, err := repository.ListSearchableRecords(
		s.ctx,
		s.Neo4jDB,
		model.ResourceTypeOrganization,
		page[0].ID.String(),
		10,
	)
	s.Require().NoError(err)
	s.Require().Len(rest, 1)
	s.Assert().NotEqual(page[0].ID, rest[0].ID)

	ids := []model.ID{page[0].ID, rest[0].ID}
	s.Assert().ElementsMatch([]model.ID{first.ID, second.ID}, ids)
}

func TestSearchIndexIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(SearchIndexIntegrationTestSuite))
}
