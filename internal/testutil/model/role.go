package model

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
)

// NewRole creates a new Role. It does not create the role in the database.
func NewRole() *model.Role {
	role, err := model.NewRole(pkg.GenerateRandomString(10))
	if err != nil {
		panic(err)
	}

	role.Description = pkg.GenerateRandomString(10)

	return role
}
