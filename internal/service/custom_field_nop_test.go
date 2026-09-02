package service_test

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/service"
)

// nopCustomFieldService is a test double for issue paths that do not assert
// custom-field behavior.
type nopCustomFieldService struct{}

func (nopCustomFieldService) CreateDefinition(context.Context, service.CreateCustomFieldOpts) (*model.CustomFieldDefinition, error) {
	return nil, nil
}

func (nopCustomFieldService) GetDefinition(context.Context, model.ID) (*model.CustomFieldDefinition, error) {
	return nil, nil
}

func (nopCustomFieldService) ListDefinitions(context.Context, model.ID, model.ResourceType, bool) ([]*model.CustomFieldDefinition, error) {
	return nil, nil
}

func (nopCustomFieldService) UpdateDefinition(context.Context, model.ID, service.UpdateCustomFieldOpts) (*model.CustomFieldDefinition, error) {
	return nil, nil
}

func (nopCustomFieldService) ArchiveDefinition(context.Context, model.ID) (*model.CustomFieldDefinition, error) {
	return nil, nil
}

func (nopCustomFieldService) DeleteDefinition(context.Context, model.ID) error {
	return nil
}

func (nopCustomFieldService) ListEffective(context.Context, model.ID) ([]service.CustomFieldEntry, error) {
	return nil, nil
}

func (nopCustomFieldService) SetValue(context.Context, model.ID, model.ID, model.CustomFieldTypedValue) error {
	return nil
}

func (nopCustomFieldService) DeleteValue(context.Context, model.ID, model.ID) error {
	return nil
}

func (nopCustomFieldService) Search(context.Context, service.CustomFieldSearchQuery) ([]model.ID, error) {
	return nil, nil
}

func (nopCustomFieldService) StageForResource(context.Context, model.ID, model.ID, []service.CustomFieldWrite) error {
	return nil
}

func (nopCustomFieldService) CommitForResource(context.Context, model.ID) error {
	return nil
}

func (nopCustomFieldService) AbortForResource(context.Context, model.ID) error {
	return nil
}

func (nopCustomFieldService) DeleteForResource(context.Context, model.ID) error {
	return nil
}

func (nopCustomFieldService) ReconcilePending(context.Context) error {
	return nil
}
