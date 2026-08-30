package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

const assignmentSyncPageSize = 1000

type IssueListOptions struct {
	Filter repository.IssueListFilter
	Sort   repository.IssueListSort
}

func (o IssueListOptions) withDefaults() IssueListOptions {
	if o.Sort.Field == "" {
		o.Sort.Field = repository.IssueListSortFieldRank
		o.Sort.Direction = repository.SortDirectionAsc
	}
	return o
}

// PartialAssignee is a lean assignment of a user to an issue.
type PartialAssignee struct {
	ID        model.ID
	Kind      model.AssignmentKind
	FirstName string
	LastName  string
	Picture   string
}

// PartialIssue represents a simplified issue within a project.
type PartialIssue struct {
	ID          model.ID
	Key         string
	NumericID   uint
	Parent      *PartialIssue
	Kind        model.IssueKind
	Title       string
	Description string
	Status      model.IssueStatus
	Priority    model.IssuePriority
	Assignments []PartialAssignee
	Labels      []PartialLabel
	Project     *PartialProject
	Namespace   *PartialNamespace
	ReportedBy  *PartialUser
	DueDate     *time.Time
	StartDate   *time.Time
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// Issue represents an issue returned by the service.
type Issue struct {
	ID              model.ID
	Key             string
	NumericID       uint
	Parent          *PartialIssue
	Kind            model.IssueKind
	Title           string
	Description     string
	Status          model.IssueStatus
	Priority        model.IssuePriority
	Resolution      model.IssueResolution
	ReportedBy      *PartialUser
	Assignments     []PartialAssignee
	Labels          []PartialLabel
	Project         *PartialProject
	Namespace       *PartialNamespace
	CommentCount    *int64
	DocumentCount   *int64
	AttachmentCount *int64
	WatcherCount    *int64
	RelationCount   *int64
	Links           []model.IssueLink
	DueDate         *time.Time
	StartDate       *time.Time
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
}

const (
	IssueRelationDirectionUnknown  IssueRelationDirection = iota
	IssueRelationDirectionOutgoing                        // outgoing
	IssueRelationDirectionIncoming                        // incoming
)

// IssueRelationDirection is whether a relation edge leaves or enters the path issue.
//
//go:generate go tool enumer -type=IssueRelationDirection -text -transform=noop -linecomment -output=issue_relation_direction_gen.go
type IssueRelationDirection uint8

// IssueRelation is a relation between the path issue and another issue.
type IssueRelation struct {
	ID        model.ID
	Kind      model.IssueRelationKind
	Direction IssueRelationDirection
	Related   *PartialIssue
	CreatedAt *time.Time
}

// CreateIssueOpts holds the data required to create an issue.
type CreateIssueOpts struct {
	Parent       *model.ID             `json:"parent" validate:"omitempty"`
	Kind         model.IssueKind       `json:"kind" validate:"required,min=1,max=4"`
	Title        string                `json:"title" validate:"required,min=3,max=120"`
	Description  string                `json:"description" validate:"omitempty,min=3"`
	Status       model.IssueStatus     `json:"status" validate:"omitempty,min=1,max=6"`
	Priority     model.IssuePriority   `json:"priority" validate:"omitempty,min=1,max=5"`
	Resolution   model.IssueResolution `json:"resolution" validate:"omitempty,min=1,max=7"`
	Links        []model.IssueLink     `json:"links" validate:"omitempty,dive"`
	DueDate      *time.Time            `json:"due_date" validate:"omitempty"`
	StartDate    *time.Time            `json:"start_date" validate:"omitempty"`
	CustomFields []CustomFieldWrite
}

// Validate validates the create options.
func (o *CreateIssueOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidIssueDetails, err)
	}
	if o.Parent != nil {
		if err := o.Parent.Validate(); err != nil {
			return errors.Join(model.ErrInvalidIssueDetails, err)
		}
	}
	return nil
}

// UpdateIssueOpts holds the fields that can be updated on an issue.
// Undefined fields (Defined == false) are left unchanged.
type UpdateIssueOpts struct {
	Kind        optional.Optional[model.IssueKind]
	Title       optional.Optional[string]
	Description optional.Optional[string]
	Status      optional.Optional[model.IssueStatus]
	Priority    optional.Optional[model.IssuePriority]
	Resolution  optional.Optional[model.IssueResolution]
	Links       optional.Optional[[]model.IssueLink]
	DueDate     optional.Optional[time.Time]
	StartDate   optional.Optional[time.Time]
	Assignees   optional.Optional[[]model.ID]
	Reviewers   optional.Optional[[]model.ID]
	Labels      optional.Optional[[]model.ID]
	Parent      optional.Optional[model.ID]
}

