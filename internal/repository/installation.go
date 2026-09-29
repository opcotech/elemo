package repository

import (
	"context"
	"errors"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrInstallationRead   = errors.New("failed to read installation")
	ErrInstallationUpdate = errors.New("failed to update installation")
)

//go:generate go tool mockgen -source=installation.go -destination=mock/mock_installation_gen.go -package=mockrepo
type InstallationRepository interface {
	// EnsureLogicalID returns the stable logical installation UUID, generating
	// it once if the singleton Installation node does not yet have one.
	EnsureLogicalID(ctx context.Context) (string, error)
}

type Neo4jInstallationRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jInstallationRepository) EnsureLogicalID(ctx context.Context) (string, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.InstallationRepository/EnsureLogicalID")
	defer span.End()

	cypher := `
	MATCH (i:` + model.ResourceTypeInstallation.String() + ` {id: $id})
	SET i.logical_id = coalesce(i.logical_id, randomUUID())
	RETURN i.logical_id AS logical_id
	`
	params := map[string]any{
		"id": model.InstallationID().String(),
	}

	id, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*string, error) {
		val, _, err := neo4j.GetRecordValue[string](rec, "logical_id")
		if err != nil {
			return nil, err
		}
		return &val, nil
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", errors.Join(ErrInstallationRead, ErrNotFound)
		}
		return "", errors.Join(ErrInstallationUpdate, err)
	}

	return *id, nil
}

func NewNeo4jInstallationRepository(opts ...Neo4jRepositoryOption) (*Neo4jInstallationRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jInstallationRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}
