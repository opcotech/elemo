package model

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
)

// NewAttachment creates a new attachment with random values. It does not
// create the db record.
func NewAttachment(createdBy model.ID) *model.Attachment {
	attachment, err := model.NewAttachment(
		pkg.GenerateRandomString(10),
		pkg.GenerateRandomString(10),
		createdBy,
	)
	if err != nil {
		panic(err)
	}
	return attachment
}
