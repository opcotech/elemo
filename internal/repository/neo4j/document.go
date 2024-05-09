package neo4j

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// DocumentRepository is a repository for managing documents.
type DocumentRepository struct {
	*baseRepository
}

func (r *DocumentRepository) scan(dp, cp, lp, commp, ap string) func(rec *neo4j.Record) (*model.Document, error) {
	return func(rec *neo4j.Record) (*model.Document, error) {
		doc := new(model.Document)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, dp)
		if err != nil {
			return nil, err
		}

		createdBy, err := ParseValueFromRecord[string](rec, cp)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &doc, []string{"id", "created_by"}); err != nil {
			return nil, err
		}

		doc.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeDocument.String())
		doc.CreatedBy, _ = model.NewIDFromString(createdBy, model.ResourceTypeUser.String())

		if doc.Labels, err = ParseIDsFromRecord(rec, lp, model.ResourceTypeLabel.String()); err != nil {
			return nil, err
		}

		if doc.Comments, err = ParseIDsFromRecord(rec, commp, model.ResourceTypeComment.String()); err != nil {
			return nil, err
		}

		if doc.Attachments, err = ParseIDsFromRecord(rec, ap, model.ResourceTypeAttachment.String()); err != nil {
			return nil, err
		}

		if err := doc.Validate(); err != nil {
			return nil, err
		}

		return doc, nil
	}
}

func (r *DocumentRepository) Create(ctx context.Context, belongsTo model.ID, document *model.Document) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/Create")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return errors.Join(repository.ErrDocumentCreate, err)
	}

	if err := document.Validate(); err != nil {
		return errors.Join(repository.ErrDocumentCreate, err)
	}

	createdAt := time.Now().UTC()

	document.ID = model.MustNewID(model.ResourceTypeDocument)
	document.CreatedAt = convert.ToPointer(createdAt)
	document.UpdatedAt = nil

	cypher := `
	MATCH (b:` + belongsTo.Label() + ` {id: $belong_to_id}), (o:` + document.CreatedBy.Label() + ` {id: $created_by_id})
	CREATE
		(d:` + document.ID.Label() + ` {
			id: $id, name: $name, excerpt: $excerpt, file_id: $file_id, created_by: $created_by_id,
			created_at: datetime($created_at)
		}),
		(d)-[:` + EdgeKindBelongsTo.String() + ` {id: $belongs_to_rel_id, created_at: datetime($created_at)}]->(b),
		(o)-[:` + EdgeKindCreated.String() + ` {id: $created_rel_id, created_at: datetime($created_at)}]->(d)`

	params := map[string]any{
		"belong_to_id":      belongsTo.String(),
		"belongs_to_rel_id": model.NewRawID(),
		"created_by_id":     document.CreatedBy.String(),
		"created_rel_id":    model.NewRawID(),
		"id":                document.ID.String(),
		"name":              document.Name,
		"excerpt":           document.Excerpt,
		"file_id":           document.FileID,
		"created_at":        createdAt.Format(time.RFC3339Nano),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrDocumentCreate, err)
	}

	return nil
}

func (r *DocumentRepository) Get(ctx context.Context, id model.ID) (*model.Document, error) {
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

	doc, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("d", "c", "l", "comm", "att"))
	if err != nil {
		return nil, errors.Join(repository.ErrDocumentRead, err)
	}

	return doc, nil
}

func (r *DocumentRepository) GetByCreator(ctx context.Context, createdBy model.ID, offset, limit int) ([]*model.Document, error) {
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

	docs, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("d", "c", "l", "comm", "att"))
	if err != nil {
		return nil, errors.Join(repository.ErrDocumentRead, err)
	}

	return docs, nil
}

func (r *DocumentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Document, error) {
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

	docs, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("d", "c", "l", "comm", "att"))
	if err != nil {
		return nil, errors.Join(repository.ErrDocumentRead, err)
	}

	return docs, nil
}

func (r *DocumentRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Document, error) {
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
		"id":         id.String(),
		"patch":      patch,
	}

	doc, err := ExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("d", "c", "l", "comm", "att"))
	if err != nil {
		return nil, errors.Join(repository.ErrDocumentUpdate, err)
	}

	return doc, nil
}

func (r *DocumentRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/Delete")
	defer span.End()

	cypher := `MATCH (d:` + id.Label() + ` {id: $id}) DETACH DELETE d`
	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrDocumentDelete, err)
	}

	return nil
}

// NewDocumentRepository creates a new document baseRepository.
func NewDocumentRepository(opts ...RepositoryOption) (*DocumentRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &DocumentRepository{
		baseRepository: baseRepo,
	}, nil
}
