package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreateRoleOpts creates repository.CreateRoleOpts for tests.
func NewCreateRoleOpts(createdBy, belongsTo model.ID) repository.CreateRoleOpts {
	return repository.CreateRoleOpts{
		Name:        pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(10),
		CreatedBy:   createdBy,
		BelongsTo:   belongsTo,
	}
}

// NewRepositoryRole creates a repository.Role for mock returns.
func NewRepositoryRole() *repository.Role {
	return &repository.Role{
		ID:          model.MustNewID(model.ResourceTypeRole),
		Name:        pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(10),
		Members:     make([]model.ID, 0),
		Permissions: make([]model.ID, 0),
		CreatedAt:   convert.ToPointer(time.Now().UTC()),
	}
}
