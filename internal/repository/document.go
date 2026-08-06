package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

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

// Document represents a document persisted by the repository.
type Document struct {
	ID          model.ID   `json:"id"`
	Name        string     `json:"name"`
	Excerpt     string     `json:"excerpt"`
	FileID      string     `json:"file_id"`
	CreatedBy   model.ID   `json:"created_by"`
	Labels      []model.ID `json:"labels"`
	Comments    []model.ID `json:"comments"`
	Attachments []model.ID `json:"attachments"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
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

//go:generate mockgen -source=document.go -destination=document_mock_gen.go -package=repository -mock_names "DocumentRepository=MockDocumentRepository"
type DocumentRepository interface {
	Create(ctx context.Context, opts CreateDocumentOpts) (*Document, error)
	Get(ctx context.Context, id model.ID) (*Document, error)
	GetByCreator(ctx context.Context, createdBy model.ID, offset, limit int) ([]*Document, error)
	GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*Document, error)
	Update(ctx context.Context, id model.ID, opts UpdateDocumentOpts) (*Document, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jDocumentRepository is a repository for managing documents.
type Neo4jDocumentRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jDocumentRepository) scan(dp, cp, lp, commp, ap string) func(rec *neo4j.Record) (*Document, error) {
	return func(rec *neo4j.Record) (*Document, error) {
		doc := new(Document)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, dp)
		if err != nil {
			return nil, err
		}

		createdBy, err := Neo4jParseValueFromRecord[string](rec, cp)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &doc, []string{"id", "created_by"}); err != nil {
			return nil, err
		}

		doc.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeDocument.String())
		doc.CreatedBy, _ = model.NewIDFromString(createdBy, model.ResourceTypeUser.String())

		if doc.Labels, err = Neo4jParseIDsFromRecord(rec, lp, model.ResourceTypeLabel.String()); err != nil {
			return nil, err
		}

		if doc.Comments, err = Neo4jParseIDsFromRecord(rec, commp, model.ResourceTypeComment.String()); err != nil {
			return nil, err
		}

		if doc.Attachments, err = Neo4jParseIDsFromRecord(rec, ap, model.ResourceTypeAttachment.String()); err != nil {
			return nil, err
		}

		return doc, nil
	}
}

func (r *Neo4jDocumentRepository) Create(ctx context.Context, opts CreateDocumentOpts) (*Document, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/Create")
	defer span.End()

	createdAt := convert.ToPointer(time.Now().UTC())

	document := &Document{
		ID:          model.MustNewID(model.ResourceTypeDocument),
		Name:        opts.Name,
		Excerpt:     opts.Excerpt,
		FileID:      opts.FileID,
		CreatedBy:   opts.CreatedBy,
		Labels:      make([]model.ID, 0),
		Comments:    make([]model.ID, 0),
		Attachments: make([]model.ID, 0),
		CreatedAt:   createdAt,
		UpdatedAt:   nil,
	}

	cypher := `
	MATCH (b:` + opts.BelongsTo.Label() + ` {id: $belong_to_id})
	MATCH (o:` + document.CreatedBy.Label() + ` {id: $created_by_id})
	CREATE
		(d:` + document.ID.Label() + ` {
			id: $id, name: $name, excerpt: $excerpt, file_id: $file_id, created_by: $created_by_id,
			created_at: datetime($created_at)
		}),
		(d)-[:` + EdgeKindBelongsTo.String() + ` {id: $belongs_to_rel_id, created_at: datetime($created_at)}]->(b),
		(o)-[:` + EdgeKindCreated.String() + ` {id: $created_rel_id, created_at: datetime($created_at)}]->(d)`

	params := map[string]any{
		"belong_to_id":      opts.BelongsTo.String(),
		"belongs_to_rel_id": model.NewRawID(),
		"created_by_id":     document.CreatedBy.String(),
		"created_rel_id":    model.NewRawID(),
		"id":                document.ID.String(),
		"name":              document.Name,
		"excerpt":           document.Excerpt,
		"file_id":           document.FileID,
		"created_at":        createdAt.Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrDocumentCreate, err)
	}

	return document, nil
}

func (r *Neo4jDocumentRepository) Get(ctx context.Context, id model.ID) (*Document, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/Get")
	defer span.End()

	cypher := `
	MATCH (d:` + id.Label() + ` {id: $id}), (d)<-[:` + EdgeKindCreated.String() + `]-(c:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (d)-[:` + EdgeKindHasLabel.String() + `]->(l:` + model.ResourceTypeLabel.String() + `)
	OPTIONAL MATCH (d)-[:` + EdgeKindHasComment.String() + `]->(comm:` + model.ResourceTypeComment.String() + `)
	OPTIONAL MATCH (d)-[:` + EdgeKindHasAttachment.String() + `]->(att:` + model.ResourceTypeAttachment.String() + `)
	RETURN d, c.id AS c, collect(DISTINCT l.id) AS l, collect(DISTINCT comm.id) AS comm, collect(DISTINCT att.id) AS att`

	params := map[string]any{
		"id": id.String(),
	}

	doc, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("d", "c", "l", "comm", "att"))
	if err != nil {
		return nil, errors.Join(ErrDocumentRead, err)
	}

	return doc, nil
}

func (r *Neo4jDocumentRepository) GetByCreator(ctx context.Context, createdBy model.ID, offset, limit int) ([]*Document, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/GetByCreator")
	defer span.End()

	cypher := `
	MATCH (d:` + model.ResourceTypeDocument.String() + `)<-[:` + EdgeKindCreated.String() + `]-(c:` + createdBy.Label() + ` {id: $id})
	OPTIONAL MATCH (d)-[:` + EdgeKindHasLabel.String() + `]->(l:` + model.ResourceTypeLabel.String() + `)
	OPTIONAL MATCH (d)-[:` + EdgeKindHasComment.String() + `]->(comm:` + model.ResourceTypeComment.String() + `)
	OPTIONAL MATCH (d)-[:` + EdgeKindHasAttachment.String() + `]->(att:` + model.ResourceTypeAttachment.String() + `)
	RETURN d, c.id AS c, collect(DISTINCT l.id) AS l, collect(DISTINCT comm.id) AS comm, collect(DISTINCT att.id) AS att
	ORDER BY d.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"id":     createdBy.String(),
		"offset": offset,
		"limit":  limit,
	}

	docs, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("d", "c", "l", "comm", "att"))
	if err != nil {
		return nil, errors.Join(ErrDocumentRead, err)
	}

	return docs, nil
}

