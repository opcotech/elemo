package repository

import (
	"fmt"
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

// ProjectProjection selects bounded fields/relations for project reads.
type ProjectProjection struct {
	Teams         bool
	DocumentCount bool
	IssueCount    bool
}

// ProjectListProjection is the allowlisted projection for project list pages.
func ProjectListProjection() ProjectProjection {
	return ProjectProjection{
		DocumentCount: true,
		IssueCount:    true,
	}
}

// ProjectDetailProjection is the allowlisted projection for project detail.
func ProjectDetailProjection() ProjectProjection {
	return ProjectProjection{
		Teams:         true,
		DocumentCount: true,
		IssueCount:    true,
	}
}

// ProjectGetQuery compiles a single-project read.
type ProjectGetQuery struct {
	ID         model.ID
	Projection ProjectProjection
}

// ProjectGetByKeyQuery compiles a single-project read by key.
type ProjectGetByKeyQuery struct {
	Key        string
	Projection ProjectProjection
}

// ProjectListQuery compiles a cursor-paginated project list for a namespace.
type ProjectListQuery struct {
	NamespaceID model.ID
	ActorID     model.ID
	ScopeIDs    []model.ID
	Page        CursorPage
	Order       SortDirection
	Projection  ProjectProjection
}

func (q ProjectGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	return compileProjectRoot(projectRootQueryInput{
		Name:       "project.get",
		Match:      `MATCH (p:` + q.ID.Label() + ` {id: $id})`,
		Params:     map[string]any{"id": q.ID.String()},
		Alias:      "p",
		Projection: q.Projection,
	})
}

func (q ProjectGetByKeyQuery) Compile() (QueryPlan, error) {
	if strings.TrimSpace(q.Key) == "" {
		return QueryPlan{}, ErrQueryCompile
	}
	return compileProjectRoot(projectRootQueryInput{
		Name:       "project.get_by_key",
		Match:      `MATCH (p:` + model.ResourceTypeProject.String() + ` {key: $key})`,
		Params:     map[string]any{"key": q.Key},
		Alias:      "p",
		Projection: q.Projection,
	})
}

func (q ProjectListQuery) Compile() (QueryPlan, error) {
	if err := q.NamespaceID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	for _, scopeID := range q.ScopeIDs {
		if err := scopeID.Validate(); err != nil {
			return QueryPlan{}, err
		}
	}
	params := map[string]any{
		"namespace_id": q.NamespaceID.String(),
		"user_id":      q.ActorID.String(),
	}
	bounds, err := compileCursorBounds("p", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	authz := applyListScopeAuthz("p", q.ScopeIDs, params)
	match := `
	MATCH (:` + q.NamespaceID.Label() + ` {id: $namespace_id})-[:` + EdgeKindHasProject.String() + `]->(p)` + whereClause(" WHERE ", authz, bounds.Where) + `
	WITH p
	ORDER BY p.id ` + bounds.Order.Cypher() + `
	LIMIT $limit`

	return compileProjectRoot(projectRootQueryInput{
		Name:       "project.list",
		Match:      match,
		Params:     params,
		Alias:      "p",
		Projection: q.Projection,
	})
}

type projectRootQueryInput struct {
	Name       string
	Match      string
	Params     map[string]any
	Alias      string
	Projection ProjectProjection
}

func compileProjectRoot(in projectRootQueryInput) (QueryPlan, error) {
	returns := []string{in.Alias}
	if in.Projection.IssueCount {
		returns = append(returns, fmt.Sprintf(
			"COUNT { (%s)<-[:%s]-(:%s) } AS issue_count",
			in.Alias,
			EdgeKindBelongsTo.String(),
			model.ResourceTypeIssue.String(),
		))
	}
	if in.Projection.DocumentCount {
		returns = append(returns, fmt.Sprintf(
			"COUNT { (:%s)-[:%s]->(%s) } AS document_count",
			model.ResourceTypeDocument.String(),
			EdgeKindRelatedTo.String(),
			in.Alias,
		))
	}

	root := CompiledQuery{
		Name:   in.Name,
		Cypher: in.Match + "\nRETURN " + strings.Join(returns, ", "),
		Params: in.Params,
	}

	loaders := make([]CompiledQuery, 0, 1)
	if in.Projection.Teams {
		loaders = append(loaders, CompiledQuery{
			Name: in.Name + ".teams",
			Cypher: `
			UNWIND $ids AS pid
			MATCH (p:` + model.ResourceTypeProject.String() + ` {id: pid})-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeTeam.String() + `)
			RETURN pid AS project_id, collect(DISTINCT t.id) AS team_ids`,
			Params: map[string]any{},
		})
	}

	plan := QueryPlan{Root: root, Loaders: loaders}
	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}
	return plan, nil
}
