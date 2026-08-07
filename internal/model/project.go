package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	ProjectStatusActive  ProjectStatus = iota + 1 // active
	ProjectStatusPending                          // pending
)

// ProjectStatus represents the status of a project.
//
//go:generate go tool enumer -type=ProjectStatus -text -transform=noop -linecomment -output=project_status_gen.go
type ProjectStatus uint8

// Project represents a project that is used to group tasks together.
type Project struct {
	ID          ID            `json:"id" validate:"required"`
	Key         string        `json:"key" validate:"required,alpha,min=3,max=6"`
	Name        string        `json:"name" validate:"required,min=3,max=120"`
	Description string        `json:"description" validate:"omitempty,min=10,max=500"`
	Logo        string        `json:"logo" validate:"omitempty,url"`
	Status      ProjectStatus `json:"status" validate:"required,min=1,max=2"`
	Teams       []ID          `json:"teams" validate:"omitempty,dive"`
	Documents   []ID          `json:"documents" validate:"omitempty,dive"`
	Issues      []ID          `json:"issues" validate:"omitempty,dive"`
	CreatedAt   *time.Time    `json:"created_at" validate:"omitempty"`
	UpdatedAt   *time.Time    `json:"updated_at" validate:"omitempty"`
}

func (p *Project) Validate() error {
	if err := validate.Struct(p); err != nil {
		return errors.Join(ErrInvalidProjectDetails, err)
	}
	if err := p.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidProjectDetails, err)
	}
	for _, team := range p.Teams {
		if err := team.Validate(); err != nil {
			return errors.Join(ErrInvalidProjectDetails, err)
		}
	}
	for _, document := range p.Documents {
		if err := document.Validate(); err != nil {
			return errors.Join(ErrInvalidProjectDetails, err)
		}
	}
	for _, issue := range p.Issues {
		if err := issue.Validate(); err != nil {
			return errors.Join(ErrInvalidProjectDetails, err)
		}
	}
	return nil
}

// NewProject creates a new project.
func NewProject(key, name string) (*Project, error) {
	project := &Project{
		ID:        MustNewNilID(ResourceTypeProject),
		Key:       key,
		Name:      name,
		Status:    ProjectStatusActive,
		Teams:     make([]ID, 0),
		Documents: make([]ID, 0),
		Issues:    make([]ID, 0),
	}

	if err := project.Validate(); err != nil {
		return nil, err
	}

	return project, nil
}
