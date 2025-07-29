package model

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
)

// NewProject creates a new Project instance. It does not create a new project
// in the database.
func NewProject() *model.Project {
	project, err := model.NewProject(pkg.GenerateRandomStringAlpha(3), pkg.GenerateRandomString(10))
	if err != nil {
		panic(err)
	}

	project.Description = pkg.GenerateRandomString(10)
	project.Logo = imageURL

	return project
}
