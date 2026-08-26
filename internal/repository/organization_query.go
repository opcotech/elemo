package repository

import (
	"fmt"
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type OrganizationProjection struct {
	NamespaceCount bool
	TeamCount      bool
	MemberCount    bool
	DocumentCount  bool
}

func OrganizationListProjection() OrganizationProjection {
	return OrganizationProjection{
		NamespaceCount: true,
		TeamCount:      true,
		MemberCount:    true,
		DocumentCount:  true,
	}
}

func OrganizationDetailProjection() OrganizationProjection {
	return OrganizationProjection{
		NamespaceCount: true,
		TeamCount:      true,
		MemberCount:    true,
		DocumentCount:  true,
	}
}

type OrganizationGetQuery struct {
	ID         model.ID
	Projection OrganizationProjection
}

// OrganizationGetByRefQuery compiles an organization read by global slug.
type OrganizationGetByRefQuery struct {
	Slug       string
	Projection OrganizationProjection
}

type OrganizationListQuery struct {
	UserID     model.ID
	Action     model.Action
	Page       CursorPage
	Order      SortDirection
	Projection OrganizationProjection
}

type OrganizationMemberListQuery struct {
	OrgID model.ID
	Page  CursorPage
	Order SortDirection
}

func (q OrganizationGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return compileOrganizationRoot(organizationRootQueryInput{
		Name:       "organization.get",
		Match:      `MATCH (o:` + q.ID.Label() + ` {id: $id})`,
		Params:     map[string]any{"id": q.ID.String()},
		Alias:      "o",
		Projection: q.Projection,
	})
}

func (q OrganizationGetByRefQuery) Compile() (QueryPlan, error) {
	if strings.TrimSpace(q.Slug) == "" {
		return QueryPlan{}, ErrQueryCompile
	}

	return compileOrganizationRoot(organizationRootQueryInput{
		Name:       "organization.get_by_ref",
		Match:      `MATCH (o:` + model.ResourceTypeOrganization.String() + ` {slug: $slug})`,
		Params:     map[string]any{"slug": q.Slug},
		Alias:      "o",
		Projection: q.Projection,
	})
}

func (q OrganizationListQuery) Compile() (QueryPlan, error) {
	if err := q.UserID.Validate(); err != nil {
		return QueryPlan{}, err
	}

	params := map[string]any{
		"user_id": q.UserID.String(),
	}
	bounds, err := compileCursorBounds("o", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	authz := applyAuthzVisible(q.UserID, q.Action, "o", "$user_id", params)
	match := `
	MATCH (o:` + model.ResourceTypeOrganization.String() + `)` + whereClause(" WHERE ", authz, bounds.Where) + `
	WITH o
	ORDER BY o.id ` + bounds.Order.Cypher() + `
	LIMIT $limit`

	return compileOrganizationRoot(organizationRootQueryInput{
		Name:       "organization.list",
		Match:      match,
		Params:     params,
		Alias:      "o",
		Projection: q.Projection,
	})
}

func (q OrganizationMemberListQuery) Compile() (QueryPlan, error) {
	if err := q.OrgID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"org_id": q.OrgID.String(),
	}
	bounds, err := compileCursorBounds("u", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	root := CompiledQuery{
		Name: "organization.list_members",
		Cypher: `
		MATCH (o:` + q.OrgID.Label() + ` {id: $org_id})
		MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindMemberOf.String() + `|` + EdgeKindInvitedTo.String() + `]->(o)
		WITH DISTINCT u, o` + cursorWherePrefix(bounds.Where, " WHERE ") + `
		WITH u, o
		ORDER BY u.id ` + bounds.Order.Cypher() + `
		LIMIT $limit
		WITH u, o, EXISTS { (u)-[:` + EdgeKindMemberOf.String() + `]->(o) } AS isMember
		RETURN u, isMember`,
		Params: params,
	}

	plan := QueryPlan{
		Root: root,
		Loaders: []CompiledQuery{
			{
				Name: "organization.list_members.roles",
				Cypher: `
				UNWIND $ids AS uid
				MATCH (o:` + model.ResourceTypeOrganization.String() + ` {id: $org_id})
				MATCH (u:` + model.ResourceTypeUser.String() + ` {id: uid})
				OPTIONAL MATCH (u)-[:` + EdgeKindMemberOf.String() + `]->(r:` + model.ResourceTypeRole.String() + `)<-[:` + EdgeKindHasTeam.String() + `]-(o)
				RETURN uid AS user_id, collect(DISTINCT r.name) AS roles`,
				Params: map[string]any{"org_id": q.OrgID.String()},
			},
		},
	}
	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}
	return plan, nil
}

type organizationRootQueryInput struct {
	Name       string
	Match      string
	Params     map[string]any
	Alias      string
	Projection OrganizationProjection
}

func compileOrganizationRoot(in organizationRootQueryInput) (QueryPlan, error) {
	returns := []string{in.Alias}
	if in.Projection.NamespaceCount {
		returns = append(returns, fmt.Sprintf(
			"COUNT { (%s)-[:%s]->(:%s) } AS namespace_count",
			in.Alias,
			EdgeKindHasNamespace.String(),
			model.ResourceTypeNamespace.String(),
		))
	}
	if in.Projection.TeamCount {
		returns = append(returns, fmt.Sprintf(
			"COUNT { (%s)-[:%s]->(:%s) } AS team_count",
			in.Alias,
			EdgeKindHasTeam.String(),
			model.ResourceTypeTeam.String(),
		))
	}
	if in.Projection.MemberCount {
		returns = append(returns, fmt.Sprintf(
			"COUNT { (:%s)-[:%s]->(%s) } AS member_count",
			model.ResourceTypeUser.String(),
			EdgeKindMemberOf.String(),
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
