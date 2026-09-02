package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrAttachmentCreate = errors.New("failed to create attachment") // the attachment could not be created
	ErrAttachmentDelete = errors.New("failed to delete attachment") // the attachment could not be deleted
	ErrAttachmentRead   = errors.New("failed to read attachment")   // the attachment could not be retrieved
	ErrAttachmentUpdate = errors.New("failed to update attachment") // the attachment could not be updated
)

// Attachment represents an attachment persisted by the repository.
type Attachment struct {
	ID        model.ID   `json:"id"`
	Name      string     `json:"name"`
	FileID    string     `json:"file_id"`
	CreatedBy model.ID   `json:"created_by"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// CreateAttachmentOpts holds the data required to create an attachment.
type CreateAttachmentOpts struct {
	BelongsTo model.ID
	Name      string
	FileID    string
	CreatedBy model.ID
}

// UpdateAttachmentOpts holds the fields that can be updated on an attachment.
type UpdateAttachmentOpts struct {
	Name string
}

//go:generate go tool mockgen -source=attachment.go -destination=mock/mock_attachment_gen.go -package=mockrepo
type AttachmentRepository interface {
	Create(ctx context.Context, opts CreateAttachmentOpts) (*Attachment, error)
	Get(ctx context.Context, id model.ID, proj AttachmentProjection) (*Attachment, error)
	ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage, proj AttachmentProjection) (Page[*Attachment], error)
	Update(ctx context.Context, id model.ID, opts UpdateAttachmentOpts) (*Attachment, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jAttachmentRepository is a repository for managing attachments.
type Neo4jAttachmentRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jAttachmentRepository) scan(cp, op string) func(rec *neo4j.Record) (*Attachment, error) {
	return func(rec *neo4j.Record) (*Attachment, error) {
		attachment := new(Attachment)

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

		return attachment, nil
	}
}

func (r *Neo4jAttachmentRepository) Create(ctx context.Context, opts CreateAttachmentOpts) (*Attachment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AttachmentRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeAttachment)

	cypher := `
	MATCH (b:` + opts.BelongsTo.Label() + ` {id: $belong_to_id})
	MATCH (o:` + opts.CreatedBy.Label() + ` {id: $created_by_id})
	CREATE
		(a:` + id.Label() + ` {
			id: $id, name: $name, file_id: $file_id, created_by: $created_by_id, created_at: datetime($created_at)
		}),
		(b)-[:` + EdgeKindHasAttachment.String() + ` {id: $has_attachment_rel_id, created_at: datetime($created_at)}]->(a),
		(o)-[:` + EdgeKindCreated.String() + ` {id: $attachment_rel_id, created_at: datetime($created_at)}]->(a)`

	params := map[string]any{
		"belong_to_id":          opts.BelongsTo.String(),
		"has_attachment_rel_id": model.NewRawID(),
		"created_by_id":         opts.CreatedBy.String(),
		"attachment_rel_id":     model.NewRawID(),
		"id":                    id.String(),
		"name":                  opts.Name,
		"file_id":               opts.FileID,
		"created_at":            createdAt.Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrAttachmentCreate, err)
	}

	return r.Get(ctx, id, AttachmentDetailProjection())
}

func (r *Neo4jAttachmentRepository) Get(ctx context.Context, id model.ID, proj AttachmentProjection) (*Attachment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AttachmentRepository/Get")
	defer span.End()

	plan, err := CompileQuery(AttachmentGetQuery{
		ID:         id,
		Projection: proj,
	})
	if err != nil {
		return nil, errors.Join(ErrAttachmentRead, err)
	}

	var attachment *Attachment
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var runErr error
		attachment, _, runErr = Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scan("a", "o"))
		return runErr
	})
	if err != nil {
		return nil, errors.Join(ErrAttachmentRead, err)
	}

	return attachment, nil
}

func (r *Neo4jAttachmentRepository) ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage, proj AttachmentProjection) (Page[*Attachment], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AttachmentRepository/ListBelongsTo")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Attachment]{}, errors.Join(ErrAttachmentRead, err)
	}
	plan, err := CompileQuery(AttachmentListBelongsToQuery{
		BelongsTo:  belongsTo,
		Page:       normalized,
		Order:      SortDirectionDesc,
		Projection: proj,
	})
	if err != nil {
		return Page[*Attachment]{}, errors.Join(ErrAttachmentRead, err)
	}

	attachments := make([]*Attachment, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var runErr error
		attachments, _, runErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan("a", "o"))
		return runErr
	})
	if err != nil {
		return Page[*Attachment]{}, errors.Join(ErrAttachmentRead, err)
	}

	return PaginateSlice(attachments, normalized.Size, func(attachment *Attachment) model.ID {
		return attachment.ID
	})
}

func (r *Neo4jAttachmentRepository) Update(ctx context.Context, id model.ID, opts UpdateAttachmentOpts) (*Attachment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AttachmentRepository/Update")
	defer span.End()

	cypher := `
	MATCH (a:` + id.Label() + ` {id: $id})
	SET a.name = $name
	SET a.updated_at = datetime.statement()
	RETURN a.id AS id`

	params := map[string]any{
		"id":   id.String(),
		"name": opts.Name,
	}

	if _, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	}); err != nil {
		return nil, errors.Join(ErrAttachmentUpdate, err)
	}

	return r.Get(ctx, id, AttachmentDetailProjection())
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
	return clearAttachmentsPattern(ctx, r, "Get", id.String(), "*")
}

func clearAttachmentsPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeAttachment.String(), pattern))
}

func clearAttachmentBelongsTo(ctx context.Context, r *redisBaseRepository, resourceID model.ID) error {
	return clearAttachmentsPattern(ctx, r, "ListBelongsTo", resourceID.String(), "*", "*")
}

func clearAttachmentAllBelongsTo(ctx context.Context, r *redisBaseRepository) error {
	return clearAttachmentsPattern(ctx, r, "ListBelongsTo", "*", "*", "*")
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

func (r *RedisCachedAttachmentRepository) Create(ctx context.Context, opts CreateAttachmentOpts) (*Attachment, error) {
	if err := clearAttachmentBelongsTo(ctx, r.cacheRepo, opts.BelongsTo); err != nil {
		return nil, err
	}

	if err := clearAttachmentAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return r.attachmentRepo.Create(ctx, opts)
}

func (r *RedisCachedAttachmentRepository) Get(ctx context.Context, id model.ID, proj AttachmentProjection) (*Attachment, error) {
	var attachment *Attachment
	var err error

	key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), projectionCacheValue(proj))
	if err = r.cacheRepo.Get(ctx, key, &attachment); err != nil {
		return nil, err
	}

	if attachment != nil {
		return attachment, nil
	}

	if attachment, err = r.attachmentRepo.Get(ctx, id, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, attachment); err != nil {
		return nil, err
	}

	return attachment, nil
}

func (r *RedisCachedAttachmentRepository) ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage, proj AttachmentProjection) (Page[*Attachment], error) {
	var attachments Page[*Attachment]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*Attachment]{}, err
	}

	key := composeCacheKey(
		model.ResourceTypeAttachment.String(),
		"ListBelongsTo",
		belongsTo.String(),
		projectionCacheValue(proj),
		pageTokenValue(normalized.Token),
		normalized.Size,
	)
	if err = r.cacheRepo.Get(ctx, key, &attachments); err != nil {
		return Page[*Attachment]{}, err
	}

	if attachments.Items != nil {
		return attachments, nil
	}

	if attachments, err = r.attachmentRepo.ListBelongsTo(ctx, belongsTo, normalized, proj); err != nil {
		return Page[*Attachment]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, attachments); err != nil {
		return Page[*Attachment]{}, err
	}

	return attachments, nil
}

func (r *RedisCachedAttachmentRepository) Update(ctx context.Context, id model.ID, opts UpdateAttachmentOpts) (*Attachment, error) {
	attachment, err := r.attachmentRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeAttachment.String(), "Get", id.String(), projectionCacheValue(AttachmentDetailProjection()))
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
