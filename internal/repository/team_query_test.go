package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
)

func TestTeamGetQuery_Compile(t *testing.T) {
	t.Parallel()

	teamID := model.MustNewID(model.ResourceTypeTeam)
	orgID := model.MustNewID(model.ResourceTypeOrganization)

	t.Run("root query matches team under parent", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(TeamGetQuery{
			ID:         teamID,
			BelongsTo:  orgID,
			Projection: TeamDetailProjection(),
		})
		require.NoError(t, err)
		assert.Equal(t, "team.get", plan.Root.Name)
		assert.Contains(t, plan.Root.Cypher, EdgeKindHasTeam.String())
		assert.Contains(t, plan.Root.Cypher, model.ResourceTypeTeam.String())
		assert.Contains(t, plan.Root.Cypher, "member_count")
		assert.Equal(t, teamID.String(), plan.Root.Params["id"])
		assert.Equal(t, orgID.String(), plan.Root.Params["belongs_to_id"])
	})

	t.Run("invalid team id", func(t *testing.T) {
		t.Parallel()

		_, err := CompileQuery(TeamGetQuery{
			ID:        model.ID{},
			BelongsTo: orgID,
		})
		require.Error(t, err)
	})

	t.Run("invalid belongs to id", func(t *testing.T) {
		t.Parallel()

		_, err := CompileQuery(TeamGetQuery{
			ID:        teamID,
			BelongsTo: model.ID{},
		})
		require.Error(t, err)
	})
}

func TestTeamListBelongsToQuery_Compile(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)

	t.Run("root query lists teams for parent", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(TeamListBelongsToQuery{
			BelongsTo:  orgID,
			Page:       CursorPage{Size: 10},
			Projection: TeamListProjection(),
		})
		require.NoError(t, err)
		assert.Equal(t, "team.list_belongs_to", plan.Root.Name)
		assert.Contains(t, plan.Root.Cypher, EdgeKindHasTeam.String())
		assert.Contains(t, plan.Root.Cypher, model.ResourceTypeTeam.String())
		assert.Contains(t, plan.Root.Cypher, "ORDER BY t.id")
		assert.Equal(t, orgID.String(), plan.Root.Params["id"])
		assert.Equal(t, 11, plan.Root.Params["limit"])
	})

	t.Run("invalid belongs to id", func(t *testing.T) {
		t.Parallel()

		_, err := CompileQuery(TeamListBelongsToQuery{
			BelongsTo: model.ID{},
			Page:      CursorPage{Size: 10},
		})
		require.Error(t, err)
	})

	t.Run("invalid page size", func(t *testing.T) {
		t.Parallel()

		_, err := CompileQuery(TeamListBelongsToQuery{
			BelongsTo: orgID,
			Page:      CursorPage{Size: MaxPageSize + 1},
		})
		require.ErrorIs(t, err, ErrInvalidPageSize)
	})
}

func TestTeamMemberListQuery_Compile(t *testing.T) {
	t.Parallel()

	teamID := model.MustNewID(model.ResourceTypeTeam)
	orgID := model.MustNewID(model.ResourceTypeOrganization)

	t.Run("root query lists members", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(TeamMemberListQuery{
			TeamID:    teamID,
			BelongsTo: orgID,
			Page:      CursorPage{Size: 10},
		})
		require.NoError(t, err)
		assert.Equal(t, "team.list_members", plan.Root.Name)
		assert.Contains(t, plan.Root.Cypher, EdgeKindHasTeam.String())
		assert.Contains(t, plan.Root.Cypher, EdgeKindMemberOf.String())
		assert.Equal(t, teamID.String(), plan.Root.Params["team_id"])
		assert.Equal(t, orgID.String(), plan.Root.Params["belongs_to_id"])
		assert.Equal(t, 11, plan.Root.Params["limit"])
	})

	t.Run("invalid team id", func(t *testing.T) {
		t.Parallel()

		_, err := CompileQuery(TeamMemberListQuery{
			TeamID:    model.ID{},
			BelongsTo: orgID,
			Page:      CursorPage{Size: 10},
		})
		require.Error(t, err)
	})
}
