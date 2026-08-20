package repository

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	searchPrimaryKey     = "id"
	searchTaskPollEvery  = 50 * time.Millisecond
	searchDefaultTimeout = 5 * time.Second
)

var (
	searchSearchableAttributes = []string{"title", "content", "key"}
	searchFilterableAttributes = []any{
		"type",
		"scope_ids",
		"organization_id",
		"namespace_id",
		"project_id",
	}
	searchSortableAttributes  = []string{"created_at", "updated_at"}
	searchDisplayedAttributes = []string{
		"id",
		"type",
		"title",
		"content",
		"key",
		"organization_id",
		"namespace_id",
		"project_id",
		"created_at",
		"updated_at",
	}
)

// SearchDocument is the Meilisearch projection of an Elemo resource.
// ID is the Meilisearch primary key (ResourceType_xid). Scope and parent
// fields use ResourceType:xid composite identifiers.
type SearchDocument struct {
	ID             string   `json:"id"`
	Type           string   `json:"type"`
	Title          string   `json:"title"`
	Content        string   `json:"content,omitempty"`
	Key            string   `json:"key,omitempty"`
	OrganizationID string   `json:"organization_id,omitempty"`
	NamespaceID    string   `json:"namespace_id,omitempty"`
	ProjectID      string   `json:"project_id,omitempty"`
	ScopeIDs       []string `json:"scope_ids"`
	CreatedAt      int64    `json:"created_at"`
	UpdatedAt      int64    `json:"updated_at"`
}

// SearchTypeFilter restricts hits of one resource type to grant scopes.
type SearchTypeFilter struct {
	Type     string
	ScopeIDs []string
}

// SearchQuery is a permission-aware search request built by the service layer.
type SearchQuery struct {
	Text           string
	TypeFilters    []SearchTypeFilter
	OrganizationID string
	NamespaceID    string
	ProjectID      string
	Offset         int64
	Limit          int64
}

// SearchHits is a page of indexed documents returned by the search engine.
type SearchHits struct {
	Documents []SearchDocument
	Offset    int64
	Limit     int64
}

//go:generate go tool mockgen -source=search.go -destination=search_mock_gen.go -package=repository -mock_names "SearchRepository=MockSearchRepository"
type SearchRepository interface {
	// Upsert writes or replaces searchable documents and waits for the task.
	Upsert(ctx context.Context, docs ...SearchDocument) error
	// Delete removes documents by primary key and waits for the task.
	Delete(ctx context.Context, ids ...string) error
	// DeleteByScope removes every document whose scope_ids contain scopeID.
	DeleteByScope(ctx context.Context, scopeID string) error
	// Search runs a filtered full-text query against the shared index.
	Search(ctx context.Context, q SearchQuery) (*SearchHits, error)
	// Ping reports whether the search engine is available.
	Ping(ctx context.Context) error
	// EnsureIndex creates the index if needed and applies search settings.
	EnsureIndex(ctx context.Context) error
}

// SearchDatabaseOption configures a SearchDatabase.
type SearchDatabaseOption func(*SearchDatabase) error

// WithSearchClient sets the Meilisearch client.
func WithSearchClient(client meilisearch.ServiceManager) SearchDatabaseOption {
	return func(db *SearchDatabase) error {
		if client == nil {
			return ErrNoClient
		}
		db.client = client
		return nil
	}
}

// WithSearchDatabaseLogger sets the logger for a SearchDatabase.
func WithSearchDatabaseLogger(logger log.Logger) SearchDatabaseOption {
	return func(db *SearchDatabase) error {
		if logger == nil {
			return log.ErrNoLogger
		}
		db.logger = logger
		return nil
	}
}

// WithSearchDatabaseTracer sets the tracer for a SearchDatabase.
func WithSearchDatabaseTracer(tracer tracing.Tracer) SearchDatabaseOption {
	return func(db *SearchDatabase) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}
		db.tracer = tracer
		return nil
	}
}

