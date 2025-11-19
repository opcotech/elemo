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
	ErrCommentCreate = errors.New("failed to create comment") // the comment could not be created
	ErrCommentDelete = errors.New("failed to delete comment") // the comment could not be deleted
	ErrCommentRead   = errors.New("failed to read comment")   // the comment could not be retrieved
	ErrCommentUpdate = errors.New("failed to update comment") // the comment could not be updated
)

//go:generate mockgen -source=comment.go -destination=../testutil/mock/comment_repo_gen.go -package=mock -mock_names "CommentRepository=CommentRepository"
type CommentRepository interface {
	Create(ctx context.Context, belongsTo model.ID, comment *model.Comment) error
	Get(ctx context.Context, id model.ID) (*model.Comment, error)
	GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Comment, error)
	Update(ctx context.Context, id model.ID, content string) (*model.Comment, error)
	Delete(ctx context.Context, id model.ID) error
}

// CommentRepository is a repository for managing comments.
type Neo4jCommentRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jCommentRepository) scan(cp, op string) func(rec *neo4j.Record) (*model.Comment, error) {
	return func(rec *neo4j.Record) (*model.Comment, error) {
		comment := new(model.Comment)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, cp)
		if err != nil {
			return nil, err
		}

		createdBy, err := Neo4jParseValueFromRecord[string](rec, op)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &comment, []string{"id", "created_by"}); err != nil {
			return nil, err
		}

		comment.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeComment.String())
		comment.CreatedBy, _ = model.NewIDFromString(createdBy, model.ResourceTypeUser.String())

		if err := comment.Validate(); err != nil {
			return nil, err
		}

		return comment, nil
	}
}

func (r *Neo4jCommentRepository) Create(ctx context.Context, belongsTo model.ID, comment *model.Comment) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/Create")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return errors.Join(ErrCommentCreate, err)
	}

	if err := comment.Validate(); err != nil {
		return errors.Join(ErrCommentCreate, err)
	}

	createdAt := time.Now().UTC()

	comment.ID = model.MustNewID(model.ResourceTypeComment)
	comment.CreatedAt = convert.ToPointer(createdAt)
	comment.UpdatedAt = nil

	cypher := `
	MATCH (b:` + belongsTo.Label() + ` {id: $belong_to_id})
	MATCH (o:` + comment.CreatedBy.Label() + ` {id: $created_by_id})
	CREATE
		(c:` + comment.ID.Label() + ` {id: $id, content: $content, created_by: $created_by_id, created_at: datetime($created_at)}),
		(b)-[:` + EdgeKindHasComment.String() + ` {id: $has_comment_rel_id, created_at: datetime($created_at)}]->(c),
		(o)-[:` + EdgeKindCommented.String() + ` {id: $commented_rel_id, created_at: datetime($created_at)}]->(c),
		(o)-[:` + EdgeKindHasPermission.String() + ` {id: $comment_perm_rel_id, kind: $perm_kind, created_at: datetime($created_at)}]->(c)`

	params := map[string]any{
		"belong_to_id":        belongsTo.String(),
		"has_comment_rel_id":  model.NewRawID(),
		"created_by_id":       comment.CreatedBy.String(),
		"commented_rel_id":    model.NewRawID(),
		"comment_perm_rel_id": model.NewRawID(),
		"perm_kind":           model.PermissionKindAll.String(),
		"id":                  comment.ID.String(),
		"content":             comment.Content,
		"created_at":          createdAt.Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrCommentCreate, err)
	}

	return nil
}

func (r *Neo4jCommentRepository) Get(ctx context.Context, id model.ID) (*model.Comment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/Get")
	defer span.End()

	cypher := `
	MATCH (c:` + id.Label() + ` {id: $id})<-[:` + EdgeKindCommented.String() + `]-(o:` + model.ResourceTypeUser.String() + `)
	RETURN c, o.id AS o`

	params := map[string]any{
		"id": id.String(),
	}

	doc, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("c", "o"))
	if err != nil {
		return nil, errors.Join(ErrCommentRead, err)
	}

	return doc, nil
}

