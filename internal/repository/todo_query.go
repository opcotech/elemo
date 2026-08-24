package repository

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type TodoProjection struct{}

func TodoListProjection() TodoProjection {
	return TodoProjection{}
}

func TodoDetailProjection() TodoProjection {
	return TodoProjection{}
}

type TodoGetQuery struct {
	ID         model.ID
	Projection TodoProjection
}

type TodoListByOwnerQuery struct {
	OwnerID    model.ID
	Page       CursorPage
	Order      SortDirection
	Completed  *bool
	Projection TodoProjection
}

func (q TodoGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: "todo.get",
			Cypher: `
				MATCH (t:` + q.ID.Label() + ` {id: $id})
				OPTIONAL MATCH (t)-[:` + EdgeKindBelongsTo.String() + `]->(o)
				OPTIONAL MATCH (t)<-[:` + EdgeKindCreated.String() + `]-(c)
				RETURN t, o.id AS o, c.id AS c`,
			Params: map[string]any{"id": q.ID.String()},
		},
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}

func (q TodoListByOwnerQuery) Compile() (QueryPlan, error) {
	if err := q.OwnerID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"owner_id":  q.OwnerID.String(),
		"completed": q.Completed,
	}
	bounds, err := compileCursorBounds("t", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	where := "WHERE ($completed IS NULL OR t.completed = $completed)"
	if bounds.Where != "" {
		where += " AND " + bounds.Where
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: "todo.list_by_owner",
			Cypher: strings.TrimSpace(`
				MATCH (t:` + model.ResourceTypeTodo.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(o:` + q.OwnerID.Label() + ` {id: $owner_id})
				` + where + `
				OPTIONAL MATCH (t)<-[:` + EdgeKindCreated.String() + `]-(c)
				RETURN t, o.id AS o, c.id AS c
				ORDER BY t.id ` + bounds.Order.Cypher() + `
				LIMIT $limit`),
			Params: params,
		},
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}
