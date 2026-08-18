package repository

import (
	"testing"

	"github.com/opcotech/elemo/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationGetQuery_Compile(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)

	t.Run("detail projection counts documents", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(OrganizationGetQuery{
			ID:         orgID,
			Projection: OrganizationDetailProjection(),
		})
		require.NoError(t, err)
		assert.Equal(t, "organization.get", plan.Root.Name)
		assert.Contains(t, plan.Root.Cypher, "COUNT { (:Document)-[:SCOPED_TO]->(o) } AS document_count")
		assert.Contains(t, plan.Root.Cypher, "AS member_count")
		assert.Empty(t, plan.Loaders)
		assert.Equal(t, orgID.String(), plan.Root.Params["id"])
	})

	t.Run("empty projection omits document count", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(OrganizationGetQuery{
			ID:         orgID,
			Projection: OrganizationProjection{},
		})
		require.NoError(t, err)
		assert.NotContains(t, plan.Root.Cypher, "document_count")
		assert.NotContains(t, plan.Root.Cypher, "member_count")
	})

	t.Run("invalid organization id", func(t *testing.T) {
		t.Parallel()

		_, err := CompileQuery(OrganizationGetQuery{
			ID:         model.ID{},
			Projection: OrganizationDetailProjection(),
		})
		require.Error(t, err)
	})
}

func TestOrganizationListQuery_Compile(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)

	t.Run("list projection counts documents", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(OrganizationListQuery{
			UserID:     userID,
			Page:       CursorPage{Size: 10},
			Projection: OrganizationListProjection(),
		})
		require.NoError(t, err)
		assert.Equal(t, "organization.list", plan.Root.Name)
		assert.Contains(t, plan.Root.Cypher, "COUNT { (:Document)-[:SCOPED_TO]->(o) } AS document_count")
		assert.Equal(t, userID.String(), plan.Root.Params["user_id"])
		assert.Equal(t, 11, plan.Root.Params["limit"])
	})

	t.Run("invalid user id", func(t *testing.T) {
		t.Parallel()

		_, err := CompileQuery(OrganizationListQuery{
			UserID: model.ID{},
			Page:   CursorPage{Size: 10},
		})
		require.Error(t, err)
	})
}
