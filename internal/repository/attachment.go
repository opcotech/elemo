package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

var (
	ErrAttachmentCreate = errors.New("failed to create attachment") // the attachment could not be created
	ErrAttachmentDelete = errors.New("failed to delete attachment") // the attachment could not be deleted
	ErrAttachmentRead   = errors.New("failed to read attachment")   // the attachment could not be retrieved
	ErrAttachmentUpdate = errors.New("failed to update attachment") // the attachment could not be updated
)

// AttachmentRepository is a repository for managing attachments.
//
//go:generate mockgen -source=attachment.go -destination=../testutil/mock/attachment_repo_gen.go -package=mock -mock_names "AttachmentRepository=AttachmentRepository"
type AttachmentRepository interface {
	Create(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) error
	Get(ctx context.Context, id model.ID) (*model.Attachment, error)
	GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Attachment, error)
	Update(ctx context.Context, id model.ID, name string) (*model.Attachment, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jAttachmentRepository is a repository for managing attachments.
type Neo4jAttachmentRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jAttachmentRepository) scan(cp, op string) func(rec *neo4j.Record) (*model.Attachment, error) {
	return func(rec *neo4j.Record) (*model.Attachment, error) {
		attachment := new(model.Attachment)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, cp)
		if err != nil {
			return nil, err
		}

		createdBy, err := Neo4jParseValueFromRecord[string](rec, op)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &attachment, []string{"id", "created_by"}); err != nil {
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

func (r *Neo4jAttachmentRepository) Create(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AttachmentRepository/Create")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return errors.Join(ErrAttachmentCreate, err)
	}

	if err := attachment.Validate(); err != nil {
		return errors.Join(ErrAttachmentCreate, err)
	}

	createdAt := time.Now().UTC()

	attachment.ID = model.MustNewID(model.ResourceTypeAttachment)
	attachment.CreatedAt = convert.ToPointer(createdAt)
	attachment.UpdatedAt = nil

	cypher := `
	MATCH (b:` + belongsTo.Label() + ` {id: $belong_to_id})
	MATCH (o:` + attachment.CreatedBy.Label() + ` {id: $created_by_id})
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

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrAttachmentCreate, err)
	}

	return nil
}

func (r *Neo4jAttachmentRepository) Get(ctx context.Context, id model.ID) (*model.Attachment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AttachmentRepository/Get")
	defer span.End()

	cypher := `
	MATCH (a:` + id.Label() + ` {id: $id})<-[:` + EdgeKindCreated.String() + `]-(o:` + model.ResourceTypeUser.String() + `)
	RETURN a, o.id AS o`

	params := map[string]any{
		"id": id.String(),
	}

	doc, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("a", "o"))
	if err != nil {
		return nil, errors.Join(ErrAttachmentRead, err)
	}

	return doc, nil
}

func (r *Neo4jAttachmentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Attachment, error) {
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

	docs, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("a", "o"))
	if err != nil {
		return nil, errors.Join(ErrAttachmentRead, err)
	}

	return docs, nil
}

func (r *Neo4jAttachmentRepository) Update(ctx context.Context, id model.ID, name string) (*model.Attachment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AttachmentRepository/Update")
	defer span.End()

	cypher := `
	MATCH (a:` + id.Label() + ` {id: $id})
	SET a.name = $name, a.updated_at = datetime()
	WITH a
	MATCH (o:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCreated.String() + `]->(a)
	RETURN a, o.id AS o`

	params := map[string]any{
		"id":   id.String(),
		"name": name,
	}

	doc, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("a", "o"))
	if err != nil {
		return nil, errors.Join(ErrAttachmentUpdate, err)
	}

	return doc, nil
}

func (r *Neo4jAttachmentRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AttachmentRepository/Delete")
	defer span.End()

	cypher := `MATCH (a:` + id.Label() + ` {id: $id}) DETACH DELETE a`
	params := map[string]any{
		"id": id.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrAttachmentDelete, err)
	}

	return nil
}

// NewNeo4jAttachmentRepository creates a new attachment neo4jBaseRepository.
func NewNeo4jAttachmentRepository(opts ...Neo4jRepositoryOption) (*Neo4jAttachmentRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jAttachmentRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

func clearAttachmentsKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeAttachment.String(), id.String()))
}

func clearAttachmentsPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeAttachment.String(), pattern))
}

func clearAttachmentBelongsTo(ctx context.Context, r *redisBaseRepository, resourceID model.ID) error {
	return clearAttachmentsPattern(ctx, r, "GetAllBelongsTo", resourceID.String(), "*")
}

func clearAttachmentAllBelongsTo(ctx context.Context, r *redisBaseRepository) error {
	return clearAttachmentsPattern(ctx, r, "GetAllBelongsTo", "*")
}

func clearAttachmentAllCrossCache(ctx context.Context, r *redisBaseRepository) error {
	deleteFns := []func(context.Context, *redisBaseRepository, ...string) error{
		clearDocumentsPattern,
		clearIssuesPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// RedisCachedAttachmentRepository implements caching on the AttachmentRepository.
type RedisCachedAttachmentRepository struct {
	cacheRepo      *redisBaseRepository
	attachmentRepo AttachmentRepository
}

func (r *RedisCachedAttachmentRepository) Create(ctx context.Context, belongsTo model.ID, attachment *model.Attachment) error {
	if err := clearAttachmentBelongsTo(ctx, r.cacheRepo, belongsTo); err != nil {
		return err
	}

	if err := clearAttachmentAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.attachmentRepo.Create(ctx, belongsTo, attachment)
}

func (r *RedisCachedAttachmentRepository) Get(ctx context.Context, id model.ID) (*model.Attachment, error) {
	var attachment *model.Attachment
	var err error

	key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &attachment); err != nil {
		return nil, err
	}

	if attachment != nil {
		return attachment, nil
	}

	if attachment, err = r.attachmentRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, attachment); err != nil {
		return nil, err
	}

	return attachment, nil
}

func (r *RedisCachedAttachmentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Attachment, error) {
	var attachments []*model.Attachment
	var err error

	key := composeCacheKey(model.ResourceTypeAttachment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &attachments); err != nil {
		return nil, err
	}

	if attachments != nil {
		return attachments, nil
	}

	if attachments, err = r.attachmentRepo.GetAllBelongsTo(ctx, belongsTo, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, attachments); err != nil {
		return nil, err
	}

	return attachments, nil
}

func (r *RedisCachedAttachmentRepository) Update(ctx context.Context, id model.ID, name string) (*model.Attachment, error) {
	var attachment *model.Attachment
	var err error

	attachment, err = r.attachmentRepo.Update(ctx, id, name)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeAttachment.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, attachment); err != nil {
		return nil, err
	}

	if err = clearAttachmentAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return attachment, nil
}

func (r *RedisCachedAttachmentRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearAttachmentsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearAttachmentAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearAttachmentAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.attachmentRepo.Delete(ctx, id)
}

// NewCachedAttachmentRepository returns a new CachedAttachmentRepository.
func NewCachedAttachmentRepository(repo AttachmentRepository, opts ...RedisRepositoryOption) (*RedisCachedAttachmentRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedAttachmentRepository{
		cacheRepo:      r,
		attachmentRepo: repo,
	}, nil
}
