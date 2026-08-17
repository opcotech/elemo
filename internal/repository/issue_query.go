package repository

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type IssueProjection struct {
	Parent          bool
	Assignments     bool
	Labels          bool
	CommentCount    bool
	DocumentCount   bool
	AttachmentCount bool
	WatcherCount    bool
	RelationCount   bool
}

func IssueDetailProjection() IssueProjection {
	return IssueProjection{
		Parent:          true,
		Assignments:     true,
		Labels:          true,
		CommentCount:    true,
		DocumentCount:   true,
		AttachmentCount: true,
		WatcherCount:    true,
		RelationCount:   true,
	}
}

// IssueListProjection selects bounded relation loaders for issue list pages.
type IssueListProjection struct {
	Parent      bool
	Assignments bool
	Labels      bool
}

// IssueListForProjectProjection is the allowlisted projection for project issues.
func IssueListForProjectProjection() IssueListProjection {
	return IssueListProjection{
		Parent:      true,
		Assignments: true,
		Labels:      true,
	}
}

// IssueListForNamespaceProjection is the allowlisted projection for namespace issues.
func IssueListForNamespaceProjection() IssueListProjection {
	return IssueListForProjectProjection()
}

func (p IssueListProjection) issueProjection() IssueProjection {
	return IssueProjection{
		Parent:      p.Parent,
		Assignments: p.Assignments,
		Labels:      p.Labels,
	}
}

type IssueGetQuery struct {
	ID         model.ID
	Projection IssueProjection
}

type IssueGetByKeyQuery struct {
	NamespaceID model.ID
	IssueKey    string
	Projection  IssueProjection
}

type IssueListForIssueQuery struct {
	IssueID    model.ID
	Page       CursorPage
	Order      SortDirection
	Projection IssueProjection
}

// IssueListQuery compiles a cursor-paginated issue list for a project.
type IssueListQuery struct {
	ProjectID  model.ID
	Page       CursorPage
	Order      SortDirection
	Projection IssueListProjection
}

// IssueListForNamespaceQuery compiles a cursor-paginated issue list for a namespace.
type IssueListForNamespaceQuery struct {
	NamespaceID model.ID
	Page        CursorPage
	Order       SortDirection
	Projection  IssueListProjection
}

// IssueListForUserQuery compiles a cursor-paginated list of issues assigned to a user.
type IssueListForUserQuery struct {
	UserID     model.ID
	Page       CursorPage
	Order      SortDirection
	Projection IssueListProjection
}

// IssueListForUserProjection is the allowlisted projection for a user's assigned issues.
func IssueListForUserProjection() IssueListProjection {
	return IssueListForNamespaceProjection()
}

func (q IssueGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return compileIssueRootQuery(issueRootQueryInput{
		Root: CompiledQuery{
			Name: "issue.get",
			Cypher: `
				MATCH (i:` + q.ID.Label() + ` {id: $id})-[:` + EdgeKindBelongsTo.String() + `]->(p:` + model.ResourceTypeProject.String() + `)
				OPTIONAL MATCH (n:` + model.ResourceTypeNamespace.String() + `)-[:` + EdgeKindHasProject.String() + `]->(p)
				MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(i)
				RETURN i, p, n, u`,
			Params: map[string]any{"id": q.ID.String()},
		},
		Projection: q.Projection,
	})
}

func (q IssueGetByKeyQuery) Compile() (QueryPlan, error) {
	if err := q.NamespaceID.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return compileIssueRootQuery(issueRootQueryInput{
		Root: CompiledQuery{
			Name: "issue.get_by_key",
			Cypher: `
				MATCH (n:` + q.NamespaceID.Label() + ` {id: $namespace_id})-[:` + EdgeKindHasProject.String() + `]->(p:` + model.ResourceTypeProject.String() + `)
				MATCH (p)<-[:` + EdgeKindBelongsTo.String() + `]-(i:` + model.ResourceTypeIssue.String() + ` {numeric_id: toInteger(split($issue_key, "-")[1])})
				MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(i)
				WHERE $issue_key = p.key + "-" + toString(i.numeric_id)
				RETURN i, p, n, u`,
			Params: map[string]any{
				"namespace_id": q.NamespaceID.String(),
				"issue_key":    q.IssueKey,
			},
		},
		Projection: q.Projection,
	})
}

