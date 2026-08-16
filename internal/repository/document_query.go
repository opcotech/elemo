package repository

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type DocumentProjection struct {
	Labels          bool
	CommentCount    bool
	AttachmentCount bool
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
	}
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

type DocumentListBelongsToQuery struct {
	BelongsTo  model.ID
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
				RETURN d, c`,
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

func (q DocumentListBelongsToQuery) Compile() (QueryPlan, error) {
	if err := q.BelongsTo.Validate(); err != nil {
		return QueryPlan{}, err
	}
	params := map[string]any{
		"id": q.BelongsTo.String(),
	}
	bounds, err := compileCursorBounds("d", q.Page, q.Order, params)
	if err != nil {
		return QueryPlan{}, err
	}

	return compileDocumentRootQuery(documentRootQueryInput{
		Root: CompiledQuery{
			Name: "document.list_belongs_to",
			Cypher: strings.TrimSpace(`
				MATCH (:` + q.BelongsTo.Label() + ` {id: $id})<-[:` + EdgeKindBelongsTo.String() + `]-(d:` + model.ResourceTypeDocument.String() + `)
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

	if err := plan.Validate(); err != nil {
		return QueryPlan{}, err
	}

	return plan, nil
}
