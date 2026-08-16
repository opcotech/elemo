package repository

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type CommentGetQuery struct {
	ID model.ID
}

type CommentListBelongsToQuery struct {
	BelongsTo model.ID
	Page      CursorPage
	Order     SortDirection
}

func (q CommentGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: "comment.get",
			Cypher: `
				MATCH (c:` + q.ID.Label() + ` {id: $id})<-[:` + EdgeKindCommented.String() + `]-(o:` + model.ResourceTypeUser.String() + `)
				RETURN c, o.id AS o`,
			Params: map[string]any{"id": q.ID.String()},
		},
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}

func (q CommentListBelongsToQuery) Compile() (QueryPlan, error) {
	if err := q.BelongsTo.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"id": q.BelongsTo.String(),
	}
	bounds, err := compileCursorBounds("c", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: "comment.list_belongs_to",
			Cypher: strings.TrimSpace(`
				MATCH (:` + q.BelongsTo.Label() + ` {id: $id})-[:` + EdgeKindHasComment.String() + `]->(c:` + model.ResourceTypeComment.String() + `)
				MATCH (o:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCommented.String() + `]->(c)
				` + cursorWherePrefix(bounds.Where, "WHERE ") + `
				RETURN c, o.id AS o
				ORDER BY c.id ` + bounds.Order.Cypher() + `
				LIMIT $limit`),
			Params: params,
		},
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}
