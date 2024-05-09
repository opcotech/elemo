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

// AttachmentRepository is a repository for managing attachments.
type AttachmentRepository struct {
	*baseRepository
}

func (r *AttachmentRepository) scan(cp, op string) func(rec *neo4j.Record) (*model.Attachment, error) {
	return func(rec *neo4j.Record) (*model.Attachment, error) {
		attachment := new(model.Attachment)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, cp)
		if err != nil {
			return nil, err
		}

		createdBy, err := ParseValueFromRecord[string](rec, op)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &attachment, []string{"id", "created_by"}); err != nil {
			return nil, err
		}

		attachment.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeAttachment.String())
		attachment.CreatedBy, _ = model.NewIDFromString(createdBy, model.ResourceTypeUser.String())

		if err := attachment.Validate(); err != nil {
			return nil, err
		}

		return attachment, nil
	}
}

func (r *AttachmentRepository) Create(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AttachmentRepository/Create")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return errors.Join(repository.ErrAttachmentCreate, err)
	}

	if err := attachment.Validate(); err != nil {
		return errors.Join(repository.ErrAttachmentCreate, err)
	}

	createdAt := time.Now().UTC()

	attachment.ID = model.MustNewID(model.ResourceTypeAttachment)
	attachment.CreatedAt = convert.ToPointer(createdAt)
	attachment.UpdatedAt = nil

	cypher := `
	MATCH (b:` + belongsTo.Label() + ` {id: $belong_to_id}), (o:` + attachment.CreatedBy.Label() + ` {id: $created_by_id})
	CREATE
		(a:` + attachment.ID.Label() + ` {
			id: $id, name: $name, file_id: $file_id, created_by: $created_by_id, created_at: datetime($created_at)
		}),
		(b)-[:` + EdgeKindHasAttachment.String() + ` {id: $has_attachment_rel_id, created_at: datetime($created_at)}]->(a),
		(o)-[:` + EdgeKindCreated.String() + ` {id: $attachment_rel_id, created_at: datetime($created_at)}]->(a)`

	params := map[string]any{
		"belong_to_id":          belongsTo.String(),
		"has_attachment_rel_id": model.NewRawID(),
		"created_by_id":         attachment.CreatedBy.String(),
		"attachment_rel_id":     model.NewRawID(),
		"id":                    attachment.ID.String(),
		"name":                  attachment.Name,
		"file_id":               attachment.FileID,
		"created_at":            createdAt.Format(time.RFC3339Nano),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrAttachmentCreate, err)
	}

	return nil
}

func (r *AttachmentRepository) Get(ctx context.Context, id model.ID) (*model.Attachment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AttachmentRepository/Get")
	defer span.End()

	cypher := `
	MATCH (a:` + id.Label() + ` {id: $id})<-[:` + EdgeKindCreated.String() + `]-(o:` + model.ResourceTypeUser.String() + `)
	RETURN a, o.id AS o`

	params := map[string]any{
		"id": id.String(),
	}

	doc, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("a", "o"))
	if err != nil {
		return nil, errors.Join(repository.ErrAttachmentRead, err)
	}

	return doc, nil
}

func (r *AttachmentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Attachment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AttachmentRepository/GetAllBelongsTo")
	defer span.End()

	cypher := `
	MATCH
		(:` + belongsTo.Label() + ` {id: $id})-[:` + EdgeKindHasAttachment.String() + `]->(a:` + model.ResourceTypeAttachment.String() + `),
		(o:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(a)
	RETURN a, o.id AS o
	ORDER BY a.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"id":     belongsTo.String(),
		"offset": offset,
		"limit":  limit,
	}

	docs, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("a", "o"))
	if err != nil {
		return nil, errors.Join(repository.ErrAttachmentRead, err)
	}

	return docs, nil
}

func (r *AttachmentRepository) Update(ctx context.Context, id model.ID, name string) (*model.Attachment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AttachmentRepository/Update")
	defer span.End()

	cypher := `
	MATCH (a:` + id.Label() + ` {id: $id})
	SET a.name = $name, a.updated_at = datetime()
	WITH a
	MATCH (o:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(a)
	RETURN a, o.id AS o`

	params := map[string]any{
		"id":         id.String(),
		"name":       name,
	}

	doc, err := ExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("a", "o"))
	if err != nil {
		return nil, errors.Join(repository.ErrAttachmentUpdate, err)
	}

	return doc, nil
}

func (r *AttachmentRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AttachmentRepository/Delete")
	defer span.End()

	cypher := `MATCH (a:` + id.Label() + ` {id: $id}) DETACH DELETE a`
	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrAttachmentDelete, err)
	}

	return nil
}

// NewAttachmentRepository creates a new attachment baseRepository.
func NewAttachmentRepository(opts ...RepositoryOption) (*AttachmentRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &AttachmentRepository{
		baseRepository: baseRepo,
	}, nil
}