// IssueService serves the business logic of interacting with issues.
//
//go:generate go tool mockgen -destination=mock/mock_issue_gen.go -package=mocksvc . IssueService
type IssueService interface {
	// Create creates a new issue in a project. If the project does not exist,
	// an error is returned.
	Create(ctx context.Context, projectID model.ID, opts CreateIssueOpts) (*Issue, error)
	// Get returns an issue by its ID. If the issue does not exist, an error is
	// returned.
	Get(ctx context.Context, id model.ID) (*Issue, error)
	// GetByKey returns an issue by its composite key (e.g. MOB-1).
	GetByKey(ctx context.Context, namespaceID model.ID, key string) (*Issue, error)
	// List returns a cursor-paginated page of issues for a project.
	List(ctx context.Context, projectID model.ID, page CursorPage, opts IssueListOptions) (Page[*PartialIssue], error)
	// ListByNamespace returns a cursor-paginated page of issues across
	// projects in a namespace.
	ListByNamespace(ctx context.Context, namespaceID model.ID, page CursorPage, opts IssueListOptions) (Page[*PartialIssue], error)
	// ListByUser returns a cursor-paginated page of issues assigned to a user.
	ListByUser(ctx context.Context, userID model.ID, page CursorPage, opts IssueListOptions) (Page[*PartialIssue], error)
	// Update updates an issue. If the issue does not exist, an error is
	// returned.
	Update(ctx context.Context, id model.ID, opts UpdateIssueOpts) (*Issue, error)
	// Delete deletes an issue. If the issue does not exist, an error is
	// returned.
	Delete(ctx context.Context, id model.ID) error
	// ListRelations returns a cursor-paginated page of relations for an issue.
	ListRelations(ctx context.Context, issueID model.ID, page CursorPage) (Page[*IssueRelation], error)
	// AddRelation creates an outgoing relation from issueID to relatedID.
	AddRelation(ctx context.Context, issueID, relatedID model.ID, kind model.IssueRelationKind) (*IssueRelation, error)
	// UpdateRelation replaces a relation with a new outgoing edge of the given kind.
	UpdateRelation(ctx context.Context, issueID, relationID model.ID, kind model.IssueRelationKind) (*IssueRelation, error)
	// RemoveRelation deletes a relation of an issue by relation ID.
	RemoveRelation(ctx context.Context, issueID, relationID model.ID) error
}

// issueService is the concrete implementation of IssueService.
type issueService struct {
	runtime
	issueRepo          repository.IssueRepository
	assignmentRepo     repository.AssignmentRepository
	labelRepo          repository.LabelRepository
	permissionService  PermissionService
	licenseService     LicenseService
	searchService      SearchService
	customFieldService CustomFieldService
}

func partialAssigneesFromRepository(assignees []repository.PartialAssignee) []PartialAssignee {
	out := make([]PartialAssignee, len(assignees))
	for i, assignee := range assignees {
		out[i] = PartialAssignee{
			ID:        assignee.ID,
			Kind:      assignee.Kind,
			FirstName: assignee.FirstName,
			LastName:  assignee.LastName,
			Picture:   assignee.Picture,
		}
	}
	return out
}

func partialLabelsFromRepository(labels []repository.PartialLabel) []PartialLabel {
	out := make([]PartialLabel, len(labels))
	for i, label := range labels {
		out[i] = PartialLabel{ID: label.ID, Name: label.Name}
	}
	return out
}

func labelIDsFromPartial(labels []repository.PartialLabel) []model.ID {
	ids := make([]model.ID, len(labels))
	for i, label := range labels {
		ids[i] = label.ID
	}
	return ids
}

func partialIssueFromRepository(i *repository.PartialIssue) *PartialIssue {
	if i == nil {
		return nil
	}

	return &PartialIssue{
		ID:          i.ID,
		Key:         i.Key,
		NumericID:   i.NumericID,
		Parent:      partialIssueFromRepository(i.Parent),
		Kind:        i.Kind,
		Title:       i.Title,
		Description: i.Description,
		Status:      i.Status,
		Priority:    i.Priority,
		Assignments: partialAssigneesFromRepository(i.Assignments),
		Labels:      partialLabelsFromRepository(i.Labels),
		Project:     partialProjectFromRepository(i.Project),
		Namespace:   partialNamespaceFromRepository(i.Namespace),
		ReportedBy:  partialUserFromRepository(i.ReportedBy),
		DueDate:     i.DueDate,
		StartDate:   i.StartDate,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}
}

