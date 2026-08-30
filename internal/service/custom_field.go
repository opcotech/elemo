package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

const customFieldReconcileStale = 30 * time.Second

// CustomFieldWrite is a typed value addressed to one definition.
type CustomFieldWrite struct {
	DefinitionID model.ID
	Value        model.CustomFieldTypedValue
}

// CustomFieldEntry is a definition plus its stored value, if any.
type CustomFieldEntry struct {
	Definition *model.CustomFieldDefinition
	Value      *model.CustomFieldTypedValue
}

// CreateCustomFieldOpts holds data required to create a definition.
type CreateCustomFieldOpts struct {
	Key           string                `validate:"required,min=2,max=63"`
	Name          string                `validate:"required,min=3,max=120"`
	Description   string                `validate:"omitempty,max=500"`
	Kind          model.CustomFieldKind `validate:"required,min=1,max=11"`
	Scope         model.ID              `validate:"required"`
	TargetType    model.ResourceType    `validate:"required"`
	Required      bool
	IndexExact    bool
	IndexRange    bool
	IndexFullText bool
	SortOrder     optional.Optional[int]
	Schema        model.CustomFieldSchema
}

func (o *CreateCustomFieldOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidCustomFieldDetails, err)
	}
	if err := o.Scope.Validate(); err != nil {
		return errors.Join(model.ErrInvalidCustomFieldDetails, err)
	}
	return nil
}

// UpdateCustomFieldOpts holds constrained definition updates.
type UpdateCustomFieldOpts struct {
	Name          optional.Optional[string]
	Description   optional.Optional[string]
	Required      optional.Optional[bool]
	Archived      optional.Optional[bool]
	IndexExact    optional.Optional[bool]
	IndexRange    optional.Optional[bool]
	IndexFullText optional.Optional[bool]
	SortOrder     optional.Optional[int]
	Schema        optional.Optional[model.CustomFieldSchema]
}

// CustomFieldSearchQuery looks up resources by an indexed field.
type CustomFieldSearchQuery struct {
	DefinitionID model.ID
	Predicate    repository.CustomFieldPredicate
	Limit        int
}

//go:generate go tool mockgen -destination=mock/mock_custom_field_gen.go -package=mocksvc . CustomFieldService
type CustomFieldService interface {
	CreateDefinition(ctx context.Context, opts CreateCustomFieldOpts) (*model.CustomFieldDefinition, error)
	GetDefinition(ctx context.Context, id model.ID) (*model.CustomFieldDefinition, error)
	ListDefinitions(ctx context.Context, scope model.ID, target model.ResourceType, includeArchived bool) ([]*model.CustomFieldDefinition, error)
	UpdateDefinition(ctx context.Context, id model.ID, opts UpdateCustomFieldOpts) (*model.CustomFieldDefinition, error)
	ArchiveDefinition(ctx context.Context, id model.ID) (*model.CustomFieldDefinition, error)
	DeleteDefinition(ctx context.Context, id model.ID) error

	ListEffective(ctx context.Context, resourceID model.ID) ([]CustomFieldEntry, error)
	SetValue(ctx context.Context, resourceID, definitionID model.ID, value model.CustomFieldTypedValue) error
	DeleteValue(ctx context.Context, resourceID, definitionID model.ID) error
	Search(ctx context.Context, query CustomFieldSearchQuery) ([]model.ID, error)

	StageForResource(ctx context.Context, scope, resourceID model.ID, writes []CustomFieldWrite) error
	CommitForResource(ctx context.Context, resourceID model.ID) error
	AbortForResource(ctx context.Context, resourceID model.ID) error
	DeleteForResource(ctx context.Context, resourceID model.ID) error
	ReconcilePending(ctx context.Context) error
}

type customFieldService struct {
	runtime
	repo              repository.CustomFieldRepository
	permissionService PermissionService
	licenseService    LicenseService
}

func (s *customFieldService) requireFeature(ctx context.Context, wrap error) error {
	ok, err := s.licenseService.HasFeature(ctx, license.FeatureCustomFields)
	if err != nil {
		return errors.Join(wrap, err)
	}
	if !ok {
		return errors.Join(wrap, ErrFeatureDisabled)
	}
	return nil
}

