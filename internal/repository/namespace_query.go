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

	match := `
	MATCH (org:` + q.OrgID.Label() + ` {id: $org_id})-[:` + EdgeKindHasNamespace.String() + `]->(ns:` + model.ResourceTypeNamespace.String() + `)` + cursorWherePrefix(bounds.Where, " WHERE ") + `
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

type namespaceRootQueryInput struct {
	Name       string
	Match      string
	Params     map[string]any
	Alias      string
	Projection NamespaceProjection
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