func partialIssueFromRepoIssue(i *repository.Issue) *PartialIssue {
	if i == nil {
		return nil
	}

	return partialIssueFromRepository(&repository.PartialIssue{
		ID:          i.ID,
		Key:         i.Key,
		NumericID:   i.NumericID,
		Parent:      i.Parent,
		Kind:        i.Kind,
		Title:       i.Title,
		Description: i.Description,
		Status:      i.Status,
		Priority:    i.Priority,
		Assignments: i.Assignments,
		Labels:      i.Labels,
		Project:     i.Project,
		Namespace:   i.Namespace,
		ReportedBy:  i.ReportedBy,
		DueDate:     i.DueDate,
		StartDate:   i.StartDate,
	})
}

func issueRelationFromItem(issueID model.ID, item *repository.IssueRelationItem) *IssueRelation {
	direction := IssueRelationDirectionIncoming
	if item.Source == issueID {
		direction = IssueRelationDirectionOutgoing
	}

	return &IssueRelation{
		ID:        item.ID,
		Kind:      item.Kind,
		Direction: direction,
		Related:   partialIssueFromRepository(item.Related),
		CreatedAt: item.CreatedAt,
	}
}

func relatedIDFromRelation(issueID model.ID, rel *repository.IssueRelation) (model.ID, bool) {
	switch issueID {
	case rel.Source:
		return rel.Target, true
	case rel.Target:
		return rel.Source, true
	default:
		return model.ID{}, false
	}
}

func validateEditableRelationKind(kind model.IssueRelationKind) error {
	if !kind.IsAIssueRelationKind() {
		return model.ErrInvalidIssueRelationKind
	}

	switch kind {
	case model.IssueRelationKindSubtaskOf, model.IssueRelationKindDependsOn:
		return ErrIssueReservedRelationKind
	}

	return nil
}

func issueFromRepository(i *repository.Issue) *Issue {
	if i == nil {
		return nil
	}

	return &Issue{
		ID:              i.ID,
		Key:             i.Key,
		NumericID:       i.NumericID,
		Parent:          partialIssueFromRepository(i.Parent),
		Kind:            i.Kind,
		Title:           i.Title,
		Description:     i.Description,
		Status:          i.Status,
		Priority:        i.Priority,
		Resolution:      i.Resolution,
		ReportedBy:      partialUserFromRepository(i.ReportedBy),
		Assignments:     partialAssigneesFromRepository(i.Assignments),
		Labels:          partialLabelsFromRepository(i.Labels),
		Project:         partialProjectFromRepository(i.Project),
		Namespace:       partialNamespaceFromRepository(i.Namespace),
		CommentCount:    i.CommentCount,
		DocumentCount:   i.DocumentCount,
		AttachmentCount: i.AttachmentCount,
		WatcherCount:    i.WatcherCount,
		RelationCount:   i.RelationCount,
		Links:           i.Links,
		DueDate:         i.DueDate,
		StartDate:       i.StartDate,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
	}
}

func partialIssuesFromRepository(issues []*repository.PartialIssue) []*PartialIssue {
	out := make([]*PartialIssue, len(issues))
	for i, issue := range issues {
		out[i] = partialIssueFromRepository(issue)
	}
	return out
}

