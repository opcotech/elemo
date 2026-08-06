package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
)

// NewCreateOrganizationOpts creates repository.CreateOrganizationOpts for tests.
func NewCreateOrganizationOpts(owner model.ID) repository.CreateOrganizationOpts {
	return repository.CreateOrganizationOpts{
		Owner:   owner,
		Name:    pkg.GenerateRandomString(10),
		Email:   testutil.GenerateEmail(10),
		Logo:    imageURL,
		Website: "https://example.com/",
		Status:  model.OrganizationStatusActive,
	}
}

// NewRepositoryOrganization creates a repository.Organization for mock returns.
func NewRepositoryOrganization() *repository.Organization {
	return &repository.Organization{
		ID:         model.MustNewID(model.ResourceTypeOrganization),
		Name:       pkg.GenerateRandomString(10),
		Email:      testutil.GenerateEmail(10),
		Logo:       imageURL,
		Website:    "https://example.com/",
		Status:     model.OrganizationStatusActive,
		Namespaces: make([]model.ID, 0),
		Teams:      make([]model.ID, 0),
		Members:    make([]model.ID, 0),
		CreatedAt:  convert.ToPointer(time.Now().UTC()),
	}
}

// NewRepositoryOrganizationMember creates a repository.OrganizationMember for mock returns.
func NewRepositoryOrganizationMember() *repository.OrganizationMember {
	return &repository.OrganizationMember{
		ID:        model.MustNewID(model.ResourceTypeUser),
		FirstName: pkg.GenerateRandomString(8),
		LastName:  pkg.GenerateRandomString(8),
		Email:     testutil.GenerateEmail(10),
		Status:    model.UserStatusActive,
		Roles:     []string{},
	}
}
