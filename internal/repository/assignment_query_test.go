package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
)

func TestAssignmentListByUserQuery_Compile(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)

	t.Run("page 2 uses assignment cursor alias", func(t *testing.T) {
		t.Parallel()

		assignmentID := model.MustNewID(model.ResourceTypeAssignment)
		token, err := EncodeCursor(assignmentID)
		require.NoError(t, err)

		plan, err := CompileQuery(AssignmentListByUserQuery{
			UserID: userID,
			Page:   CursorPage{Size: 10, Token: &token},
		})
		require.NoError(t, err)
		assert.Equal(t, "assignment.list_by_user", plan.Root.Name)
		assert.Contains(t, plan.Root.Cypher, "a.id < $cursor_id")
		assert.NotContains(t, plan.Root.Cypher, "u.id <")
		assert.NotContains(t, plan.Root.Params, "cursor_created_at")
		assert.Equal(t, assignmentID.String(), plan.Root.Params["cursor_id"])
		assert.Equal(t, 11, plan.Root.Params["limit"])
	})
}
