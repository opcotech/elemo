package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
)

var (
	ErrDocumentCreate = errors.New("failed to create document") // the document could not be created
	ErrDocumentDelete = errors.New("failed to delete document") // the document could not be deleted
	ErrDocumentRead   = errors.New("failed to read document")   // the document could not be retrieved
	ErrDocumentUpdate = errors.New("failed to update document") // the document could not be updated
)

// PartialDocument represents a simplified document that can be used in list views.
type PartialDocument struct {
	ID        model.ID    `json:"id"`
	Name      string      `json:"name"`
	Excerpt   string      `json:"excerpt"`
	CreatedBy PartialUser `json:"created_by"`
	CreatedAt *time.Time  `json:"created_at"`
}

// Document represents a document persisted by the repository.
type Document struct {
	ID              model.ID       `json:"id"`
	Name            string         `json:"name"`
	Excerpt         string         `json:"excerpt"`
	FileID          string         `json:"file_id"`
	CreatedBy       PartialUser    `json:"created_by"`
	Labels          []PartialLabel `json:"labels"`
	CommentCount    *int64         `json:"comment_count"`
	AttachmentCount *int64         `json:"attachment_count"`
	CreatedAt       *time.Time     `json:"created_at"`
	UpdatedAt       *time.Time     `json:"updated_at"`
}

// CreateDocumentOpts holds the data required to create a document.
type CreateDocumentOpts struct {
	BelongsTo model.ID
	Name      string
	Excerpt   string
	FileID    string
	CreatedBy model.ID
}

// UpdateDocumentOpts holds the fields that can be updated on a document.
// Undefined fields (Defined == false) are left unchanged.
type UpdateDocumentOpts struct {
	Name    optional.Optional[string]
	Excerpt optional.Optional[string]
	FileID  optional.Optional[string]
}

// patch builds a Neo4j property map from defined optional fields.
func (o UpdateDocumentOpts) patch() map[string]any {
	p := make(map[string]any)

	if o.Name.Defined {
		p["name"] = *o.Name.Value
	}
	if o.Excerpt.Defined {
		p["excerpt"] = *o.Excerpt.Value
	}
	if o.FileID.Defined {
		p["file_id"] = *o.FileID.Value
	}

	return p
}

//go:generate go tool mockgen -source=document.go -destination=document_mock_gen.go -package=repository -mock_names "DocumentRepository=MockDocumentRepository"
type DocumentRepository interface {
	Create(ctx context.Context, opts CreateDocumentOpts) (*Document, error)
	Get(ctx context.Context, id model.ID, proj DocumentProjection) (*Document, error)
	ListByCreator(ctx context.Context, createdBy model.ID, page CursorPage, proj DocumentProjection) (Page[*Document], error)
	ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage, proj DocumentProjection) (Page[*Document], error)
	Update(ctx context.Context, id model.ID, opts UpdateDocumentOpts) (*Document, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jDocumentRepository is a repository for managing documents.
type Neo4jDocumentRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jDocumentRepository) scan(proj DocumentProjection) func(rec *neo4j.Record) (*Document, error) {
	return func(rec *neo4j.Record) (*Document, error) {
		node, err := Neo4jRecordNode(rec, "d")
		if err != nil {
			return nil, err
		}

		createdBy, err := Neo4jRecordPartialUser(rec, "c")
		if err != nil {
			return nil, err
		}
		if createdBy == nil {
			return nil, ErrMalformedResult
		}

		doc := new(Document)
		if err := Neo4jScanIntoStruct(&node, &doc, []string{"id", "created_by"}); err != nil {
			return nil, err
		}

		doc.ID, err = Neo4jDecodeID(node, model.ResourceTypeDocument)
		if err != nil {
			return nil, err
		}
		doc.CreatedBy = *createdBy
		if proj.Labels {
			doc.Labels = make([]PartialLabel, 0)
		}
		if proj.CommentCount {
			doc.CommentCount = convert.ToPointer(int64(0))
		}
		if proj.AttachmentCount {
			doc.AttachmentCount = convert.ToPointer(int64(0))
		}

		return doc, nil
	}
}

