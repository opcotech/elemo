package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	IssueKindEpic  IssueKind = iota + 1 // epic
	IssueKindStory                      // story
	IssueKindTask                       // task
	IssueKindBug                        // bug
)

const (
	IssueStatusOpen       IssueStatus = iota + 1 // open
	IssueStatusInProgress                        // in progress
	IssueStatusBlocked                           // blocked
	IssueStatusReview                            // review
	IssueStatusDone                              // done
	IssueStatusClosed                            // closed
)

const (
	IssueResolutionNone            IssueResolution = iota + 1 // none
	IssueResolutionFixed                                      // fixed
	IssueResolutionDuplicate                                  // duplicate
	IssueResolutionWontFix                                    // won't fix
	IssueResolutionInvalid                                    // invalid
	IssueResolutionIncomplete                                 // incomplete
	IssueResolutionCannotReproduce                            // cannot reproduce
)

const (
	IssuePriorityLow      IssuePriority = iota + 1 // low
	IssuePriorityMedium                            // medium
	IssuePriorityHigh                              // high
	IssuePriorityCritical                          // critical
)

const (
	IssueRelationKindBlockedBy    IssueRelationKind = iota + 1 // blocked by
	IssueRelationKindBlocks                                    // blocks
	IssueRelationKindDependsOn                                 // depends on
	IssueRelationKindDuplicatedBy                              // duplicated by
	IssueRelationKindDuplicates                                // duplicates
	IssueRelationKindRelatedTo                                 // related to
	IssueRelationKindSubtaskOf                                 // subtask of
)

// IssueKind represents a kind of issue.
//
//go:generate go tool enumer -type=IssueKind -text -transform=noop -linecomment -output=issue_kind_gen.go
type IssueKind uint8

// IssueStatus represents the status of an issue.
//
//go:generate go tool enumer -type=IssueStatus -text -transform=noop -linecomment -output=issue_status_gen.go
type IssueStatus uint8

// IssueResolution represents the resolution of an issue.
//
//go:generate go tool enumer -type=IssueResolution -text -transform=noop -linecomment -output=issue_resolution_gen.go
type IssueResolution uint8

// IssuePriority represents the priority of an issue.
//
//go:generate go tool enumer -type=IssuePriority -text -transform=noop -linecomment -output=issue_priority_gen.go
type IssuePriority uint8

// IssueRelationKind represents the kind of relation between two issues.
//
//go:generate go tool enumer -type=IssueRelationKind -text -transform=noop -linecomment -output=issue_relation_kind_gen.go
type IssueRelationKind uint8

// IssueRelation represents a relation between two issues.
type IssueRelation struct {
	ID        ID                `json:"id" validate:"required"`
	Source    ID                `json:"source" validate:"required"`
	Target    ID                `json:"target" validate:"required"`
	Kind      IssueRelationKind `json:"kind" validate:"required,min=1,max=7"`
	CreatedAt *time.Time        `json:"created_at" validate:"omitempty"`
	UpdatedAt *time.Time        `json:"updated_at" validate:"omitempty"`
}

// Validate validates the issue relation details.
func (i *IssueRelation) Validate() error {
	if err := validate.Struct(i); err != nil {
		return errors.Join(ErrInvalidIssueRelationDetails, err)
	}
	if err := i.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidIssueRelationDetails, err)
	}
	if err := i.Source.Validate(); err != nil {
		return errors.Join(ErrInvalidIssueRelationDetails, err)
	}
	if err := i.Target.Validate(); err != nil {
		return errors.Join(ErrInvalidIssueRelationDetails, err)
	}
	return nil
}

