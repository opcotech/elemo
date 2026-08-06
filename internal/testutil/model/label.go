package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreateLabelOpts creates repository.CreateLabelOpts for tests.
func NewCreateLabelOpts() repository.CreateLabelOpts {
	return repository.CreateLabelOpts{
		Name:        pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(10),
	}
}

// NewRepositoryLabel creates a repository.Label for mock returns.
func NewRepositoryLabel() *repository.Label {
	opts := NewCreateLabelOpts()
	return &repository.Label{
		ID:          model.MustNewID(model.ResourceTypeLabel),
		Name:        opts.Name,
		Description: opts.Description,
		CreatedAt:   convert.ToPointer(time.Now().UTC()),
	}
}

// NewLabel creates a new label with random values. It does not create the db
// record.
//
// Deprecated: prefer NewCreateLabelOpts / NewRepositoryLabel.
func NewLabel() *model.Label {
	label, err := model.NewLabel(pkg.GenerateRandomString(10))
	if err != nil {
		panic(err)
	}

	label.Description = pkg.GenerateRandomString(10)

	return label
}
