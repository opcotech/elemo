package repository

import (
	"testing"

	"github.com/opcotech/elemo/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueRelationListQuery_Compile(t *testing.T) {
	t.Parallel()

	issueID := model.MustNewID(model.ResourceTypeIssue)

	t.Run("root query uses start and end nodes", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(IssueRelationListQuery{
			IssueID: issueID,
			Page:    CursorPage{Size: 10},
		})
		require.NoError(t, err)
		assert.Equal(t, "issue.list_relations", plan.Root.Name)
		assert.Empty(t, plan.Loaders)
		assert.Contains(t, plan.Root.Cypher, "startNode(r).id AS source_id")
		assert.Contains(t, plan.Root.Cypher, "endNode(r).id AS target_id")
		assert.Contains(t, plan.Root.Cypher, "ORDER BY r.id DESC")
		assert.Equal(t, issueID.String(), plan.Root.Params["issue_id"])
		assert.Equal(t, 11, plan.Root.Params["limit"])
	})

	t.Run("invalid issue id", func(t *testing.T) {
		t.Parallel()

		_, err := CompileQuery(IssueRelationListQuery{
			IssueID: model.ID{},
			Page:    CursorPage{Size: 10},
		})
		require.Error(t, err)
	})

	t.Run("invalid page size", func(t *testing.T) {
		t.Parallel()

		_, err := CompileQuery(IssueRelationListQuery{
			IssueID: issueID,
			Page:    CursorPage{Size: MaxPageSize + 1},
		})
		require.ErrorIs(t, err, ErrInvalidPageSize)
	})
}

func TestIssueListForNamespaceQuery_Compile(t *testing.T) {
	t.Parallel()

	namespaceID := model.MustNewID(model.ResourceTypeNamespace)

	t.Run("root query matches issues across namespace projects", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(IssueListForNamespaceQuery{
			NamespaceID: namespaceID,
			Page:        CursorPage{Size: 10},
			Projection:  IssueListForNamespaceProjection(),
		})
		require.NoError(t, err)
		assert.Equal(t, "issue.list_for_namespace", plan.Root.Name)
		assert.Contains(t, plan.Root.Cypher, namespaceID.Label())
		assert.Contains(t, plan.Root.Cypher, "RETURN i, p, n, u")
		assert.Equal(t, namespaceID.String(), plan.Root.Params["namespace_id"])
		assert.Equal(t, 11, plan.Root.Params["limit"])
		require.Len(t, plan.Loaders, 3)
	})

	t.Run("invalid namespace id", func(t *testing.T) {
		t.Parallel()

		_, err := CompileQuery(IssueListForNamespaceQuery{
			NamespaceID: model.ID{},
			Page:        CursorPage{Size: 10},
		})
		require.Error(t, err)
	})
}

func TestIssueListForUserQuery_Compile(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)

	t.Run("root query matches issues assigned to the user", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(IssueListForUserQuery{
			UserID:     userID,
			Page:       CursorPage{Size: 10},
			Projection: IssueListForUserProjection(),
		})
		require.NoError(t, err)
		assert.Equal(t, "issue.list_for_user", plan.Root.Name)
		assert.Contains(t, plan.Root.Cypher, "ASSIGNED_TO")
		assert.Contains(t, plan.Root.Cypher, "{kind: $assignee_kind}")
		assert.Contains(t, plan.Root.Cypher, "RETURN i, p, n, u")
		assert.Equal(t, userID.String(), plan.Root.Params["user_id"])
		assert.Equal(t, model.AssignmentKindAssignee.String(), plan.Root.Params["assignee_kind"])
		assert.Equal(t, 11, plan.Root.Params["limit"])
		require.Len(t, plan.Loaders, 3)
	})

	t.Run("authz filter does not rebind the namespace alias", func(t *testing.T) {
		t.Parallel()

		actorID := model.MustNewID(model.ResourceTypeUser)
		plan, err := CompileQuery(IssueListForUserQuery{
			UserID:     userID,
			ActorID:    actorID,
			Action:     model.ActionIssueRead,
			Page:       CursorPage{Size: 10},
			Projection: IssueListForUserProjection(),
		})
		require.NoError(t, err)
		assert.Contains(t, plan.Root.Cypher, "EXISTS {")
		assert.Contains(t, plan.Root.Cypher, "ALL(authz_node IN nodes(path)")
		assert.NotContains(t, plan.Root.Cypher, "ALL(n IN")
		assert.Equal(t, userID.String(), plan.Root.Params["user_id"])
		assert.Equal(t, actorID.String(), plan.Root.Params["actor_id"])
	})

	t.Run("invalid user id", func(t *testing.T) {
		t.Parallel()

		_, err := CompileQuery(IssueListForUserQuery{
			UserID: model.ID{},
			Page:   CursorPage{Size: 10},
		})
		require.Error(t, err)
	})
}

