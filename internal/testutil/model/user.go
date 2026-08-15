package model

import (
	"strings"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
)

// NewCreateUserOpts creates repository.CreateUserOpts for tests.
func NewCreateUserOpts() repository.CreateUserOpts {
	return repository.CreateUserOpts{
		Username:  strings.ToLower(pkg.GenerateRandomString(10)),
		FirstName: "Test",
		LastName:  "User",
		Email:     testutil.GenerateEmail(10),
		Password:  password.HashPassword(pkg.GenerateRandomString(10)),
		Status:    model.UserStatusActive,
		Picture:   imageURL,
		Title:     "Senior Test User",
		Bio:       "I am a test user.",
		Phone:     "+1234567890",
		Address:   "1234 Main St, Anytown, USA",
		Links:     []string{"https://example.com/"},
		Languages: []model.Language{
			model.LanguageHU,
			model.LanguageEN,
			model.LanguageES,
		},
	}
}

// NewRepositoryUser creates a repository.User for mock returns.
func NewRepositoryUser() *repository.User {
	return &repository.User{
		ID:          model.MustNewID(model.ResourceTypeUser),
		Username:    strings.ToLower(pkg.GenerateRandomString(10)),
		FirstName:   "Test",
		LastName:    "User",
		Email:       testutil.GenerateEmail(10),
		Password:    password.HashPassword(pkg.GenerateRandomString(10)),
		Status:      model.UserStatusActive,
		Picture:     imageURL,
		Title:       "Senior Test User",
		Bio:         "I am a test user.",
		Phone:       "+1234567890",
		Address:     "1234 Main St, Anytown, USA",
		Links:       []string{"https://example.com/"},
		Languages:   []model.Language{model.LanguageHU, model.LanguageEN, model.LanguageES},
		Permissions: make([]model.ID, 0),
		CreatedAt:   convert.ToPointer(time.Now().UTC()),
	}
}

// NewUser creates a repository.User with random values. It does not create the db
// record. Prefer NewCreateUserOpts / NewRepositoryUser for new tests.
func NewUser() *repository.User {
	return NewRepositoryUser()
}
