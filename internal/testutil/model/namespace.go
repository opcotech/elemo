package model

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/testutil"
)

// NewNamespace creates a new Namespace instance. It does not create a
// namespace in the database.
func NewNamespace() *model.Namespace {
	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	if err != nil {
		panic(err)
	}

	namespace.Description = testutil.GenerateRandomString(10)

	return namespace
}
