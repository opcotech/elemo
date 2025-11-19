package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrTokenCreate = errors.New("failed to create token") // token cannot be created
	ErrTokenDelete = errors.New("failed to delete token") // token cannot be deleted
	ErrTokenRead   = errors.New("failed to read token")   // token cannot be read
)

//go:generate mockgen -source=auth.go -destination=../testutil/mock/user_token_repo_gen.go -package=mock -mock_names "UserTokenRepository=UserTokenRepository"
type UserTokenRepository interface {
	Create(ctx context.Context, token *model.UserToken) error
	Get(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) (*model.UserToken, error)
	Delete(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) error
}

// UserTokenRepository is a repository for managing user tokens.
type PGUserTokenRepository struct {
	*pgBaseRepository
}

func (r *PGUserTokenRepository) Create(ctx context.Context, token *model.UserToken) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.UserTokenRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC().Round(time.Microsecond)

	token.ID = model.MustNewID(model.ResourceTypeUserToken)
	token.CreatedAt = &createdAt

	query := `
	INSERT INTO user_tokens (id, user_id, sent_to, token, context, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.pool.Exec(ctx, query,
		token.ID, token.UserID, token.SentTo, token.Token,
		token.Context.String(), createdAt,
	)
	if err != nil {
		return errors.Join(ErrTokenCreate, err)
	}

	return nil
}

func (r *PGUserTokenRepository) Get(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) (*model.UserToken, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.UserTokenRepository/Get")
	defer span.End()

	query := `
	SELECT id, user_id, sent_to, token, context, created_at
	FROM user_tokens
	WHERE user_id = $1 AND context = $2`

	var t model.UserToken
	row := r.db.pool.QueryRow(ctx, query, userID, tokenCtx.String())
	if err := row.Scan(&t.ID, &t.UserID, &t.SentTo, &t.Token, &t.Context, &t.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(ErrTokenRead, err)
	}

	return &t, nil
}

func (r *PGUserTokenRepository) Delete(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.UserTokenRepository/DeleteByWorkspaceID")
	defer span.End()

	query := "DELETE FROM user_tokens WHERE user_id = $1 AND context = $2"

	_, err := r.db.pool.Exec(ctx, query, userID, tokenCtx.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return errors.Join(ErrTokenDelete, err)
	}

	return nil
}

// NewUserTokenRepository creates a new UserTokenRepository.
func NewUserTokenRepository(opts ...PGRepositoryOption) (*PGUserTokenRepository, error) {
	baseRepo, err := newPGRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &PGUserTokenRepository{
		pgBaseRepository: baseRepo,
	}, nil
}
