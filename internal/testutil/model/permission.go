package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreatePermissionOpts creates repository.CreatePermissionOpts for tests.
func NewCreatePermissionOpts(subject, target model.ID, kind model.PermissionKind) repository.CreatePermissionOpts {
	return repository.CreatePermissionOpts{
		Kind:    kind,
		Subject: subject,
		Target:  target,
	}
}

// NewRepositoryPermission creates a repository.Permission for mock returns.
func NewRepositoryPermission(subject, target model.ID, kind model.PermissionKind) *repository.Permission {
	return &repository.Permission{
		ID:        model.MustNewID(model.ResourceTypePermission),
		Kind:      kind,
		Subject:   subject,
		Target:    target,
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}
}