func optionalIDs(opt optional.Optional[[]model.ID]) ([]model.ID, error) {
	if !opt.Defined {
		return nil, nil
	}
	if opt.Value == nil {
		return make([]model.ID, 0), nil
	}

	ids := make([]model.ID, 0, len(*opt.Value))
	for _, id := range *opt.Value {
		if err := id.Validate(); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *issueService) syncAssignments(ctx context.Context, issueID model.ID, kind model.AssignmentKind, userIDs []model.ID) error {
	page := repository.CursorPage{Size: assignmentSyncPageSize}
	existing := make([]*repository.Assignment, 0)
	for {
		assignments, err := s.assignmentRepo.ListByResource(
			ctx,
			issueID,
			page,
			repository.AssignmentListProjection(),
		)
		if err != nil {
			return err
		}

		existing = append(existing, assignments.Items...)
		if !assignments.PageInfo.HasMore || assignments.PageInfo.NextPageToken == nil {
			break
		}

		page.Token = assignments.PageInfo.NextPageToken
	}

	desired := make(map[string]model.ID, len(userIDs))
	for _, id := range userIDs {
		desired[id.String()] = id
	}

	current := make(map[string]*repository.Assignment)
	for _, assignment := range existing {
		if assignment.Kind != kind {
			continue
		}
		current[assignment.User.String()] = assignment
	}

	for userID, assignment := range current {
		if _, ok := desired[userID]; ok {
			continue
		}
		if err := s.assignmentRepo.Delete(ctx, assignment.ID); err != nil {
			return err
		}
	}

	for userID, id := range desired {
		if _, ok := current[userID]; ok {
			continue
		}
		if _, err := s.assignmentRepo.Create(ctx, repository.CreateAssignmentOpts{
			Kind:     kind,
			User:     id,
			Resource: issueID,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *issueService) syncLabels(ctx context.Context, issueID model.ID, current, desired []model.ID) error {
	desiredSet := make(map[string]model.ID, len(desired))
	for _, id := range desired {
		desiredSet[id.String()] = id
	}

	currentSet := make(map[string]model.ID, len(current))
	for _, id := range current {
		currentSet[id.String()] = id
	}

	for idStr, id := range currentSet {
		if _, ok := desiredSet[idStr]; ok {
			continue
		}
		if err := s.labelRepo.DetachFrom(ctx, id, issueID); err != nil {
			return err
		}
	}

	for idStr, id := range desiredSet {
		if _, ok := currentSet[idStr]; ok {
			continue
		}
		if err := s.labelRepo.AttachTo(ctx, id, issueID); err != nil {
			return err
		}
	}

	return nil
}

func (s *issueService) validateParentUpdate(ctx context.Context, issueID model.ID, parent optional.Optional[model.ID]) error {
	if !parent.Defined || parent.Value == nil {
		return nil
	}

	parentID := *parent.Value
	if err := parentID.Validate(); err != nil {
		return err
	}
	if parentID == issueID {
		return ErrIssueSelfRelation
	}
	if err := requireAction(ctx, s.permissionService, parentID, model.ActionIssueRead); err != nil {
		return err
	}

	return nil
}

func (s *issueService) syncParent(ctx context.Context, issueID model.ID, current *repository.PartialIssue, parent optional.Optional[model.ID]) error {
	var currentID *model.ID
	if current != nil {
		id := current.ID
		currentID = &id
	}

	desired := parent.Value
	if currentID == nil && desired == nil {
		return nil
	}
	if currentID != nil && desired != nil && *currentID == *desired {
		return nil
	}

	if currentID != nil {
		if err := s.issueRepo.RemoveRelation(ctx, issueID, *currentID, model.IssueRelationKindSubtaskOf); err != nil {
			return err
		}
	}

	if desired != nil {
		if _, err := s.issueRepo.AddRelation(ctx, repository.CreateIssueRelationOpts{
			Source: issueID,
			Target: *desired,
			Kind:   model.IssueRelationKindSubtaskOf,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *issueService) Create(ctx context.Context, projectID model.ID, opts CreateIssueOpts) (*Issue, error) {
	ctx, span := s.tracer.Start(ctx, "service.issueService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrIssueCreate, license.ErrLicenseExpired)
	}

	if err := projectID.Validate(); err != nil {
		return nil, errors.Join(ErrIssueCreate, err)
	}

	if err := opts.Validate(); err != nil {
		return nil, errors.Join(ErrIssueCreate, err)
	}

	if err := requireAction(ctx, s.permissionService, projectID, model.ActionIssueCreate); err != nil {
		return nil, errors.Join(ErrIssueCreate, err)
	}

	parent := optional.None[model.ID]()
	if opts.Parent != nil {
		parent = optional.Some(*opts.Parent)
	}
	if err := s.validateParentUpdate(ctx, model.ID{}, parent); err != nil {
		return nil, errors.Join(ErrIssueCreate, err)
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return nil, errors.Join(ErrIssueCreate, err)
	}

	issueID := model.MustNewID(model.ResourceTypeIssue)
	if err := s.customFieldService.StageForResource(ctx, projectID, issueID, opts.CustomFields); err != nil {
		return nil, errors.Join(ErrIssueCreate, err)
	}

	status := opts.Status
	if status == 0 {
		status = model.IssueStatusOpen
	}

	priority := opts.Priority
	if priority == 0 {
		priority = model.IssuePriorityNormal
	}

	resolution := opts.Resolution
	if resolution == 0 {
		resolution = model.IssueResolutionNone
	}

	links := opts.Links
	if links == nil {
		links = make([]model.IssueLink, 0)
	}

	issue, err := s.issueRepo.Create(ctx, repository.CreateIssueOpts{
		ID:          &issueID,
		ProjectID:   projectID,
		Parent:      opts.Parent,
		Kind:        opts.Kind,
		Title:       opts.Title,
		Description: opts.Description,
		Status:      status,
		Priority:    priority,
		Resolution:  resolution,
		ReportedBy:  userID,
		Links:       links,
		DueDate:     opts.DueDate,
		StartDate:   opts.StartDate,
	})
	if err != nil {
		if abortErr := s.customFieldService.AbortForResource(ctx, issueID); abortErr != nil {
			return nil, errors.Join(ErrIssueCreate, err, abortErr)
		}
		return nil, errors.Join(ErrIssueCreate, err)
	}
	if err := s.customFieldService.CommitForResource(ctx, issueID); err != nil {
		s.logger.Warn(ctx, "failed to commit custom field values",
			log.WithError(err),
			log.WithValue(issueID.Composite()),
		)
	}

	actions, err := roleTemplateActions(model.RoleKeyIssueMaintainer)
	if err != nil {
		return nil, errors.Join(ErrIssueCreate, err)
	}
	if err := s.permissionService.BootstrapCreator(ctx, userID, issue.ID, actions); err != nil {
		return nil, errors.Join(ErrIssueCreate, err)
	}

	out := issueFromRepository(issue)
	enqueueSearchIndex(ctx, s.logger, s.searchService, out.ID)
	return out, nil
}

func (s *issueService) Get(ctx context.Context, id model.ID) (*Issue, error) {
	ctx, span := s.tracer.Start(ctx, "service.issueService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrIssueGet, err)
	}

	if err := requireAction(ctx, s.permissionService, id, model.ActionIssueRead); err != nil {
		return nil, errors.Join(ErrIssueGet, err)
	}

	issue, err := s.issueRepo.Get(ctx, id, repository.IssueDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrIssueGet, err)
	}

	return issueFromRepository(issue), nil
}

func (s *issueService) GetByKey(ctx context.Context, namespaceID model.ID, key string) (*Issue, error) {
	ctx, span := s.tracer.Start(ctx, "service.issueService/GetByKey")
	defer span.End()

	if err := namespaceID.Validate(); err != nil {
		return nil, errors.Join(ErrIssueGet, err)
	}

	if _, _, err := model.ParseIssueKey(key); err != nil {
		return nil, errors.Join(ErrIssueGet, err)
	}

	issue, err := s.issueRepo.GetByKey(ctx, namespaceID, key, repository.IssueDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrIssueGet, err)
	}

	if err := requireAction(ctx, s.permissionService, issue.ID, model.ActionIssueRead); err != nil {
		return nil, errors.Join(ErrIssueGet, err)
	}

	return issueFromRepository(issue), nil
}

func (s *issueService) List(ctx context.Context, projectID model.ID, page CursorPage, opts IssueListOptions) (Page[*PartialIssue], error) {
	ctx, span := s.tracer.Start(ctx, "service.issueService/List")
	defer span.End()

	if err := projectID.Validate(); err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}
	scopeIDs, allowed, err := resolvedListScopeIDs(ctx, s.permissionService, projectID, model.ActionIssueRead)
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}
	if !allowed {
		return repository.EmptyPage[*PartialIssue](), nil
	}
	listOpts := opts.withDefaults()

	issues, err := s.issueRepo.ListForProject(ctx, repository.IssueListQuery{
		ProjectID:  projectID,
		ActorID:    userID,
		Action:     model.ActionIssueRead,
		ScopeIDs:   scopeIDs,
		SortField:  listOpts.Sort.Field,
		Page:       normalized,
		Order:      listOpts.Sort.Direction,
		Filter:     listOpts.Filter,
		Projection: repository.IssueListForProjectProjection(),
	})
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}

	return mapPage(issues, partialIssueFromRepository), nil
}

func (s *issueService) ListByNamespace(ctx context.Context, namespaceID model.ID, page CursorPage, opts IssueListOptions) (Page[*PartialIssue], error) {
	ctx, span := s.tracer.Start(ctx, "service.issueService/ListByNamespace")
	defer span.End()

	if err := namespaceID.Validate(); err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}
	scopeIDs, allowed, err := resolvedListScopeIDs(ctx, s.permissionService, namespaceID, model.ActionIssueRead)
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}
	if !allowed {
		return repository.EmptyPage[*PartialIssue](), nil
	}
	listOpts := opts.withDefaults()

	issues, err := s.issueRepo.ListForNamespace(ctx, repository.IssueListForNamespaceQuery{
		NamespaceID: namespaceID,
		ActorID:     userID,
		Action:      model.ActionIssueRead,
		ScopeIDs:    scopeIDs,
		SortField:   listOpts.Sort.Field,
		Page:        normalized,
		Order:       listOpts.Sort.Direction,
		Filter:      listOpts.Filter,
		Projection:  repository.IssueListForNamespaceProjection(),
	})
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}

	return mapPage(issues, partialIssueFromRepository), nil
}

