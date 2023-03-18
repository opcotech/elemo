package neo4j

import (
	"context"
	"errors"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

type LicenseRepository struct {
	*baseRepository
}

func (r *LicenseRepository) scan(cp string) func(rec *neo4j.Record) (*int, error) {
	return func(rec *neo4j.Record) (*int, error) {
		val, _, err := neo4j.GetRecordValue[int64](rec, cp)
		if err != nil {
			return nil, err
		}

		return convert.ToPointer(int(val)), nil
	}
}

func (r *LicenseRepository) count(ctx context.Context, cypher string, params map[string]any) (int, error) {
	count, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("c"))
	if err != nil {
		return 0, errors.Join(repository.ErrReadResourceCount, err)
	}

	return *count, nil
}

func (r *LicenseRepository) ActiveUserCount(ctx context.Context) (int, error) {
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

func (r *LicenseRepository) ActiveOrganizationCount(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LicenseRepository/ActiveOrganizationCount")
	defer span.End()

	cypher := `MATCH (n:` + model.ResourceTypeOrganization.String() + ` {status: $status}) RETURN count(n) as c`
	params := map[string]any{
		"status": model.OrganizationStatusActive.String(),
	}

	return r.count(ctx, cypher, params)
}

func (r *LicenseRepository) DocumentCount(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LicenseRepository/DocumentCount")
	defer span.End()

	return r.count(ctx, `MATCH (n:`+model.ResourceTypeDocument.String()+`) RETURN count(n) as c`, nil)
}

func (r *LicenseRepository) NamespaceCount(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LicenseRepository/NamespaceCount")
	defer span.End()

	return r.count(ctx, `MATCH (n:`+model.ResourceTypeNamespace.String()+`) RETURN count(n) as c`, nil)
}

func (r *LicenseRepository) ProjectCount(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LicenseRepository/ProjectCount")
	defer span.End()

	return r.count(ctx, `MATCH (n:`+model.ResourceTypeProject.String()+`) RETURN count(n) as c`, nil)
}

func (r *LicenseRepository) RoleCount(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LicenseRepository/RoleCount")
	defer span.End()

	return r.count(ctx, `MATCH (n:`+model.ResourceTypeRole.String()+`) WHERE n.system IS NULL OR n.system = false RETURN count(n) as c`, nil)
}

// NewLicenseRepository creates a new LicenseRepository
func NewLicenseRepository(opts ...RepositoryOption) (*LicenseRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &LicenseRepository{
		baseRepository: baseRepo,
	}, nil
}
