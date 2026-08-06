package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

var (
	ErrTokenCreate = errors.New("failed to create token") // token cannot be created
	ErrTokenDelete = errors.New("failed to delete token") // token cannot be deleted
	ErrTokenRead   = errors.New("failed to read token")   // token cannot be read
)

// UserToken represents a user token persisted by the repository.
type UserToken struct {
	ID        model.ID               `json:"id"`
	UserID    model.ID               `json:"user_id"`
	SentTo    string                 `json:"sent_to"`
	Token     string                 `json:"token"`
	Context   model.UserTokenContext `json:"context"`
	CreatedAt *time.Time             `json:"created_at"`
}

// CreateUserTokenOpts holds the data required to create a user token.
type CreateUserTokenOpts struct {
	UserID  model.ID
	SentTo  string
	Token   string
	Context model.UserTokenContext
}

//go:generate go tool mockgen -source=auth.go -destination=auth_mock_gen.go -package=repository -mock_names "UserTokenRepository=MockUserTokenRepository"
type UserTokenRepository interface {
	Create(ctx context.Context, opts CreateUserTokenOpts) (*UserToken, error)
	Get(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) (*UserToken, error)
	Delete(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) error
}

// PGUserTokenRepository is a repository for managing user tokens.
type PGUserTokenRepository struct {
	*pgBaseRepository
}

func (r *PGUserTokenRepository) Create(ctx context.Context, opts CreateUserTokenOpts) (*UserToken, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.UserTokenRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC().Round(time.Microsecond)

	token := &UserToken{
		ID:        model.MustNewID(model.ResourceTypeUserToken),
		UserID:    opts.UserID,
		SentTo:    opts.SentTo,
		Token:     opts.Token,
		Context:   opts.Context,
		CreatedAt: convert.ToPointer(createdAt),
	}

	query := `
	INSERT INTO user_tokens (id, user_id, sent_to, token, context, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.pool.Exec(ctx, query,
		token.ID, token.UserID, token.SentTo, token.Token,
		token.Context.String(), createdAt,
	)
	if err != nil {
		return nil, errors.Join(ErrTokenCreate, err)
	}

	return token, nil
}

func (r *PGUserTokenRepository) Get(ctx context.Context, userID model.ID, tokenCtx model.UserTokenContext) (*UserToken, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.UserTokenRepository/Get")
	defer span.End()

	query := `
	SELECT id, user_id, sent_to, token, context, created_at
	FROM user_tokens
	WHERE user_id = $1 AND context = $2`

	var t UserToken
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