func (s *issueService) ListByUser(ctx context.Context, userID model.ID, page CursorPage, opts IssueListOptions) (Page[*PartialIssue], error) {
	ctx, span := s.tracer.Start(ctx, "service.issueService/ListByUser")
	defer span.End()

	if err := userID.Validate(); err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}

	actorID, err := ctxUserID(ctx)
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}

	if actorID != userID {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, ErrNoPermission)
	}
	scopeIDs, err := s.permissionService.CtxUserListGrantScopes(ctx, model.ActionIssueRead)
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}
	if len(scopeIDs) == 0 {
		return repository.EmptyPage[*PartialIssue](), nil
	}
	listOpts := opts.withDefaults()

	issues, err := s.issueRepo.ListForUser(ctx, repository.IssueListForUserQuery{
		UserID:     userID,
		ActorID:    actorID,
		Action:     model.ActionIssueRead,
		ScopeIDs:   scopeIDs,
		SortField:  listOpts.Sort.Field,
		Page:       normalized,
		Order:      listOpts.Sort.Direction,
		Filter:     listOpts.Filter,
		Projection: repository.IssueListForUserProjection(),
	})
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueList, err)
	}

	return mapPage(issues, partialIssueFromRepository), nil
}

func (s *issueService) Update(ctx context.Context, id model.ID, opts UpdateIssueOpts) (*Issue, error) {
	ctx, span := s.tracer.Start(ctx, "service.issueService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrIssueUpdate, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrIssueUpdate, err)
	}

	if err := requireAction(ctx, s.permissionService, id, model.ActionIssueUpdate); err != nil {
		return nil, errors.Join(ErrIssueUpdate, err)
	}

	assignees, err := optionalIDs(opts.Assignees)
	if err != nil {
		return nil, errors.Join(ErrIssueUpdate, err)
	}
	if opts.Assignees.Defined && len(assignees) > 1 {
		ok, err := s.licenseService.HasFeature(ctx, license.FeatureMultipleAssignees)
		if err != nil {
			return nil, errors.Join(ErrIssueUpdate, err)
		}
		if !ok {
			return nil, errors.Join(ErrIssueUpdate, ErrQuotaExceeded)
		}
	}

	reviewers, err := optionalIDs(opts.Reviewers)
	if err != nil {
		return nil, errors.Join(ErrIssueUpdate, err)
	}

	labels, err := optionalIDs(opts.Labels)
	if err != nil {
		return nil, errors.Join(ErrIssueUpdate, err)
	}

	if err := s.validateParentUpdate(ctx, id, opts.Parent); err != nil {
		return nil, errors.Join(ErrIssueUpdate, err)
	}

	issue, err := s.issueRepo.Update(ctx, id, repository.UpdateIssueOpts{
		Kind:        opts.Kind,
		Title:       opts.Title,
		Description: opts.Description,
		Status:      opts.Status,
		Priority:    opts.Priority,
		Resolution:  opts.Resolution,
		Links:       opts.Links,
		DueDate:     opts.DueDate,
		StartDate:   opts.StartDate,
	}, repository.IssueDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrIssueUpdate, err)
	}

	if opts.Assignees.Defined {
		if err := s.syncAssignments(ctx, id, model.AssignmentKindAssignee, assignees); err != nil {
			return nil, errors.Join(ErrIssueUpdate, err)
		}
	}

	if opts.Reviewers.Defined {
		if err := s.syncAssignments(ctx, id, model.AssignmentKindReviewer, reviewers); err != nil {
			return nil, errors.Join(ErrIssueUpdate, err)
		}
	}

	if opts.Labels.Defined {
		if err := s.syncLabels(ctx, id, labelIDsFromPartial(issue.Labels), labels); err != nil {
			return nil, errors.Join(ErrIssueUpdate, err)
		}
	}

	if opts.Parent.Defined {
		if err := s.syncParent(ctx, id, issue.Parent, opts.Parent); err != nil {
			return nil, errors.Join(ErrIssueUpdate, err)
		}
	}

	if opts.Assignees.Defined || opts.Reviewers.Defined || opts.Labels.Defined || opts.Parent.Defined {
		issue, err = s.issueRepo.Get(ctx, id, repository.IssueDetailProjection())
		if err != nil {
			return nil, errors.Join(ErrIssueUpdate, err)
		}
	}

	out := issueFromRepository(issue)
	enqueueSearchIndex(ctx, s.logger, s.searchService, out.ID)
	return out, nil
}

