package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// NewCreateTeamOpts creates repository.CreateTeamOpts for tests.
func NewCreateTeamOpts(createdBy, belongsTo model.ID) repository.CreateTeamOpts {
	return repository.CreateTeamOpts{
		Name:        pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(10),
		CreatedBy:   createdBy,
		BelongsTo:   belongsTo,
	}
}

// NewRepositoryTeam creates a repository.Team for mock returns.
func NewRepositoryTeam() *repository.Team {
	return &repository.Team{
		ID:          model.MustNewID(model.ResourceTypeTeam),
		Name:        pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(10),
		MemberCount: convert.ToPointer(int64(0)),
		CreatedAt:   convert.ToPointer(time.Now().UTC()),
	}
}
