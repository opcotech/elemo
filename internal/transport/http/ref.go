package http

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/service"
)

func parseOrganizationRef(value string) (model.ID, string, error) {
	idStr, slug, err := validate.ParseRef(value)
	if err != nil {
		return model.ID{}, "", err
	}
	if idStr != "" {
		id, err := model.NewIDFromString(idStr, model.ResourceTypeOrganization.String())
		if err != nil {
			return model.ID{}, "", err
		}
		return id, "", nil
	}
	return model.ID{}, slug, nil
}

func parseNamespaceRef(value string) (model.ID, string, error) {
	idStr, slug, err := validate.ParseRef(value)
	if err != nil {
		return model.ID{}, "", err
	}
	if idStr != "" {
		id, err := model.NewIDFromString(idStr, model.ResourceTypeNamespace.String())
		if err != nil {
			return model.ID{}, "", err
		}
		return id, "", nil
	}
	return model.ID{}, slug, nil
}

func resolveOrganizationID(ctx context.Context, orgs service.OrganizationService, ref string) (model.ID, error) {
	id, slug, err := parseOrganizationRef(ref)
	if err != nil {
		return model.ID{}, err
	}
	org, err := orgs.Resolve(ctx, id, slug)
	if err != nil {
		return model.ID{}, err
	}
	return org.ID, nil
}

func resolveNamespaceID(ctx context.Context, orgs service.OrganizationService, namespaces service.NamespaceService, orgRef, nsRef string) (model.ID, error) {
	orgID, err := resolveOrganizationID(ctx, orgs, orgRef)
	if err != nil {
		return model.ID{}, err
	}
	id, slug, err := parseNamespaceRef(nsRef)
	if err != nil {
		return model.ID{}, err
	}
	namespace, err := namespaces.Resolve(ctx, orgID, id, slug)
	if err != nil {
		return model.ID{}, err
	}
	return namespace.ID, nil
}