func TestIssueListQuery_Compile(t *testing.T) {
	t.Parallel()

	projectID := model.MustNewID(model.ResourceTypeProject)

	t.Run("root query returns the reporter id", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(IssueListQuery{
			ProjectID:  projectID,
			Page:       CursorPage{Size: 10},
			Projection: IssueListForProjectProjection(),
		})
		require.NoError(t, err)
		assert.Equal(t, "issue.list_for_project", plan.Root.Name)
		assert.Contains(t, plan.Root.Cypher, "RETURN i, p, n, u")
		assert.Equal(t, projectID.String(), plan.Root.Params["project_id"])
		require.Len(t, plan.Loaders, 3)
		assert.Equal(t, "issue.load_parent", plan.Loaders[0].Name)
		assert.Contains(t, plan.Loaders[0].Cypher, "pp.key AS parent_project_key")
	})
}

func TestIssueListForIssueQuery_Compile(t *testing.T) {
	t.Parallel()

	issueID := model.MustNewID(model.ResourceTypeIssue)

	t.Run("optional creator match keeps related issues without a reporter", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(IssueListForIssueQuery{
			IssueID:    issueID,
			Page:       CursorPage{Size: 10},
			Projection: IssueDetailProjection(),
		})
		require.NoError(t, err)
		assert.Equal(t, "issue.list_for_issue", plan.Root.Name)
		assert.Contains(t, plan.Root.Cypher, "OPTIONAL MATCH (u:")
		assert.Contains(t, plan.Root.Cypher, "RETURN i, p, n, u")
		assert.Equal(t, issueID.String(), plan.Root.Params["issue_id"])
	})

	t.Run("invalid issue id", func(t *testing.T) {
		t.Parallel()

		_, err := CompileQuery(IssueListForIssueQuery{
			IssueID: model.ID{},
			Page:    CursorPage{Size: 10},
		})
		require.Error(t, err)
	})
}

func TestIssueGetQuery_Compile(t *testing.T) {
	t.Parallel()

	issueID := model.MustNewID(model.ResourceTypeIssue)

	t.Run("parent loader returns the parent project key", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(IssueGetQuery{
			ID:         issueID,
			Projection: IssueProjection{Parent: true},
		})
		require.NoError(t, err)
		require.Len(t, plan.Loaders, 1)
		assert.Equal(t, "issue.load_parent", plan.Loaders[0].Name)
		assert.Contains(t, plan.Loaders[0].Cypher, "pp.key AS parent_project_key")
	})

	t.Run("detail projection loads document count", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(IssueGetQuery{
			ID:         issueID,
			Projection: IssueDetailProjection(),
		})
		require.NoError(t, err)
		names := make([]string, 0, len(plan.Loaders))
		for _, loader := range plan.Loaders {
			names = append(names, loader.Name)
		}
		assert.Contains(t, names, "issue.load_document_count")
		assert.Contains(t, names, "issue.load_comment_count")
		for _, loader := range plan.Loaders {
			if loader.Name == "issue.load_document_count" {
				assert.Contains(t, loader.Cypher, "COUNT { (:Document)-[:RELATED_TO]->(i) } AS document_count")
			}
		}
	})
}

func TestIssueListQuery_CompileOmitsDocumentCount(t *testing.T) {
	t.Parallel()

	projectID := model.MustNewID(model.ResourceTypeProject)

	t.Run("list projection omits document count", func(t *testing.T) {
		t.Parallel()

		plan, err := CompileQuery(IssueListQuery{
			ProjectID:  projectID,
			Page:       CursorPage{Size: 10},
			Projection: IssueListForProjectProjection(),
		})
		require.NoError(t, err)
		for _, loader := range plan.Loaders {
			assert.NotEqual(t, "issue.load_document_count", loader.Name)
			assert.NotContains(t, loader.Cypher, "document_count")
		}
	})
}

func TestIssueWatchersQuery(t *testing.T) {
	t.Parallel()

	issueID := model.MustNewID(model.ResourceTypeIssue)

	t.Run("empty user projection omits document count", func(t *testing.T) {
		t.Parallel()

		root, err := IssueWatchersQuery(issueID)
		require.NoError(t, err)

		plan, err := compileUserRootQuery(userRootQueryInput{
			Name:       root.Name,
			Match:      root.Cypher,
			Params:     root.Params,
			Projection: UserProjection{},
		})
		require.NoError(t, err)
		assert.Equal(t, "issue.get_watchers", plan.Root.Name)
		assert.NotContains(t, plan.Root.Cypher, "document_count")
		assert.Empty(t, plan.Loaders)
	})

	t.Run("list projection over-fetches document count", func(t *testing.T) {
		t.Parallel()

		root, err := IssueWatchersQuery(issueID)
		require.NoError(t, err)

		plan, err := compileUserRootQuery(userRootQueryInput{
			Name:       root.Name,
			Match:      root.Cypher,
			Params:     root.Params,
			Projection: UserListProjection(),
		})
		require.NoError(t, err)
		assert.Contains(t, plan.Root.Cypher, "document_count")
	})
}

func TestIssueRelationByIDQuery(t *testing.T) {
	t.Parallel()

	relationID := model.MustNewID(model.ResourceTypeIssueRelation)
	query, err := IssueRelationByIDQuery(relationID)
	require.NoError(t, err)
	assert.Equal(t, "issue.get_relation", query.Name)
	assert.Contains(t, query.Cypher, "startNode(r).id AS source_id")
	assert.Contains(t, query.Cypher, "endNode(r).id AS target_id")
	assert.Equal(t, relationID.String(), query.Params["id"])
}
