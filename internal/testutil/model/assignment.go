package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreateAssignmentOpts creates repository.CreateAssignmentOpts for tests.
func NewCreateAssignmentOpts(user, resource model.ID, kind model.AssignmentKind) repository.CreateAssignmentOpts {
	return repository.CreateAssignmentOpts{
		Kind:     kind,
		User:     user,
		Resource: resource,
	}
}

// NewRepositoryAssignment creates a repository.Assignment for mock returns.
func NewRepositoryAssignment(user, resource model.ID, kind model.AssignmentKind) *repository.Assignment {
	return &repository.Assignment{
		ID:        model.MustNewID(model.ResourceTypeAssignment),
		Kind:      kind,
		User:      user,
		Resource:  resource,
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}
}

// NewAssignment creates a new assignment between a user and a resource. It
// does not create the db record.
//
// Deprecated: prefer NewCreateAssignmentOpts / NewRepositoryAssignment.
func NewAssignment(createdBy model.ID, documentID model.ID, kind model.AssignmentKind) *model.Assignment {
	assignment, err := model.NewAssignment(createdBy, documentID, kind)
	if err != nil {
		panic(err)
	}
	return assignment
}
