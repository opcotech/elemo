package model

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
)

// NewLabel creates a new label with random values. It does not create the db
// record.
func NewLabel() *model.Label {
	label, err := model.NewLabel(pkg.GenerateRandomString(10))
	if err != nil {
		panic(err)
	}

	label.Description = pkg.GenerateRandomString(10)

	return label
}
