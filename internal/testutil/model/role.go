package model

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/testutil"
)

// NewRole creates a new Role. It does not create the role in the database.
func NewRole() *model.Role {
	role, err := model.NewRole(testutil.GenerateRandomString(10))
	if err != nil {
		panic(err)
	}

	role.Description = testutil.GenerateRandomString(10)

	return role
}
