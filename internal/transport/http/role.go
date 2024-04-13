package http

import (
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func roleToDTO(role *model.Role) api.Role {
	dto := api.Role{
		Id:          role.ID.String(),
		Description: &role.Description,
		Name:        role.Name,
		Members:     make([]api.Id, len(role.Members)),
		Permissions: make([]api.Id, len(role.Permissions)),
		CreatedAt:   *role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}

	for i, memberID := range role.Members {
		dto.Permissions[i] = api.Id(memberID.String())
	}

	for i, permissionID := range role.Permissions {
		dto.Permissions[i] = api.Id(permissionID.String())
	}

	return dto
}
