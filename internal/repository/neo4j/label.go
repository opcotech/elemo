package neo4j

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// LabelRepository is a repository for managing labels.
type LabelRepository struct {
	*baseRepository
}

func (r *LabelRepository) scan(lp string) func(rec *neo4j.Record) (*model.Label, error) {
	return func(rec *neo4j.Record) (*model.Label, error) {
		l := new(model.Label)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, lp)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &l, []string{"id"}); err != nil {
			return nil, err
		}

		l.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeLabel.String())

		if err := l.Validate(); err != nil {
			return nil, err
		}

		return l, nil
	}
}

func (r *LabelRepository) Create(ctx context.Context, label *model.Label) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/Create")
	defer span.End()

	if err := label.Validate(); err != nil {
		return errors.Join(repository.ErrLabelCreate, err)
	}

	createdAt := time.Now()

	label.ID = model.MustNewID(model.ResourceTypeLabel)
	label.CreatedAt = convert.ToPointer(createdAt)
	label.UpdatedAt = nil

	cypher := `CREATE (l:` + label.ID.Label() + ` {id: $id, name: $name, description: $description, created_at: datetime($created_at)})`
	params := map[string]any{
		"id":          label.ID.String(),
		"name":        label.Name,
		"description": label.Description,
		"created_at":  createdAt.Format(time.RFC3339Nano),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrLabelCreate, err)
	}

	return nil
}

func (r *LabelRepository) Get(ctx context.Context, id model.ID) (*model.Label, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/Get")
	defer span.End()

	cypher := `MATCH (l:` + id.Label() + ` {id: $id}) RETURN l`
	params := map[string]any{
		"id": id.String(),
	}

	label, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("l"))
	if err != nil {
		return nil, errors.Join(repository.ErrLabelRead, err)
	}

	return label, nil
}

func (r *LabelRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Label, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/Update")
	defer span.End()

	cypher := `
	MATCH (l:` + id.Label() + ` {id: $id})
	SET l += $patch, l.updated_at = datetime($updated_at)
	RETURN l`

	params := map[string]any{
		"id":         id.String(),
		"patch":      patch,
		"updated_at": time.Now().Format(time.RFC3339Nano),
	}

	label, err := ExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("l"))
	if err != nil {
		return nil, errors.Join(repository.ErrLabelUpdate, err)
	}

	return label, nil
}

func (r *LabelRepository) AttachTo(ctx context.Context, labelID, attachTo model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/AttachTo")
	defer span.End()

	if err := attachTo.Validate(); err != nil {
		return errors.Join(repository.ErrLabelAttach, err)
	}

	if err := labelID.Validate(); err != nil {
		return errors.Join(repository.ErrLabelAttach, err)
	}

	cypher := `
	MATCH (l:` + labelID.Label() + ` {id: $label_id}), (n:` + attachTo.Label() + ` {id: $node_id})
	CREATE (n)-[:` + EdgeKindHasLabel.String() + `]->(l)`

	params := map[string]any{
		"label_id": labelID.String(),
		"node_id":  attachTo.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrLabelAttach, err)
	}

	return nil
}

func (r *LabelRepository) DetachFrom(ctx context.Context, labelID, detachFrom model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/DetachFrom")
	defer span.End()

	if err := detachFrom.Validate(); err != nil {
		return errors.Join(repository.ErrLabelDetach, err)
	}

	if err := labelID.Validate(); err != nil {
		return errors.Join(repository.ErrLabelDetach, err)
	}

	cypher := `
	MATCH (l:` + labelID.Label() + ` {id: $label_id})-[r:` + EdgeKindHasLabel.String() + `]->(n:` + detachFrom.Label() + ` {id: $node_id})
	DELETE r`

	params := map[string]any{
		"label_id": labelID.String(),
		"node_id":  detachFrom.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrLabelDetach, err)
	}

	return nil
}

func (r *LabelRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/Delete")
	defer span.End()

	cypher := `MATCH (l:` + id.Label() + ` {id: $id}) DETACH DELETE l`
	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrLabelDelete, err)
	}

	return nil
}

// NewLabelRepository creates a new label baseRepository.
func NewLabelRepository(opts ...RepositoryOption) (*LabelRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &LabelRepository{
		baseRepository: baseRepo,
	}, nil
}
