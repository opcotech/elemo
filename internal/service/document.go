package service

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
)

// PartialDocument represents a simplified document within a namespace.
type PartialDocument struct {
	ID        model.ID
	Name      string
	Excerpt   string
	CreatedBy model.ID
	CreatedAt *time.Time
}
