package repository

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type FolderGetQuery struct {
	ID model.ID
}

type FolderListQuery struct {
	LibraryID model.ID
	ParentID  *model.ID
	Page      CursorPage
	Order     SortDirection
}

func (q FolderGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: "folder.get",
			Cypher: strings.TrimSpace(`
				MATCH (f:` + q.ID.Label() + ` {id: $id})-[:` + EdgeKindScopedTo.String() + `]->(lib)
				MATCH (c:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(f)
				OPTIONAL MATCH (f)-[:` + EdgeKindLocatedIn.String() + `]->(parent:` + model.ResourceTypeFolder.String() + `)
				RETURN f, lib, c, parent`),
			Params: map[string]any{"id": q.ID.String()},
		},
	}
	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}
	return plan, nil
}

func (q FolderListQuery) Compile() (QueryPlan, error) {
	if err := q.LibraryID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	if q.ParentID != nil {
		if err := q.ParentID.Validate(); err != nil {
			return QueryPlan{}, err
		}
	}

	params := map[string]any{
		"library_id": q.LibraryID.String(),
	}
	bounds, err := compileCursorBounds("f", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	var match string
	cursorPrefix := "WHERE "
	if q.ParentID != nil {
		params["parent_id"] = q.ParentID.String()
		match = `
				MATCH (:` + q.LibraryID.Label() + ` {id: $library_id})<-[:` + EdgeKindScopedTo.String() + `]-(parent:` + model.ResourceTypeFolder.String() + ` {id: $parent_id})
				MATCH (parent)<-[:` + EdgeKindLocatedIn.String() + `]-(f:` + model.ResourceTypeFolder.String() + `)
				MATCH (f)-[:` + EdgeKindScopedTo.String() + `]->(lib)
				MATCH (c:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(f)`
	} else {
		cursorPrefix = "AND "
		match = `
				MATCH (:` + q.LibraryID.Label() + ` {id: $library_id})<-[:` + EdgeKindScopedTo.String() + `]-(f:` + model.ResourceTypeFolder.String() + `)
				MATCH (f)-[:` + EdgeKindScopedTo.String() + `]->(lib)
				MATCH (c:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(f)
				OPTIONAL MATCH (f)-[:` + EdgeKindLocatedIn.String() + `]->(parent:` + model.ResourceTypeFolder.String() + `)
				WITH f, lib, c, parent
				WHERE parent IS NULL`
	}

	plan := QueryPlan{
		Root: CompiledQuery{
			Name: "folder.list",
			Cypher: strings.TrimSpace(match + `
				` + cursorWherePrefix(bounds.Where, cursorPrefix) + `
				RETURN f, lib, c, parent
				ORDER BY f.id ` + bounds.Order.Cypher() + `
				LIMIT $limit`),
			Params: params,
		},
	}
	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}
	return plan, nil
}
