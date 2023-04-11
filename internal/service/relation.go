package service

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// RelationRepository is a repository for querying relations.
type RelationRepository interface {
	HasAnyRelation(ctx context.Context, source, target model.ID) (bool, error)
}
