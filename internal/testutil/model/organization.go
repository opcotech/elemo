package model

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/testutil"
)

// NewOrganization creates a new organization with random values. It does not
// create the db record.
func NewOrganization() *model.Organization {
	org, err := model.NewOrganization(pkg.GenerateRandomString(10), testutil.GenerateEmail(10))
	if err != nil {
		panic(err)
	}

	org.Logo = imageURL
	org.Website = "https://example.com/"

	return org
}
