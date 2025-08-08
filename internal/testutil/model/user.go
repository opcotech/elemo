package model

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/testutil"
)

// NewUser creates a new user with random values. It does not create the db
// record.
func NewUser() *model.User {
	user, err := model.NewUser(
		strings.ToLower(pkg.GenerateRandomString(10)),
		testutil.GenerateEmail(10),
		password.HashPassword(pkg.GenerateRandomString(10)),
	)
	if err != nil {
		panic(err)
	}

	user.FirstName = "Test"
	user.LastName = "User"
	user.Picture = imageURL
	user.Title = "Senior Test User"
	user.Bio = "I am a test user."
	user.Phone = "+1234567890"
	user.Address = "1234 Main St, Anytown, USA"
	user.Links = []string{"https://example.com/"}
	user.Languages = []model.Language{
		model.LanguageHU,
		model.LanguageEN,
		model.LanguageES,
	}

	return user
}