func (q IssueListForIssueQuery) Compile() (QueryPlan, error) {
	if err := q.IssueID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"issue_id": q.IssueID.String(),
	}
	bounds, err := compileCursorBounds("i", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	return compileIssueRootQuery(issueRootQueryInput{
		Root: CompiledQuery{
			Name: "issue.list_for_issue",
			Cypher: strings.TrimSpace(`
				MATCH (anchor:` + q.IssueID.Label() + ` {id: $issue_id})
				MATCH (i:` + model.ResourceTypeIssue.String() + `)-[:` + EdgeKindRelatedTo.String() + `]-(anchor)
				MATCH (i)-[:` + EdgeKindBelongsTo.String() + `]->(p:` + model.ResourceTypeProject.String() + `)
				OPTIONAL MATCH (n:` + model.ResourceTypeNamespace.String() + `)-[:` + EdgeKindHasProject.String() + `]->(p)
				OPTIONAL MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(i)
				` + cursorWherePrefix(bounds.Where, "WHERE ") + `
				RETURN i, p, n, u
				ORDER BY i.id ` + bounds.Order.Cypher() + `
				LIMIT $limit`),
			Params: params,
		},
		Projection: q.Projection,
	})
}

func (q IssueListQuery) Compile() (QueryPlan, error) {
	if err := q.ProjectID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"project_id": q.ProjectID.String(),
	}
	bounds, err := compileCursorBounds("i", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	root := CompiledQuery{
		Name: "issue.list_for_project",
		Cypher: `
		MATCH (p:` + q.ProjectID.Label() + ` {id: $project_id})<-[:` + EdgeKindBelongsTo.String() + `]-(i:` + model.ResourceTypeIssue.String() + `)
		WHERE true` + cursorWherePrefix(bounds.Where, " AND ") + `
		WITH p, i
		ORDER BY i.id ` + bounds.Order.Cypher() + `
		LIMIT $limit
		OPTIONAL MATCH (n:` + model.ResourceTypeNamespace.String() + `)-[:` + EdgeKindHasProject.String() + `]->(p)
		OPTIONAL MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(i)
		RETURN i, p, n, u`,
		Params: params,
	}

	plan := QueryPlan{Root: root, Loaders: issueRelationLoaders(q.Projection.issueProjection())}
	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}
	return plan, nil
}

func (q IssueListForNamespaceQuery) Compile() (QueryPlan, error) {
	if err := q.NamespaceID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"namespace_id": q.NamespaceID.String(),
	}
	bounds, err := compileCursorBounds("i", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	root := CompiledQuery{
		Name: "issue.list_for_namespace",
		Cypher: `
		MATCH (n:` + q.NamespaceID.Label() + ` {id: $namespace_id})-[:` + EdgeKindHasProject.String() + `]->(p:` + model.ResourceTypeProject.String() + `)<-[:` + EdgeKindBelongsTo.String() + `]-(i:` + model.ResourceTypeIssue.String() + `)
		WHERE true` + cursorWherePrefix(bounds.Where, " AND ") + `
		WITH n, p, i
		ORDER BY i.id ` + bounds.Order.Cypher() + `
		LIMIT $limit
		OPTIONAL MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(i)
		RETURN i, p, n, u`,
		Params: params,
	}

	plan := QueryPlan{Root: root, Loaders: issueRelationLoaders(q.Projection.issueProjection())}
	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}
	return plan, nil
}

