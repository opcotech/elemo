package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreateProjectOpts creates repository.CreateProjectOpts for tests.
func NewCreateProjectOpts(namespaceID model.ID) repository.CreateProjectOpts {
	return repository.CreateProjectOpts{
		NamespaceID: namespaceID,
		Key:         pkg.GenerateRandomStringAlpha(3),
		Name:        pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(10),
		Logo:        imageURL,
		Status:      model.ProjectStatusActive,
	}
}

// NewRepositoryProject creates a repository.Project for mock returns.
func NewRepositoryProject() *repository.Project {
	return &repository.Project{
		ID:          model.MustNewID(model.ResourceTypeProject),
		Key:         pkg.GenerateRandomStringAlpha(3),
		Name:        pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(10),
		Logo:        imageURL,
		Status:      model.ProjectStatusActive,
		Teams:       make([]model.ID, 0),
		Documents:   make([]model.ID, 0),
		Issues:      make([]model.ID, 0),
		CreatedAt:   convert.ToPointer(time.Now().UTC()),
	}
}

// NewProject creates a new Project instance. It does not create a new project
// in the database.
//
// Deprecated: prefer NewCreateProjectOpts / NewRepositoryProject.
func NewProject() *model.Project {
	project, err := model.NewProject(pkg.GenerateRandomStringAlpha(3), pkg.GenerateRandomString(10))
	if err != nil {
		panic(err)
	}

	project.Description = pkg.GenerateRandomString(10)
	project.Logo = imageURL

	return project
}
