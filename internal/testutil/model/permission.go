package model

import "github.com/opcotech/elemo/internal/model"

// NewPermission creates a new permission with the given subject and target. It
// does not create the permission in the database.
func NewPermission(subject, target model.ID, kind model.PermissionKind) *model.Permission {
	permission, err := model.NewPermission(subject, target, kind)
	if err != nil {
		panic(err)
	}

	return permission
}