func (s *issueService) Delete(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.issueService/Delete")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrIssueDelete, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return errors.Join(ErrIssueDelete, err)
	}

	if err := requireAction(ctx, s.permissionService, id, model.ActionIssueDelete); err != nil {
		return errors.Join(ErrIssueDelete, err)
	}

	if err := s.issueRepo.Delete(ctx, id); err != nil {
		return errors.Join(ErrIssueDelete, err)
	}
	if err := s.customFieldService.DeleteForResource(ctx, id); err != nil {
		s.logger.Warn(ctx, "failed to delete custom field values",
			log.WithError(err),
			log.WithValue(id.Composite()),
		)
	}

	if err := s.searchService.Delete(ctx, id); err != nil {
		s.logger.Warn(ctx, "failed to delete search document",
			log.WithError(err),
			log.WithValue(id.Composite()),
		)
	}
	return nil
}

func (s *issueService) ListRelations(ctx context.Context, issueID model.ID, page CursorPage) (Page[*IssueRelation], error) {
	ctx, span := s.tracer.Start(ctx, "service.issueService/ListRelations")
	defer span.End()

	if err := issueID.Validate(); err != nil {
		return Page[*IssueRelation]{}, errors.Join(ErrIssueGetRelations, err)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*IssueRelation]{}, errors.Join(ErrIssueGetRelations, err)
	}

	if err := requireAction(ctx, s.permissionService, issueID, model.ActionIssueRead); err != nil {
		return Page[*IssueRelation]{}, errors.Join(ErrIssueGetRelations, err)
	}

	relations, err := s.issueRepo.ListRelations(ctx, repository.IssueRelationListQuery{
		IssueID: issueID,
		Page:    normalized,
	})
	if err != nil {
		return Page[*IssueRelation]{}, errors.Join(ErrIssueGetRelations, err)
	}

	return mapPage(relations, func(item *repository.IssueRelationItem) *IssueRelation {
		return issueRelationFromItem(issueID, item)
	}), nil
}