// Issue represents an issue in the system that can be assigned to a
// user and belong to a project or another Issue.
type Issue struct {
	ID          ID              `json:"id" validate:"required"`
	NumericID   uint            `json:"numeric_id" validate:"required"`
	Parent      *ID             `json:"parent" validate:"omitempty"`
	Kind        IssueKind       `json:"kind" validate:"required,min=1,max=4"`
	Title       string          `json:"title" validate:"required,min=3,max=120"`
	Description string          `json:"description" validate:"omitempty,min=10"`
	Status      IssueStatus     `json:"status" validate:"required,min=1,max=6"`
	Priority    IssuePriority   `json:"priority" validate:"required,min=1,max=5"`
	Resolution  IssueResolution `json:"resolution" validate:"required,min=1,max=7"`
	ReportedBy  ID              `json:"reported_by" validate:"required"`
	Assignees   []ID            `json:"assignees" validate:"omitempty,dive"`
	Labels      []ID            `json:"labels" validate:"omitempty,dive"`
	Comments    []ID            `json:"comments" validate:"omitempty,dive"`
	Attachments []ID            `json:"attachments" validate:"omitempty,dive"`
	Watchers    []ID            `json:"watchers" validate:"omitempty,dive"`
	Relations   []ID            `json:"relations" validate:"omitempty,dive"`
	Links       []string        `json:"links" validate:"omitempty,dive,url"`
	DueDate     *time.Time      `json:"due_date" validate:"omitempty"`
	CreatedAt   *time.Time      `json:"created_at" validate:"omitempty"`
	UpdatedAt   *time.Time      `json:"updated_at" validate:"omitempty"`
}

// Validate validates the issue details.
func (i *Issue) Validate() error {
	if err := validate.Struct(i); err != nil {
		return errors.Join(ErrInvalidIssueDetails, err)
	}
	if err := i.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidIssueDetails, err)
	}
	if err := i.ReportedBy.Validate(); err != nil {
		return errors.Join(ErrInvalidIssueDetails, err)
	}
	if i.Parent != nil {
		if err := i.Parent.Validate(); err != nil {
			return errors.Join(ErrInvalidIssueDetails, err)
		}
	}
	for _, id := range i.Assignees {
		if err := id.Validate(); err != nil {
			return errors.Join(ErrInvalidIssueDetails, err)
		}
	}
	for _, id := range i.Labels {
		if err := id.Validate(); err != nil {
			return errors.Join(ErrInvalidIssueDetails, err)
		}
	}
	for _, id := range i.Comments {
		if err := id.Validate(); err != nil {
			return errors.Join(ErrInvalidIssueDetails, err)
		}
	}
	for _, id := range i.Attachments {
		if err := id.Validate(); err != nil {
			return errors.Join(ErrInvalidIssueDetails, err)
		}
	}
	for _, id := range i.Watchers {
		if err := id.Validate(); err != nil {
			return errors.Join(ErrInvalidIssueDetails, err)
		}
	}
	for _, id := range i.Relations {
		if err := id.Validate(); err != nil {
			return errors.Join(ErrInvalidIssueDetails, err)
		}
	}
	return nil
}

// NewIssueRelation creates a relation between issues.
func NewIssueRelation(source, target ID, kind IssueRelationKind) (*IssueRelation, error) {
	issueRelation := &IssueRelation{
		ID:     MustNewNilID(ResourceTypeIssueRelation),
		Source: source,
		Target: target,
		Kind:   kind,
	}

	if err := issueRelation.Validate(); err != nil {
		return nil, err
	}
	return issueRelation, nil
}

// NewIssue creates a new issue with the given details.
func NewIssue(numericID uint, title string, kind IssueKind, reportedBy ID) (*Issue, error) {
	issue := &Issue{
		ID:          MustNewNilID(ResourceTypeIssue),
		NumericID:   numericID,
		Kind:        kind,
		Title:       title,
		Status:      IssueStatusOpen,
		Priority:    IssuePriorityMedium,
		Resolution:  IssueResolutionNone,
		ReportedBy:  reportedBy,
		Assignees:   make([]ID, 0),
		Labels:      make([]ID, 0),
		Comments:    make([]ID, 0),
		Attachments: make([]ID, 0),
		Watchers:    make([]ID, 0),
		Relations:   make([]ID, 0),
		Links:       make([]string, 0),
	}

	if err := issue.Validate(); err != nil {
		return nil, err
	}

	return issue, nil
}
