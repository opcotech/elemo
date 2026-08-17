package repository

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type DocumentProjection struct {
	Labels          bool
	CommentCount    bool
	AttachmentCount bool
	Folder          bool
	Relations       bool
}

func DocumentSummaryProjection() DocumentProjection {
	return DocumentProjection{}
}

func DocumentListProjection() DocumentProjection {
	return DocumentSummaryProjection()
}

func DocumentDetailProjection() DocumentProjection {
	return DocumentProjection{
		Labels:          true,
		CommentCount:    true,
		AttachmentCount: true,
		Folder:          true,
		Relations:       true,
	}
}

// LibraryListFilter selects which documents to return from a library.
// All lists every document. FolderID lists documents in that folder.
// Neither lists unfiled documents at the library root.
type LibraryListFilter struct {
	FolderID *model.ID
	All      bool
}

func (f LibraryListFilter) cacheValue() string {
	if f.All {
		return "all"
	}
	if f.FolderID != nil {
		return f.FolderID.String()
	}
	return "root"
}

type DocumentGetQuery struct {
	ID         model.ID
	Projection DocumentProjection
}

type DocumentListByCreatorQuery struct {
	CreatedBy  model.ID
	Page       CursorPage
	Order      SortDirection
	Projection DocumentProjection
}

type DocumentListLibraryQuery struct {
	LibraryID  model.ID
	Filter     LibraryListFilter
	Page       CursorPage
	Order      SortDirection
	Projection DocumentProjection
}

type DocumentListRelatedQuery struct {
	RelatedTo  model.ID
	Page       CursorPage
	Order      SortDirection
	Projection DocumentProjection
}

func (q DocumentGetQuery) Compile() (QueryPlan, error) {
	if err := q.ID.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return compileDocumentRootQuery(documentRootQueryInput{
		Root: CompiledQuery{
			Name: "document.get",
			Cypher: `
				MATCH (d:` + q.ID.Label() + ` {id: $id})<-[:` + EdgeKindCreated.String() + `]-(c:` + model.ResourceTypeUser.String() + `)
				MATCH (d)-[:` + EdgeKindScopedTo.String() + `]->(lib)
				RETURN d, c, lib`,
			Params: map[string]any{"id": q.ID.String()},
		},
		Projection: q.Projection,
	})
}

func (q DocumentListByCreatorQuery) Compile() (QueryPlan, error) {
	if err := q.CreatedBy.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"id": q.CreatedBy.String(),
	}
	bounds, err := compileCursorBounds("d", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	return compileDocumentRootQuery(documentRootQueryInput{
		Root: CompiledQuery{
			Name: "document.list_by_creator",
			Cypher: strings.TrimSpace(`
				MATCH (d:` + model.ResourceTypeDocument.String() + `)<-[:` + EdgeKindCreated.String() + `]-(c:` + q.CreatedBy.Label() + ` {id: $id})
				` + cursorWherePrefix(bounds.Where, "WHERE ") + `
				RETURN d, c
				ORDER BY d.id ` + bounds.Order.Cypher() + `
				LIMIT $limit`),
			Params: params,
		},
		Projection: q.Projection,
	})
}

