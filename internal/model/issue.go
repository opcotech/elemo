package model

import (
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	IssueIDType = "Issue"
)

const (
	IssueKindEpic  IssueKind = iota + 1 // an epic
	IssueKindStory                      // a story
	IssueKindTask                       // a task
	IssueKindBug                        // a bug
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
)

var (
	ErrInvalidIssueKind         = errors.New("invalid issue kind")          // the issue kind is invalid
	ErrInvalidIssueStatus       = errors.New("invalid issue status")        // the issue status is invalid
	ErrInvalidIssueResolution   = errors.New("invalid issue resolution")    // the issue resolution is invalid
	ErrInvalidIssueRelationKind = errors.New("invalid issue relation kind") // the issue relation kind is invalid
	ErrInvalidIssuePriority     = errors.New("invalid issue priority")      // the issue priority is invalid
	ErrInvalidIssueDetails      = errors.New("invalid issue details")       // the issue details are invalid

	issueKindKeys = map[IssueKind]string{
		IssueKindEpic:  "epic",
		IssueKindStory: "story",
		IssueKindTask:  "task",
		IssueKindBug:   "bug",
	}
	issueKindValues = map[string]IssueKind{
		"epic":  IssueKindEpic,
		"story": IssueKindStory,
		"task":  IssueKindTask,
		"bug":   IssueKindBug,
	}

	issueStatusKeys = map[IssueStatus]string{
		IssueStatusOpen:       "open",
		IssueStatusInProgress: "in progress",
		IssueStatusBlocked:    "blocked",
		IssueStatusReview:     "review",
		IssueStatusDone:       "done",
		IssueStatusClosed:     "closed",
	}
	issueStatusValues = map[string]IssueStatus{
		"open":        IssueStatusOpen,
		"in progress": IssueStatusInProgress,
		"blocked":     IssueStatusBlocked,
		"review":      IssueStatusReview,
		"done":        IssueStatusDone,
		"closed":      IssueStatusClosed,
	}

	issueResolutionKeys = map[IssueResolution]string{
		IssueResolutionNone:            "none",
		IssueResolutionFixed:           "fixed",
		IssueResolutionDuplicate:       "duplicate",
		IssueResolutionWontFix:         "won't fix",
		IssueResolutionInvalid:         "invalid",
		IssueResolutionIncomplete:      "incomplete",
		IssueResolutionCannotReproduce: "cannot reproduce",
	}
	issueResolutionValues = map[string]IssueResolution{
		"none":             IssueResolutionNone,
		"fixed":            IssueResolutionFixed,
		"duplicate":        IssueResolutionDuplicate,
		"won't fix":        IssueResolutionWontFix,
		"invalid":          IssueResolutionInvalid,
		"incomplete":       IssueResolutionIncomplete,
		"cannot reproduce": IssueResolutionCannotReproduce,
	}

	issuePriorityKeys = map[IssuePriority]string{
		IssuePriorityLow:      "low",
		IssuePriorityMedium:   "medium",
		IssuePriorityHigh:     "high",
		IssuePriorityCritical: "critical",
	}
	issuePriorityValues = map[string]IssuePriority{
		"low":      IssuePriorityLow,
		"medium":   IssuePriorityMedium,
		"high":     IssuePriorityHigh,
		"critical": IssuePriorityCritical,
	}

	issueRelationKindKeys = map[IssueRelationKind]string{
		IssueRelationKindBlockedBy:    "blocked by",
		IssueRelationKindBlocks:       "blocks",
		IssueRelationKindDependsOn:    "depends on",
		IssueRelationKindDuplicatedBy: "duplicated by",
		IssueRelationKindDuplicates:   "duplicates",
		IssueRelationKindRelatedTo:    "related to",
	}
	issueRelationKindValues = map[string]IssueRelationKind{
		"blocked by":    IssueRelationKindBlockedBy,
		"blocks":        IssueRelationKindBlocks,
		"depends on":    IssueRelationKindDependsOn,
		"duplicated by": IssueRelationKindDuplicatedBy,
		"duplicates":    IssueRelationKindDuplicates,
		"related to":    IssueRelationKindRelatedTo,
	}
)

// IssueKind represents a kind of issue.
type IssueKind uint8

