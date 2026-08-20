package repository

import (
	"fmt"
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type NamespaceProjection struct {
	ProjectCount  bool
	DocumentCount bool
}

func NamespaceListProjection() NamespaceProjection {
	return NamespaceProjection{
		ProjectCount:  true,
		DocumentCount: true,
	}
}

func NamespaceDetailProjection() NamespaceProjection {
	return NamespaceProjection{
		ProjectCount:  true,
		DocumentCount: true,
	}
}

type NamespaceGetQuery struct {
	ID         model.ID
	Projection NamespaceProjection
}

type NamespaceListQuery struct {
	OrgID      model.ID
	ActorID    model.ID
	Page       CursorPage
	Order      SortDirection
	Projection NamespaceProjection
}

// NamespaceListAccessibleQuery compiles reachable namespaces for an actor.
type NamespaceListAccessibleQuery struct {
	ActorID    model.ID
	Page       CursorPage
	Order      SortDirection
	Projection NamespaceProjection
}

func (q NamespaceGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return compileNamespaceRoot(namespaceRootQueryInput{
		Name:       "namespace.get",
		Match:      `MATCH (ns:` + q.ID.Label() + ` {id: $id})`,
		Params:     map[string]any{"id": q.ID.String()},
		Alias:      "ns",
		Projection: q.Projection,
	})
}

func (q NamespaceListQuery) Compile() (QueryPlan, error) {
	if err := q.OrgID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"org_id": q.OrgID.String(),
	}
	bounds, err := compileCursorBounds("ns", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	authz := applyAuthzReachableNamespace(q.ActorID, "ns", "$user_id", params)
	match := `
	MATCH (org:` + q.OrgID.Label() + ` {id: $org_id})-[:` + EdgeKindHasNamespace.String() + `]->(ns:` + model.ResourceTypeNamespace.String() + `)` + whereClause(" WHERE ", authz, bounds.Where) + `
	WITH ns
	ORDER BY ns.id ` + bounds.Order.Cypher() + `
	LIMIT $limit`

	return compileNamespaceRoot(namespaceRootQueryInput{
		Name:       "namespace.list",
		Match:      match,
		Params:     params,
		Alias:      "ns",
		Projection: q.Projection,
	})
}

func (q NamespaceListAccessibleQuery) Compile() (QueryPlan, error) {
	if err := q.ActorID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{}
	bounds, err := compileCursorBounds("ns", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	authz := applyAuthzReachableNamespace(q.ActorID, "ns", "$user_id", params)
	match := `
	MATCH (ns:` + model.ResourceTypeNamespace.String() + `)` + whereClause(" WHERE ", authz, bounds.Where) + `
	WITH ns
	ORDER BY ns.id ` + bounds.Order.Cypher() + `
	LIMIT $limit
	OPTIONAL MATCH (org:` + model.ResourceTypeOrganization.String() + `)-[:` + EdgeKindHasNamespace.String() + `]->(ns)`

	return compileNamespaceRoot(namespaceRootQueryInput{
		Name:         "namespace.list_accessible",
		Match:        match,
		Params:       params,
		Alias:        "ns",
		Projection:   q.Projection,
		Organization: true,
	})
}

type namespaceRootQueryInput struct {
	Name         string
	Match        string
	Params       map[string]any
	Alias        string
	Projection   NamespaceProjection
	Organization bool
}

func compileNamespaceRoot(in namespaceRootQueryInput) (QueryPlan, error) {
	returns := []string{in.Alias}
	if in.Projection.ProjectCount {
		returns = append(returns, fmt.Sprintf(
			"COUNT { (:%s)<-[:%s]-(%s) } AS project_count",
			model.ResourceTypeProject.String(),
			EdgeKindHasProject.String(),
			in.Alias,
		))
	}
	if in.Projection.DocumentCount {
		returns = append(returns, fmt.Sprintf(
			"COUNT { (:%s)-[:%s]->(%s) } AS document_count",
			model.ResourceTypeDocument.String(),
			EdgeKindScopedTo.String(),
			in.Alias,
		))
	}
	if in.Organization {
		returns = append(returns, "org")
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name:   in.Name,
			Cypher: in.Match + "\nRETURN " + strings.Join(returns, ", "),
			Params: in.Params,
		},
	}
	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}
	return plan, nil
}
