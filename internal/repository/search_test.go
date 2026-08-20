package repository

import (
	"context"
	"testing"

	"github.com/meilisearch/meilisearch-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func testSearchClient() meilisearch.ServiceManager {
	return meilisearch.New("http://127.0.0.1:1")
}

func TestNewSearchDatabase(t *testing.T) {
	t.Parallel()

	client := testSearchClient()
	logger := mock.NewMockLogger(nil)
	tracer := mock.NewMockTracer(nil)

	tests := []struct {
		name    string
		opts    []SearchDatabaseOption
		wantErr error
	}{
		{
			name: "create new database",
			opts: []SearchDatabaseOption{
				WithSearchClient(client),
				WithSearchDatabaseLogger(logger),
				WithSearchDatabaseTracer(tracer),
			},
		},
		{
			name: "create new database with nil client",
			opts: []SearchDatabaseOption{
				WithSearchClient(nil),
				WithSearchDatabaseLogger(logger),
				WithSearchDatabaseTracer(tracer),
			},
			wantErr: ErrNoClient,
		},
		{
			name: "create new database with nil logger",
			opts: []SearchDatabaseOption{
				WithSearchClient(client),
				WithSearchDatabaseLogger(nil),
				WithSearchDatabaseTracer(tracer),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new database with nil tracer",
			opts: []SearchDatabaseOption{
				WithSearchClient(client),
				WithSearchDatabaseLogger(logger),
				WithSearchDatabaseTracer(nil),
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, err := NewSearchDatabase(tt.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				assert.Nil(t, db)
				return
			}
			require.NotNil(t, db)
			assert.Equal(t, client, db.client)
			assert.Equal(t, logger, db.logger)
			assert.Equal(t, tracer, db.tracer)
		})
	}
}

func TestNewMeilisearchClient(t *testing.T) {
	t.Parallel()

	t.Run("nil config", func(t *testing.T) {
		t.Parallel()

		client, err := NewMeilisearchClient(nil)
		require.ErrorIs(t, err, config.ErrNoConfig)
		assert.Nil(t, client)
	})

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		client, err := NewMeilisearchClient(&config.SearchConfig{
			Host: "127.0.0.1",
			Port: 7700,
		})
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestNewMeilisearchSearchRepository(t *testing.T) {
	t.Parallel()

	db, err := NewSearchDatabase(WithSearchClient(testSearchClient()))
	require.NoError(t, err)

	logger := mock.NewMockLogger(nil)
	tracer := mock.NewMockTracer(nil)

	tests := []struct {
		name    string
		opts    []SearchRepositoryOption
		wantErr error
	}{
		{
			name: "create new repository",
			opts: []SearchRepositoryOption{
				WithSearchDatabase(db),
				WithSearchIndex("elemo"),
				WithSearchRepositoryLogger(logger),
				WithSearchRepositoryTracer(tracer),
			},
		},
		{
			name: "create new repository with nil database",
			opts: []SearchRepositoryOption{
				WithSearchDatabase(nil),
				WithSearchIndex("elemo"),
				WithSearchRepositoryLogger(logger),
				WithSearchRepositoryTracer(tracer),
			},
			wantErr: ErrNoDriver,
		},
		{
			name: "create new repository with empty index",
			opts: []SearchRepositoryOption{
				WithSearchDatabase(db),
				WithSearchIndex(""),
				WithSearchRepositoryLogger(logger),
				WithSearchRepositoryTracer(tracer),
			},
			wantErr: ErrNoBucket,
		},
		{
			name: "create new repository with nil logger",
			opts: []SearchRepositoryOption{
				WithSearchDatabase(db),
				WithSearchIndex("elemo"),
				WithSearchRepositoryLogger(nil),
				WithSearchRepositoryTracer(tracer),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new repository with nil tracer",
			opts: []SearchRepositoryOption{
				WithSearchDatabase(db),
				WithSearchIndex("elemo"),
				WithSearchRepositoryLogger(logger),
				WithSearchRepositoryTracer(nil),
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo, err := NewMeilisearchSearchRepository(tt.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				assert.Nil(t, repo)
				return
			}
			require.NotNil(t, repo)
			assert.Equal(t, db, repo.db)
			assert.Equal(t, "elemo", repo.indexUID)
			assert.Equal(t, logger, repo.logger)
			assert.Equal(t, tracer, repo.tracer)
		})
	}
}

func TestMeilisearchSearchRepository_Upsert(t *testing.T) {
	t.Parallel()

	t.Run("empty documents is a no-op", func(t *testing.T) {
		t.Parallel()

		repo := &MeilisearchSearchRepository{
			logger: log.DefaultLogger(),
			tracer: tracing.NoopTracer(),
		}
		require.NoError(t, repo.Upsert(context.Background()))
	})
}

func TestMeilisearchSearchRepository_Delete(t *testing.T) {
	t.Parallel()

	t.Run("empty ids is a no-op", func(t *testing.T) {
		t.Parallel()

		repo := &MeilisearchSearchRepository{
			logger: log.DefaultLogger(),
			tracer: tracing.NoopTracer(),
		}
		require.NoError(t, repo.Delete(context.Background()))
	})
}

func TestMeilisearchSearchRepository_DeleteByScope(t *testing.T) {
	t.Parallel()

	t.Run("empty scope fails closed", func(t *testing.T) {
		t.Parallel()

		repo := &MeilisearchSearchRepository{
			logger: log.DefaultLogger(),
			tracer: tracing.NoopTracer(),
		}
		err := repo.DeleteByScope(context.Background(), "")
		require.ErrorIs(t, err, ErrSearchDelete)
		require.ErrorIs(t, err, ErrSearchFilter)
	})
}