func (s *customFieldService) CreateDefinition(ctx context.Context, opts CreateCustomFieldOpts) (*model.CustomFieldDefinition, error) {
	ctx, span := s.tracer.Start(ctx, "service.customFieldService/CreateDefinition")
	defer span.End()

	if err := s.requireFeature(ctx, ErrCustomFieldCreate); err != nil {
		return nil, err
	}
	if err := opts.Validate(); err != nil {
		return nil, errors.Join(ErrCustomFieldCreate, err)
	}
	if err := requireAction(ctx, s.permissionService, opts.Scope, model.ActionCustomFieldManage); err != nil {
		return nil, errors.Join(ErrCustomFieldCreate, err)
	}
	if _, err := s.permissionService.ListScopeAncestry(ctx, opts.Scope); err != nil {
		return nil, errors.Join(ErrCustomFieldCreate, err)
	}

	owner, err := ctxUserID(ctx)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldCreate, err)
	}
	clientID, _ := pkg.CtxOAuthClientID(ctx)

	def, err := model.NewCustomFieldDefinition(
		opts.Key,
		opts.Name,
		opts.Kind,
		opts.Scope,
		owner,
		opts.TargetType,
		opts.Schema,
	)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldCreate, err)
	}
	def.Description = opts.Description
	def.Required = opts.Required
	def.IndexExact = opts.IndexExact
	def.IndexRange = opts.IndexRange
	def.IndexFullText = opts.IndexFullText
	def.RegistrarClientID = clientID
	if opts.SortOrder.Defined && opts.SortOrder.Value != nil {
		def.SortOrder = *opts.SortOrder.Value
	} else {
		next, err := s.repo.NextSortOrder(ctx, opts.Scope, opts.TargetType)
		if err != nil {
			return nil, errors.Join(ErrCustomFieldCreate, err)
		}
		def.SortOrder = next
	}
	if err := def.Validate(); err != nil {
		return nil, errors.Join(ErrCustomFieldCreate, err)
	}

	created, err := s.repo.CreateDefinition(ctx, def)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldCreate, err)
	}
	return created, nil
}

func (s *customFieldService) GetDefinition(ctx context.Context, id model.ID) (*model.CustomFieldDefinition, error) {
	ctx, span := s.tracer.Start(ctx, "service.customFieldService/GetDefinition")
	defer span.End()

	def, err := s.repo.GetDefinition(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldGet, err)
	}
	if err := requireAction(ctx, s.permissionService, def.Scope, model.ActionCustomFieldManage); err != nil {
		read, ok := model.ReadActionFor(def.Scope.Type)
		if !ok {
			return nil, errors.Join(ErrCustomFieldGet, err)
		}
		if err := requireAction(ctx, s.permissionService, def.Scope, read); err != nil {
			return nil, errors.Join(ErrCustomFieldGet, err)
		}
	}
	return def, nil
}

func (s *customFieldService) ListDefinitions(
	ctx context.Context,
	scope model.ID,
	target model.ResourceType,
	includeArchived bool,
) ([]*model.CustomFieldDefinition, error) {
	ctx, span := s.tracer.Start(ctx, "service.customFieldService/ListDefinitions")
	defer span.End()

	if err := requireAction(ctx, s.permissionService, scope, model.ActionCustomFieldManage); err != nil {
		read, ok := model.ReadActionFor(scope.Type)
		if !ok {
			return nil, errors.Join(ErrCustomFieldList, err)
		}
		if err := requireAction(ctx, s.permissionService, scope, read); err != nil {
			return nil, errors.Join(ErrCustomFieldList, err)
		}
		includeArchived = false
	}

	ancestry, err := s.permissionService.ListScopeAncestry(ctx, scope)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldList, err)
	}
	return s.repo.ListDefinitions(ctx, ancestry, target, includeArchived)
}

