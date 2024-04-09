package http

import (
	"context"
	"errors"

	oapiTypes "github.com/oapi-codegen/runtime/types"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// OrganizationController is a controller for organization endpoints.
type OrganizationController interface {
	V1OrganizationsGet(ctx context.Context, request api.V1OrganizationsGetRequestObject) (api.V1OrganizationsGetResponseObject, error)
	V1OrganizationsCreate(ctx context.Context, request api.V1OrganizationsCreateRequestObject) (api.V1OrganizationsCreateResponseObject, error)
	V1OrganizationDelete(ctx context.Context, request api.V1OrganizationDeleteRequestObject) (api.V1OrganizationDeleteResponseObject, error)
	V1OrganizationGet(ctx context.Context, request api.V1OrganizationGetRequestObject) (api.V1OrganizationGetResponseObject, error)
	V1OrganizationUpdate(ctx context.Context, request api.V1OrganizationUpdateRequestObject) (api.V1OrganizationUpdateResponseObject, error)
	V1OrganizationMembersGet(ctx context.Context, request api.V1OrganizationMembersGetRequestObject) (api.V1OrganizationMembersGetResponseObject, error)
	V1OrganizationMembersAdd(ctx context.Context, request api.V1OrganizationMembersAddRequestObject) (api.V1OrganizationMembersAddResponseObject, error)
	V1OrganizationMembersRemove(ctx context.Context, request api.V1OrganizationMembersRemoveRequestObject) (api.V1OrganizationMembersRemoveResponseObject, error)
}

// organizationController is the concrete implementation of OrganizationController.
type organizationController struct {
	*baseController
}

func (c *organizationController) V1OrganizationsCreate(ctx context.Context, request api.V1OrganizationsCreateRequestObject) (api.V1OrganizationsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationsCreate")
	defer span.End()

	ownedBy, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return api.V1OrganizationsCreate400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	organization, err := createOrganizationJSONRequestBodyToOrganization(request.Body)
	if err != nil {
		return api.V1OrganizationsCreate400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	if err := c.organizationService.Create(ctx, ownedBy, organization); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1OrganizationsCreate500JSONResponse{
			N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}

	return api.V1OrganizationsCreate201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: organization.ID.String(),
	}}, nil
}

func (c *organizationController) V1OrganizationGet(ctx context.Context, request api.V1OrganizationGetRequestObject) (api.V1OrganizationGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationGet400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	organization, err := c.organizationService.Get(ctx, organizationID)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return api.V1OrganizationGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationGet200JSONResponse(organizationToDTO(organization)), nil
}

func (c *organizationController) V1OrganizationsGet(ctx context.Context, request api.V1OrganizationsGetRequestObject) (api.V1OrganizationsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationsGet")
	defer span.End()

	organizations, err := c.organizationService.GetAll(ctx,
		pkg.GetDefaultPtr(request.Params.Offset, DefaultOffset),
		pkg.GetDefaultPtr(request.Params.Limit, DefaultLimit),
	)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1OrganizationsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	organizationsDTO := make([]api.Organization, len(organizations))
	for i, organization := range organizations {
		organizationsDTO[i] = organizationToDTO(organization)
	}

	return api.V1OrganizationsGet200JSONResponse(organizationsDTO), nil
}

func (c *organizationController) V1OrganizationUpdate(ctx context.Context, request api.V1OrganizationUpdateRequestObject) (api.V1OrganizationUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationUpdate")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationUpdate400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	patch := make(map[string]any)
	if err := convert.AnyToAny(request.Body, &patch); err != nil {
		return api.V1OrganizationUpdate400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	organization, err := c.organizationService.Update(ctx, organizationID, patch)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return api.V1OrganizationUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		}
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1OrganizationUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationUpdate200JSONResponse(organizationToDTO(organization)), nil
}

func (c *organizationController) V1OrganizationDelete(ctx context.Context, request api.V1OrganizationDeleteRequestObject) (api.V1OrganizationDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationDelete")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationDelete400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	if err := c.organizationService.Delete(ctx, organizationID, pkg.GetDefaultPtr(request.Params.Force, false)); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return api.V1OrganizationDelete404JSONResponse{N404JSONResponse: notFound}, nil
		}
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1OrganizationDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationDelete204Response{}, nil
}

func (c *organizationController) V1OrganizationMembersGet(ctx context.Context, request api.V1OrganizationMembersGetRequestObject) (api.V1OrganizationMembersGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationMembersGet")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationMembersGet400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	users, err := c.organizationService.GetMembers(ctx, organizationID)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationMembersGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return api.V1OrganizationMembersGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationMembersGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	// TODO: This could be a method on the organization service.
	usersDTO := make([]api.User, len(users))
	for i, user := range users {
		usersDTO[i] = userToDTO(user)
	}

	return api.V1OrganizationMembersGet200JSONResponse(usersDTO), nil
}

func (c *organizationController) V1OrganizationMembersAdd(ctx context.Context, request api.V1OrganizationMembersAddRequestObject) (api.V1OrganizationMembersAddResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationMembersAdd")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationMembersAdd400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	userID, err := model.NewIDFromString(request.Id, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1OrganizationMembersAdd400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	if err := c.organizationService.AddMember(ctx, organizationID, userID); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationMembersAdd403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return api.V1OrganizationMembersAdd404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationMembersAdd500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationMembersAdd201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: userID.String(),
	}}, nil

}

func (c *organizationController) V1OrganizationMembersRemove(ctx context.Context, request api.V1OrganizationMembersRemoveRequestObject) (api.V1OrganizationMembersRemoveResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1OrganizationMembersRemove")
	defer span.End()

	organizationID, err := model.NewIDFromString(request.Id, model.ResourceTypeOrganization.String())
	if err != nil {
		return api.V1OrganizationMembersRemove400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	userID, err := model.NewIDFromString(request.UserId, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1OrganizationMembersRemove400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	if err := c.organizationService.RemoveMember(ctx, organizationID, userID); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1OrganizationMembersRemove403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return api.V1OrganizationMembersRemove404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1OrganizationMembersRemove500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1OrganizationMembersRemove204Response{}, nil
}

// NewOrganizationController creates a new OrganizationController.
func NewOrganizationController(opts ...ControllerOption) (OrganizationController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	controller := &organizationController{
		baseController: c,
	}

	if controller.organizationService == nil {
		return nil, ErrNoOrganizationService
	}

	return controller, nil
}

func createOrganizationJSONRequestBodyToOrganization(body *api.V1OrganizationsCreateJSONRequestBody) (*model.Organization, error) {
	return model.NewOrganization(body.Name, string(body.Email))
}

func organizationToDTO(organization *model.Organization) api.Organization {
	o := api.Organization{
		Email:      oapiTypes.Email(organization.Email),
		Id:         organization.ID.String(),
		Name:       organization.Name,
		Logo:       &organization.Logo,
		Website:    &organization.Website,
		Status:     api.OrganizationStatus(organization.Status.String()),
		Members:    make([]api.Id, len(organization.Members)),
		Namespaces: make([]api.Id, len(organization.Namespaces)),
		Teams:      make([]api.Id, len(organization.Teams)),
		UpdatedAt:  organization.UpdatedAt,
		CreatedAt:  *organization.CreatedAt,
	}

	for i, member := range organization.Members {
		o.Members[i] = api.Id(member.String())
	}

	for i, namespace := range organization.Namespaces {
		o.Namespaces[i] = api.Id(namespace.String())
	}

	for i, team := range organization.Teams {
		o.Teams[i] = api.Id(team.String())
	}

	return o
}