func (r *Neo4jDocumentRepository) applyDocumentLoaders(ctx context.Context, tx neo4j.ManagedTransaction, plan QueryPlan, documents []*Document) error {
	if len(plan.Loaders) == 0 || len(documents) == 0 {
		return nil
	}

	documentByID := make(map[string]*Document, len(documents))
	ids := make([]string, 0, len(documents))
	for _, document := range documents {
		if document == nil {
			continue
		}
		id := document.ID.String()
		documentByID[id] = document
		ids = append(ids, id)
	}

	for _, loader := range plan.Loaders {
		query := loader
		query.Params = cloneParams(loader.Params)
		query.Params["ids"] = ids
		switch loader.Name {
		case "document.load_labels":
			rows, _, err := Neo4jRunQuery(ctx, tx, query, func(rec *neo4j.Record) (struct {
				DocumentID string
				Labels     []PartialLabel
			}, error) {
				documentID, err := Neo4jParseValueFromRecord[string](rec, "document_id")
				if err != nil {
					return struct {
						DocumentID string
						Labels     []PartialLabel
					}{}, err
				}
				labels, err := Neo4jRecordPartialLabels(rec, "labels")
				if err != nil {
					return struct {
						DocumentID string
						Labels     []PartialLabel
					}{}, err
				}
				return struct {
					DocumentID string
					Labels     []PartialLabel
				}{DocumentID: documentID, Labels: labels}, nil
			})
			if err != nil {
				return err
			}
			for _, row := range rows {
				if document := documentByID[row.DocumentID]; document != nil {
					document.Labels = row.Labels
				}
			}
		case "document.load_comment_count":
			if err := applyDocumentCountLoader(ctx, tx, query, documentByID, "comment_count", func(document *Document, count int64) {
				document.CommentCount = convert.ToPointer(count)
			}); err != nil {
				return err
			}
		case "document.load_attachment_count":
			if err := applyDocumentCountLoader(ctx, tx, query, documentByID, "attachment_count", func(document *Document, count int64) {
				document.AttachmentCount = convert.ToPointer(count)
			}); err != nil {
				return err
			}
		default:
			return ErrQueryCompile
		}
	}

	return nil
}

func applyDocumentCountLoader(
	ctx context.Context,
	tx neo4j.ManagedTransaction,
	query CompiledQuery,
	documentByID map[string]*Document,
	field string,
	assign func(document *Document, count int64),
) error {
	rows, _, err := Neo4jRunQuery(ctx, tx, query, func(rec *neo4j.Record) (struct {
		DocumentID string
		Count      int64
	}, error) {
		documentID, err := Neo4jParseValueFromRecord[string](rec, "document_id")
		if err != nil {
			return struct {
				DocumentID string
				Count      int64
			}{}, err
		}
		count, err := Neo4jParseValueFromRecord[int64](rec, field)
		if err != nil {
			return struct {
				DocumentID string
				Count      int64
			}{}, err
		}
		return struct {
			DocumentID string
			Count      int64
		}{DocumentID: documentID, Count: count}, nil
	})
	if err != nil {
		return err
	}
	for _, row := range rows {
		if document := documentByID[row.DocumentID]; document != nil {
			assign(document, row.Count)
		}
	}
	return nil
}

func (r *Neo4jDocumentRepository) Create(ctx context.Context, opts CreateDocumentOpts) (*Document, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeDocument)

	cypher := `
	MATCH (b:` + opts.BelongsTo.Label() + ` {id: $belong_to_id})
	MATCH (o:` + opts.CreatedBy.Label() + ` {id: $created_by_id})
	CREATE
		(d:` + id.Label() + ` {
			id: $id, name: $name, excerpt: $excerpt, file_id: $file_id, created_by: $created_by_id,
			created_at: datetime($created_at)
		}),
		(d)-[:` + EdgeKindBelongsTo.String() + ` {id: $belongs_to_rel_id, created_at: datetime($created_at)}]->(b),
		(o)-[:` + EdgeKindCreated.String() + ` {id: $created_rel_id, created_at: datetime($created_at)}]->(d)`

	params := map[string]any{
		"belong_to_id":      opts.BelongsTo.String(),
		"belongs_to_rel_id": model.NewRawID(),
		"created_by_id":     opts.CreatedBy.String(),
		"created_rel_id":    model.NewRawID(),
		"id":                id.String(),
		"name":              opts.Name,
		"excerpt":           opts.Excerpt,
		"file_id":           opts.FileID,
		"created_at":        createdAt.Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrDocumentCreate, err)
	}

	return r.Get(ctx, id, DocumentDetailProjection())
}

