package pg

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// UserTokenRepository is a repository for managing user tokens.
type UserTokenRepository struct {
	*baseRepository
}

func (r *UserTokenRepository) Create(ctx context.Context, token *model.UserToken) error {
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
		return errors.Join(repository.ErrTokenCreate, err)
	}

	return nil
}

func (r *UserTokenRepository) Get(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) (*model.UserToken, error) {
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
			return nil, repository.ErrNotFound
		}
		return nil, errors.Join(repository.ErrTokenRead, err)
	}

	return &t, nil
}

func (r *UserTokenRepository) Delete(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.UserTokenRepository/DeleteByWorkspaceID")
	defer span.End()

	query := "DELETE FROM user_tokens WHERE user_id = $1 AND context = $2"

	_, err := r.db.pool.Exec(ctx, query, userID, tokenCtx.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrNotFound
		}
		return errors.Join(repository.ErrTokenDelete, err)
	}

	return nil
}

// NewUserTokenRepository creates a new UserTokenRepository.
func NewUserTokenRepository(opts ...RepositoryOption) (*UserTokenRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &UserTokenRepository{
		baseRepository: baseRepo,
	}, nil
}