func (r *Neo4jCommentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Comment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/GetAllBelongsTo")
	defer span.End()

	cypher := `
	MATCH
		(:` + belongsTo.Label() + ` {id: $id})-[:` + EdgeKindHasComment.String() + `]->(c:` + model.ResourceTypeComment.String() + `),
		(o:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCommented.String() + `]->(c)
	RETURN c, o.id AS o
	ORDER BY c.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"id":     belongsTo.String(),
		"offset": offset,
		"limit":  limit,
	}

	docs, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("c", "o"))
	if err != nil {
		return nil, errors.Join(ErrCommentRead, err)
	}

	return docs, nil
}

func (r *Neo4jCommentRepository) Update(ctx context.Context, id model.ID, content string) (*model.Comment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/Update")
	defer span.End()

	cypher := `
	MATCH (c:` + id.Label() + ` {id: $id})
	SET c.content = $content, c.updated_at = datetime()
	WITH c
	MATCH (o:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindCommented.String() + `]->(c)
	RETURN c, o.id AS o`

	params := map[string]any{
		"id":      id.String(),
		"content": content,
	}

	doc, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("c", "o"))
	if err != nil {
		return nil, errors.Join(ErrCommentUpdate, err)
	}

	return doc, nil
}

func (r *Neo4jCommentRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/Delete")
	defer span.End()

	cypher := `MATCH (d:` + id.Label() + ` {id: $id}) DETACH DELETE d`
	params := map[string]any{
		"id": id.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrCommentDelete, err)
	}

	return nil
}

// NewNeo4jCommentRepository creates a new comment neo4jBaseRepository.
func NewNeo4jCommentRepository(opts ...Neo4jRepositoryOption) (*Neo4jCommentRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jCommentRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

func clearCommentsKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeComment.String(), id.String()))
}

func clearCommentsPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeComment.String(), pattern))
}

func clearCommentBelongsTo(ctx context.Context, r *redisBaseRepository, resourceID model.ID) error {
	switch resourceID.Type {
	case model.ResourceTypeDocument:
		if err := clearDocumentsPattern(ctx, r, "*"); err != nil {
			return err
		}
	case model.ResourceTypeIssue:
		if err := clearIssuesPattern(ctx, r, "*"); err != nil {
			return err
		}
	}

	return clearCommentsPattern(ctx, r, "GetAllBelongsTo", resourceID.String(), "*")
}

func clearCommentAllBelongsTo(ctx context.Context, r *redisBaseRepository) error {
	return clearCommentsPattern(ctx, r, "GetAllBelongsTo", "*")
}

func clearCommentAllCrossCache(ctx context.Context, r *redisBaseRepository) error {
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

// CachedCommentRepository implements caching on the
// repository.CommentRepository.
type RedisCachedCommentRepository struct {
	cacheRepo   *redisBaseRepository
	commentRepo CommentRepository
}

func (r *RedisCachedCommentRepository) Create(ctx context.Context, belongsTo model.ID, comment *model.Comment) error {
	if err := clearCommentBelongsTo(ctx, r.cacheRepo, belongsTo); err != nil {
		return err
	}
	return r.commentRepo.Create(ctx, belongsTo, comment)
}

func (r *RedisCachedCommentRepository) Get(ctx context.Context, id model.ID) (*model.Comment, error) {
	var comment *model.Comment
	var err error

	key := composeCacheKey(model.ResourceTypeComment.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &comment); err != nil {
		return nil, err
	}

	if comment != nil {
		return comment, nil
	}

	if comment, err = r.commentRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (r *RedisCachedCommentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Comment, error) {
	var comments []*model.Comment
	var err error

	key := composeCacheKey(model.ResourceTypeComment.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &comments); err != nil {
		return nil, err
	}

	if comments != nil {
		return comments, nil
	}

	if comments, err = r.commentRepo.GetAllBelongsTo(ctx, belongsTo, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, comments); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *RedisCachedCommentRepository) Update(ctx context.Context, id model.ID, content string) (*model.Comment, error) {
	var comment *model.Comment
	var err error

	comment, err = r.commentRepo.Update(ctx, id, content)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeComment.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, comment); err != nil {
		return nil, err
	}

	if err := clearCommentAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return comment, nil
}

func (r *RedisCachedCommentRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearCommentsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearCommentAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearCommentAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.commentRepo.Delete(ctx, id)
}

// NewCachedCommentRepository returns a new CachedCommentRepository.
func NewCachedCommentRepository(repo CommentRepository, opts ...RedisRepositoryOption) (*RedisCachedCommentRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedCommentRepository{
		cacheRepo:   r,
		commentRepo: repo,
	}, nil
}
