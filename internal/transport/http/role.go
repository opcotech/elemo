package http

import (
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func createRoleJSONRequestBodyToCreateRoleOpts(body *api.V1OrganizationRolesCreateJSONRequestBody) service.CreateRoleOpts {
	opts := service.CreateRoleOpts{
		Name: body.Name,
	}

	if body.Description != nil {
		opts.Description = *body.Description
	}

	return opts
}

func updateRoleJSONRequestBodyToUpdateRoleOpts(body *api.V1OrganizationRoleUpdateJSONRequestBody) service.UpdateRoleOpts {
	opts := service.UpdateRoleOpts{}

	if body.Name != nil {
		opts.Name = optional.Some(*body.Name)
	}
	if body.Description.Defined {
		opts.Description = body.Description
	}

	return opts
}

func roleToDTO(role *service.Role) api.Role {
	dto := api.Role{
		Id:          role.ID.String(),
		Description: &role.Description,
		Name:        role.Name,
		MemberCount: role.MemberCount,
		Permissions: make([]api.Id, len(role.Permissions)),
		CreatedAt:   *role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}

	for i, permissionID := range role.Permissions {
		dto.Permissions[i] = api.Id(permissionID.String())
	}

	return dto
}