func (r *Neo4jDocumentRepository) Get(ctx context.Context, id model.ID, proj DocumentProjection) (*Document, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/Get")
	defer span.End()

	plan, err := CompileQuery(DocumentGetQuery{ID: id, Projection: proj})
	if err != nil {
		return nil, errors.Join(ErrDocumentRead, err)
	}

	var document *Document
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		document, _, readErr = Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scan(proj))
		if readErr != nil {
			return readErr
		}
		return r.applyDocumentLoaders(ctx, tx, plan, []*Document{document})
	})
	if err != nil {
		return nil, errors.Join(ErrDocumentRead, err)
	}

	return document, nil
}

func (r *Neo4jDocumentRepository) ListByCreator(ctx context.Context, createdBy model.ID, page CursorPage, proj DocumentProjection) (Page[*Document], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/ListByCreator")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Document]{}, errors.Join(ErrDocumentRead, err)
	}
	plan, err := CompileQuery(DocumentListByCreatorQuery{
		CreatedBy:  createdBy,
		Page:       normalized,
		Order:      SortDirectionDesc,
		Projection: proj,
	})
	if err != nil {
		return Page[*Document]{}, errors.Join(ErrDocumentRead, err)
	}

	documents := make([]*Document, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		documents, _, readErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan(proj))
		if readErr != nil {
			return readErr
		}
		return r.applyDocumentLoaders(ctx, tx, plan, documents)
	})
	if err != nil {
		return Page[*Document]{}, errors.Join(ErrDocumentRead, err)
	}

	return PaginateSlice(documents, normalized.Size, func(document *Document) model.ID {
		return document.ID
	})
}

func (r *Neo4jDocumentRepository) ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage, proj DocumentProjection) (Page[*Document], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/ListBelongsTo")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Document]{}, errors.Join(ErrDocumentRead, err)
	}
	plan, err := CompileQuery(DocumentListBelongsToQuery{
		BelongsTo:  belongsTo,
		Page:       normalized,
		Order:      SortDirectionDesc,
		Projection: proj,
	})
	if err != nil {
		return Page[*Document]{}, errors.Join(ErrDocumentRead, err)
	}

	documents := make([]*Document, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		documents, _, readErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan(proj))
		if readErr != nil {
			return readErr
		}
		return r.applyDocumentLoaders(ctx, tx, plan, documents)
	})
	if err != nil {
		return Page[*Document]{}, errors.Join(ErrDocumentRead, err)
	}

	return PaginateSlice(documents, normalized.Size, func(document *Document) model.ID {
		return document.ID
	})
}

func (r *Neo4jDocumentRepository) Update(ctx context.Context, id model.ID, opts UpdateDocumentOpts) (*Document, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/Update")
	defer span.End()

	cypher := `
	MATCH (d:` + id.Label() + ` {id: $id})
	SET d += $patch, d.updated_at = datetime()
	RETURN d.id AS id`

	params := map[string]any{
		"id":    id.String(),
		"patch": opts.patch(),
	}

	_, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	})
	if err != nil {
		return nil, errors.Join(ErrDocumentUpdate, err)
	}

	return r.Get(ctx, id, DocumentDetailProjection())
}

func (r *Neo4jDocumentRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/Delete")
	defer span.End()

	cypher := `MATCH (d:` + id.Label() + ` {id: $id}) DETACH DELETE d`
	params := map[string]any{
		"id": id.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrDocumentDelete, err)
	}

	return nil
}

// NewNeo4jDocumentRepository creates a new document neo4jBaseRepository.
func NewNeo4jDocumentRepository(opts ...Neo4jRepositoryOption) (*Neo4jDocumentRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jDocumentRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

func clearDocumentsPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeDocument.String(), pattern))
}

func clearDocumentsKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return clearDocumentsPattern(ctx, r, "Get", id.String(), "*")
}

func clearDocumentBelongsTo(ctx context.Context, r *redisBaseRepository, belongsToID model.ID) error {
	return clearDocumentsPattern(ctx, r, "ListBelongsTo", belongsToID.String(), "*", "*", "*")
}

func clearDocumentAllBelongsTo(ctx context.Context, r *redisBaseRepository) error {
	return clearDocumentsPattern(ctx, r, "ListBelongsTo", "*")
}