func (q IssueListForUserQuery) Compile() (QueryPlan, error) {
	if err := q.UserID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"user_id":       q.UserID.String(),
		"assignee_kind": model.AssignmentKindAssignee.String(),
	}
	bounds, err := compileCursorBounds("i", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	root := CompiledQuery{
		Name: "issue.list_for_user",
		Cypher: `
		MATCH (assignee:` + q.UserID.Label() + ` {id: $user_id})-[a:` + EdgeKindAssignedTo.String() + ` {kind: $assignee_kind}]->(i:` + model.ResourceTypeIssue.String() + `)
		MATCH (i)-[:` + EdgeKindBelongsTo.String() + `]->(p:` + model.ResourceTypeProject.String() + `)
		OPTIONAL MATCH (n:` + model.ResourceTypeNamespace.String() + `)-[:` + EdgeKindHasProject.String() + `]->(p)
		WHERE true` + cursorWherePrefix(bounds.Where, " AND ") + `
		WITH n, p, i
		ORDER BY i.id ` + bounds.Order.Cypher() + `
		LIMIT $limit
		OPTIONAL MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(i)
		RETURN i, p, n, u`,
		Params: params,
	}

	plan := QueryPlan{Root: root, Loaders: issueRelationLoaders(q.Projection.issueProjection())}
	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}
	return plan, nil
}

type issueRootQueryInput struct {
	Root       CompiledQuery
	Projection IssueProjection
}