func (q DocumentListLibraryQuery) Compile() (QueryPlan, error) {
	if err := q.LibraryID.Validate(); err != nil {
		return QueryPlan{}, err
	}
	if q.Filter.FolderID != nil {
		if err := q.Filter.FolderID.Validate(); err != nil {
			return QueryPlan{}, err
		}
	}

	params := map[string]any{
		"id": q.LibraryID.String(),
	}
	bounds, err := compileCursorBounds("d", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	var match string
	cursorPrefix := "WHERE "
	switch {
	case q.Filter.All:
		match = `
				MATCH (:` + q.LibraryID.Label() + ` {id: $id})<-[:` + EdgeKindScopedTo.String() + `]-(d:` + model.ResourceTypeDocument.String() + `)
				MATCH (c:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(d)`
	case q.Filter.FolderID != nil:
		params["folder_id"] = q.Filter.FolderID.String()
		match = `
				MATCH (:` + q.LibraryID.Label() + ` {id: $id})<-[:` + EdgeKindScopedTo.String() + `]-(:` + model.ResourceTypeFolder.String() + ` {id: $folder_id})<-[:` + EdgeKindLocatedIn.String() + `]-(d:` + model.ResourceTypeDocument.String() + `)
				MATCH (c:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(d)`
	default:
		cursorPrefix = "AND "
		match = `
				MATCH (:` + q.LibraryID.Label() + ` {id: $id})<-[:` + EdgeKindScopedTo.String() + `]-(d:` + model.ResourceTypeDocument.String() + `)
				MATCH (c:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(d)
				WHERE NOT (d)-[:` + EdgeKindLocatedIn.String() + `]->(:` + model.ResourceTypeFolder.String() + `)`
	}

	return compileDocumentRootQuery(documentRootQueryInput{
		Root: CompiledQuery{
			Name: "document.list_library",
			Cypher: strings.TrimSpace(match + `
				` + cursorWherePrefix(bounds.Where, cursorPrefix) + `
				RETURN d, c
				ORDER BY d.id ` + bounds.Order.Cypher() + `
				LIMIT $limit`),
			Params: params,
		},
		Projection: q.Projection,
	})
}

func (q DocumentListRelatedQuery) Compile() (QueryPlan, error) {
	if err := q.RelatedTo.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"id": q.RelatedTo.String(),
	}
	bounds, err := compileCursorBounds("d", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	return compileDocumentRootQuery(documentRootQueryInput{
		Root: CompiledQuery{
			Name: "document.list_related",
			Cypher: strings.TrimSpace(`
				MATCH (:` + q.RelatedTo.Label() + ` {id: $id})<-[:` + EdgeKindRelatedTo.String() + `]-(d:` + model.ResourceTypeDocument.String() + `)
				MATCH (c:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(d)
				` + cursorWherePrefix(bounds.Where, "WHERE ") + `
				RETURN d, c
				ORDER BY d.id ` + bounds.Order.Cypher() + `
				LIMIT $limit`),
			Params: params,
		},
		Projection: q.Projection,
	})
}

type documentRootQueryInput struct {
	Root       CompiledQuery
	Projection DocumentProjection
}

func compileDocumentRootQuery(in documentRootQueryInput) (QueryPlan, error) {
	plan := QueryPlan{
		Root: in.Root,
	}

	if in.Projection.Labels {
		plan.Loaders = append(plan.Loaders, CompiledQuery{
			Name: "document.load_labels",
			Cypher: `
				UNWIND $ids AS document_id
				MATCH (d:` + model.ResourceTypeDocument.String() + ` {id: document_id})
				OPTIONAL MATCH (d)-[:` + EdgeKindHasLabel.String() + `]->(l:` + model.ResourceTypeLabel.String() + `)
				RETURN document_id, collect(DISTINCT l) AS labels`,
			Params: map[string]any{},
		})
	}

	if in.Projection.CommentCount {
		plan.Loaders = append(plan.Loaders, CompiledQuery{
			Name: "document.load_comment_count",
			Cypher: `
				UNWIND $ids AS document_id
				MATCH (d:` + model.ResourceTypeDocument.String() + ` {id: document_id})
				RETURN document_id, COUNT { (d)-[:` + EdgeKindHasComment.String() + `]->(:` + model.ResourceTypeComment.String() + `) } AS comment_count`,
			Params: map[string]any{},
		})
	}

	if in.Projection.AttachmentCount {
		plan.Loaders = append(plan.Loaders, CompiledQuery{
			Name: "document.load_attachment_count",
			Cypher: `
				UNWIND $ids AS document_id
				MATCH (d:` + model.ResourceTypeDocument.String() + ` {id: document_id})
				RETURN document_id, COUNT { (d)-[:` + EdgeKindHasAttachment.String() + `]->(:` + model.ResourceTypeAttachment.String() + `) } AS attachment_count`,
			Params: map[string]any{},
		})
	}

	if in.Projection.Folder {
		plan.Loaders = append(plan.Loaders, CompiledQuery{
			Name: "document.load_folder",
			Cypher: `
				UNWIND $ids AS document_id
				MATCH (d:` + model.ResourceTypeDocument.String() + ` {id: document_id})
				OPTIONAL MATCH (d)-[:` + EdgeKindLocatedIn.String() + `]->(f:` + model.ResourceTypeFolder.String() + `)
				OPTIONAL MATCH (f)-[:` + EdgeKindLocatedIn.String() + `]->(parent:` + model.ResourceTypeFolder.String() + `)
				RETURN document_id, f, parent`,
			Params: map[string]any{},
		})
	}

	if in.Projection.Relations {
		plan.Loaders = append(plan.Loaders, CompiledQuery{
			Name: "document.load_relations",
			Cypher: `
				UNWIND $ids AS document_id
				MATCH (d:` + model.ResourceTypeDocument.String() + ` {id: document_id})-[:` + EdgeKindRelatedTo.String() + `]->(rel)
				OPTIONAL MATCH (rel:` + model.ResourceTypeIssue.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(p:` + model.ResourceTypeProject.String() + `)
				RETURN document_id, rel, p`,
			Params: map[string]any{},
		})
	}

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}
