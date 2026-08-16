package repository

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type AssignmentProjection struct{}

func AssignmentListProjection() AssignmentProjection {
	return AssignmentProjection{}
}

func AssignmentDetailProjection() AssignmentProjection {
	return AssignmentProjection{}
}

type AssignmentGetQuery struct {
	ID         model.ID
	Projection AssignmentProjection
}

type AssignmentListByUserQuery struct {
	UserID     model.ID
	Page       CursorPage
	Order      SortDirection
	Projection AssignmentProjection
}

type AssignmentListByResourceQuery struct {
	ResourceID model.ID
	Page       CursorPage
	Order      SortDirection
	Projection AssignmentProjection
}

func (q AssignmentGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: "assignment.get",
			Cypher: `
				MATCH (u)-[a:` + EdgeKindAssignedTo.String() + ` {id: $id}]->(r)
				RETURN u, a, r`,
			Params: map[string]any{"id": q.ID.String()},
		},
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}

func (q AssignmentListByUserQuery) Compile() (QueryPlan, error) {
	return compileAssignmentListQuery(assignmentListQueryInput{
		Name:        "assignment.list_by_user",
		CursorAlias: "a",
		ID:          q.UserID,
		Match:       "MATCH (u:" + q.UserID.Label() + " {id: $user_id})-[a:" + EdgeKindAssignedTo.String() + "]->(r)",
		Params:      map[string]any{"user_id": q.UserID.String()},
		Page:        q.Page,
		Order:       q.Order,
	})
}

func (q AssignmentListByResourceQuery) Compile() (QueryPlan, error) {
	return compileAssignmentListQuery(assignmentListQueryInput{
		Name:        "assignment.list_by_resource",
		CursorAlias: "a",
		ID:          q.ResourceID,
		Match:       "MATCH (u)-[a:" + EdgeKindAssignedTo.String() + "]->(r:" + q.ResourceID.Label() + " {id: $resource_id})",
		Params:      map[string]any{"resource_id": q.ResourceID.String()},
		Page:        q.Page,
		Order:       q.Order,
	})
}

type assignmentListQueryInput struct {
	Name        string
	CursorAlias string
	ID          model.ID
	Match       string
	Params      map[string]any
	Page        CursorPage
	Order       SortDirection
}

func compileAssignmentListQuery(in assignmentListQueryInput) (QueryPlan, error) {
	if err := in.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	bounds, err := compileCursorBounds(in.CursorAlias, in.Page, in.Order, in.Params)
	if err != nil {
		return QueryPlan{}, err
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: in.Name,
			Cypher: strings.TrimSpace(in.Match + `
				` + cursorWherePrefix(bounds.Where, " WHERE ") + `
				RETURN u, a, r
				ORDER BY a.id ` + bounds.Order.Cypher() + `
				LIMIT $limit`),
			Params: in.Params,
		},
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}
