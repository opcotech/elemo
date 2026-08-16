package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
)

var (
	ErrLabelAttach = errors.New("failed to attach label") // the label could not be attached
	ErrLabelCreate = errors.New("failed to create label") // the label could not be created
	ErrLabelDelete = errors.New("failed to delete label") // the label could not be deleted
	ErrLabelDetach = errors.New("failed to detach label") // the label could not be detached
	ErrLabelRead   = errors.New("failed to read label")   // the label could not be retrieved
	ErrLabelUpdate = errors.New("failed to update label") // the label could not be updated
)

// PartialLabel is a lean label used on issue and document reads.
type PartialLabel struct {
	ID   model.ID `json:"id"`
	Name string   `json:"name"`
}

// Label represents a label persisted by the repository.
type Label struct {
	ID          model.ID   `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// CreateLabelOpts holds the data required to create a label.
type CreateLabelOpts struct {
	Name        string
	Description string
}

// UpdateLabelOpts holds the fields that can be updated on a label.
// Undefined fields (Defined == false) are left unchanged.
type UpdateLabelOpts struct {
	Name        optional.Optional[string]
	Description optional.Optional[string]
}

// patch builds a Neo4j property map from defined optional fields.
func (o UpdateLabelOpts) patch() map[string]any {
	p := make(map[string]any)

	if o.Name.Defined {
		p["name"] = *o.Name.Value
	}
	if o.Description.Defined {
		p["description"] = *o.Description.Value
	}

	return p
}

//go:generate go tool mockgen -source=label.go -destination=label_mock_gen.go -package=repository -mock_names "LabelRepository=MockLabelRepository"
type LabelRepository interface {
	Create(ctx context.Context, opts CreateLabelOpts) (*Label, error)
	Get(ctx context.Context, id model.ID, proj LabelProjection) (*Label, error)
	List(ctx context.Context, page CursorPage, proj LabelProjection) (Page[*Label], error)
	Update(ctx context.Context, id model.ID, opts UpdateLabelOpts) (*Label, error)
	AttachTo(ctx context.Context, labelID, attachTo model.ID) error
	DetachFrom(ctx context.Context, labelID, detachFrom model.ID) error
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jLabelRepository is a repository for managing labels.
type Neo4jLabelRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jLabelRepository) scan(lp string) func(rec *neo4j.Record) (*Label, error) {
	return func(rec *neo4j.Record) (*Label, error) {
		l := new(Label)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, lp)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &l, []string{"id"}); err != nil {
			return nil, err
		}

		l.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeLabel.String())

		return l, nil
	}
}

func (r *Neo4jLabelRepository) Create(ctx context.Context, opts CreateLabelOpts) (*Label, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeLabel)

	cypher := `CREATE (l:` + id.Label() + ` {id: $id, name: $name, description: $description, created_at: datetime($created_at)})`
	params := map[string]any{
		"id":          id.String(),
		"name":        opts.Name,
		"description": opts.Description,
		"created_at":  createdAt.Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrLabelCreate, err)
	}

	return r.Get(ctx, id, LabelDetailProjection())
}

func (r *Neo4jLabelRepository) Get(ctx context.Context, id model.ID, proj LabelProjection) (*Label, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/Get")
	defer span.End()

	plan, err := CompileQuery(LabelGetQuery{
		ID:         id,
		Projection: proj,
	})
	if err != nil {
		return nil, errors.Join(ErrLabelRead, err)
	}

	var label *Label
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var runErr error
		label, _, runErr = Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scan("l"))
		return runErr
	})
	if err != nil {
		return nil, errors.Join(ErrLabelRead, err)
	}

	return label, nil
}

func (r *Neo4jLabelRepository) List(ctx context.Context, page CursorPage, proj LabelProjection) (Page[*Label], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/List")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Label]{}, errors.Join(ErrLabelRead, err)
	}
	plan, err := CompileQuery(LabelListQuery{
		Page:       normalized,
		Order:      SortDirectionDesc,
		Projection: proj,
	})
	if err != nil {
		return Page[*Label]{}, errors.Join(ErrLabelRead, err)
	}

	items := make([]*Label, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var runErr error
		items, _, runErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan("l"))
		return runErr
	})
	if err != nil {
		return Page[*Label]{}, errors.Join(ErrLabelRead, err)
	}

	return PaginateSlice(items, normalized.Size, func(label *Label) model.ID {
		return label.ID
	})
}

func (r *Neo4jLabelRepository) Update(ctx context.Context, id model.ID, opts UpdateLabelOpts) (*Label, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/Update")
	defer span.End()

	cypher := `
	MATCH (l:` + id.Label() + ` {id: $id})
	SET l += $patch, l.updated_at = datetime()
	RETURN l.id AS id`

	params := map[string]any{
		"id":    id.String(),
		"patch": opts.patch(),
	}

	if _, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	}); err != nil {
		return nil, errors.Join(ErrLabelUpdate, err)
	}

	return r.Get(ctx, id, LabelDetailProjection())
}

func (r *Neo4jLabelRepository) AttachTo(ctx context.Context, labelID, attachTo model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/AttachTo")
	defer span.End()

	cypher := `
	MATCH (l:` + labelID.Label() + ` {id: $label_id})
	MATCH (n:` + attachTo.Label() + ` {id: $node_id})
	CREATE (n)-[:` + EdgeKindHasLabel.String() + `]->(l)`

	params := map[string]any{
		"label_id": labelID.String(),
		"node_id":  attachTo.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrLabelAttach, err)
	}

	return nil
}

func (r *Neo4jLabelRepository) DetachFrom(ctx context.Context, labelID, detachFrom model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/DetachFrom")
	defer span.End()

	cypher := `
	MATCH (n:` + detachFrom.Label() + ` {id: $node_id})-[r:` + EdgeKindHasLabel.String() + `]->(l:` + labelID.Label() + ` {id: $label_id})
	DELETE r`

	params := map[string]any{
		"label_id": labelID.String(),
		"node_id":  detachFrom.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrLabelDetach, err)
	}

	return nil
}

func (r *Neo4jLabelRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/Delete")
	defer span.End()

	cypher := `MATCH (l:` + id.Label() + ` {id: $id}) DETACH DELETE l`
	params := map[string]any{
		"id": id.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrLabelDelete, err)
	}

	return nil
}

// NewNeo4jLabelRepository creates a new label neo4jBaseRepository.
func NewNeo4jLabelRepository(opts ...Neo4jRepositoryOption) (*Neo4jLabelRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jLabelRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

func clearLabelsPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeLabel.String(), pattern))
}

func clearLabelsKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return clearLabelsPattern(ctx, r, "Get", id.String(), "*")
}

func clearLabelAllLists(ctx context.Context, r *redisBaseRepository) error {
	return clearLabelsPattern(ctx, r, "List", "*", "*", "*")
}

func clearLabelAllCrossCache(ctx context.Context, r *redisBaseRepository) error {
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

// RedisCachedLabelRepository implements caching on the LabelRepository.
type RedisCachedLabelRepository struct {
	cacheRepo *redisBaseRepository
	labelRepo LabelRepository
}

func (r *RedisCachedLabelRepository) Create(ctx context.Context, opts CreateLabelOpts) (*Label, error) {
	if err := clearLabelAllLists(ctx, r.cacheRepo); err != nil {
		return nil, err
	}
	if err := clearLabelAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return r.labelRepo.Create(ctx, opts)
}

func (r *RedisCachedLabelRepository) Get(ctx context.Context, id model.ID, proj LabelProjection) (*Label, error) {
	var label *Label
	var err error

	key := composeCacheKey(model.ResourceTypeLabel.String(), "Get", id.String(), projectionCacheValue(proj))
	if err = r.cacheRepo.Get(ctx, key, &label); err != nil {
		return nil, err
	}

	if label != nil {
		return label, nil
	}

	if label, err = r.labelRepo.Get(ctx, id, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, label); err != nil {
		return nil, err
	}

	return label, nil
}

func (r *RedisCachedLabelRepository) List(ctx context.Context, page CursorPage, proj LabelProjection) (Page[*Label], error) {
	var labels Page[*Label]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*Label]{}, err
	}

	key := composeCacheKey(
		model.ResourceTypeLabel.String(),
		"List",
		projectionCacheValue(proj),
		pageTokenValue(normalized.Token),
		normalized.Size,
	)
	if err = r.cacheRepo.Get(ctx, key, &labels); err != nil {
		return Page[*Label]{}, err
	}

	if labels.Items != nil {
		return labels, nil
	}

	if labels, err = r.labelRepo.List(ctx, normalized, proj); err != nil {
		return Page[*Label]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, labels); err != nil {
		return Page[*Label]{}, err
	}

	return labels, nil
}

func (r *RedisCachedLabelRepository) Update(ctx context.Context, id model.ID, opts UpdateLabelOpts) (*Label, error) {
	label, err := r.labelRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeLabel.String(), "Get", id.String(), projectionCacheValue(LabelDetailProjection()))
	if err := r.cacheRepo.Set(ctx, key, label); err != nil {
		return nil, err
	}

	if err := clearLabelAllLists(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return label, nil
}

func (r *RedisCachedLabelRepository) AttachTo(ctx context.Context, labelID, attachTo model.ID) error {
	if err := clearLabelsKey(ctx, r.cacheRepo, labelID); err != nil {
		return err
	}

	if err := clearLabelAllLists(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearLabelAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.labelRepo.AttachTo(ctx, labelID, attachTo)
}

func (r *RedisCachedLabelRepository) DetachFrom(ctx context.Context, labelID, detachFrom model.ID) error {
	if err := clearLabelsKey(ctx, r.cacheRepo, labelID); err != nil {
		return err
	}

	if err := clearLabelAllLists(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearLabelAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.labelRepo.DetachFrom(ctx, labelID, detachFrom)
}

func (r *RedisCachedLabelRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearLabelsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}
	if err := clearLabelAllLists(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := clearLabelAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.labelRepo.Delete(ctx, id)
}

// NewCachedLabelRepository returns a new CachedLabelRepository.
func NewCachedLabelRepository(repo LabelRepository, opts ...RedisRepositoryOption) (*RedisCachedLabelRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedLabelRepository{
		cacheRepo: r,
		labelRepo: repo,
	}, nil
}
