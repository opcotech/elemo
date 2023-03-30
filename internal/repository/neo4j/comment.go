package neo4j

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
	ErrCommentRead   = errors.New("failed to read comment")   // the comment could not be retrieved
	ErrCommentUpdate = errors.New("failed to update comment") // the comment could not be updated
	ErrCommentDelete = errors.New("failed to delete comment") // the comment could not be deleted
)

// CommentRepository is a repository for managing comments.
type CommentRepository struct {
	*repository
}

func (r *CommentRepository) scan(cp, op string) func(rec *neo4j.Record) (*model.Comment, error) {
	return func(rec *neo4j.Record) (*model.Comment, error) {
		comment := new(model.Comment)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, cp)
		if err != nil {
			return nil, err
		}

		createdBy, err := ParseValueFromRecord[string](rec, op)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &comment, []string{"id", "created_by"}); err != nil {
			return nil, err
		}

		comment.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.CommentIDType)
		comment.CreatedBy, _ = model.NewIDFromString(createdBy, model.UserIDType)

		if err := comment.Validate(); err != nil {
			return nil, err
		}

		return comment, nil
	}
}

func (r *CommentRepository) Create(ctx context.Context, belongsTo model.ID, comment *model.Comment) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/Create")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return errors.Join(ErrCommentCreate, err)
	}

	if err := comment.Validate(); err != nil {
		return errors.Join(ErrCommentCreate, err)
	}

	createdAt := time.Now()

	hasCommentRelID := model.MustNewID(EdgeKindHasComment.String())
	commentedRelID := model.MustNewID(EdgeKindCommented.String())
	commentPermRelID := model.MustNewID(EdgeKindHasPermission.String())

	comment.ID = model.MustNewID(model.CommentIDType)
	comment.CreatedAt = convert.ToPointer(createdAt)
	comment.UpdatedAt = nil

	cypher := `
	MATCH (b:` + belongsTo.Label() + ` {id: $belong_to_id}), (o:` + comment.CreatedBy.Label() + ` {id: $created_by_id})
	CREATE
		(c:` + comment.ID.Label() + ` {id: $id, content: $content, created_by: $created_by_id, created_at: datetime($created_at)}),
		(b)-[:` + hasCommentRelID.Label() + ` {id: $has_comment_rel_id, created_at: datetime($created_at)}]->(c),
		(o)-[:` + commentedRelID.Label() + ` {id: $commented_rel_id, created_at: datetime($created_at)}]->(c),
		(o)-[:` + commentPermRelID.Label() + ` {id: $comment_perm_rel_id, kind: $perm_kind, created_at: datetime($created_at)}]->(c)`

	params := map[string]any{
		"belong_to_id":        belongsTo.String(),
		"has_comment_rel_id":  hasCommentRelID.String(),
		"created_by_id":       comment.CreatedBy.String(),
		"commented_rel_id":    commentedRelID.String(),
		"comment_perm_rel_id": commentPermRelID.String(),
		"perm_kind":           model.PermissionKindAll,
		"id":                  comment.ID.String(),
		"content":             comment.Content,
		"created_at":          createdAt.Format(time.RFC3339Nano),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrCommentCreate, err)
	}

	return nil
}

func (r *CommentRepository) Get(ctx context.Context, id model.ID) (*model.Comment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/Get")
	defer span.End()

	cypher := `
	MATCH (c:` + id.Label() + ` {id: $id})<-[:` + EdgeKindCommented.String() + `]-(o:` + model.UserIDType + `)
	RETURN c, o.id AS o`

	params := map[string]any{
		"id": id.String(),
	}

	doc, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("c", "o"))
	if err != nil {
		return nil, errors.Join(ErrCommentRead, err)
	}

	return doc, nil
}

func (r *CommentRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Comment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/GetAllBelongsTo")
	defer span.End()

	cypher := `
	MATCH
		(:` + belongsTo.Label() + ` {id: $id})-[:` + EdgeKindHasComment.String() + `]->(c:` + model.CommentIDType + `),
		(o:` + model.UserIDType + `)-[:` + EdgeKindCommented.String() + `]->(c)
	RETURN c, o.id AS o
	ORDER BY c.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"id":     belongsTo.String(),
		"offset": offset,
		"limit":  limit,
	}

	docs, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("c", "o"))
	if err != nil {
		return nil, errors.Join(ErrCommentRead, err)
	}

	return docs, nil
}

func (r *CommentRepository) Update(ctx context.Context, id model.ID, content string) (*model.Comment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/Update")
	defer span.End()

	cypher := `
	MATCH (c:` + id.Label() + ` {id: $id})
	SET c.content = $content, c.updated_at = datetime($updated_at)
	WITH c
	MATCH (o:` + model.UserIDType + `)-[:` + EdgeKindCommented.String() + `]->(c)
	RETURN c, o.id AS o`

	params := map[string]any{
		"id":         id.String(),
		"content":    content,
		"updated_at": time.Now().Format(time.RFC3339Nano),
	}

	doc, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("c", "o"))
	if err != nil {
		return nil, errors.Join(ErrCommentUpdate, err)
	}

	return doc, nil
}

func (r *CommentRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.CommentRepository/Delete")
	defer span.End()

	cypher := `MATCH (d:` + id.Label() + ` {id: $id}) DETACH DELETE d`
	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrCommentDelete, err)
	}

	return nil
}

// NewCommentRepository creates a new comment repository.
func NewCommentRepository(opts ...RepositoryOption) (*CommentRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CommentRepository{
		repository: baseRepo,
	}, nil
}
