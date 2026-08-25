package model

import (
	"strings"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
)

// UniqueSlug returns a canonical kebab-case slug that is not xid-shaped.
func UniqueSlug() string {
	return strings.ToLower(pkg.GenerateRandomStringAlpha(8))
}

// NewCreateOrganizationOpts creates repository.CreateOrganizationOpts for tests.
func NewCreateOrganizationOpts(owner model.ID) repository.CreateOrganizationOpts {
	return repository.CreateOrganizationOpts{
		Owner:   owner,
		Slug:    UniqueSlug(),
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
		ID:             model.MustNewID(model.ResourceTypeOrganization),
		Slug:           UniqueSlug(),
		Name:           pkg.GenerateRandomString(10),
		Email:          testutil.GenerateEmail(10),
		Logo:           imageURL,
		Website:        "https://example.com/",
		Status:         model.OrganizationStatusActive,
		NamespaceCount: convert.ToPointer(int64(0)),
		TeamCount:      convert.ToPointer(int64(0)),
		MemberCount:    convert.ToPointer(int64(0)),
		DocumentCount:  convert.ToPointer(int64(0)),
		CreatedAt:      convert.ToPointer(time.Now().UTC()),
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
