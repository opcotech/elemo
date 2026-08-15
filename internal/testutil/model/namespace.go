package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreateNamespaceOpts creates repository.CreateNamespaceOpts for tests.
func NewCreateNamespaceOpts(creatorID, orgID model.ID) repository.CreateNamespaceOpts {
	return repository.CreateNamespaceOpts{
		Name:        pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(10),
		CreatorID:   creatorID,
		OrgID:       orgID,
	}
}

// NewRepositoryNamespace creates a repository.Namespace for mock returns.
func NewRepositoryNamespace() *repository.Namespace {
	return &repository.Namespace{
		ID:            model.MustNewID(model.ResourceTypeNamespace),
		Name:          pkg.GenerateRandomString(10),
		Description:   pkg.GenerateRandomString(10),
		ProjectCount:  convert.ToPointer(int64(0)),
		DocumentCount: convert.ToPointer(int64(0)),
		CreatedAt:     convert.ToPointer(time.Now().UTC()),
	}
}