func (s *issueService) AddRelation(ctx context.Context, issueID, relatedID model.ID, kind model.IssueRelationKind) (*IssueRelation, error) {
	ctx, span := s.tracer.Start(ctx, "service.issueService/AddRelation")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrIssueAddRelation, license.ErrLicenseExpired)
	}

	if err := issueID.Validate(); err != nil {
		return nil, errors.Join(ErrIssueAddRelation, err)
	}
	if err := relatedID.Validate(); err != nil {
		return nil, errors.Join(ErrIssueAddRelation, err)
	}
	if issueID == relatedID {
		return nil, errors.Join(ErrIssueAddRelation, ErrIssueSelfRelation)
	}
	if err := validateEditableRelationKind(kind); err != nil {
		return nil, errors.Join(ErrIssueAddRelation, err)
	}

	if err := requireAction(ctx, s.permissionService, issueID, model.ActionIssueUpdate); err != nil {
		return nil, errors.Join(ErrIssueAddRelation, err)
	}
	if err := requireAction(ctx, s.permissionService, relatedID, model.ActionIssueRead); err != nil {
		return nil, errors.Join(ErrIssueAddRelation, err)
	}

	created, err := s.issueRepo.AddRelation(ctx, repository.CreateIssueRelationOpts{
		Source: issueID,
		Target: relatedID,
		Kind:   kind,
	})
	if err != nil {
		return nil, errors.Join(ErrIssueAddRelation, err)
	}

	related, err := s.issueRepo.Get(ctx, relatedID, repository.IssueProjection{})
	if err != nil {
		return nil, errors.Join(ErrIssueAddRelation, err)
	}

	return &IssueRelation{
		ID:        created.ID,
		Kind:      created.Kind,
		Direction: IssueRelationDirectionOutgoing,
		Related:   partialIssueFromRepoIssue(related),
		CreatedAt: created.CreatedAt,
	}, nil
}