// String returns the string representation of IssueKind.
func (k IssueKind) String() string {
	return issueKindKeys[k]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (k IssueKind) MarshalText() (text []byte, err error) {
	if k < 1 || k > 4 {
		return nil, ErrInvalidIssueKind
	}
	return []byte(k.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (k *IssueKind) UnmarshalText(text []byte) error {
	if v, ok := issueKindValues[string(text)]; ok {
		*k = v
		return nil
	}
	return ErrInvalidIssueKind
}

// IssueStatus represents the status of an issue.
type IssueStatus uint8

// String returns the string representation of IssueStatus.
func (s IssueStatus) String() string {
	return issueStatusKeys[s]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (s IssueStatus) MarshalText() (text []byte, err error) {
	if s < 1 || s > 6 {
		return nil, ErrInvalidIssueStatus
	}
	return []byte(s.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (s *IssueStatus) UnmarshalText(text []byte) error {
	if v, ok := issueStatusValues[string(text)]; ok {
		*s = v
		return nil
	}
	return ErrInvalidIssueStatus
}

// IssueResolution represents the resolution of an issue.
type IssueResolution uint8

// String returns the string representation of IssueResolution.
func (r IssueResolution) String() string {
	return issueResolutionKeys[r]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (r IssueResolution) MarshalText() (text []byte, err error) {
	if r < 1 || r > 7 {
		return nil, ErrInvalidIssueResolution
	}
	return []byte(r.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (r *IssueResolution) UnmarshalText(text []byte) error {
	if v, ok := issueResolutionValues[string(text)]; ok {
		*r = v
		return nil
	}
	return ErrInvalidIssueResolution
}

// IssuePriority represents the priority of an issue.
type IssuePriority uint8

// String returns the string representation of IssuePriority.
func (p IssuePriority) String() string {
	return issuePriorityKeys[p]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (p IssuePriority) MarshalText() (text []byte, err error) {
	if p < 1 || p > 5 {
		return nil, ErrInvalidIssuePriority
	}
	return []byte(p.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (p *IssuePriority) UnmarshalText(text []byte) error {
	if v, ok := issuePriorityValues[string(text)]; ok {
		*p = v
		return nil
	}
	return ErrInvalidIssuePriority
}

// IssueRelationKind represents the kind of relation between two issues.
type IssueRelationKind uint8

// String returns the string representation of IssueKind.
func (r IssueRelationKind) String() string {
	return issueRelationKindKeys[r]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (r IssueRelationKind) MarshalText() (text []byte, err error) {
	if r < 1 || r > 6 {
		return nil, ErrInvalidIssueRelationKind
	}
	return []byte(r.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (r *IssueRelationKind) UnmarshalText(text []byte) error {
	if v, ok := issueRelationKindValues[string(text)]; ok {
		*r = v
		return nil
	}
	return ErrInvalidIssueRelationKind
}

// IssueRelation represents a relation between two issues.
type IssueRelation struct {
	ID        ID                `json:"id" validate:"required,dive"`
	Kind      IssueRelationKind `json:"kind" validate:"required,min=1,max=3"`
	CreatedAt *time.Time        `json:"created_at" validate:"omitempty"`
	UpdatedAt *time.Time        `json:"updated_at" validate:"omitempty"`
}

// Issue represents an issue in the system that can be assigned to a
// user and belong to a project or another Issue.
type Issue struct {
	ID          ID              `json:"id" validate:"required,dive"`
	NumericID   uint            `json:"numeric_id" validate:"required"`
	Parent      *ID             `json:"parent" validate:"omitempty,dive"`
	Kind        IssueKind       `json:"kind" validate:"required,min=1,max=4"`
	Title       string          `json:"title" validate:"required,min=3,max=120"`
	Description string          `json:"description" validate:"omitempty,min=10"`
	Status      IssueStatus     `json:"status" validate:"required,min=1,max=6"`
	Priority    IssuePriority   `json:"priority" validate:"required,min=1,max=5"`
	Resolution  IssueResolution `json:"resolution" validate:"required,min=1,max=7"`
	ReportedBy  ID              `json:"reported_by" validate:"required,dive"`
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

// NewIssue creates a new issue with the given details.
func NewIssue(numericID uint, title string, kind IssueKind, reportedBy ID) (*Issue, error) {
	issue := &Issue{
		ID:          MustNewNilID(IssueIDType),
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
