package repository_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
)

type SearchRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.SearchContainerIntegrationTestSuite
}

func (s *SearchRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupSearch(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *SearchRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupSearch(&s.ContainerIntegrationTestSuite)
}

func (s *SearchRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *SearchRepositoryIntegrationTestSuite) newDoc(id model.ID, title string, scopes ...model.ID) repository.SearchDocument {
	s.T().Helper()

	now := time.Now().Unix()
	doc := repository.SearchDocument{
		ID:        id.SearchKey(),
		Type:      id.Type.String(),
		Title:     title,
		ScopeIDs:  make([]string, 0, len(scopes)),
		CreatedAt: now,
		UpdatedAt: now,
	}
	for _, scope := range scopes {
		doc.ScopeIDs = append(doc.ScopeIDs, scope.Composite())
		switch scope.Type {
		case model.ResourceTypeOrganization:
			doc.OrganizationID = scope.Composite()
		case model.ResourceTypeNamespace:
			doc.NamespaceID = scope.Composite()
		case model.ResourceTypeProject:
			doc.ProjectID = scope.Composite()
		}
	}
	return doc
}

func (s *SearchRepositoryIntegrationTestSuite) search(text string, typ model.ResourceType, scopes []string, orgID, projectID string) *repository.SearchHits {
	s.T().Helper()

	hits, err := s.SearchRepo.Search(context.Background(), repository.SearchQuery{
		Text: text,
		TypeFilters: []repository.SearchTypeFilter{{
			Type:     typ.String(),
			ScopeIDs: scopes,
		}},
		OrganizationID: orgID,
		ProjectID:      projectID,
		Limit:          20,
	})
	s.Require().NoError(err)
	return hits
}

func (s *SearchRepositoryIntegrationTestSuite) hitIDs(hits *repository.SearchHits) []string {
	s.T().Helper()

	ids := make([]string, 0, len(hits.Documents))
	for _, doc := range hits.Documents {
		ids = append(ids, doc.ID)
	}
	return ids
}

func (s *SearchRepositoryIntegrationTestSuite) TestPing() {
	s.Require().NoError(s.SearchRepo.Ping(context.Background()))
}

func (s *SearchRepositoryIntegrationTestSuite) TestUpsertAndSearch() {
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	issueID := model.MustNewID(model.ResourceTypeIssue)
	doc := s.newDoc(issueID, "alpha unique search title", orgID, issueID)

	s.Require().NoError(s.SearchRepo.Upsert(context.Background(), doc))

	hits := s.search("alpha unique search title", model.ResourceTypeIssue, []string{orgID.Composite()}, "", "")
	s.Require().Len(hits.Documents, 1)
	s.Assert().Equal(issueID.SearchKey(), hits.Documents[0].ID)
	s.Assert().Equal("alpha unique search title", hits.Documents[0].Title)
	s.Assert().Equal(model.ResourceTypeIssue.String(), hits.Documents[0].Type)
	s.Assert().Equal(orgID.Composite(), hits.Documents[0].OrganizationID)
}

func (s *SearchRepositoryIntegrationTestSuite) TestUpsertUpdatesInPlace() {
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	issueID := model.MustNewID(model.ResourceTypeIssue)
	doc := s.newDoc(issueID, "alpha-before-token", orgID, issueID)

	s.Require().NoError(s.SearchRepo.Upsert(context.Background(), doc))

	doc.Title = "omega-after-token"
	s.Require().NoError(s.SearchRepo.Upsert(context.Background(), doc))

	updated := s.search("omega-after-token", model.ResourceTypeIssue, []string{orgID.Composite()}, "", "")
	s.Require().Len(updated.Documents, 1)
	s.Assert().Equal("omega-after-token", updated.Documents[0].Title)

	original := s.search("alpha-before-token", model.ResourceTypeIssue, []string{orgID.Composite()}, "", "")
	s.Assert().Empty(original.Documents)
}

func (s *SearchRepositoryIntegrationTestSuite) TestSearchFailClosed() {
	_, err := s.SearchRepo.Search(context.Background(), repository.SearchQuery{})
	s.Require().ErrorIs(err, repository.ErrSearchQuery)
	s.Require().ErrorIs(err, repository.ErrSearchFilter)

	_, err = s.SearchRepo.Search(context.Background(), repository.SearchQuery{
		TypeFilters: []repository.SearchTypeFilter{{Type: model.ResourceTypeIssue.String()}},
	})
	s.Require().ErrorIs(err, repository.ErrSearchQuery)
	s.Require().ErrorIs(err, repository.ErrSearchFilter)
}

func (s *SearchRepositoryIntegrationTestSuite) TestSearchClientFilters() {
	orgA := model.MustNewID(model.ResourceTypeOrganization)
	orgB := model.MustNewID(model.ResourceTypeOrganization)
	projectA := model.MustNewID(model.ResourceTypeProject)
	projectB := model.MustNewID(model.ResourceTypeProject)
	issueA := model.MustNewID(model.ResourceTypeIssue)
	issueB := model.MustNewID(model.ResourceTypeIssue)

	s.Require().NoError(s.SearchRepo.Upsert(context.Background(),
		s.newDoc(issueA, "client filter issue", orgA, projectA, issueA),
		s.newDoc(issueB, "client filter issue", orgB, projectB, issueB),
	))

	scopes := []string{orgA.Composite(), orgB.Composite()}
	hits := s.search("client filter issue", model.ResourceTypeIssue, scopes, orgA.Composite(), projectA.Composite())
	s.Require().Len(hits.Documents, 1)
	s.Assert().Equal(issueA.SearchKey(), hits.Documents[0].ID)
}

func (s *SearchRepositoryIntegrationTestSuite) TestDelete() {
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	issueID := model.MustNewID(model.ResourceTypeIssue)
	doc := s.newDoc(issueID, "delete by primary key", orgID, issueID)

	s.Require().NoError(s.SearchRepo.Upsert(context.Background(), doc))
	s.Require().NoError(s.SearchRepo.Delete(context.Background(), issueID.SearchKey()))

	hits := s.search("delete by primary key", model.ResourceTypeIssue, []string{orgID.Composite()}, "", "")
	s.Assert().Empty(hits.Documents)
}

func (s *SearchRepositoryIntegrationTestSuite) TestDeleteByScope() {
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	projectID := model.MustNewID(model.ResourceTypeProject)
	otherProjectID := model.MustNewID(model.ResourceTypeProject)
	inScope := model.MustNewID(model.ResourceTypeIssue)
	outOfScope := model.MustNewID(model.ResourceTypeIssue)

	s.Require().NoError(s.SearchRepo.Upsert(context.Background(),
		s.newDoc(inScope, "scoped descendant", orgID, projectID, inScope),
		s.newDoc(outOfScope, "unrelated issue", orgID, otherProjectID, outOfScope),
	))

	s.Require().NoError(s.SearchRepo.DeleteByScope(context.Background(), projectID.Composite()))

	scopes := []string{orgID.Composite(), projectID.Composite(), otherProjectID.Composite()}
	hits := s.search("", model.ResourceTypeIssue, scopes, "", "")
	s.Assert().Equal([]string{outOfScope.SearchKey()}, s.hitIDs(hits))
}

func (s *SearchRepositoryIntegrationTestSuite) TestEmptyUpsertAndDelete() {
	s.Require().NoError(s.SearchRepo.Upsert(context.Background()))
	s.Require().NoError(s.SearchRepo.Delete(context.Background()))
}

func (s *SearchRepositoryIntegrationTestSuite) TestDeleteAll() {
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	issueID := model.MustNewID(model.ResourceTypeIssue)

	s.Require().NoError(s.SearchRepo.Upsert(context.Background(),
		s.newDoc(issueID, "wipe entire index", orgID, issueID),
	))
	s.Require().NoError(s.SearchRepo.DeleteAll(context.Background()))

	hits := s.search("wipe entire index", model.ResourceTypeIssue, []string{orgID.Composite()}, "", "")
	s.Assert().Empty(hits.Documents)
}

func TestSearchRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(SearchRepositoryIntegrationTestSuite))
}
