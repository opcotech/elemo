package model

import "github.com/opcotech/elemo/internal/model"

// NewAssignment creates a new assignment between a user and a resource. It
// does not create the db record.
func NewAssignment(createdBy model.ID, documentID model.ID, kind model.AssignmentKind) *model.Assignment {
	assignment, err := model.NewAssignment(createdBy, documentID, kind)
	if err != nil {
		panic(err)
	}
	return assignment
}
