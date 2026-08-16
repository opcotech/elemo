package repository

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type UserProjection struct {
	DocumentCount bool
	Permissions   bool
}

func UserListProjection() UserProjection {
	return UserProjection{
		DocumentCount: true,
	}
}

func UserDetailProjection() UserProjection {
	return UserProjection{
		DocumentCount: true,
	}
}

type UserGetQuery struct {
	ID         model.ID
	Projection UserProjection
}

type UserGetByEmailQuery struct {
	Email      string
	Projection UserProjection
}

type UserListQuery struct {
	Page       CursorPage
	Order      SortDirection
	Projection UserProjection
}

func (q UserGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return compileUserRootQuery(userRootQueryInput{
		Name:       "user.get",
		Match:      "MATCH (u:" + model.ResourceTypeUser.String() + " {id: $id})",
		Params:     map[string]any{"id": q.ID.String()},
		Projection: q.Projection,
	})
}

func (q UserGetByEmailQuery) Compile() (QueryPlan, error) {
	return compileUserRootQuery(userRootQueryInput{
		Name:       "user.get_by_email",
		Match:      "MATCH (u:" + model.ResourceTypeUser.String() + " {email: $email})",
		Params:     map[string]any{"email": q.Email},
		Projection: q.Projection,
	})
}

func (q UserListQuery) Compile() (QueryPlan, error) {
	params := map[string]any{}
	bounds, err := compileCursorBounds("u", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	match := strings.TrimSpace(`
				MATCH (u:` + model.ResourceTypeUser.String() + `)
				` + cursorWherePrefix(bounds.Where, "WHERE ") + `
				WITH u
				ORDER BY u.id ` + bounds.Order.Cypher() + `
				LIMIT $limit`)

	return compileUserRootQuery(userRootQueryInput{
		Name:       "user.list",
		Match:      match,
		Params:     params,
		Projection: q.Projection,
	})
}

type userRootQueryInput struct {
	Name       string
	Match      string
	Params     map[string]any
	Projection UserProjection
}

func compileUserRootQuery(in userRootQueryInput) (QueryPlan, error) {
	returns := []string{"u"}
	if in.Projection.DocumentCount {
		returns = append(returns, "COUNT { (u)<-[:"+EdgeKindBelongsTo.String()+"]-(:"+model.ResourceTypeDocument.String()+") } AS document_count")
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
			Name: "user.load_permissions",
			Cypher: `
				UNWIND $ids AS user_id
				MATCH (u:` + model.ResourceTypeUser.String() + ` {id: user_id})-[p:` + EdgeKindHasPermission.String() + `]->()
				RETURN user_id AS user_id, collect(p.id) AS permission_ids`,
			Params: map[string]any{},
		})
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}
