package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
)

func TestBuildSearchFilter(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	projectID := model.MustNewID(model.ResourceTypeProject)

	t.Run("empty type filters fail closed", func(t *testing.T) {
		t.Parallel()
		_, err := buildSearchFilter(SearchQuery{})
		assert.ErrorIs(t, err, ErrSearchFilter)
	})

	t.Run("empty scope ids omit the branch and fail closed", func(t *testing.T) {
		t.Parallel()
		_, err := buildSearchFilter(SearchQuery{
			TypeFilters: []SearchTypeFilter{{Type: "Issue"}},
		})
		assert.ErrorIs(t, err, ErrSearchFilter)
	})

	t.Run("single type with scopes", func(t *testing.T) {
		t.Parallel()
		got, err := buildSearchFilter(SearchQuery{
			TypeFilters: []SearchTypeFilter{{
				Type:     "Issue",
				ScopeIDs: []string{projectID.Composite()},
			}},
		})
		require.NoError(t, err)
		assert.Equal(t, `(type = "Issue" AND scope_ids IN ["`+projectID.Composite()+`"])`, got)
	})

	t.Run("client filters are anded onto authz", func(t *testing.T) {
		t.Parallel()
		got, err := buildSearchFilter(SearchQuery{
			TypeFilters: []SearchTypeFilter{{
				Type:     "Issue",
				ScopeIDs: []string{orgID.Composite()},
			}},
			OrganizationID: orgID.Composite(),
			ProjectID:      projectID.Composite(),
		})
		require.NoError(t, err)
		assert.Contains(t, got, `type = "Issue"`)
		assert.Contains(t, got, `organization_id = "`+orgID.Composite()+`"`)
		assert.Contains(t, got, `project_id = "`+projectID.Composite()+`"`)
		assert.NotContains(t, got, "AND type")
	})

	t.Run("quotes in values are escaped", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, `"foo\"bar"`, quoteFilterValue(`foo"bar`))
	})
}
