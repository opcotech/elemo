package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func newSearchServiceForTest(perm PermissionService, repo repository.SearchRepository) *searchService {
	return &searchService{
		baseService: &baseService{
			logger:            log.DefaultLogger(),
			tracer:            tracing.NoopTracer(),
			permissionService: perm,
		},
		searchRepo: repo,
	}
}

func TestSearchService_Search(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	issueID := model.MustNewID(model.ResourceTypeIssue)
	projectID := model.MustNewID(model.ResourceTypeProject)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	t.Run("empty grant scopes return without querying", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		perm := NewMockPermissionService(ctrl)
		repo := repository.NewMockSearchRepository(ctrl)
		perm.EXPECT().ListGrantScopes(gomock.Any(), userID, model.ActionIssueRead).Return([]model.ID{}, nil)
		perm.EXPECT().ListGrantScopes(gomock.Any(), userID, model.ActionProjectRead).Return([]model.ID{}, nil)
		perm.EXPECT().ListGrantScopes(gomock.Any(), userID, model.ActionOrganizationRead).Return([]model.ID{}, nil)
		perm.EXPECT().ListGrantScopes(gomock.Any(), userID, model.ActionNamespaceRead).Return([]model.ID{}, nil)
		perm.EXPECT().ListGrantScopes(gomock.Any(), userID, model.ActionDocumentRead).Return([]model.ID{}, nil)

		page, err := newSearchServiceForTest(perm, repo).Search(ctx, SearchQuery{})
		require.NoError(t, err)
		assert.Empty(t, page.Items)
		assert.False(t, page.PageInfo.HasMore)
		assert.Nil(t, page.PageInfo.TotalCount)
	})

	t.Run("missing user fails closed", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		perm := NewMockPermissionService(ctrl)
		repo := repository.NewMockSearchRepository(ctrl)
		_, err := newSearchServiceForTest(perm, repo).Search(context.Background(), SearchQuery{})
		assert.ErrorIs(t, err, ErrNoUser)
	})

	t.Run("post filter drops unauthorized hits and omits totals", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		perm := NewMockPermissionService(ctrl)
		repo := repository.NewMockSearchRepository(ctrl)
		perm.EXPECT().ListGrantScopes(gomock.Any(), userID, model.ActionIssueRead).Return([]model.ID{projectID}, nil)
		repo.EXPECT().Search(gomock.Any(), gomock.Any()).Return(&repository.SearchHits{
			Documents: []repository.SearchDocument{{
				ID:        issueID.SearchKey(),
				Type:      model.ResourceTypeIssue.String(),
				Title:     "secret",
				Key:       "ELE-1",
				ScopeIDs:  []string{projectID.Composite()},
				CreatedAt: time.Now().Unix(),
			}},
			Limit: 20,
		}, nil)
		perm.EXPECT().Has(gomock.Any(), userID, issueID, model.ActionIssueRead).Return(false, nil)

		page, err := newSearchServiceForTest(perm, repo).Search(ctx, SearchQuery{
			Types: []model.ResourceType{model.ResourceTypeIssue},
		})
		require.NoError(t, err)
		assert.Empty(t, page.Items)
		assert.Nil(t, page.PageInfo.TotalCount)
	})

	t.Run("client organization filter is anded onto authz", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		perm := NewMockPermissionService(ctrl)
		repo := repository.NewMockSearchRepository(ctrl)
		perm.EXPECT().ListGrantScopes(gomock.Any(), userID, model.ActionIssueRead).Return([]model.ID{orgID}, nil)
		repo.EXPECT().Search(gomock.Any(), gomock.AssignableToTypeOf(repository.SearchQuery{})).DoAndReturn(
			func(_ context.Context, q repository.SearchQuery) (*repository.SearchHits, error) {
				assert.Equal(t, orgID.Composite(), q.OrganizationID)
				assert.NotEmpty(t, q.TypeFilters)
				assert.Equal(t, "Issue", q.TypeFilters[0].Type)
				assert.Contains(t, q.TypeFilters[0].ScopeIDs, orgID.Composite())
				return &repository.SearchHits{Documents: []repository.SearchDocument{}}, nil
			},
		)

		page, err := newSearchServiceForTest(perm, repo).Search(ctx, SearchQuery{
			Types:          []model.ResourceType{model.ResourceTypeIssue},
			OrganizationID: &orgID,
		})
		require.NoError(t, err)
		assert.Empty(t, page.Items)
		assert.Nil(t, page.PageInfo.TotalCount)
	})
}

func TestSearchService_Index(t *testing.T) {
	t.Parallel()

	issueID := model.MustNewID(model.ResourceTypeIssue)
	projectID := model.MustNewID(model.ResourceTypeProject)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	now := time.Now().UTC().Truncate(time.Second)

	ctrl := gomock.NewController(t)
	perm := NewMockPermissionService(ctrl)
	repo := repository.NewMockSearchRepository(ctrl)
	perm.EXPECT().ListScopeAncestry(gomock.Any(), issueID).Return([]model.ID{issueID, projectID, orgID}, nil)
	repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, docs ...repository.SearchDocument) error {
			require.Len(t, docs, 1)
			assert.Equal(t, issueID.SearchKey(), docs[0].ID)
			assert.Contains(t, docs[0].ScopeIDs, issueID.Composite())
			assert.Contains(t, docs[0].ScopeIDs, projectID.Composite())
			assert.Equal(t, orgID.Composite(), docs[0].OrganizationID)
			assert.Equal(t, projectID.Composite(), docs[0].ProjectID)
			return nil
		},
	)

	err := newSearchServiceForTest(perm, repo).Index(context.Background(), IndexInput{
		ID:        issueID,
		Title:     "Bug",
		Key:       "ELE-1",
		CreatedAt: &now,
	})
	require.NoError(t, err)
}

func TestNewSearchService(t *testing.T) {
	t.Parallel()

	t.Run("requires search repository", func(t *testing.T) {
		t.Parallel()
		_, err := NewSearchService(nil, WithPermissionService(NewMockPermissionService(gomock.NewController(t))))
		assert.ErrorIs(t, err, ErrNoSearchRepository)
	})

	t.Run("requires permission service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		_, err := NewSearchService(repository.NewMockSearchRepository(ctrl))
		assert.ErrorIs(t, err, ErrNoPermissionService)
	})
}

func TestSearchService_Reindex(t *testing.T) {
	t.Parallel()

	t.Run("nil database", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc := newSearchServiceForTest(
			NewMockPermissionService(ctrl),
			repository.NewMockSearchRepository(ctrl),
		)

		err := svc.Reindex(context.Background(), SearchReindexSources{})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSearchReindex)
		assert.ErrorIs(t, err, repository.ErrNoDriver)
	})

	t.Run("skips nil sources", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		logger := mock.NewMockLogger(ctrl)
		logger.EXPECT().Warn(gomock.Any(), "skipping nil search reindex source", gomock.Any()).Times(5)

		svc := newSearchServiceForTest(
			NewMockPermissionService(ctrl),
			repository.NewMockSearchRepository(ctrl),
		)
		svc.logger = logger

		err := svc.Reindex(context.Background(), SearchReindexSources{
			DB: &repository.Neo4jDatabase{},
		})
		require.NoError(t, err)
	})
}
