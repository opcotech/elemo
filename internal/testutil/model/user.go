package model

import (
	"strings"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/testutil"
)

// NewUser creates a new user with random values. It does not create the db
// record.
func NewUser() *model.User {
	return &model.User{
		ID:          model.MustNewID(model.ResourceTypeUser),
		Username:    strings.ToLower(testutil.GenerateRandomString(10)),
		Email:       testutil.GenerateEmail(10),
		Password:    password.HashPassword(testutil.GenerateRandomString(10)),
		Status:      model.UserStatusActive,
		FirstName:   testutil.GenerateRandomString(5),
		LastName:    testutil.GenerateRandomString(5),
		Picture:     "https://www.gravatar.com/avatar",
		Title:       "Senior Test User",
		Bio:         "I am a test user.",
		Phone:       "+1234567890",
		Address:     "1234 Main St, Anytown, USA",
		Links:       []string{"https://example.com/"},
		Languages:   []model.Language{model.LanguageHU, model.LanguageEN, model.LanguageAR},
		Documents:   make([]model.ID, 0),
		Permissions: make([]model.ID, 0),
		CreatedAt:   convert.ToPointer(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
		UpdatedAt:   nil,
	}
}