func (s *customFieldService) UpdateDefinition(
	ctx context.Context,
	id model.ID,
	opts UpdateCustomFieldOpts,
) (*model.CustomFieldDefinition, error) {
	ctx, span := s.tracer.Start(ctx, "service.customFieldService/UpdateDefinition")
	defer span.End()

	if err := s.requireFeature(ctx, ErrCustomFieldUpdate); err != nil {
		return nil, err
	}
	current, err := s.repo.GetDefinition(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldUpdate, err)
	}
	if err := requireAction(ctx, s.permissionService, current.Scope, model.ActionCustomFieldManage); err != nil {
		return nil, errors.Join(ErrCustomFieldUpdate, err)
	}

	updated := *current
	if opts.Name.Defined && opts.Name.Value != nil {
		updated.Name = *opts.Name.Value
	}
	if opts.Description.Defined && opts.Description.Value != nil {
		updated.Description = *opts.Description.Value
	}
	if opts.Required.Defined && opts.Required.Value != nil {
		updated.Required = *opts.Required.Value
	}
	if opts.Archived.Defined && opts.Archived.Value != nil {
		updated.Archived = *opts.Archived.Value
	}
	if opts.IndexExact.Defined && opts.IndexExact.Value != nil {
		updated.IndexExact = *opts.IndexExact.Value
	}
	if opts.IndexRange.Defined && opts.IndexRange.Value != nil {
		updated.IndexRange = *opts.IndexRange.Value
	}
	if opts.IndexFullText.Defined && opts.IndexFullText.Value != nil {
		updated.IndexFullText = *opts.IndexFullText.Value
	}
	if opts.SortOrder.Defined && opts.SortOrder.Value != nil {
		updated.SortOrder = *opts.SortOrder.Value
	}
	if opts.Schema.Defined && opts.Schema.Value != nil {
		if err := evolveCustomFieldSchema(current, *opts.Schema.Value); err != nil {
			return nil, errors.Join(ErrCustomFieldUpdate, err)
		}
		updated.Schema = *opts.Schema.Value
	}
	if !current.IdentityEquals(&updated) {
		return nil, errors.Join(ErrCustomFieldUpdate, model.ErrCustomFieldIdentityImmutable)
	}

	out, err := s.repo.UpdateDefinition(ctx, &updated)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldUpdate, err)
	}
	return out, nil
}

func (s *customFieldService) ArchiveDefinition(ctx context.Context, id model.ID) (*model.CustomFieldDefinition, error) {
	return s.UpdateDefinition(ctx, id, UpdateCustomFieldOpts{
		Archived: optional.Some(true),
	})
}

func (s *customFieldService) DeleteDefinition(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.customFieldService/DeleteDefinition")
	defer span.End()

	if err := s.requireFeature(ctx, ErrCustomFieldDelete); err != nil {
		return err
	}
	def, err := s.repo.GetDefinition(ctx, id)
	if err != nil {
		return errors.Join(ErrCustomFieldDelete, err)
	}
	if err := requireAction(ctx, s.permissionService, def.Scope, model.ActionCustomFieldManage); err != nil {
		return errors.Join(ErrCustomFieldDelete, err)
	}
	count, err := s.repo.CountValues(ctx, id)
	if err != nil {
		return errors.Join(ErrCustomFieldDelete, err)
	}
	if err := def.CanHardDelete(count > 0); err != nil {
		return errors.Join(ErrCustomFieldDelete, err)
	}
	if err := s.repo.DeleteDefinition(ctx, id); err != nil {
		return errors.Join(ErrCustomFieldDelete, err)
	}
	return nil
}