// SearchDatabase wraps a Meilisearch client.
type SearchDatabase struct {
	client meilisearch.ServiceManager `validate:"required"`
	logger log.Logger                 `validate:"required"`
	tracer tracing.Tracer             `validate:"required"`
}

// Client returns the Meilisearch client.
func (db *SearchDatabase) Client() meilisearch.ServiceManager {
	return db.client
}

// Ping checks that the search engine is available.
func (db *SearchDatabase) Ping(ctx context.Context) error {
	health, err := db.client.HealthWithContext(ctx)
	if err != nil {
		return errors.Join(ErrSearchPing, err)
	}
	if health == nil || health.Status != "available" {
		return ErrSearchPing
	}
	return nil
}

// Close releases search engine resources. The SDK owns the HTTP client.
func (db *SearchDatabase) Close() error {
	return nil
}

// NewSearchDatabase creates a SearchDatabase.
func NewSearchDatabase(opts ...SearchDatabaseOption) (*SearchDatabase, error) {
	db := &SearchDatabase{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(db); err != nil {
			return nil, err
		}
	}

	if err := validate.Struct(db); err != nil {
		return nil, errors.Join(ErrInvalidDatabase, err)
	}

	return db, nil
}

// NewMeilisearchClient creates a Meilisearch client from config.
func NewMeilisearchClient(conf *config.SearchConfig) (meilisearch.ServiceManager, error) {
	if conf == nil {
		return nil, config.ErrNoConfig
	}

	timeout := conf.ReadTimeout * time.Second
	if timeout <= 0 {
		timeout = searchDefaultTimeout
	}

	return meilisearch.New(
		conf.URL(),
		meilisearch.WithAPIKey(conf.APIKey),
		meilisearch.WithCustomClient(&http.Client{Timeout: timeout}),
	), nil
}

// SearchRepositoryOption configures a MeilisearchSearchRepository.
type SearchRepositoryOption func(*MeilisearchSearchRepository) error

// WithSearchDatabase sets the search database.
func WithSearchDatabase(db *SearchDatabase) SearchRepositoryOption {
	return func(r *MeilisearchSearchRepository) error {
		if db == nil {
			return ErrNoDriver
		}
		r.db = db
		return nil
	}
}

// WithSearchIndex sets the Meilisearch index UID.
func WithSearchIndex(uid string) SearchRepositoryOption {
	return func(r *MeilisearchSearchRepository) error {
		if uid == "" {
			return ErrNoBucket
		}
		r.indexUID = uid
		return nil
	}
}

// WithSearchRepositoryLogger sets the logger.
func WithSearchRepositoryLogger(logger log.Logger) SearchRepositoryOption {
	return func(r *MeilisearchSearchRepository) error {
		if logger == nil {
			return log.ErrNoLogger
		}
		r.logger = logger
		return nil
	}
}

// WithSearchRepositoryTracer sets the tracer.
func WithSearchRepositoryTracer(tracer tracing.Tracer) SearchRepositoryOption {
	return func(r *MeilisearchSearchRepository) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}
		r.tracer = tracer
		return nil
	}
}

// MeilisearchSearchRepository stores and queries the searchable projection.
type MeilisearchSearchRepository struct {
	db       *SearchDatabase `validate:"required"`
	indexUID string          `validate:"required"`
	logger   log.Logger      `validate:"required"`
	tracer   tracing.Tracer  `validate:"required"`
}

func (r *MeilisearchSearchRepository) index() meilisearch.IndexManager {
	return r.db.Client().Index(r.indexUID)
}

func (r *MeilisearchSearchRepository) waitForTask(ctx context.Context, info *meilisearch.TaskInfo) error {
	if info == nil {
		return nil
	}
	task, err := r.db.Client().WaitForTaskWithContext(ctx, info.TaskUID, searchTaskPollEvery)
	if err != nil {
		return err
	}
	if task.Status != meilisearch.TaskStatusSucceeded {
		if task.Error.Message != "" {
			return errors.New(task.Error.Message)
		}
		return fmt.Errorf("search task %d status %s", task.UID, task.Status)
	}
	return nil
}

