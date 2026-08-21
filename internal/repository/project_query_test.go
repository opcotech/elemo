package repository

import (
	"testing"

	"github.com/opcotech/elemo/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectListQuery_Compile(t *testing.T) {
	t.Parallel()

	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	actorID := model.MustNewID(model.ResourceTypeUser)

	t.Run("covering grant skips per-row EXISTS", func(t *testing.T) {
		t.Parallel()
		plan, err := CompileQuery(ProjectListQuery{
			NamespaceID: namespaceID,
			ActorID:     actorID,
			Page:        CursorPage{Size: 10},
			Projection:  ProjectListProjection(),
		})
		require.NoError(t, err)
		assert.Equal(t, "project.list", plan.Root.Name)
		assert.NotContains(t, plan.Root.Cypher, "scope.id IN $scope_ids")
		assert.NotContains(t, plan.Root.Cypher, "GRANTED")
	})

	t.Run("narrow grants use scope_ids EXISTS", func(t *testing.T) {
		t.Parallel()
		scopeID := model.MustNewID(model.ResourceTypeProject)
		plan, err := CompileQuery(ProjectListQuery{
			NamespaceID: namespaceID,
			ActorID:     actorID,
			ScopeIDs:    []model.ID{scopeID},
			Page:        CursorPage{Size: 10},
			Projection:  ProjectListProjection(),
		})
		require.NoError(t, err)
		assert.Contains(t, plan.Root.Cypher, "scope.id IN $scope_ids")
		assert.Equal(t, []string{scopeID.String()}, plan.Root.Params["scope_ids"])
	})
}