func (r *Neo4jDocumentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*Document, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/GetAllBelongsTo")
	defer span.End()

	cypher := `
	MATCH
		(d:` + model.ResourceTypeDocument.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(b:` + belongsTo.Label() + ` {id: $id}),
		(c:` + model.ResourceTypeUser.String() + `)-[` + EdgeKindCreated.String() + `]->(d)
	OPTIONAL MATCH (d)-[:` + EdgeKindHasLabel.String() + `]->(l:` + model.ResourceTypeLabel.String() + `)
	OPTIONAL MATCH (d)-[:` + EdgeKindHasComment.String() + `]->(comm:` + model.ResourceTypeComment.String() + `)
	OPTIONAL MATCH (d)-[:` + EdgeKindHasAttachment.String() + `]->(att:` + model.ResourceTypeAttachment.String() + `)
	RETURN d, c.id AS c, collect(DISTINCT l.id) AS l, collect(DISTINCT comm.id) AS comm, collect(DISTINCT att.id) AS att
	ORDER BY d.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"id":     belongsTo.String(),
		"offset": offset,
		"limit":  limit,
	}

	docs, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("d", "c", "l", "comm", "att"))
	if err != nil {
		return nil, errors.Join(ErrDocumentRead, err)
	}

	return docs, nil
}