func (s *customFieldService) ListEffective(ctx context.Context, resourceID model.ID) ([]CustomFieldEntry, error) {
	ctx, span := s.tracer.Start(ctx, "service.customFieldService/ListEffective")
	defer span.End()

	read, ok := model.ReadActionFor(resourceID.Type)
	if !ok {
		return nil, errors.Join(ErrCustomFieldValueGet, model.ErrInvalidResourceType)
	}
	if err := requireAction(ctx, s.permissionService, resourceID, read); err != nil {
		return nil, errors.Join(ErrCustomFieldValueGet, err)
	}

	ancestry, err := s.permissionService.ListScopeAncestry(ctx, resourceID)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldValueGet, err)
	}
	if len(ancestry) == 0 {
		return nil, errors.Join(ErrCustomFieldValueGet, repository.ErrNotFound)
	}

	defs, err := s.repo.ListDefinitions(ctx, ancestry, resourceID.Type, false)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldValueGet, err)
	}
	stored, err := s.repo.ListValues(ctx, resourceID, false)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldValueGet, err)
	}
	byDef := groupStoredValues(stored)

	var entries []CustomFieldEntry
	for _, def := range defs {
		if def.Archived {
			continue
		}
		entry := CustomFieldEntry{Definition: def}
		if atoms := byDef[def.ID.Composite()]; len(atoms) > 0 {
			typed, err := model.TypedValueFromAtomics(def.Kind, atoms)
			if err != nil {
				return nil, errors.Join(ErrCustomFieldValueGet, err)
			}
			entry.Value = &typed
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (s *customFieldService) SetValue(
	ctx context.Context,
	resourceID, definitionID model.ID,
	value model.CustomFieldTypedValue,
) error {
	ctx, span := s.tracer.Start(ctx, "service.customFieldService/SetValue")
	defer span.End()

	if err := s.requireFeature(ctx, ErrCustomFieldValueSet); err != nil {
		return err
	}
	def, err := s.writableDefinition(ctx, resourceID, definitionID, ErrCustomFieldValueSet)
	if err != nil {
		return err
	}
	atoms, err := value.Atomics()
	if err != nil {
		return errors.Join(ErrCustomFieldValueSet, err)
	}
	if err := model.ValidateAgainst(def, atoms); err != nil {
		return errors.Join(ErrCustomFieldValueSet, err)
	}
	if err := s.validateReferences(ctx, def, atoms); err != nil {
		return errors.Join(ErrCustomFieldValueSet, err)
	}
	if err := s.repo.ReplaceValues(ctx, def, resourceID, atoms, true); err != nil {
		return errors.Join(ErrCustomFieldValueSet, err)
	}
	return nil
}

func (s *customFieldService) DeleteValue(ctx context.Context, resourceID, definitionID model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.customFieldService/DeleteValue")
	defer span.End()

	if err := s.requireFeature(ctx, ErrCustomFieldValueDelete); err != nil {
		return err
	}
	def, err := s.writableDefinition(ctx, resourceID, definitionID, ErrCustomFieldValueDelete)
	if err != nil {
		return err
	}
	if def.Required {
		return errors.Join(ErrCustomFieldValueDelete, model.ErrCustomFieldRequired)
	}
	if err := s.repo.DeleteValues(ctx, definitionID, resourceID); err != nil {
		return errors.Join(ErrCustomFieldValueDelete, err)
	}
	return nil
}

func (s *customFieldService) Search(ctx context.Context, query CustomFieldSearchQuery) ([]model.ID, error) {
	ctx, span := s.tracer.Start(ctx, "service.customFieldService/Search")
	defer span.End()

	def, err := s.repo.GetDefinition(ctx, query.DefinitionID)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldSearch, err)
	}
	read, ok := model.ReadActionFor(def.TargetType)
	if !ok {
		return nil, errors.Join(ErrCustomFieldSearch, model.ErrInvalidResourceType)
	}

	candidates, err := s.repo.Search(ctx, query.DefinitionID, query.Predicate, query.Limit)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldSearch, err)
	}
	var allowed []model.ID
	for _, id := range candidates {
		if err := id.Validate(); err != nil {
			continue
		}
		if err := requireAction(ctx, s.permissionService, id, read); err != nil {
			continue
		}
		allowed = append(allowed, id)
	}
	return allowed, nil
}

