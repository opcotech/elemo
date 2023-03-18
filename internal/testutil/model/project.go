package model

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/testutil"
)

// NewProject creates a new Project instance. It does not create a new project
// in the database.
func NewProject() *model.Project {
	project, err := model.NewProject(testutil.GenerateRandomStringAlpha(3), testutil.GenerateRandomString(10))
	if err != nil {
		panic(err)
	}

	project.Description = testutil.GenerateRandomString(10)
	project.Logo = imageURL

	return project
}
