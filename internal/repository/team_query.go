package repository

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type TeamProjection struct {
	MemberCount bool
}

func TeamListProjection() TeamProjection {
	return TeamProjection{
		MemberCount: true,
	}
}

func TeamDetailProjection() TeamProjection {
	return TeamProjection{
		MemberCount: true,
	}
}

type TeamGetQuery struct {
	ID         model.ID
	BelongsTo  model.ID
	Projection TeamProjection
}

type TeamListBelongsToQuery struct {
	BelongsTo  model.ID
	Page       CursorPage
	Order      SortDirection
	Projection TeamProjection
}

func (q TeamGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	if err := q.BelongsTo.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return compileTeamRootQuery(teamRootQueryInput{
		Name: "team.get",
		Match: `
				MATCH (:` + q.BelongsTo.Label() + ` {id: $belongs_to_id})-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeTeam.String() + ` {id: $id})`,
		Params: map[string]any{
			"id":            q.ID.String(),
			"belongs_to_id": q.BelongsTo.String(),
		},
		Projection: q.Projection,
	})
}

func (q TeamListBelongsToQuery) Compile() (QueryPlan, error) {
	if err := q.BelongsTo.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"id": q.BelongsTo.String(),
	}
	bounds, err := compileCursorBounds("t", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	match := strings.TrimSpace(`
				MATCH (:` + q.BelongsTo.Label() + ` {id: $id})-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeTeam.String() + `)
				` + cursorWherePrefix(bounds.Where, "WHERE ") + `
				WITH t
				ORDER BY t.id ` + bounds.Order.Cypher() + `
				LIMIT $limit`)

	return compileTeamRootQuery(teamRootQueryInput{
		Name:       "team.list_belongs_to",
		Match:      match,
		Params:     params,
		Projection: q.Projection,
	})
}

type teamRootQueryInput struct {
	Name       string
	Match      string
	Params     map[string]any
	Projection TeamProjection
}

func compileTeamRootQuery(in teamRootQueryInput) (QueryPlan, error) {
	returns := []string{"t"}
	if in.Projection.MemberCount {
		returns = append(returns, `COUNT { (:`+model.ResourceTypeUser.String()+`)-[:`+EdgeKindMemberOf.String()+`]->(t) } AS member_count`)
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name:   in.Name,
			Cypher: strings.TrimSpace(in.Match) + "\nRETURN " + strings.Join(returns, ", "),
			Params: in.Params,
		},
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}

type TeamMemberListQuery struct {
	TeamID    model.ID
	BelongsTo model.ID
	Page      CursorPage
	Order     SortDirection
}

func (q TeamMemberListQuery) Compile() (QueryPlan, error) {
	if err := q.TeamID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	if err := q.BelongsTo.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"team_id":       q.TeamID.String(),
		"belongs_to_id": q.BelongsTo.String(),
	}
	bounds, err := compileCursorBounds("u", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: "team.list_members",
			Cypher: `
			MATCH (:` + q.BelongsTo.Label() + ` {id: $belongs_to_id})-[:` + EdgeKindHasTeam.String() + `]->(t:` + q.TeamID.Label() + ` {id: $team_id})
			MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindMemberOf.String() + `]->(t)` + cursorWherePrefix(bounds.Where, " WHERE ") + `
			WITH u
			ORDER BY u.id ` + bounds.Order.Cypher() + `
			LIMIT $limit
			RETURN u`,
			Params: params,
		},
	}
	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}
	return plan, nil
}
