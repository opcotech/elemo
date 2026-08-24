package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
)

func TestNamespaceListAccessibleQuery_Compile(t *testing.T) {
	t.Parallel()

	actorID := model.MustNewID(model.ResourceTypeUser)

	t.Run("starts from grant scopes instead of every namespace", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(NamespaceListAccessibleQuery{
			ActorID:    actorID,
			Page:       CursorPage{Size: 10},
			Projection: NamespaceListProjection(),
		})
		require.NoError(t, err)
		assert.Equal(t, "namespace.list_accessible", plan.Root.Name)
		assert.Contains(t, plan.Root.Cypher, "GRANTED")
		assert.Contains(t, plan.Root.Cypher, "IN_SCOPE_OF*0..4")
		assert.NotContains(t, plan.Root.Cypher, "MATCH (ns:Namespace)")
		assert.NotContains(t, plan.Root.Cypher, "authz_issue:Issue")
		assert.Equal(t, actorID.String(), plan.Root.Params["user_id"])
		assert.Equal(t, namespaceReachableActions(), plan.Root.Params["reachable_actions"])
	})

	t.Run("invalid actor id", func(t *testing.T) {
		t.Parallel()

		_, err := CompileQuery(NamespaceListAccessibleQuery{
			ActorID: model.ID{},
			Page:    CursorPage{Size: 10},
		})
		require.Error(t, err)
	})
}

func TestNamespaceListQuery_Compile(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	actorID := model.MustNewID(model.ResourceTypeUser)

	t.Run("intersects grant-reachable namespaces with the organization", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(NamespaceListQuery{
			OrgID:      orgID,
			ActorID:    actorID,
			Page:       CursorPage{Size: 10},
			Projection: NamespaceListProjection(),
		})
		require.NoError(t, err)
		assert.Equal(t, "namespace.list", plan.Root.Name)
		assert.Contains(t, plan.Root.Cypher, "GRANTED")
		assert.Contains(t, plan.Root.Cypher, "HAS_NAMESPACE")
		assert.NotContains(t, plan.Root.Cypher, "authz_issue:Issue")
		assert.Equal(t, orgID.String(), plan.Root.Params["org_id"])
		assert.Equal(t, actorID.String(), plan.Root.Params["user_id"])
	})
}
