package model

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/testutil"
)

// NewAttachment creates a new attachment with random values. It does not
// create the db record.
func NewAttachment(createdBy model.ID) *model.Attachment {
	attachment, err := model.NewAttachment(
		testutil.GenerateRandomString(10),
		testutil.GenerateRandomString(10),
		createdBy,
	)
	if err != nil {
		panic(err)
	}
	return attachment
}
