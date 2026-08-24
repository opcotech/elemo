package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrCommentCreate = errors.New("failed to create comment") // the comment could not be created
	ErrCommentDelete = errors.New("failed to delete comment") // the comment could not be deleted
	ErrCommentRead   = errors.New("failed to read comment")   // the comment could not be retrieved
	ErrCommentUpdate = errors.New("failed to update comment") // the comment could not be updated
)

// Comment represents a comment persisted by the repository.
type Comment struct {
	ID        model.ID   `json:"id"`
	Content   string     `json:"content"`
	CreatedBy model.ID   `json:"created_by"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// CreateCommentOpts holds the data required to create a comment.
type CreateCommentOpts struct {
	BelongsTo model.ID
	Content   string
	CreatedBy model.ID
}

// UpdateCommentOpts holds the fields that can be updated on a comment.
type UpdateCommentOpts struct {
	Content string
}

//go:generate go tool mockgen -source=comment.go -destination=mock/mock_comment_gen.go -package=mockrepo
type CommentRepository interface {
	Create(ctx context.Context, opts CreateCommentOpts) (*Comment, error)
	Get(ctx context.Context, id model.ID) (*Comment, error)
	ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage) (Page[*Comment], error)
	Update(ctx context.Context, id model.ID, opts UpdateCommentOpts) (*Comment, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jCommentRepository is a repository for managing comments.
type Neo4jCommentRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jCommentRepository) scan(cp, op string) func(rec *neo4j.Record) (*Comment, error) {
	return func(rec *neo4j.Record) (*Comment, error) {
		comment := new(Comment)

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

		return comment, nil
	}
}

func (r *Neo4jCommentRepository) Create(ctx context.Context, opts CreateCommentOpts) (*Comment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeComment)

	cypher := `
	MATCH (b:` + opts.BelongsTo.Label() + ` {id: $belong_to_id})
	MATCH (o:` + opts.CreatedBy.Label() + ` {id: $created_by_id})
	CREATE
		(c:` + id.Label() + ` {id: $id, content: $content, created_by: $created_by_id, created_at: datetime($created_at)}),
		(b)-[:` + EdgeKindHasComment.String() + ` {id: $has_comment_rel_id, created_at: datetime($created_at)}]->(c),
		(c)-[:` + EdgeKindInScopeOf.String() + ` {id: $scope_id, created_at: datetime($created_at)}]->(b),
		(o)-[:` + EdgeKindCommented.String() + ` {id: $commented_rel_id, created_at: datetime($created_at)}]->(c)`

	params := map[string]any{
		"belong_to_id":       opts.BelongsTo.String(),
		"has_comment_rel_id": model.NewRawID(),
		"scope_id":           model.NewRawID(),
		"created_by_id":      opts.CreatedBy.String(),
		"commented_rel_id":   model.NewRawID(),
		"id":                 id.String(),
		"content":            opts.Content,
		"created_at":         createdAt.Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrCommentCreate, err)
	}

	return r.Get(ctx, id)
}

func (r *Neo4jCommentRepository) Get(ctx context.Context, id model.ID) (*Comment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/Get")
	defer span.End()

	plan, err := CompileQuery(CommentGetQuery{ID: id})
	if err != nil {
		return nil, errors.Join(ErrCommentRead, err)
	}

	var comment *Comment
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var runErr error
		comment, _, runErr = Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scan("c", "o"))
		return runErr
	})
	if err != nil {
		return nil, errors.Join(ErrCommentRead, err)
	}

	return comment, nil
}

func (r *Neo4jCommentRepository) ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage) (Page[*Comment], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/ListBelongsTo")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Comment]{}, errors.Join(ErrCommentRead, err)
	}
	plan, err := CompileQuery(CommentListBelongsToQuery{
		BelongsTo: belongsTo,
		Page:      normalized,
		Order:     SortDirectionDesc,
	})
	if err != nil {
		return Page[*Comment]{}, errors.Join(ErrCommentRead, err)
	}

	comments := make([]*Comment, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var runErr error
		comments, _, runErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan("c", "o"))
		return runErr
	})
	if err != nil {
		return Page[*Comment]{}, errors.Join(ErrCommentRead, err)
	}

	return PaginateSlice(comments, normalized.Size, func(comment *Comment) model.ID {
		return comment.ID
	})
}

func (r *Neo4jCommentRepository) Update(ctx context.Context, id model.ID, opts UpdateCommentOpts) (*Comment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/Update")
	defer span.End()

	cypher := `
	MATCH (c:` + id.Label() + ` {id: $id})
	SET c.content = $content, c.updated_at = datetime()
	RETURN c.id AS id`

	params := map[string]any{
		"id":      id.String(),
		"content": opts.Content,
	}

	if _, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	}); err != nil {
		return nil, errors.Join(ErrCommentUpdate, err)
	}

	return r.Get(ctx, id)
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
	return clearCommentsPattern(ctx, r, "Get", id.String()+"*")
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

	return clearCommentsPattern(ctx, r, "ListBelongsTo", resourceID.String(), "*", "*")
}

func clearCommentAllBelongsTo(ctx context.Context, r *redisBaseRepository) error {
	return clearCommentsPattern(ctx, r, "ListBelongsTo", "*", "*", "*")
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

// RedisCachedCommentRepository implements caching on the CommentRepository.
type RedisCachedCommentRepository struct {
	cacheRepo   *redisBaseRepository
	commentRepo CommentRepository
}

func (r *RedisCachedCommentRepository) Create(ctx context.Context, opts CreateCommentOpts) (*Comment, error) {
	if err := clearCommentBelongsTo(ctx, r.cacheRepo, opts.BelongsTo); err != nil {
		return nil, err
	}
	return r.commentRepo.Create(ctx, opts)
}

func (r *RedisCachedCommentRepository) Get(ctx context.Context, id model.ID) (*Comment, error) {
	var comment *Comment
	var err error

	key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String())
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

func (r *RedisCachedCommentRepository) ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage) (Page[*Comment], error) {
	var comments Page[*Comment]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*Comment]{}, err
	}

	key := composeCacheKey(
		model.ResourceTypeComment.String(),
		"ListBelongsTo",
		belongsTo.String(),
		pageTokenValue(normalized.Token),
		normalized.Size,
	)
	if err = r.cacheRepo.Get(ctx, key, &comments); err != nil {
		return Page[*Comment]{}, err
	}

	if comments.Items != nil {
		return comments, nil
	}

	if comments, err = r.commentRepo.ListBelongsTo(ctx, belongsTo, normalized); err != nil {
		return Page[*Comment]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, comments); err != nil {
		return Page[*Comment]{}, err
	}

	return comments, nil
}

func (r *RedisCachedCommentRepository) Update(ctx context.Context, id model.ID, opts UpdateCommentOpts) (*Comment, error) {
	comment, err := r.commentRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeComment.String(), "Get", id.String())
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