func compileIssueRootQuery(in issueRootQueryInput) (QueryPlan, error) {
	plan := QueryPlan{
		Root:    in.Root,
		Loaders: issueRelationLoaders(in.Projection),
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}

func issueRelationLoaders(proj IssueProjection) []CompiledQuery {
	loaders := make([]CompiledQuery, 0, 8)

	if proj.Parent {
		loaders = append(loaders, CompiledQuery{
			Name: "issue.load_parent",
			Cypher: `
				UNWIND $ids AS issue_id
				MATCH (i:` + model.ResourceTypeIssue.String() + ` {id: issue_id})
				OPTIONAL MATCH (i)-[:` + EdgeKindRelatedTo.String() + ` {kind: $parent_kind}]->(p:` + model.ResourceTypeIssue.String() + `)
				OPTIONAL MATCH (p)-[:` + EdgeKindBelongsTo.String() + `]->(pp:` + model.ResourceTypeProject.String() + `)
				RETURN issue_id, p AS parent, pp.key AS parent_project_key`,
			Params: map[string]any{"parent_kind": model.IssueRelationKindSubtaskOf.String()},
		})
	}

	if proj.Assignments {
		loaders = append(loaders, CompiledQuery{
			Name: "issue.load_assignments",
			Cypher: `
				UNWIND $ids AS issue_id
				MATCH (i:` + model.ResourceTypeIssue.String() + ` {id: issue_id})
				RETURN issue_id, ` + issueAssignmentsPattern("i") + ` AS assignments`,
			Params: map[string]any{},
		})
	}

	if proj.Labels {
		loaders = append(loaders, CompiledQuery{
			Name: "issue.load_labels",
			Cypher: `
				UNWIND $ids AS issue_id
				MATCH (i:` + model.ResourceTypeIssue.String() + ` {id: issue_id})
				OPTIONAL MATCH (i)-[:` + EdgeKindHasLabel.String() + `]->(l:` + model.ResourceTypeLabel.String() + `)
				RETURN issue_id, collect(DISTINCT l) AS labels`,
			Params: map[string]any{},
		})
	}

	if proj.CommentCount {
		loaders = append(loaders, CompiledQuery{
			Name: "issue.load_comment_count",
			Cypher: `
				UNWIND $ids AS issue_id
				MATCH (i:` + model.ResourceTypeIssue.String() + ` {id: issue_id})
				RETURN issue_id, COUNT { (i)-[:` + EdgeKindHasComment.String() + `]->(:` + model.ResourceTypeComment.String() + `) } AS comment_count`,
			Params: map[string]any{},
		})
	}

	if proj.DocumentCount {
		loaders = append(loaders, CompiledQuery{
			Name: "issue.load_document_count",
			Cypher: `
				UNWIND $ids AS issue_id
				MATCH (i:` + model.ResourceTypeIssue.String() + ` {id: issue_id})
				RETURN issue_id, COUNT { (:` + model.ResourceTypeDocument.String() + `)-[:` + EdgeKindRelatedTo.String() + `]->(i) } AS document_count`,
			Params: map[string]any{},
		})
	}

	if proj.AttachmentCount {
		loaders = append(loaders, CompiledQuery{
			Name: "issue.load_attachment_count",
			Cypher: `
				UNWIND $ids AS issue_id
				MATCH (i:` + model.ResourceTypeIssue.String() + ` {id: issue_id})
				RETURN issue_id, COUNT { (i)-[:` + EdgeKindHasAttachment.String() + `]->(:` + model.ResourceTypeAttachment.String() + `) } AS attachment_count`,
			Params: map[string]any{},
		})
	}

	if proj.WatcherCount {
		loaders = append(loaders, CompiledQuery{
			Name: "issue.load_watcher_count",
			Cypher: `
				UNWIND $ids AS issue_id
				MATCH (i:` + model.ResourceTypeIssue.String() + ` {id: issue_id})
				RETURN issue_id, COUNT { (:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindWatches.String() + `]->(i) } AS watcher_count`,
			Params: map[string]any{},
		})
	}

	if proj.RelationCount {
		loaders = append(loaders, CompiledQuery{
			Name: "issue.load_relation_count",
			Cypher: `
				UNWIND $ids AS issue_id
				MATCH (i:` + model.ResourceTypeIssue.String() + ` {id: issue_id})
				RETURN issue_id, COUNT { (i)-[:` + EdgeKindRelatedTo.String() + `]-(:` + model.ResourceTypeIssue.String() + `) } AS relation_count`,
			Params: map[string]any{},
		})
	}

	return loaders
}

func IssueWatchersQuery(issueID model.ID) (CompiledQuery, error) {
	if err := issueID.Validate(); err != nil {
		return CompiledQuery{}, err
	}

	return CompiledQuery{
		Name: "issue.get_watchers",
		Cypher: `
			MATCH (i:` + issueID.Label() + ` {id: $issue_id})<-[:` + EdgeKindWatches.String() + `]-(u:` + model.ResourceTypeUser.String() + `)`,
		Params: map[string]any{"issue_id": issueID.String()},
	}, nil
}

func IssueRelationsQuery(issueID model.ID) (CompiledQuery, error) {
	if err := issueID.Validate(); err != nil {
		return CompiledQuery{}, err
	}

	return CompiledQuery{
		Name: "issue.get_relations",
		Cypher: `
			MATCH (i:` + issueID.Label() + ` {id: $issue_id})-[r:` + EdgeKindRelatedTo.String() + `]-(n:` + model.ResourceTypeIssue.String() + `)
			RETURN i.id AS issue_id, r, n.id AS related_issue_id`,
		Params: map[string]any{"issue_id": issueID.String()},
	}, nil
}

func IssueRelationByIDQuery(relationID model.ID) (CompiledQuery, error) {
	if err := relationID.Validate(); err != nil {
		return CompiledQuery{}, err
	}

	return CompiledQuery{
		Name: "issue.get_relation",
		Cypher: `
			MATCH ()-[r:` + EdgeKindRelatedTo.String() + ` {id: $id}]->()
			RETURN r, startNode(r).id AS source_id, endNode(r).id AS target_id`,
		Params: map[string]any{"id": relationID.String()},
	}, nil
}

// IssueRelationListQuery compiles a cursor-paginated relation list for an issue.
type IssueRelationListQuery struct {
	IssueID model.ID
	Page    CursorPage
	Order   SortDirection
}

func (q IssueRelationListQuery) Compile() (QueryPlan, error) {
	if err := q.IssueID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"issue_id": q.IssueID.String(),
	}
	bounds, err := compileCursorBounds("r", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: "issue.list_relations",
			Cypher: strings.TrimSpace(`
				MATCH (i:` + q.IssueID.Label() + ` {id: $issue_id})-[r:` + EdgeKindRelatedTo.String() + `]-(n:` + model.ResourceTypeIssue.String() + `)
				MATCH (n)-[:` + EdgeKindBelongsTo.String() + `]->(p:` + model.ResourceTypeProject.String() + `)
				OPTIONAL MATCH (ns:` + model.ResourceTypeNamespace.String() + `)-[:` + EdgeKindHasProject.String() + `]->(p)
				` + cursorWherePrefix(bounds.Where, "WHERE ") + `
				RETURN r, n, p, ns, startNode(r).id AS source_id, endNode(r).id AS target_id
				ORDER BY r.id ` + bounds.Order.Cypher() + `
				LIMIT $limit`),
			Params: params,
		},
	}
	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}
