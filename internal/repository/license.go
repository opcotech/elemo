package repository

import (
	"context"
	"errors"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

// LicenseRepository is the repository for retrieving license information.
//
//go:generate mockgen -source=license.go -destination=../testutil/mock/license_repo_gen.go -package=mock -mock_names "LicenseRepository=LicenseRepository"
type LicenseRepository interface {
	ActiveUserCount(ctx context.Context) (int, error)
	ActiveOrganizationCount(ctx context.Context) (int, error)
	DocumentCount(ctx context.Context) (int, error)
	NamespaceCount(ctx context.Context) (int, error)
	ProjectCount(ctx context.Context) (int, error)
	RoleCount(ctx context.Context) (int, error)
}

type Neo4jLicenseRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jLicenseRepository) scan(cp string) func(rec *neo4j.Record) (*int, error) {
	return func(rec *neo4j.Record) (*int, error) {
		val, _, err := neo4j.GetRecordValue[int64](rec, cp)
		if err != nil {
			return nil, err
		}

		return convert.ToPointer(int(val)), nil
	}
}

func (r *Neo4jLicenseRepository) count(ctx context.Context, cypher string, params map[string]any) (int, error) {
	count, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("c"))
	if err != nil {
		return 0, errors.Join(ErrReadResourceCount, err)
	}

	return *count, nil
}

func (r *Neo4jLicenseRepository) ActiveUserCount(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LicenseRepository/ActiveUserCount")
	defer span.End()

	cypher := `MATCH (n:` + model.ResourceTypeUser.String() + `) WHERE n.status IN $status RETURN count(n) as c`
	params := map[string]any{
		"status": []string{
			model.UserStatusActive.String(),
			model.UserStatusPending.String(),
		},
	}

	return r.count(ctx, cypher, params)
}

func (r *Neo4jLicenseRepository) ActiveOrganizationCount(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LicenseRepository/ActiveOrganizationCount")
	defer span.End()

	cypher := `MATCH (n:` + model.ResourceTypeOrganization.String() + ` {status: $status}) RETURN count(n) as c`
	params := map[string]any{
		"status": model.OrganizationStatusActive.String(),
	}

	return r.count(ctx, cypher, params)
}

func (r *Neo4jLicenseRepository) DocumentCount(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LicenseRepository/DocumentCount")
	defer span.End()

	return r.count(ctx, `MATCH (n:`+model.ResourceTypeDocument.String()+`) RETURN count(n) as c`, nil)
}

func (r *Neo4jLicenseRepository) NamespaceCount(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LicenseRepository/NamespaceCount")
	defer span.End()

	return r.count(ctx, `MATCH (n:`+model.ResourceTypeNamespace.String()+`) RETURN count(n) as c`, nil)
}

func (r *Neo4jLicenseRepository) ProjectCount(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LicenseRepository/ProjectCount")
	defer span.End()

	return r.count(ctx, `MATCH (n:`+model.ResourceTypeProject.String()+`) RETURN count(n) as c`, nil)
}

func (r *Neo4jLicenseRepository) RoleCount(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LicenseRepository/RoleCount")
	defer span.End()

	return r.count(ctx, `MATCH (n:`+model.ResourceTypeRole.String()+`) WHERE n.system IS NULL OR n.system = false RETURN count(n) as c`, nil)
}

// NewNeo4jLicenseRepository creates a new LicenseRepository
func NewNeo4jLicenseRepository(opts ...Neo4jRepositoryOption) (*Neo4jLicenseRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jLicenseRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}
