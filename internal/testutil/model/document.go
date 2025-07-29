package model

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
)

// NewDocument creates a new document with random values. It does not create
// the db record.
func NewDocument(createdBy model.ID) *model.Document {
	doc, err := model.NewDocument(
		pkg.GenerateRandomString(10),
		pkg.GenerateRandomString(10),
		createdBy,
	)
	if err != nil {
		panic(err)
	}

	doc.Excerpt = pkg.GenerateRandomString(10)

	return doc
}