func (s *customFieldService) StageForResource(
	ctx context.Context,
	scope, resourceID model.ID,
	writes []CustomFieldWrite,
) error {
	ctx, span := s.tracer.Start(ctx, "service.customFieldService/StageForResource")
	defer span.End()

	if err := s.requireFeature(ctx, ErrCustomFieldValueSet); err != nil {
		return err
	}

	ancestry, err := s.permissionService.ListScopeAncestry(ctx, scope)
	if err != nil {
		return errors.Join(ErrCustomFieldValueSet, err)
	}
	if len(ancestry) == 0 {
		return errors.Join(ErrCustomFieldValueSet, repository.ErrNotFound)
	}
	defs, err := s.repo.ListDefinitions(ctx, ancestry, resourceID.Type, false)
	if err != nil {
		return errors.Join(ErrCustomFieldValueSet, err)
	}
	byID := make(map[string]*model.CustomFieldDefinition, len(defs))
	for _, def := range defs {
		byID[def.ID.Composite()] = def
	}

	provided := make(map[string]struct{}, len(writes))
	staged := make([]repository.CustomFieldStagedWrite, 0, len(writes))
	for _, write := range writes {
		def, ok := byID[write.DefinitionID.Composite()]
		if !ok {
			return errors.Join(ErrCustomFieldValueSet, repository.ErrNotFound)
		}
		atoms, err := write.Value.Atomics()
		if err != nil {
			return errors.Join(ErrCustomFieldValueSet, err)
		}
		if err := model.ValidateAgainst(def, atoms); err != nil {
			return errors.Join(ErrCustomFieldValueSet, err)
		}
		if err := s.validateReferences(ctx, def, atoms); err != nil {
			return errors.Join(ErrCustomFieldValueSet, err)
		}
		staged = append(staged, repository.CustomFieldStagedWrite{
			Definition: def,
			Values:     atoms,
		})
		provided[def.ID.Composite()] = struct{}{}
	}

	for _, def := range defs {
		if def.Required {
			if _, ok := provided[def.ID.Composite()]; !ok {
				return errors.Join(ErrCustomFieldValueSet, model.ErrCustomFieldRequired)
			}
		}
	}

	if len(staged) == 0 {
		return nil
	}

	err = s.repo.StageValues(ctx, resourceID, staged, repository.CustomFieldOperation{
		Kind:       repository.CustomFieldOpStageValues,
		ResourceID: resourceID,
	})
	if err != nil {
		return errors.Join(ErrCustomFieldValueSet, err)
	}
	return nil
}

func (s *customFieldService) CommitForResource(ctx context.Context, resourceID model.ID) error {
	if err := s.repo.CommitValues(ctx, resourceID); err != nil {
		return errors.Join(ErrCustomFieldValueSet, err)
	}
	if err := s.repo.UpdatePendingOperations(ctx, resourceID, repository.CustomFieldOpCommitted); err != nil {
		return errors.Join(ErrCustomFieldValueSet, err)
	}
	return nil
}

func (s *customFieldService) AbortForResource(ctx context.Context, resourceID model.ID) error {
	if err := s.repo.AbortValues(ctx, resourceID); err != nil {
		return errors.Join(ErrCustomFieldValueSet, err)
	}
	if err := s.repo.UpdatePendingOperations(ctx, resourceID, repository.CustomFieldOpAborted); err != nil {
		return errors.Join(ErrCustomFieldValueSet, err)
	}
	return nil
}

func (s *customFieldService) DeleteForResource(ctx context.Context, resourceID model.ID) error {
	if err := s.repo.DeleteForResource(ctx, resourceID); err != nil {
		_, opErr := s.repo.CreateOperation(ctx, repository.CustomFieldOperation{
			Kind:       repository.CustomFieldOpDeleteResource,
			ResourceID: resourceID,
		})
		return errors.Join(ErrCustomFieldDelete, err, opErr)
	}
	return nil
}

func (s *customFieldService) ReconcilePending(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "service.customFieldService/ReconcilePending")
	defer span.End()

	ops, err := s.repo.ListPendingOperations(ctx, time.Now().UTC().Add(-customFieldReconcileStale), 100)
	if err != nil {
		return errors.Join(ErrCustomFieldReconcile, err)
	}
	for _, op := range ops {
		if op.Kind == repository.CustomFieldOpDeleteResource {
			if err := s.repo.DeleteForResource(ctx, op.ResourceID); err != nil {
				return errors.Join(ErrCustomFieldReconcile, err)
			}
			if err := s.repo.UpdateOperationStatus(ctx, op.ID, repository.CustomFieldOpCommitted); err != nil {
				return errors.Join(ErrCustomFieldReconcile, err)
			}
			continue
		}

		ancestry, existsErr := s.permissionService.ListScopeAncestry(ctx, op.ResourceID)
		if existsErr == nil && len(ancestry) == 0 {
			existsErr = repository.ErrNotFound
		}
		switch {
		case existsErr == nil:
			if err := s.repo.CommitValues(ctx, op.ResourceID); err != nil {
				return errors.Join(ErrCustomFieldReconcile, err)
			}
			if err := s.repo.UpdateOperationStatus(ctx, op.ID, repository.CustomFieldOpCommitted); err != nil {
				return errors.Join(ErrCustomFieldReconcile, err)
			}
		case errors.Is(existsErr, repository.ErrNotFound):
			if err := s.repo.AbortValues(ctx, op.ResourceID); err != nil {
				return errors.Join(ErrCustomFieldReconcile, err)
			}
			if err := s.repo.UpdateOperationStatus(ctx, op.ID, repository.CustomFieldOpAborted); err != nil {
				return errors.Join(ErrCustomFieldReconcile, err)
			}
		default:
			return errors.Join(ErrCustomFieldReconcile, existsErr)
		}
	}
	return nil
}

