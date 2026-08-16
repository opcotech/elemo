package repository

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type AttachmentProjection struct{}

func AttachmentListProjection() AttachmentProjection {
	return AttachmentProjection{}
}

func AttachmentDetailProjection() AttachmentProjection {
	return AttachmentProjection{}
}

type AttachmentGetQuery struct {
	ID         model.ID
	Projection AttachmentProjection
}

type AttachmentListBelongsToQuery struct {
	BelongsTo  model.ID
	Page       CursorPage
	Order      SortDirection
	Projection AttachmentProjection
}

func (q AttachmentGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: "attachment.get",
			Cypher: `
				MATCH (a:` + q.ID.Label() + ` {id: $id})<-[:` + EdgeKindCreated.String() + `]-(o:` + model.ResourceTypeUser.String() + `)
				RETURN a, o.id AS o`,
			Params: map[string]any{"id": q.ID.String()},
		},
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}

func (q AttachmentListBelongsToQuery) Compile() (QueryPlan, error) {
	if err := q.BelongsTo.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"id": q.BelongsTo.String(),
	}
	bounds, err := compileCursorBounds("a", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: "attachment.list_belongs_to",
			Cypher: strings.TrimSpace(`
				MATCH (:` + q.BelongsTo.Label() + ` {id: $id})-[:` + EdgeKindHasAttachment.String() + `]->(a:` + model.ResourceTypeAttachment.String() + `)
				MATCH (o:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(a)
				` + cursorWherePrefix(bounds.Where, "WHERE ") + `
				RETURN a, o.id AS o
				ORDER BY a.id ` + bounds.Order.Cypher() + `
				LIMIT $limit`),
			Params: params,
		},
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}
