package neo4j

import (
	"context"
	"errors"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// RelationRepository is a repository for querying relations.
type RelationRepository struct {
	*baseRepository
}

func (r *RelationRepository) HasAnyRelation(ctx context.Context, source, target model.ID) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RelationRepository/HasAnyRelation")
	defer span.End()

	if err := source.Validate(); err != nil {
		return false, errors.Join(repository.ErrRelationRead, err)
	}

	if err := target.Validate(); err != nil {
		return false, errors.Join(repository.ErrRelationRead, err)
	}

	cypher := `
	MATCH
		(s:` + source.Label() + ` {id: $source_id}),
		(t:` + target.Label() + ` {id: $target_id}),
		path = shortestPath((s)-[*]-(t))
	WITH path
	WHERE length(path) > 1
	RETURN count(path) > 0 AS has_relation`

	params := map[string]any{
		"source_id": source.String(),
		"target_id": target.String(),
	}

	hasRelation, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*bool, error) {
		val, _, err := neo4j.GetRecordValue[bool](rec, "has_relation")
		if err != nil {
			return nil, err
		}
		return &val, nil
	})
	if err != nil {
		return false, errors.Join(repository.ErrRelationRead, err)
	}

	return *hasRelation, nil
}

// NewRelationRepository creates a new relation repository.
func NewRelationRepository(opts ...RepositoryOption) (*RelationRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RelationRepository{
		baseRepository: baseRepo,
	}, nil
}