func (s *customFieldService) writableDefinition(
	ctx context.Context,
	resourceID, definitionID model.ID,
	wrap error,
) (*model.CustomFieldDefinition, error) {
	update, ok := model.UpdateActionFor(resourceID.Type)
	if !ok {
		return nil, errors.Join(wrap, model.ErrInvalidResourceType)
	}
	if err := requireAction(ctx, s.permissionService, resourceID, update); err != nil {
		return nil, errors.Join(wrap, err)
	}
	ancestry, err := s.permissionService.ListScopeAncestry(ctx, resourceID)
	if err != nil {
		return nil, errors.Join(wrap, err)
	}
	if len(ancestry) == 0 {
		return nil, errors.Join(wrap, repository.ErrNotFound)
	}
	def, err := s.repo.GetDefinition(ctx, definitionID)
	if err != nil {
		return nil, errors.Join(wrap, err)
	}
	if def.TargetType != resourceID.Type {
		return nil, errors.Join(wrap, model.ErrInvalidCustomFieldDetails)
	}
	if !scopeInAncestry(def.Scope, ancestry) {
		return nil, errors.Join(wrap, repository.ErrNotFound)
	}
	if err := def.AssertWritable(); err != nil {
		return nil, errors.Join(wrap, err)
	}
	return def, nil
}

func (s *customFieldService) validateReferences(
	ctx context.Context,
	_ *model.CustomFieldDefinition,
	atoms []model.CustomFieldAtomicValue,
) error {
	for _, atom := range atoms {
		if atom.UserID != nil {
			if err := s.requireExistingResource(ctx, *atom.UserID); err != nil {
				return err
			}
		}
		if atom.ResourceID != nil {
			if err := s.requireExistingResource(ctx, *atom.ResourceID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *customFieldService) requireExistingResource(ctx context.Context, id model.ID) error {
	ancestry, err := s.permissionService.ListScopeAncestry(ctx, id)
	if err != nil {
		return err
	}
	if len(ancestry) == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func scopeInAncestry(scope model.ID, ancestry []model.ID) bool {
	key := scope.Composite()
	for _, id := range ancestry {
		if id.Composite() == key {
			return true
		}
	}
	return false
}

func evolveCustomFieldSchema(current *model.CustomFieldDefinition, next model.CustomFieldSchema) error {
	if current.Schema.Select != nil {
		if next.Select == nil {
			return model.ErrCustomFieldIdentityImmutable
		}
		old := make(map[string]model.CustomFieldOption, len(current.Schema.Select.Options))
		for _, opt := range current.Schema.Select.Options {
			old[opt.Key] = opt
		}
		for _, opt := range next.Select.Options {
			delete(old, opt.Key)
		}
		if len(old) > 0 {
			return model.ErrCustomFieldOptionInUse
		}
	}
	return nil
}

func groupStoredValues(stored []repository.CustomFieldStoredValue) map[string][]model.CustomFieldAtomicValue {
	out := make(map[string][]model.CustomFieldAtomicValue)
	for _, row := range stored {
		key := row.DefinitionID.Composite()
		out[key] = append(out[key], row.Value)
	}
	return out
}

func NewCustomFieldService(
	repo repository.CustomFieldRepository,
	permissionService PermissionService,
	licenseService LicenseService,
	opts ...Option,
) (CustomFieldService, error) {
	rt, err := newRuntime(opts...)
	if err != nil {
		return nil, err
	}
	svc := &customFieldService{
		runtime:           rt,
		repo:              repo,
		permissionService: permissionService,
		licenseService:    licenseService,
	}
	if svc.repo == nil {
		return nil, ErrNoCustomFieldRepository
	}
	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}
	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}
	return svc, nil
}
