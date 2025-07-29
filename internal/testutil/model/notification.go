package model

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
)

// NewNotification creates a new Notification instance. It does not create a
// notification in the database.
func NewNotification(recipient model.ID) *model.Notification {
	notification, err := model.NewNotification(pkg.GenerateRandomString(10), recipient)
	if err != nil {
		panic(err)
	}

	notification.Description = pkg.GenerateRandomString(10)

	return notification
}
