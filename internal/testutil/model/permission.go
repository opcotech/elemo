package model

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// OrgAdminActions returns the org-admin role template actions for tests.
func OrgAdminActions() []model.Action {
	tmpl, err := model.RoleTemplateByKey(model.RoleKeyOrgAdmin)
	if err != nil {
		panic(err)
	}
	return tmpl.Actions
}

// ProjectViewerActions returns the project-viewer role template actions for tests.
func ProjectViewerActions() []model.Action {
	tmpl, err := model.RoleTemplateByKey(model.RoleKeyProjectViewer)
	if err != nil {
		panic(err)
	}
	return tmpl.Actions
}

// NewCreateGrantOpts creates repository.CreateGrantOpts for tests.
func NewCreateGrantOpts(principal, scope model.ID, actions ...model.Action) repository.CreateGrantOpts {
	if len(actions) == 0 {
		actions = []model.Action{model.ActionOrganizationRead}
	}
	return repository.CreateGrantOpts{
		Principal: principal,
		Scope:     scope,
		Actions:   actions,
	}
}

// NewRepositoryGrant creates a repository.Grant for mock returns.
func NewRepositoryGrant(principal, scope model.ID, actions ...model.Action) *repository.Grant {
	if len(actions) == 0 {
		actions = []model.Action{model.ActionOrganizationRead}
	}
	return &repository.Grant{
		ID:        model.MustNewID(model.ResourceTypePermission),
		Principal: principal,
		Scope:     scope,
		Actions:   actions,
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}
}
