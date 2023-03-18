package model

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/testutil"
)

// NewDocument creates a new document with random values. It does not create
// the db record.
func NewDocument(createdBy model.ID) *model.Document {
	doc, err := model.NewDocument(
		testutil.GenerateRandomString(10),
		testutil.GenerateRandomString(10),
		createdBy,
	)
	if err != nil {
		panic(err)
	}

	doc.Excerpt = testutil.GenerateRandomString(10)

	return doc
}