func (r *MeilisearchSearchRepository) Ping(ctx context.Context) error {
	ctx, span := r.tracer.Start(ctx, "repository.meilisearch.SearchRepository/Ping")
	defer span.End()
	return r.db.Ping(ctx)
}

func (r *MeilisearchSearchRepository) EnsureIndex(ctx context.Context) error {
	ctx, span := r.tracer.Start(ctx, "repository.meilisearch.SearchRepository/EnsureIndex")
	defer span.End()

	_, err := r.db.Client().GetIndexWithContext(ctx, r.indexUID)
	if err != nil {
		var mErr *meilisearch.Error
		if !errors.As(err, &mErr) || mErr.StatusCode != http.StatusNotFound {
			return errors.Join(ErrSearchIndex, err)
		}
		info, createErr := r.db.Client().CreateIndexWithContext(ctx, &meilisearch.IndexConfig{
			Uid:        r.indexUID,
			PrimaryKey: searchPrimaryKey,
		})
		if createErr != nil {
			if !errors.As(createErr, &mErr) || mErr.StatusCode != http.StatusConflict {
				return errors.Join(ErrSearchIndex, createErr)
			}
		} else if err := r.waitForTask(ctx, info); err != nil {
			return errors.Join(ErrSearchIndex, err)
		}
	}

	idx := r.index()

	info, err := idx.UpdateSearchableAttributesWithContext(ctx, &searchSearchableAttributes)
	if err != nil {
		return errors.Join(ErrSearchIndex, err)
	}
	if err := r.waitForTask(ctx, info); err != nil {
		return errors.Join(ErrSearchIndex, err)
	}

	info, err = idx.UpdateFilterableAttributesWithContext(ctx, &searchFilterableAttributes)
	if err != nil {
		return errors.Join(ErrSearchIndex, err)
	}
	if err := r.waitForTask(ctx, info); err != nil {
		return errors.Join(ErrSearchIndex, err)
	}

	info, err = idx.UpdateSortableAttributesWithContext(ctx, &searchSortableAttributes)
	if err != nil {
		return errors.Join(ErrSearchIndex, err)
	}
	if err := r.waitForTask(ctx, info); err != nil {
		return errors.Join(ErrSearchIndex, err)
	}

	info, err = idx.UpdateDisplayedAttributesWithContext(ctx, &searchDisplayedAttributes)
	if err != nil {
		return errors.Join(ErrSearchIndex, err)
	}
	if err := r.waitForTask(ctx, info); err != nil {
		return errors.Join(ErrSearchIndex, err)
	}

	return nil
}

