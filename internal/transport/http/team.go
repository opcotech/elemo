package http

import (
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func createTeamJSONRequestBodyToCreateTeamOpts(body *api.V1OrganizationTeamsCreateJSONRequestBody) service.CreateTeamOpts {
	opts := service.CreateTeamOpts{
		Name: body.Name,
	}

	if body.Description != nil {
		opts.Description = *body.Description
	}

	return opts
}

func updateTeamJSONRequestBodyToUpdateTeamOpts(body *api.V1OrganizationTeamUpdateJSONRequestBody) service.UpdateTeamOpts {
	opts := service.UpdateTeamOpts{}

	if body.Name != nil {
		opts.Name = optional.Some(*body.Name)
	}
	if body.Description.Defined {
		opts.Description = body.Description
	}

	return opts
}

func teamToDTO(team *service.Team) api.Team {
	return api.Team{
		Id:          team.ID.String(),
		Name:        team.Name,
		Description: &team.Description,
		MemberCount: team.MemberCount,
		CreatedAt:   *team.CreatedAt,
		UpdatedAt:   team.UpdatedAt,
	}
}
