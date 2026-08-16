package repository

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

// LabelProjection selects bounded fields for label reads.
type LabelProjection struct{}

func LabelListProjection() LabelProjection {
	return LabelProjection{}
}

func LabelDetailProjection() LabelProjection {
	return LabelProjection{}
}

type LabelGetQuery struct {
	ID         model.ID
	Projection LabelProjection
}

type LabelListQuery struct {
	Page       CursorPage
	Order      SortDirection
	Projection LabelProjection
}

func (q LabelGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name:   "label.get",
			Cypher: `MATCH (l:` + q.ID.Label() + ` {id: $id}) RETURN l`,
			Params: map[string]any{"id": q.ID.String()},
		},
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}

func (q LabelListQuery) Compile() (QueryPlan, error) {
	params := map[string]any{}
	bounds, err := compileCursorBounds("l", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: "label.list",
			Cypher: strings.TrimSpace(`
				MATCH (l:` + model.ResourceTypeLabel.String() + `)
				` + cursorWherePrefix(bounds.Where, "WHERE ") + `
				WITH l
				ORDER BY l.id ` + bounds.Order.Cypher() + `
				LIMIT $limit
				RETURN l`,
			),
			Params: params,
		},
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}
