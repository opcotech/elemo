package http

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func createRoleJSONRequestBodyToCreateRoleOpts(body *api.V1OrganizationRolesCreateJSONRequestBody) (service.CreateRoleOpts, error) {
	opts := service.CreateRoleOpts{
		Name: body.Name,
	}

	if body.Key != nil {
		opts.Key = *body.Key
	}
	if body.Description != nil {
		opts.Description = *body.Description
	}
	if body.Actions != nil {
		actions, err := model.ParseActions(*body.Actions)
		if err != nil {
			return service.CreateRoleOpts{}, err
		}
		opts.Actions = actions
	}

	return opts, nil
}

func updateRoleJSONRequestBodyToUpdateRoleOpts(body *api.V1OrganizationRoleUpdateJSONRequestBody) (service.UpdateRoleOpts, error) {
	opts := service.UpdateRoleOpts{}

	if body.Name != nil {
		opts.Name = optional.Some(*body.Name)
	}
	if body.Description.Defined {
		opts.Description = body.Description
	}
	if body.Actions.Defined {
		if body.Actions.Value == nil {
			opts.Actions = optional.Some([]model.Action{})
		} else {
			actions, err := model.ParseActions(*body.Actions.Value)
			if err != nil {
				return service.UpdateRoleOpts{}, err
			}
			opts.Actions = optional.Some(actions)
		}
	}

	return opts, nil
}

func roleToDTO(role *service.Role) api.Role {
	return api.Role{
		Id:          role.ID.String(),
		Key:         role.Key,
		Description: &role.Description,
		Name:        role.Name,
		MemberCount: role.MemberCount,
		Actions:     actionStringsOrEmpty(role.Actions),
		CreatedAt:   *role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}