func (s *issueService) UpdateRelation(ctx context.Context, issueID, relationID model.ID, kind model.IssueRelationKind) (*IssueRelation, error) {
	ctx, span := s.tracer.Start(ctx, "service.issueService/UpdateRelation")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrIssueUpdateRelation, license.ErrLicenseExpired)
	}

	if err := issueID.Validate(); err != nil {
		return nil, errors.Join(ErrIssueUpdateRelation, err)
	}
	if err := relationID.Validate(); err != nil {
		return nil, errors.Join(ErrIssueUpdateRelation, err)
	}
	if err := validateEditableRelationKind(kind); err != nil {
		return nil, errors.Join(ErrIssueUpdateRelation, err)
	}

	if err := requireAction(ctx, s.permissionService, issueID, model.ActionIssueUpdate); err != nil {
		return nil, errors.Join(ErrIssueUpdateRelation, err)
	}

	existing, err := s.issueRepo.GetRelation(ctx, relationID)
	if err != nil {
		return nil, errors.Join(ErrIssueUpdateRelation, err)
	}

	relatedID, ok := relatedIDFromRelation(issueID, existing)
	if !ok {
		return nil, errors.Join(ErrIssueUpdateRelation, repository.ErrNotFound)
	}

	if err := requireAction(ctx, s.permissionService, relatedID, model.ActionIssueRead); err != nil {
		return nil, errors.Join(ErrIssueUpdateRelation, err)
	}

	if err := s.issueRepo.RemoveRelationByID(ctx, relationID); err != nil {
		return nil, errors.Join(ErrIssueUpdateRelation, err)
	}

	created, err := s.issueRepo.AddRelation(ctx, repository.CreateIssueRelationOpts{
		Source: issueID,
		Target: relatedID,
		Kind:   kind,
	})
	if err != nil {
		return nil, errors.Join(ErrIssueUpdateRelation, err)
	}

	related, err := s.issueRepo.Get(ctx, relatedID, repository.IssueProjection{})
	if err != nil {
		return nil, errors.Join(ErrIssueUpdateRelation, err)
	}

	return &IssueRelation{
		ID:        created.ID,
		Kind:      created.Kind,
		Direction: IssueRelationDirectionOutgoing,
		Related:   partialIssueFromRepoIssue(related),
		CreatedAt: created.CreatedAt,
	}, nil
}

func (s *issueService) RemoveRelation(ctx context.Context, issueID, relationID model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.issueService/RemoveRelation")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrIssueRemoveRelation, license.ErrLicenseExpired)
	}

	if err := issueID.Validate(); err != nil {
		return errors.Join(ErrIssueRemoveRelation, err)
	}
	if err := relationID.Validate(); err != nil {
		return errors.Join(ErrIssueRemoveRelation, err)
	}

	if err := requireAction(ctx, s.permissionService, issueID, model.ActionIssueUpdate); err != nil {
		return errors.Join(ErrIssueRemoveRelation, err)
	}

	existing, err := s.issueRepo.GetRelation(ctx, relationID)
	if err != nil {
		return errors.Join(ErrIssueRemoveRelation, err)
	}

	relatedID, ok := relatedIDFromRelation(issueID, existing)
	if !ok {
		return errors.Join(ErrIssueRemoveRelation, repository.ErrNotFound)
	}

	if err := requireAction(ctx, s.permissionService, relatedID, model.ActionIssueRead); err != nil {
		return errors.Join(ErrIssueRemoveRelation, err)
	}

	if err := s.issueRepo.RemoveRelationByID(ctx, relationID); err != nil {
		return errors.Join(ErrIssueRemoveRelation, err)
	}

	return nil
}

// NewIssueService returns a new instance of the IssueService interface.
func NewIssueService(
	issueRepo repository.IssueRepository,
	assignmentRepo repository.AssignmentRepository,
	labelRepo repository.LabelRepository,
	permissionService PermissionService,
	licenseService LicenseService,
	searchService SearchService,
	customFieldService CustomFieldService,
	opts ...Option,
) (IssueService, error) {
	rt, err := newRuntime(opts...)
	if err != nil {
		return nil, err
	}

	svc := &issueService{
		runtime:            rt,
		issueRepo:          issueRepo,
		assignmentRepo:     assignmentRepo,
		labelRepo:          labelRepo,
		permissionService:  permissionService,
		licenseService:     licenseService,
		searchService:      searchService,
		customFieldService: customFieldService,
	}

	if svc.issueRepo == nil {
		return nil, ErrNoIssueRepository
	}

	if svc.assignmentRepo == nil {
		return nil, ErrNoAssignmentRepository
	}

	if svc.labelRepo == nil {
		return nil, ErrNoLabelRepository
	}

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	if svc.searchService == nil {
		return nil, ErrNoSearchService
	}

	if svc.customFieldService == nil {
		return nil, ErrNoCustomFieldService
	}

	return svc, nil
}
