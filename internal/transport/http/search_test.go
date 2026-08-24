package http

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func TestSearchController_V1SearchGet(t *testing.T) {
	t.Parallel()

	issueID := model.MustNewID(model.ResourceTypeIssue)
	now := time.Now().UTC().Truncate(time.Second)
	q := "bug"

	ctrl := gomock.NewController(t)
	searchSvc := mocksvc.NewMockSearchService(ctrl)
	searchSvc.EXPECT().Search(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, query service.SearchQuery) (service.Page[*service.SearchResult], error) {
			assert.Equal(t, "bug", query.Text)
			return service.Page[*service.SearchResult]{
				Items: []*service.SearchResult{{
					ID:        issueID,
					Type:      model.ResourceTypeIssue,
					Title:     "Bug",
					Key:       "ELE-1",
					CreatedAt: now,
				}},
				PageInfo: service.PageInfo{HasMore: false},
			}, nil
		},
	)

	c := &searchController{
		baseController: &baseController{
			logger: log.DefaultLogger(),
			tracer: tracing.NoopTracer(),
		},
		searchService: searchSvc,
	}

	resp, err := c.V1SearchGet(context.Background(), api.V1SearchGetRequestObject{
		Params: api.V1SearchGetParams{Q: &q},
	})
	require.NoError(t, err)
	got, ok := resp.(api.V1SearchGet200JSONResponse)
	require.True(t, ok)
	require.Len(t, got.Items, 1)
	assert.Equal(t, issueID.String(), got.Items[0].Id)
	assert.Equal(t, api.SearchResultTypeIssue, got.Items[0].Type)
	assert.False(t, got.PageInfo.HasMore)
	assert.Nil(t, got.PageInfo.TotalCount)
}

func TestNewSearchController(t *testing.T) {
	t.Parallel()

	t.Run("requires search service", func(t *testing.T) {
		t.Parallel()
		_, err := NewSearchController(nil)
		assert.ErrorIs(t, err, ErrNoSearchService)
	})
}