func (r *Neo4jDocumentRepository) Update(ctx context.Context, id model.ID, opts UpdateDocumentOpts) (*Document, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/Update")
	defer span.End()

	cypher := `
	MATCH (d:` + id.Label() + ` {id: $id})
	SET d += $patch, d.updated_at = datetime()
	WITH d
	MATCH (c:` + model.ResourceTypeUser.String() + `)-[` + EdgeKindCreated.String() + `]->(d)
	OPTIONAL MATCH (d)-[:` + EdgeKindHasLabel.String() + `]->(l:` + model.ResourceTypeLabel.String() + `)
	OPTIONAL MATCH (d)-[:` + EdgeKindHasComment.String() + `]->(comm:` + model.ResourceTypeComment.String() + `)
	OPTIONAL MATCH (d)-[:` + EdgeKindHasAttachment.String() + `]->(att:` + model.ResourceTypeAttachment.String() + `)
	RETURN d, c.id AS c, collect(DISTINCT l.id) AS l, collect(DISTINCT comm.id) AS comm, collect(DISTINCT att.id) AS att`

	params := map[string]any{
		"id":    id.String(),
		"patch": opts.patch(),
	}

	doc, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("d", "c", "l", "comm", "att"))
	if err != nil {
		return nil, errors.Join(ErrDocumentUpdate, err)
	}

	return doc, nil
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
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeDocument.String(), id.String()))
}

func clearDocumentBelongsTo(ctx context.Context, r *redisBaseRepository, belongsToID model.ID) error {
	return clearDocumentsPattern(ctx, r, "GetAllBelongsTo", belongsToID.String(), "*")
}

func clearDocumentAllBelongsTo(ctx context.Context, r *redisBaseRepository) error {
	return clearDocumentsPattern(ctx, r, "GetAllBelongsTo", "*")
}

func clearDocumentByCreator(ctx context.Context, r *redisBaseRepository, createdByID model.ID) error {
	return clearDocumentsPattern(ctx, r, "GetByCreator", createdByID.String(), "*")
}

func clearDocumentAllByCreator(ctx context.Context, r *redisBaseRepository) error {
	return clearDocumentsPattern(ctx, r, "GetByCreator", "*")
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

func (r *RedisCachedDocumentRepository) Get(ctx context.Context, id model.ID) (*Document, error) {
	var document *Document
	var err error

	key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &document); err != nil {
		return nil, err
	}

	if document != nil {
		return document, nil
	}

	if document, err = r.documentRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, document); err != nil {
		return nil, err
	}

	return document, nil
}

func (r *RedisCachedDocumentRepository) GetByCreator(ctx context.Context, createdBy model.ID, offset, limit int) ([]*Document, error) {
	var documents []*Document
	var err error

	key := composeCacheKey(model.ResourceTypeDocument.String(), "GetByCreator", createdBy.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &documents); err != nil {
		return nil, err
	}

	if documents != nil {
		return documents, nil
	}

	if documents, err = r.documentRepo.GetByCreator(ctx, createdBy, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, documents); err != nil {
		return nil, err
	}

	return documents, nil
}

func (r *RedisCachedDocumentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*Document, error) {
	var documents []*Document
	var err error

	key := composeCacheKey(model.ResourceTypeDocument.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &documents); err != nil {
		return nil, err
	}

	if documents != nil {
		return documents, nil
	}

	if documents, err = r.documentRepo.GetAllBelongsTo(ctx, belongsTo, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, documents); err != nil {
		return nil, err
	}

	return documents, nil
}

func (r *RedisCachedDocumentRepository) Update(ctx context.Context, id model.ID, opts UpdateDocumentOpts) (*Document, error) {
	document, err := r.documentRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeDocument.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, document); err != nil {
		return nil, err
	}

	if err := clearDocumentAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	if err = clearDocumentByCreator(ctx, r.cacheRepo, document.CreatedBy); err != nil {
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