func (r *MeilisearchSearchRepository) Upsert(ctx context.Context, docs ...SearchDocument) error {
	ctx, span := r.tracer.Start(ctx, "repository.meilisearch.SearchRepository/Upsert")
	defer span.End()

	if len(docs) == 0 {
		return nil
	}

	for i := range docs {
		if docs[i].ScopeIDs == nil {
			docs[i].ScopeIDs = []string{}
		}
	}

	primaryKey := searchPrimaryKey
	info, err := r.index().AddDocumentsWithContext(ctx, docs, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	if err != nil {
		return errors.Join(ErrSearchIndex, err)
	}
	if err := r.waitForTask(ctx, info); err != nil {
		return errors.Join(ErrSearchIndex, err)
	}
	return nil
}

func (r *MeilisearchSearchRepository) Delete(ctx context.Context, ids ...string) error {
	ctx, span := r.tracer.Start(ctx, "repository.meilisearch.SearchRepository/Delete")
	defer span.End()

	if len(ids) == 0 {
		return nil
	}

	info, err := r.index().DeleteDocumentsWithContext(ctx, ids, nil)
	if err != nil {
		return errors.Join(ErrSearchDelete, err)
	}
	if err := r.waitForTask(ctx, info); err != nil {
		return errors.Join(ErrSearchDelete, err)
	}
	return nil
}

func (r *MeilisearchSearchRepository) DeleteByScope(ctx context.Context, scopeID string) error {
	ctx, span := r.tracer.Start(ctx, "repository.meilisearch.SearchRepository/DeleteByScope")
	defer span.End()

	if scopeID == "" {
		return errors.Join(ErrSearchDelete, ErrSearchFilter)
	}

	filter := "scope_ids = " + quoteFilterValue(scopeID)
	info, err := r.index().DeleteDocumentsByFilterWithContext(ctx, filter, nil)
	if err != nil {
		return errors.Join(ErrSearchDelete, err)
	}
	if err := r.waitForTask(ctx, info); err != nil {
		return errors.Join(ErrSearchDelete, err)
	}
	return nil
}

// DeleteAll removes every document in the index. Used by tests and reindex.
func (r *MeilisearchSearchRepository) DeleteAll(ctx context.Context) error {
	ctx, span := r.tracer.Start(ctx, "repository.meilisearch.SearchRepository/DeleteAll")
	defer span.End()

	info, err := r.index().DeleteAllDocumentsWithContext(ctx, nil)
	if err != nil {
		return errors.Join(ErrSearchDelete, err)
	}
	if err := r.waitForTask(ctx, info); err != nil {
		return errors.Join(ErrSearchDelete, err)
	}
	return nil
}

func (r *MeilisearchSearchRepository) Search(ctx context.Context, q SearchQuery) (*SearchHits, error) {
	ctx, span := r.tracer.Start(ctx, "repository.meilisearch.SearchRepository/Search")
	defer span.End()

	filter, err := buildSearchFilter(q)
	if err != nil {
		return nil, errors.Join(ErrSearchQuery, err)
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}

	resp, err := r.index().SearchWithContext(ctx, q.Text, &meilisearch.SearchRequest{
		Offset:               q.Offset,
		Limit:                limit,
		Filter:               filter,
		AttributesToRetrieve: searchDisplayedAttributes,
	})
	if err != nil {
		return nil, errors.Join(ErrSearchQuery, err)
	}

	docs := make([]SearchDocument, 0, len(resp.Hits))
	for _, hit := range resp.Hits {
		var doc SearchDocument
		if err := hit.DecodeInto(&doc); err != nil {
			return nil, errors.Join(ErrSearchQuery, err)
		}
		if doc.ScopeIDs == nil {
			doc.ScopeIDs = []string{}
		}
		docs = append(docs, doc)
	}

	return &SearchHits{
		Documents: docs,
		Offset:    resp.Offset,
		Limit:     resp.Limit,
	}, nil
}

// NewMeilisearchSearchRepository creates a Meilisearch-backed SearchRepository.
func NewMeilisearchSearchRepository(opts ...SearchRepositoryOption) (*MeilisearchSearchRepository, error) {
	r := &MeilisearchSearchRepository{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	if err := validate.Struct(r); err != nil {
		return nil, errors.Join(ErrInvalidRepository, err)
	}

	return r, nil
}

// ListSearchableIDs returns every node ID of resourceType. Used by reindex.
func ListSearchableIDs(ctx context.Context, db *Neo4jDatabase, resourceType model.ResourceType) ([]model.ID, error) {
	if db == nil {
		return nil, ErrNoDriver
	}
	if !resourceType.IsAResourceType() {
		return nil, model.ErrInvalidResourceType
	}

	cypher := `MATCH (n:` + resourceType.String() + `) RETURN n.id AS id`
	ids, err := Neo4jExecuteReadAndReadAll(ctx, db, cypher, nil, func(rec *neo4j.Record) (model.ID, error) {
		raw, err := Neo4jParseValueFromRecord[string](rec, "id")
		if err != nil {
			return model.ID{}, err
		}
		return model.NewIDFromString(raw, resourceType.String())
	})
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []model.ID{}
	}
	return ids, nil
}