func clearDocumentByCreator(ctx context.Context, r *redisBaseRepository, createdByID model.ID) error {
	return clearDocumentsPattern(ctx, r, "ListByCreator", createdByID.String(), "*", "*", "*")
}

func clearDocumentAllByCreator(ctx context.Context, r *redisBaseRepository) error {
	return clearDocumentsPattern(ctx, r, "ListByCreator", "*")
}

func clearDocumentAllCrossCache(ctx context.Context, r *redisBaseRepository) error {
	deleteFns := []func(context.Context, *redisBaseRepository, ...string) error{
		clearNamespacesPattern,
		clearProjectsPattern,
		clearUsersPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// RedisCachedDocumentRepository implements caching on the DocumentRepository.
type RedisCachedDocumentRepository struct {
	cacheRepo    *redisBaseRepository
	documentRepo DocumentRepository
}

func (r *RedisCachedDocumentRepository) Create(ctx context.Context, opts CreateDocumentOpts) (*Document, error) {
	if err := clearDocumentBelongsTo(ctx, r.cacheRepo, opts.BelongsTo); err != nil {
		return nil, err
	}

	if err := clearDocumentByCreator(ctx, r.cacheRepo, opts.CreatedBy); err != nil {
		return nil, err
	}

	if err := clearDocumentAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return r.documentRepo.Create(ctx, opts)
}

func (r *RedisCachedDocumentRepository) Get(ctx context.Context, id model.ID, proj DocumentProjection) (*Document, error) {
	var document *Document
	var err error

	key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(proj))
	if err = r.cacheRepo.Get(ctx, key, &document); err != nil {
		return nil, err
	}

	if document != nil {
		return document, nil
	}

	if document, err = r.documentRepo.Get(ctx, id, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, document); err != nil {
		return nil, err
	}

	return document, nil
}

func (r *RedisCachedDocumentRepository) ListByCreator(ctx context.Context, createdBy model.ID, page CursorPage, proj DocumentProjection) (Page[*Document], error) {
	var documents Page[*Document]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*Document]{}, err
	}

	key := composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", createdBy.String(), projectionCacheValue(proj), pageTokenValue(normalized.Token), normalized.Size)
	if err = r.cacheRepo.Get(ctx, key, &documents); err != nil {
		return Page[*Document]{}, err
	}

	if documents.Items != nil {
		return documents, nil
	}

	if documents, err = r.documentRepo.ListByCreator(ctx, createdBy, normalized, proj); err != nil {
		return Page[*Document]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, documents); err != nil {
		return Page[*Document]{}, err
	}

	return documents, nil
}

func (r *RedisCachedDocumentRepository) ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage, proj DocumentProjection) (Page[*Document], error) {
	var documents Page[*Document]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*Document]{}, err
	}

	key := composeCacheKey(model.ResourceTypeDocument.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(proj), pageTokenValue(normalized.Token), normalized.Size)
	if err = r.cacheRepo.Get(ctx, key, &documents); err != nil {
		return Page[*Document]{}, err
	}

	if documents.Items != nil {
		return documents, nil
	}

	if documents, err = r.documentRepo.ListBelongsTo(ctx, belongsTo, normalized, proj); err != nil {
		return Page[*Document]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, documents); err != nil {
		return Page[*Document]{}, err
	}

	return documents, nil
}

func (r *RedisCachedDocumentRepository) Update(ctx context.Context, id model.ID, opts UpdateDocumentOpts) (*Document, error) {
	document, err := r.documentRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeDocument.String(), "Get", id.String(), projectionCacheValue(DocumentDetailProjection()))
	if err = r.cacheRepo.Set(ctx, key, document); err != nil {
		return nil, err
	}

	if err := clearDocumentAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	if err = clearDocumentByCreator(ctx, r.cacheRepo, document.CreatedBy.ID); err != nil {
		return nil, err
	}

	return document, nil
}

func (r *RedisCachedDocumentRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearDocumentsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearDocumentAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearDocumentAllByCreator(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearDocumentAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.documentRepo.Delete(ctx, id)
}

// NewCachedDocumentRepository returns a new CachedDocumentRepository.
func NewCachedDocumentRepository(repo DocumentRepository, opts ...RedisRepositoryOption) (*RedisCachedDocumentRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedDocumentRepository{
		cacheRepo:    r,
		documentRepo: repo,
	}, nil
}
