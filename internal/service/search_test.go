package service_test

import (
	"context"
	"testing"
	"time"

	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
)

func newSearchServiceForTest(perm service.PermissionService, repo repository.SearchRepository) service.SearchService {
	return func() service.SearchService {
		svc, err := service.NewSearchService(
			repo,
			perm,
			nil,
			service.WithLogger(log.DefaultLogger()),
			service.WithTracer(tracing.NoopTracer()),
		)
		if err != nil {
			panic(err)
		}
		return svc
	}()
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
		perm := mocksvc.NewMockPermissionService(ctrl)
		repo := mockrepo.NewMockSearchRepository(ctrl)
		perm.EXPECT().ListGrantScopes(gomock.Any(), userID, model.ActionIssueRead).Return([]model.ID{}, nil)
		perm.EXPECT().ListGrantScopes(gomock.Any(), userID, model.ActionProjectRead).Return([]model.ID{}, nil)
		perm.EXPECT().ListGrantScopes(gomock.Any(), userID, model.ActionOrganizationRead).Return([]model.ID{}, nil)
		perm.EXPECT().ListGrantScopes(gomock.Any(), userID, model.ActionNamespaceRead).Return([]model.ID{}, nil)
		perm.EXPECT().ListGrantScopes(gomock.Any(), userID, model.ActionDocumentRead).Return([]model.ID{}, nil)

		page, err := newSearchServiceForTest(perm, repo).Search(ctx, service.SearchQuery{})
		require.NoError(t, err)
		assert.Empty(t, page.Items)
		assert.False(t, page.PageInfo.HasMore)
		assert.Nil(t, page.PageInfo.TotalCount)
	})

	t.Run("missing user fails closed", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		perm := mocksvc.NewMockPermissionService(ctrl)
		repo := mockrepo.NewMockSearchRepository(ctrl)
		_, err := newSearchServiceForTest(perm, repo).Search(context.Background(), service.SearchQuery{})
		assert.ErrorIs(t, err, service.ErrNoUser)
	})

	t.Run("post filter drops unauthorized hits and omits totals", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		perm := mocksvc.NewMockPermissionService(ctrl)
		repo := mockrepo.NewMockSearchRepository(ctrl)
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

		page, err := newSearchServiceForTest(perm, repo).Search(ctx, service.SearchQuery{
			Types: []model.ResourceType{model.ResourceTypeIssue},
		})
		require.NoError(t, err)
		assert.Empty(t, page.Items)
		assert.Nil(t, page.PageInfo.TotalCount)
	})

	t.Run("client organization filter is anded onto authz", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		perm := mocksvc.NewMockPermissionService(ctrl)
		repo := mockrepo.NewMockSearchRepository(ctrl)
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

		page, err := newSearchServiceForTest(perm, repo).Search(ctx, service.SearchQuery{
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
	perm := mocksvc.NewMockPermissionService(ctrl)
	repo := mockrepo.NewMockSearchRepository(ctrl)
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

	err := newSearchServiceForTest(perm, repo).Index(context.Background(), service.IndexInput{
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
		_, err := service.NewSearchService(nil, mocksvc.NewMockPermissionService(gomock.NewController(t)), nil)
		assert.ErrorIs(t, err, service.ErrNoSearchRepository)
	})

	t.Run("requires permission service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		_, err := service.NewSearchService(mockrepo.NewMockSearchRepository(ctrl), nil, nil)
		assert.ErrorIs(t, err, service.ErrNoPermissionService)
	})
}

func TestSearchService_Reindex(t *testing.T) {
	t.Parallel()

	t.Run("nil database", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc := newSearchServiceForTest(
			mocksvc.NewMockPermissionService(ctrl),
			mockrepo.NewMockSearchRepository(ctrl),
		)

		err := svc.Reindex(context.Background(), service.SearchReindexSources{}, service.SearchReindexOptions{})
		require.Error(t, err)
		assert.ErrorIs(t, err, service.ErrSearchReindex)
		assert.ErrorIs(t, err, repository.ErrNoDriver)
	})

	t.Run("upserts records in batches", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mockrepo.NewMockSearchRepository(ctrl)
		issueID := model.MustNewID(model.ResourceTypeIssue)
		projectID := model.MustNewID(model.ResourceTypeProject)

		svc := newSearchServiceForTest(mocksvc.NewMockPermissionService(ctrl), repo)
		service.SetSearchServiceListRecords(svc, func(
			_ context.Context,
			_ *repository.Neo4jDatabase,
			resourceType model.ResourceType,
			after string,
			limit int,
		) ([]repository.SearchableRecord, error) {
			assert.Equal(t, 2, limit)
			if resourceType != model.ResourceTypeIssue || after != "" {
				return []repository.SearchableRecord{}, nil
			}
			return []repository.SearchableRecord{{
				ID:       issueID,
				Title:    "Bug",
				Key:      "ELE-1",
				Ancestry: []model.ID{issueID, projectID},
			}}, nil
		})
		repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, docs ...repository.SearchDocument) error {
				require.Len(t, docs, 1)
				assert.Equal(t, issueID.SearchKey(), docs[0].ID)
				assert.Equal(t, "Bug", docs[0].Title)
				assert.Contains(t, docs[0].ScopeIDs, projectID.Composite())
				return nil
			},
		)

		err := svc.Reindex(context.Background(), service.SearchReindexSources{
			DB: &repository.Neo4jDatabase{},
		}, service.SearchReindexOptions{BatchSize: 2, Concurrency: 1})
		require.NoError(t, err)
	})

	t.Run("delete all runs before listing", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mockrepo.NewMockSearchRepository(ctrl)
		listed := false
		repo.EXPECT().DeleteAll(gomock.Any()).DoAndReturn(func(context.Context) error {
			assert.False(t, listed)
			return nil
		})

		svc := newSearchServiceForTest(mocksvc.NewMockPermissionService(ctrl), repo)
		service.SetSearchServiceListRecords(svc, func(
			_ context.Context,
			_ *repository.Neo4jDatabase,
			_ model.ResourceType,
			_ string,
			_ int,
		) ([]repository.SearchableRecord, error) {
			listed = true
			return []repository.SearchableRecord{}, nil
		})

		err := svc.Reindex(context.Background(), service.SearchReindexSources{
			DB: &repository.Neo4jDatabase{},
		}, service.SearchReindexOptions{DeleteAll: true, Concurrency: 1})
		require.NoError(t, err)
		assert.True(t, listed)
	})

	t.Run("skips delete all by default", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mockrepo.NewMockSearchRepository(ctrl)
		svc := newSearchServiceForTest(mocksvc.NewMockPermissionService(ctrl), repo)
		service.SetSearchServiceListRecords(svc, func(
			_ context.Context,
			_ *repository.Neo4jDatabase,
			_ model.ResourceType,
			_ string,
			_ int,
		) ([]repository.SearchableRecord, error) {
			return []repository.SearchableRecord{}, nil
		})

		err := svc.Reindex(context.Background(), service.SearchReindexSources{
			DB: &repository.Neo4jDatabase{},
		}, service.SearchReindexOptions{Concurrency: 1})
		require.NoError(t, err)
	})

	t.Run("upserts pages concurrently", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mockrepo.NewMockSearchRepository(ctrl)
		first := model.MustNewID(model.ResourceTypeIssue)
		second := model.MustNewID(model.ResourceTypeIssue)

		svc := newSearchServiceForTest(mocksvc.NewMockPermissionService(ctrl), repo)
		service.SetSearchServiceListRecords(svc, func(
			_ context.Context,
			_ *repository.Neo4jDatabase,
			resourceType model.ResourceType,
			after string,
			limit int,
		) ([]repository.SearchableRecord, error) {
			assert.Equal(t, 1, limit)
			if resourceType != model.ResourceTypeIssue {
				return []repository.SearchableRecord{}, nil
			}
			switch after {
			case "":
				return []repository.SearchableRecord{{ID: first, Title: "one", Ancestry: []model.ID{first}}}, nil
			case first.String():
				return []repository.SearchableRecord{{ID: second, Title: "two", Ancestry: []model.ID{second}}}, nil
			default:
				return []repository.SearchableRecord{}, nil
			}
		})
		repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).Times(2).Return(nil)

		err := svc.Reindex(context.Background(), service.SearchReindexSources{
			DB: &repository.Neo4jDatabase{},
		}, service.SearchReindexOptions{BatchSize: 1, Concurrency: 2})
		require.NoError(t, err)
	})

	t.Run("joins listing errors", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mockrepo.NewMockSearchRepository(ctrl)
		svc := newSearchServiceForTest(mocksvc.NewMockPermissionService(ctrl), repo)
		service.SetSearchServiceListRecords(svc, func(
			_ context.Context,
			_ *repository.Neo4jDatabase,
			_ model.ResourceType,
			_ string,
			_ int,
		) ([]repository.SearchableRecord, error) {
			return nil, assert.AnError
		})

		err := svc.Reindex(context.Background(), service.SearchReindexSources{
			DB: &repository.Neo4jDatabase{},
		}, service.SearchReindexOptions{Concurrency: 1})
		require.Error(t, err)
		assert.ErrorIs(t, err, service.ErrSearchReindex)
		assert.ErrorIs(t, err, assert.AnError)
	})
}
