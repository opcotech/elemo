package repository

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type RoleProjection struct {
	MemberCount bool
	Permissions bool
}

func RoleListProjection() RoleProjection {
	return RoleProjection{
		MemberCount: true,
		Permissions: true,
	}
}

func RoleDetailProjection() RoleProjection {
	return RoleProjection{
		MemberCount: true,
		Permissions: true,
	}
}

type RoleGetQuery struct {
	ID         model.ID
	BelongsTo  model.ID
	Projection RoleProjection
}

type RoleListBelongsToQuery struct {
	BelongsTo  model.ID
	Page       CursorPage
	Order      SortDirection
	Projection RoleProjection
}

func (q RoleGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	if err := q.BelongsTo.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return compileRoleRootQuery(roleRootQueryInput{
		Name: "role.get",
		Match: `
				MATCH (:` + q.BelongsTo.Label() + ` {id: $belongs_to_id})-[:` + EdgeKindDefinesRole.String() + `]->(r:` + model.ResourceTypeRole.String() + ` {id: $id})`,
		Params: map[string]any{
			"id":            q.ID.String(),
			"belongs_to_id": q.BelongsTo.String(),
		},
		Projection: q.Projection,
	})
}

func (q RoleListBelongsToQuery) Compile() (QueryPlan, error) {
	if err := q.BelongsTo.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"id": q.BelongsTo.String(),
	}
	bounds, err := compileCursorBounds("r", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	match := strings.TrimSpace(`
				MATCH (:` + q.BelongsTo.Label() + ` {id: $id})-[:` + EdgeKindDefinesRole.String() + `]->(r:` + model.ResourceTypeRole.String() + `)
				` + cursorWherePrefix(bounds.Where, "WHERE ") + `
				WITH r
				ORDER BY r.id ` + bounds.Order.Cypher() + `
				LIMIT $limit`)

	return compileRoleRootQuery(roleRootQueryInput{
		Name:       "role.list_belongs_to",
		Match:      match,
		Params:     params,
		Projection: q.Projection,
	})
}

type roleRootQueryInput struct {
	Name       string
	Match      string
	Params     map[string]any
	Projection RoleProjection
}

func compileRoleRootQuery(in roleRootQueryInput) (QueryPlan, error) {
	returns := []string{"r"}
	if in.Projection.MemberCount {
		returns = append(returns, `COUNT { (:`+model.ResourceTypeUser.String()+`)-[:`+EdgeKindMemberOf.String()+`]->(r) } AS member_count`)
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name:   in.Name,
			Cypher: strings.TrimSpace(in.Match) + "\nRETURN " + strings.Join(returns, ", "),
			Params: in.Params,
		},
	}

	if in.Projection.Permissions {
		plan.Loaders = append(plan.Loaders, CompiledQuery{
			Name: "role.load_permissions",
			Cypher: `
				UNWIND $ids AS role_id
				MATCH (r:` + model.ResourceTypeRole.String() + ` {id: role_id})
				OPTIONAL MATCH (r)-[p:` + EdgeKindGranted.String() + `]->()
				RETURN role_id, collect(DISTINCT p.id) AS permission_ids`,
			Params: map[string]any{},
		})
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}

type RoleMemberListQuery struct {
	RoleID    model.ID
	BelongsTo model.ID
	Page      CursorPage
	Order     SortDirection
}

func (q RoleMemberListQuery) Compile() (QueryPlan, error) {
	if err := q.RoleID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	if err := q.BelongsTo.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"role_id":       q.RoleID.String(),
		"belongs_to_id": q.BelongsTo.String(),
	}
	bounds, err := compileCursorBounds("u", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: "role.list_members",
			Cypher: `
			MATCH (:` + q.BelongsTo.Label() + ` {id: $belongs_to_id})-[:` + EdgeKindHasTeam.String() + `]->(r:` + q.RoleID.Label() + ` {id: $role_id})
			MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindMemberOf.String() + `]->(r)` + cursorWherePrefix(bounds.Where, " WHERE ") + `
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
