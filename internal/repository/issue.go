package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/metrics"
	"github.com/opcotech/elemo/internal/pkg/optional"
)

var (
	ErrIssueAddRelation    = errors.New("failed to add relation to issue")      // the relation could not be added to the issue
	ErrIssueAddWatcher     = errors.New("failed to add watcher to issue")       // the watcher could not be added to the issue
	ErrIssueCreate         = errors.New("failed to create issue")               // the issue could not be created
	ErrIssueDelete         = errors.New("failed to delete issue")               // the issue could not be deleted
	ErrIssueGetRelation    = errors.New("failed to get issue relation")         // the relation could not be retrieved
	ErrIssueGetRelations   = errors.New("failed to get relations for issue")    // the relations could not be retrieved for the issue
	ErrIssueGetWatchers    = errors.New("failed to get watchers for issue")     // the watchers could not be retrieved for the issue
	ErrIssueRead           = errors.New("failed to read issue")                 // the issue could not be retrieved
	ErrIssueRemoveRelation = errors.New("failed to remove relation from issue") // the relation could not be removed from the issue
	ErrIssueRemoveWatcher  = errors.New("failed to remove watcher from issue")  // the watcher could not be removed from the issue
	ErrIssueUpdate         = errors.New("failed to update issue")               // the issue could not be updated
)

// PartialAssignee is a lean assignment of a user to an issue.
type PartialAssignee struct {
	ID        model.ID             `json:"id"`
	Kind      model.AssignmentKind `json:"kind"`
	FirstName string               `json:"first_name"`
	LastName  string               `json:"last_name"`
	Picture   string               `json:"picture"`
}

// PartialIssue represents a simplified issue that can be used in lists.
type PartialIssue struct {
	ID          model.ID            `json:"id"`
	Key         string              `json:"key"`
	NumericID   uint                `json:"numeric_id"`
	Parent      *PartialIssue       `json:"parent"`
	Kind        model.IssueKind     `json:"kind"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Status      model.IssueStatus   `json:"status"`
	Priority    model.IssuePriority `json:"priority"`
	Assignments []PartialAssignee   `json:"assignments"`
	Labels      []PartialLabel      `json:"labels"`
	Project     *PartialProject     `json:"project"`
	Namespace   *PartialNamespace   `json:"namespace"`
	ReportedBy  *PartialUser        `json:"reported_by"`
	DueDate     *time.Time          `json:"due_date"`
	StartDate   *time.Time          `json:"start_date"`
	CreatedAt   *time.Time          `json:"created_at"`
	UpdatedAt   *time.Time          `json:"updated_at"`
}

// Issue represents an issue persisted by the repository.
type Issue struct {
	ID              model.ID              `json:"id"`
	Key             string                `json:"key"`
	NumericID       uint                  `json:"numeric_id"`
	Parent          *PartialIssue         `json:"parent"`
	Kind            model.IssueKind       `json:"kind"`
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	Status          model.IssueStatus     `json:"status"`
	Priority        model.IssuePriority   `json:"priority"`
	Resolution      model.IssueResolution `json:"resolution"`
	ReportedBy      *PartialUser          `json:"reported_by"`
	Assignments     []PartialAssignee     `json:"assignments"`
	Labels          []PartialLabel        `json:"labels"`
	Project         *PartialProject       `json:"project"`
	Namespace       *PartialNamespace     `json:"namespace"`
	CommentCount    *int64                `json:"comment_count"`
	DocumentCount   *int64                `json:"document_count"`
	AttachmentCount *int64                `json:"attachment_count"`
	WatcherCount    *int64                `json:"watcher_count"`
	RelationCount   *int64                `json:"relation_count"`
	Links           []model.IssueLink     `json:"links"`
	DueDate         *time.Time            `json:"due_date"`
	StartDate       *time.Time            `json:"start_date"`
	CreatedAt       *time.Time            `json:"created_at"`
	UpdatedAt       *time.Time            `json:"updated_at"`
}

// IssueRelation represents a relation between issues persisted by the repository.
type IssueRelation struct {
	ID        model.ID                `json:"id"`
	Source    model.ID                `json:"source"`
	Target    model.ID                `json:"target"`
	Kind      model.IssueRelationKind `json:"kind"`
	CreatedAt *time.Time              `json:"created_at"`
	UpdatedAt *time.Time              `json:"updated_at"`
}

// IssueRelationItem is a paginated relation row with the related issue loaded
// as a lean PartialIssue. Source and Target follow startNode/endNode.
type IssueRelationItem struct {
	ID        model.ID                `json:"id"`
	Kind      model.IssueRelationKind `json:"kind"`
	Source    model.ID                `json:"source"`
	Target    model.ID                `json:"target"`
	Related   *PartialIssue           `json:"related"`
	CreatedAt *time.Time              `json:"created_at"`
}

// CreateIssueOpts holds the data required to create an issue.
// NumericID is allocated atomically from the project's next_issue_id counter.
type CreateIssueOpts struct {
	ProjectID   model.ID
	Parent      *model.ID
	Kind        model.IssueKind
	Title       string
	Description string
	Status      model.IssueStatus
	Priority    model.IssuePriority
	Resolution  model.IssueResolution
	ReportedBy  model.ID
	Links       []model.IssueLink
	DueDate     *time.Time
	StartDate   *time.Time
}

// CreateIssueRelationOpts holds the data required to create an issue relation.
type CreateIssueRelationOpts struct {
	Source model.ID
	Target model.ID
	Kind   model.IssueRelationKind
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
	Parent      optional.Optional[model.ID]
}

// patch builds a Neo4j property map from defined optional fields.
func (o UpdateIssueOpts) patch() map[string]any {
	p := make(map[string]any)

	if o.Kind.Defined {
		p["kind"] = o.Kind.Value.String()
	}
	if o.Title.Defined {
		p["title"] = *o.Title.Value
	}
	if o.Description.Defined {
		if o.Description.Value == nil {
			p["description"] = nil
		} else {
			p["description"] = *o.Description.Value
		}
	}
	if o.Status.Defined {
		p["status"] = o.Status.Value.String()
	}
	if o.Priority.Defined {
		p["priority"] = o.Priority.Value.String()
	}
	if o.Resolution.Defined {
		p["resolution"] = o.Resolution.Value.String()
	}
	if o.Links.Defined {
		links := make([]model.IssueLink, 0)
		if o.Links.Value != nil {
			links = *o.Links.Value
		}
		p["links"] = encodeIssueLinks(links)
		p["link_labels"] = nil
	}
	if o.DueDate.Defined {
		if o.DueDate.Value == nil {
			p["due_date"] = nil
		} else {
			p["due_date"] = o.DueDate.Value.Format(time.RFC3339Nano)
		}
	}
	if o.StartDate.Defined {
		if o.StartDate.Value == nil {
			p["start_date"] = nil
		} else {
			p["start_date"] = o.StartDate.Value.Format(time.RFC3339Nano)
		}
	}

	return p
}

// issueLinkLabelSep separates an optional label from the URL in a stored link
// entry. Neo4j cannot persist MAP or LIST<LIST<STRING>> properties, so each
// pair is one STRING in a homogeneous LIST<STRING>: "url" or "url\tlabel".
const issueLinkLabelSep = "\t"

func encodeIssueLinks(links []model.IssueLink) []string {
	out := make([]string, len(links))
	for i, link := range links {
		if link.Label == "" || link.Label == link.URL {
			out[i] = link.URL
			continue
		}
		out[i] = link.URL + issueLinkLabelSep + link.Label
	}
	return out
}

func neo4jStringList(val any) []string {
	switch items := val.(type) {
	case []string:
		out := make([]string, len(items))
		copy(out, items)
		return out
	case []any:
		out := make([]string, 0, len(items))
		for _, item := range items {
			s, ok := item.(string)
			if !ok || s == "" {
				continue
			}
			out = append(out, s)
		}
		return out
	default:
		return make([]string, 0)
	}
}

func decodeIssueLinkEntry(entry string) model.IssueLink {
	url, label, found := strings.Cut(entry, issueLinkLabelSep)
	if !found || label == "" {
		return model.IssueLink{URL: url, Label: url}
	}
	return model.IssueLink{URL: url, Label: label}
}

func decodeIssueLinks(props map[string]any) []model.IssueLink {
	entries := neo4jStringList(props["links"])
	legacyLabels := neo4jStringList(props["link_labels"])
	links := make([]model.IssueLink, 0, len(entries))
	for i, entry := range entries {
		link := decodeIssueLinkEntry(entry)
		if strings.Contains(entry, issueLinkLabelSep) {
			links = append(links, link)
			continue
		}
		if i < len(legacyLabels) && legacyLabels[i] != "" {
			link.Label = legacyLabels[i]
		}
		links = append(links, link)
	}
	return links
}

func applyDecodedIssueLinks(issue *Issue, props map[string]any) {
	issue.Links = decodeIssueLinks(props)
}

// issueAssignmentsPattern returns a Cypher pattern comprehension that collects
// AssignedTo edges (both assignee and reviewer) as user maps.
func issueAssignmentsPattern(issueVar string) string {
	return `[(` + issueVar + `)<-[at:` + EdgeKindAssignedTo.String() +
		`]-(u:` + model.ResourceTypeUser.String() +
		`) | {id: u.id, kind: at.kind, first_name: u.first_name, last_name: u.last_name, picture: u.picture}]`
}

// parsePartialAssignees parses [{id, kind}, ...] values from Neo4j into PartialAssignee.
func parsePartialAssignees(val any) ([]PartialAssignee, error) {
	items, ok := val.([]any)
	if !ok {
		return make([]PartialAssignee, 0), nil
	}

	assignees := make([]PartialAssignee, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		m, ok := item.(map[string]any)
		if !ok {
			return nil, ErrMalformedResult
		}

		idStr, ok := m["id"].(string)
		if !ok {
			return nil, ErrMalformedResult
		}
		kindStr, ok := m["kind"].(string)
		if !ok {
			return nil, ErrMalformedResult
		}

		id, err := model.NewIDFromString(idStr, model.ResourceTypeUser.String())
		if err != nil {
			return nil, err
		}

		var kind model.AssignmentKind
		if err := kind.UnmarshalText([]byte(kindStr)); err != nil {
			return nil, err
		}

		assignees = append(assignees, PartialAssignee{
			ID:        id,
			Kind:      kind,
			FirstName: mapString(m, "first_name"),
			LastName:  mapString(m, "last_name"),
			Picture:   mapString(m, "picture"),
		})
	}

	return assignees, nil
}

// Neo4jParsePartialAssigneesFromRecord parses PartialAssignee values from a record key.
func Neo4jParsePartialAssigneesFromRecord(record *neo4j.Record, key string) ([]PartialAssignee, error) {
	val, err := Neo4jParseValueFromRecord[[]any](record, key)
	if err != nil {
		return nil, err
	}
	return parsePartialAssignees(val)
}

//go:generate go tool mockgen -source=issue.go -destination=mock/mock_issue_gen.go -package=mockrepo
type IssueRepository interface {
	Create(ctx context.Context, opts CreateIssueOpts) (*Issue, error)
	Get(ctx context.Context, id model.ID, proj IssueProjection) (*Issue, error)
	GetByKey(ctx context.Context, namespaceID model.ID, key string, proj IssueProjection) (*Issue, error)
	ListForProject(ctx context.Context, query IssueListQuery) (Page[*PartialIssue], error)
	ListForNamespace(ctx context.Context, query IssueListForNamespaceQuery) (Page[*PartialIssue], error)
	ListForUser(ctx context.Context, query IssueListForUserQuery) (Page[*PartialIssue], error)
	ListForIssue(ctx context.Context, query IssueListForIssueQuery) (Page[*Issue], error)
	AddWatcher(ctx context.Context, issue model.ID, user model.ID) error
	GetWatchers(ctx context.Context, issue model.ID) ([]*User, error)
	RemoveWatcher(ctx context.Context, issue model.ID, user model.ID) error
	AddRelation(ctx context.Context, opts CreateIssueRelationOpts) (*IssueRelation, error)
	GetRelation(ctx context.Context, relationID model.ID) (*IssueRelation, error)
	GetRelations(ctx context.Context, issue model.ID) ([]*IssueRelation, error)
	ListRelations(ctx context.Context, query IssueRelationListQuery) (Page[*IssueRelationItem], error)
	RemoveRelation(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) error
	RemoveRelationByID(ctx context.Context, relationID model.ID) error
	Update(ctx context.Context, id model.ID, opts UpdateIssueOpts, proj IssueProjection) (*Issue, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jIssueRepository is a repository for managing user issues.
type Neo4jIssueRepository struct {
	*neo4jBaseRepository
}

type issueListRow struct {
	projectKey string
	issue      *PartialIssue
}

type issueDetailRow struct {
	projectKey string
	issue      *Issue
}

func applyIssueKeys(issue *Issue, projectKey string) {
	issue.Key = model.FormatIssueKey(projectKey, issue.NumericID)
	if issue.Parent != nil {
		issue.Parent.Key = model.FormatIssueKey(projectKey, issue.Parent.NumericID)
	}
}

func decodePartialIssueNode(node neo4j.Node, projectKey string) (*PartialIssue, error) {
	issueID, err := Neo4jDecodeID(node, model.ResourceTypeIssue)
	if err != nil {
		return nil, err
	}

	var tempIssue struct {
		NumericID   uint       `json:"numeric_id"`
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Kind        string     `json:"kind"`
		Status      string     `json:"status"`
		Priority    string     `json:"priority"`
		DueDate     *time.Time `json:"due_date"`
		StartDate   *time.Time `json:"start_date"`
		CreatedAt   *time.Time `json:"created_at"`
		UpdatedAt   *time.Time `json:"updated_at"`
	}
	if err := Neo4jScanIntoStruct(&node, &tempIssue, []string{"id"}); err != nil {
		return nil, err
	}

	var kind model.IssueKind
	if err := kind.UnmarshalText([]byte(tempIssue.Kind)); err != nil {
		return nil, err
	}

	var status model.IssueStatus
	if err := status.UnmarshalText([]byte(tempIssue.Status)); err != nil {
		return nil, err
	}

	var priority model.IssuePriority
	if err := priority.UnmarshalText([]byte(tempIssue.Priority)); err != nil {
		return nil, err
	}

	return &PartialIssue{
		ID:          issueID,
		Key:         model.FormatIssueKey(projectKey, tempIssue.NumericID),
		NumericID:   tempIssue.NumericID,
		Kind:        kind,
		Title:       tempIssue.Title,
		Description: tempIssue.Description,
		Status:      status,
		Priority:    priority,
		Assignments: make([]PartialAssignee, 0),
		Labels:      make([]PartialLabel, 0),
		DueDate:     tempIssue.DueDate,
		StartDate:   tempIssue.StartDate,
		CreatedAt:   tempIssue.CreatedAt,
		UpdatedAt:   tempIssue.UpdatedAt,
	}, nil
}

func parentProjectKeyFromRecord(record *neo4j.Record, fallback string) string {
	val, ok := record.Get("parent_project_key")
	if !ok || val == nil {
		return fallback
	}
	key, ok := val.(string)
	if !ok || key == "" {
		return fallback
	}
	return key
}

func relationCountAfterCreate(hasParent bool) *int64 {
	if hasParent {
		return convert.ToPointer(int64(1))
	}
	return convert.ToPointer(int64(0))
}

func (r *Neo4jIssueRepository) scanDetail(proj IssueProjection) func(rec *neo4j.Record) (*issueDetailRow, error) {
	return func(rec *neo4j.Record) (*issueDetailRow, error) {
		node, err := Neo4jRecordNode(rec, "i")
		if err != nil {
			return nil, err
		}
		project, err := Neo4jRecordPartialProject(rec, "p")
		if err != nil {
			return nil, err
		}
		namespace, err := Neo4jRecordOptionalPartialNamespace(rec, "n")
		if err != nil {
			return nil, err
		}
		reportedBy, err := Neo4jRecordPartialUser(rec, "u")
		if err != nil {
			return nil, err
		}

		issue := new(Issue)
		if err := Neo4jScanIntoStruct(&node, &issue, []string{"id", "key", "parent", "reported_by", "links", "link_labels"}); err != nil {
			return nil, err
		}
		applyDecodedIssueLinks(issue, node.GetProperties())
		issue.ID, err = Neo4jDecodeID(node, model.ResourceTypeIssue)
		if err != nil {
			return nil, err
		}
		issue.ReportedBy = reportedBy
		issue.Project = project
		issue.Namespace = namespace
		applyIssueKeys(issue, project.Key)
		if proj.Assignments {
			issue.Assignments = make([]PartialAssignee, 0)
		}
		if proj.Labels {
			issue.Labels = make([]PartialLabel, 0)
		}
		if proj.CommentCount {
			issue.CommentCount = convert.ToPointer(int64(0))
		}
		if proj.DocumentCount {
			issue.DocumentCount = convert.ToPointer(int64(0))
		}
		if proj.AttachmentCount {
			issue.AttachmentCount = convert.ToPointer(int64(0))
		}
		if proj.WatcherCount {
			issue.WatcherCount = convert.ToPointer(int64(0))
		}
		if proj.RelationCount {
			issue.RelationCount = convert.ToPointer(int64(0))
		}

		return &issueDetailRow{projectKey: project.Key, issue: issue}, nil
	}
}

func (r *Neo4jIssueRepository) applyIssueLoaders(ctx context.Context, tx neo4j.ManagedTransaction, plan QueryPlan, rows []*issueDetailRow) error {
	if len(plan.Loaders) == 0 || len(rows) == 0 {
		return nil
	}

	rowByID := make(map[string]*issueDetailRow, len(rows))
	targets := make(map[string]issueRelationTarget, len(rows))
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.issue == nil {
			continue
		}
		id := row.issue.ID.String()
		rowByID[id] = row
		targets[id] = issueRelationTarget{
			projectKey:  row.projectKey,
			parent:      &row.issue.Parent,
			assignments: &row.issue.Assignments,
			labels:      &row.issue.Labels,
		}
		ids = append(ids, id)
	}

	for _, loader := range plan.Loaders {
		query := loaderQueryWithIDs(loader, ids)
		handled, err := applyIssueRelationLoader(ctx, tx, query, loader.Name, targets)
		if err != nil {
			return err
		}
		if handled {
			continue
		}
		switch loader.Name {
		case "issue.load_comment_count":
			if err := applyIssueCountLoader(ctx, tx, query, rowByID, "comment_count", func(issue *Issue, count int64) {
				issue.CommentCount = convert.ToPointer(count)
			}); err != nil {
				return err
			}
		case "issue.load_document_count":
			if err := applyIssueCountLoader(ctx, tx, query, rowByID, "document_count", func(issue *Issue, count int64) {
				issue.DocumentCount = convert.ToPointer(count)
			}); err != nil {
				return err
			}
		case "issue.load_attachment_count":
			if err := applyIssueCountLoader(ctx, tx, query, rowByID, "attachment_count", func(issue *Issue, count int64) {
				issue.AttachmentCount = convert.ToPointer(count)
			}); err != nil {
				return err
			}
		case "issue.load_watcher_count":
			if err := applyIssueCountLoader(ctx, tx, query, rowByID, "watcher_count", func(issue *Issue, count int64) {
				issue.WatcherCount = convert.ToPointer(count)
			}); err != nil {
				return err
			}
		case "issue.load_relation_count":
			if err := applyIssueCountLoader(ctx, tx, query, rowByID, "relation_count", func(issue *Issue, count int64) {
				issue.RelationCount = convert.ToPointer(count)
			}); err != nil {
				return err
			}
		default:
			return ErrQueryCompile
		}
	}

	return nil
}

type issueRelationTarget struct {
	projectKey  string
	parent      **PartialIssue
	assignments *[]PartialAssignee
	labels      *[]PartialLabel
}

func loaderQueryWithIDs(loader CompiledQuery, ids []string) CompiledQuery {
	query := loader
	query.Params = cloneParams(loader.Params)
	query.Params["ids"] = ids
	return query
}

func applyIssueRelationLoader(
	ctx context.Context,
	tx neo4j.ManagedTransaction,
	query CompiledQuery,
	name string,
	targets map[string]issueRelationTarget,
) (bool, error) {
	switch name {
	case "issue.load_parent":
		return true, applyIssueParentLoader(ctx, tx, query, targets)
	case "issue.load_assignments":
		return true, applyIssueAssignmentsLoader(ctx, tx, query, targets)
	case "issue.load_labels":
		return true, applyIssueLabelsLoader(ctx, tx, query, targets)
	default:
		return false, nil
	}
}

func applyIssueParentLoader(ctx context.Context, tx neo4j.ManagedTransaction, query CompiledQuery, targets map[string]issueRelationTarget) error {
	parentRows, _, err := Neo4jRunQuery(ctx, tx, query, func(rec *neo4j.Record) (struct {
		IssueID string
		Parent  *PartialIssue
	}, error,
	) {
		issueID, err := Neo4jParseValueFromRecord[string](rec, "issue_id")
		if err != nil {
			return struct {
				IssueID string
				Parent  *PartialIssue
			}{}, err
		}

		parentVal, ok := rec.AsMap()["parent"]
		if !ok || parentVal == nil {
			return struct {
				IssueID string
				Parent  *PartialIssue
			}{IssueID: issueID}, nil
		}

		parentNode, ok := parentVal.(neo4j.Node)
		if !ok {
			return struct {
				IssueID string
				Parent  *PartialIssue
			}{}, ErrMalformedResult
		}
		target, exists := targets[issueID]
		if !exists {
			return struct {
				IssueID string
				Parent  *PartialIssue
			}{}, ErrMalformedResult
		}
		parent, err := decodePartialIssueNode(parentNode, parentProjectKeyFromRecord(rec, target.projectKey))
		if err != nil {
			return struct {
				IssueID string
				Parent  *PartialIssue
			}{}, err
		}
		return struct {
			IssueID string
			Parent  *PartialIssue
		}{IssueID: issueID, Parent: parent}, nil
	})
	if err != nil {
		return err
	}
	for _, parentRow := range parentRows {
		target, ok := targets[parentRow.IssueID]
		if !ok || target.parent == nil {
			continue
		}
		*target.parent = parentRow.Parent
	}
	return nil
}

func applyIssueAssignmentsLoader(ctx context.Context, tx neo4j.ManagedTransaction, query CompiledQuery, targets map[string]issueRelationTarget) error {
	assignmentRows, _, err := Neo4jRunQuery(ctx, tx, query, func(rec *neo4j.Record) (struct {
		IssueID     string
		Assignments []PartialAssignee
	}, error,
	) {
		issueID, err := Neo4jParseValueFromRecord[string](rec, "issue_id")
		if err != nil {
			return struct {
				IssueID     string
				Assignments []PartialAssignee
			}{}, err
		}
		assignments, err := Neo4jParsePartialAssigneesFromRecord(rec, "assignments")
		if err != nil {
			return struct {
				IssueID     string
				Assignments []PartialAssignee
			}{}, err
		}
		return struct {
			IssueID     string
			Assignments []PartialAssignee
		}{IssueID: issueID, Assignments: assignments}, nil
	})
	if err != nil {
		return err
	}
	for _, assignmentRow := range assignmentRows {
		target, ok := targets[assignmentRow.IssueID]
		if !ok || target.assignments == nil {
			continue
		}
		*target.assignments = assignmentRow.Assignments
	}
	return nil
}

func applyIssueLabelsLoader(ctx context.Context, tx neo4j.ManagedTransaction, query CompiledQuery, targets map[string]issueRelationTarget) error {
	rows, _, err := Neo4jRunQuery(ctx, tx, query, func(rec *neo4j.Record) (struct {
		IssueID string
		Labels  []PartialLabel
	}, error,
	) {
		issueID, err := Neo4jParseValueFromRecord[string](rec, "issue_id")
		if err != nil {
			return struct {
				IssueID string
				Labels  []PartialLabel
			}{}, err
		}
		labels, err := Neo4jRecordPartialLabels(rec, "labels")
		if err != nil {
			return struct {
				IssueID string
				Labels  []PartialLabel
			}{}, err
		}
		return struct {
			IssueID string
			Labels  []PartialLabel
		}{IssueID: issueID, Labels: labels}, nil
	})
	if err != nil {
		return err
	}
	for _, row := range rows {
		target, ok := targets[row.IssueID]
		if !ok || target.labels == nil {
			continue
		}
		*target.labels = row.Labels
	}
	return nil
}

func applyIssueCountLoader(ctx context.Context, tx neo4j.ManagedTransaction, query CompiledQuery, rowByID map[string]*issueDetailRow, field string, assign func(issue *Issue, count int64)) error {
	rows, _, err := Neo4jRunQuery(ctx, tx, query, func(rec *neo4j.Record) (struct {
		IssueID string
		Count   int64
	}, error,
	) {
		issueID, err := Neo4jParseValueFromRecord[string](rec, "issue_id")
		if err != nil {
			return struct {
				IssueID string
				Count   int64
			}{}, err
		}
		count, err := Neo4jParseValueFromRecord[int64](rec, field)
		if err != nil {
			return struct {
				IssueID string
				Count   int64
			}{}, err
		}
		return struct {
			IssueID string
			Count   int64
		}{IssueID: issueID, Count: count}, nil
	})
	if err != nil {
		return err
	}
	for _, row := range rows {
		if issueRow := rowByID[row.IssueID]; issueRow != nil && issueRow.issue != nil {
			assign(issueRow.issue, row.Count)
		}
	}
	return nil
}

func (r *Neo4jIssueRepository) scanRelation(ip, rp, tp string) func(rec *neo4j.Record) (*IssueRelation, error) {
	return func(rec *neo4j.Record) (*IssueRelation, error) {
		rel := new(IssueRelation)

		val, _, err := neo4j.GetRecordValue[neo4j.Relationship](rec, rp)
		if err != nil {
			return nil, err
		}

		source, err := Neo4jParseValueFromRecord[string](rec, ip)
		if err != nil {
			return nil, err
		}

		target, err := Neo4jParseValueFromRecord[string](rec, tp)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &rel, []string{"id", "source", "target"}); err != nil {
			return nil, err
		}

		rel.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeIssueRelation.String())
		rel.Source, _ = model.NewIDFromString(source, model.ResourceTypeIssue.String())
		rel.Target, _ = model.NewIDFromString(target, model.ResourceTypeIssue.String())

		return rel, nil
	}
}

func (r *Neo4jIssueRepository) scanRelationItem() func(rec *neo4j.Record) (*IssueRelationItem, error) {
	return func(rec *neo4j.Record) (*IssueRelationItem, error) {
		rel, err := r.scanRelation("source_id", "r", "target_id")(rec)
		if err != nil {
			return nil, err
		}

		relatedNode, err := Neo4jRecordNode(rec, "n")
		if err != nil {
			return nil, err
		}
		project, err := Neo4jRecordPartialProject(rec, "p")
		if err != nil {
			return nil, err
		}
		namespace, err := Neo4jRecordOptionalPartialNamespace(rec, "ns")
		if err != nil {
			return nil, err
		}

		related, err := decodePartialIssueNode(relatedNode, project.Key)
		if err != nil {
			return nil, err
		}
		related.Project = project
		related.Namespace = namespace

		return &IssueRelationItem{
			ID:        rel.ID,
			Kind:      rel.Kind,
			Source:    rel.Source,
			Target:    rel.Target,
			Related:   related,
			CreatedAt: rel.CreatedAt,
		}, nil
	}
}

func (r *Neo4jIssueRepository) Create(ctx context.Context, opts CreateIssueOpts) (*Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeIssue)

	links := opts.Links
	if links == nil {
		links = make([]model.IssueLink, 0)
	}

	cypher := `
	MATCH (p:` + opts.ProjectID.Label() + ` {id: $project_id})
	MATCH (u:` + opts.ReportedBy.Label() + ` {id: $reported_by_id})
	SET p.next_issue_id = coalesce(p.next_issue_id, 0) + 1
	WITH p, u, p.next_issue_id AS numeric_id
	CREATE
		(i:` + id.Label() + ` {
			id: $id, numeric_id: numeric_id, kind: $kind, title: $title, description: $description, status: $status,
			priority: $priority, resolution: $resolution, links: $links, due_date: datetime($due_date),
			start_date: datetime($start_date), created_at: datetime($created_at)
		}),
		(u)-[:` + EdgeKindCreated.String() + ` {id: $created_rel_id, created_at: datetime($created_at)}]->(i),
		(u)-[:` + EdgeKindWatches.String() + ` {id: $watches_rel_id, created_at: datetime($created_at)}]->(i),
		(i)-[:` + EdgeKindInScopeOf.String() + ` {id: $scope_id, created_at: datetime($created_at)}]->(p),
		(i)-[:` + EdgeKindBelongsTo.String() + ` {id: $belongs_to_rel_id, created_at: datetime($created_at)}]->(p)`

	params := map[string]any{
		"project_id":        opts.ProjectID.String(),
		"reported_by_id":    opts.ReportedBy.String(),
		"id":                id.String(),
		"kind":              opts.Kind.String(),
		"title":             opts.Title,
		"description":       opts.Description,
		"status":            opts.Status.String(),
		"priority":          opts.Priority.String(),
		"resolution":        opts.Resolution.String(),
		"links":             encodeIssueLinks(links),
		"due_date":          nil,
		"start_date":        nil,
		"created_at":        createdAt.Format(time.RFC3339Nano),
		"created_rel_id":    model.NewRawID(),
		"watches_rel_id":    model.NewRawID(),
		"scope_id":          model.NewRawID(),
		"belongs_to_rel_id": model.NewRawID(),
	}

	if opts.DueDate != nil {
		params["due_date"] = opts.DueDate.Format(time.RFC3339Nano)
	}
	if opts.StartDate != nil {
		params["start_date"] = opts.StartDate.Format(time.RFC3339Nano)
	}

	if opts.Parent != nil {
		cypher += `
		WITH i
		MATCH (parent:` + opts.Parent.Label() + ` {id: $parent_id})
		CREATE (i)-[:` + EdgeKindRelatedTo.String() + ` {id: $issue_rel_id, kind: $rel_kind, created_at: datetime($created_at)}]->(parent)`

		params["parent_id"] = opts.Parent.String()
		params["issue_rel_id"] = model.NewRawID()
		params["rel_kind"] = model.IssueRelationKindSubtaskOf.String()
	}

	cypher += `
	RETURN i.id AS id`

	if _, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	}); err != nil {
		return nil, errors.Join(ErrIssueCreate, err)
	}

	return r.Get(ctx, id, IssueDetailProjection())
}

func (r *Neo4jIssueRepository) Get(ctx context.Context, id model.ID, proj IssueProjection) (*Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Read")
	defer span.End()

	plan, err := CompileQuery(IssueGetQuery{ID: id, Projection: proj})
	if err != nil {
		return nil, errors.Join(ErrIssueRead, err)
	}

	var issue *Issue
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		row, _, readErr := Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scanDetail(proj))
		if readErr != nil {
			return readErr
		}
		if err := r.applyIssueLoaders(ctx, tx, plan, []*issueDetailRow{row}); err != nil {
			return err
		}
		issue = row.issue
		return nil
	})
	if err != nil {
		return nil, errors.Join(ErrIssueRead, err)
	}

	return issue, nil
}

func (r *Neo4jIssueRepository) GetByKey(ctx context.Context, namespaceID model.ID, key string, proj IssueProjection) (*Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/GetByKey")
	defer span.End()

	if _, _, err := model.ParseIssueKey(key); err != nil {
		return nil, errors.Join(ErrIssueRead, err)
	}

	plan, err := CompileQuery(IssueGetByKeyQuery{
		NamespaceID: namespaceID,
		IssueKey:    key,
		Projection:  proj,
	})
	if err != nil {
		return nil, errors.Join(ErrIssueRead, err)
	}

	var issue *Issue
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		row, _, readErr := Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scanDetail(proj))
		if readErr != nil {
			return readErr
		}
		if err := r.applyIssueLoaders(ctx, tx, plan, []*issueDetailRow{row}); err != nil {
			return err
		}
		issue = row.issue
		return nil
	})
	if err != nil {
		return nil, errors.Join(ErrIssueRead, err)
	}

	return issue, nil
}

func (r *Neo4jIssueRepository) ListForProject(ctx context.Context, query IssueListQuery) (Page[*PartialIssue], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/ListForProject")
	defer span.End()

	page, err := query.Page.Normalize()
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueRead, err)
	}
	query.Page = page
	sort := IssueListSort{Field: query.SortField, Direction: query.Order}.normalize()
	filter := normalizeIssueListFilter(query.Filter)
	query.SortField = sort.Field
	query.Order = sort.Direction
	query.Filter = filter

	plan, err := CompileQuery(query)
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueRead, err)
	}

	return r.executePartialIssueList(ctx, plan, page, sort, issueListCursorHash(query.ProjectID, filter, sort))
}

func (r *Neo4jIssueRepository) ListForNamespace(ctx context.Context, query IssueListForNamespaceQuery) (Page[*PartialIssue], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/ListForNamespace")
	defer span.End()

	page, err := query.Page.Normalize()
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueRead, err)
	}
	query.Page = page
	sort := IssueListSort{Field: query.SortField, Direction: query.Order}.normalize()
	filter := normalizeIssueListFilter(query.Filter)
	query.SortField = sort.Field
	query.Order = sort.Direction
	query.Filter = filter

	plan, err := CompileQuery(query)
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueRead, err)
	}

	return r.executePartialIssueList(ctx, plan, page, sort, issueListCursorHash(query.NamespaceID, filter, sort))
}

func (r *Neo4jIssueRepository) ListForUser(ctx context.Context, query IssueListForUserQuery) (Page[*PartialIssue], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/ListForUser")
	defer span.End()

	page, err := query.Page.Normalize()
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueRead, err)
	}
	query.Page = page
	sort := IssueListSort{Field: query.SortField, Direction: query.Order}.normalize()
	filter := normalizeIssueListFilter(query.Filter)
	query.SortField = sort.Field
	query.Order = sort.Direction
	query.Filter = filter

	plan, err := CompileQuery(query)
	if err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueRead, err)
	}

	return r.executePartialIssueList(ctx, plan, page, sort, issueListCursorHash(query.UserID, filter, sort))
}

func (r *Neo4jIssueRepository) executePartialIssueList(ctx context.Context, plan QueryPlan, page CursorPage, sort IssueListSort, cursorHash string) (Page[*PartialIssue], error) {
	rows := make([]*issueListRow, 0)
	if err := Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		rootRows, _, err := Neo4jRunQuery(ctx, tx, plan.Root, func(record *neo4j.Record) (*issueListRow, error) {
			issueNode, err := Neo4jRecordNode(record, "i")
			if err != nil {
				return nil, err
			}
			project, err := Neo4jRecordPartialProject(record, "p")
			if err != nil {
				return nil, err
			}
			namespace, err := Neo4jRecordOptionalPartialNamespace(record, "n")
			if err != nil {
				return nil, err
			}
			issue, err := decodePartialIssueNode(issueNode, project.Key)
			if err != nil {
				return nil, err
			}
			issue.ReportedBy, err = Neo4jRecordPartialUser(record, "u")
			if err != nil {
				return nil, err
			}
			issue.Project = project
			issue.Namespace = namespace
			return &issueListRow{
				projectKey: project.Key,
				issue:      issue,
			}, nil
		})
		if err != nil {
			return err
		}

		rows = rootRows
		if len(rows) == 0 || len(plan.Loaders) == 0 {
			return nil
		}

		ids := make([]string, 0, len(rows))
		targets := make(map[string]issueRelationTarget, len(rows))
		for _, row := range rows {
			if row == nil || row.issue == nil {
				continue
			}
			id := row.issue.ID.String()
			ids = append(ids, id)
			targets[id] = issueRelationTarget{
				projectKey:  row.projectKey,
				parent:      &row.issue.Parent,
				assignments: &row.issue.Assignments,
				labels:      &row.issue.Labels,
			}
		}

		for _, loader := range plan.Loaders {
			query := loaderQueryWithIDs(loader, ids)
			handled, err := applyIssueRelationLoader(ctx, tx, query, loader.Name, targets)
			if err != nil {
				return err
			}
			if !handled {
				return ErrQueryCompile
			}
		}

		return nil
	}); err != nil {
		return Page[*PartialIssue]{}, errors.Join(ErrIssueRead, err)
	}

	hasMore := len(rows) > page.Size
	if hasMore {
		rows = rows[:page.Size]
	}

	items := make([]*PartialIssue, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.issue)
	}

	var nextPageToken *string
	if hasMore && len(rows) > 0 {
		token, err := encodeIssueListCursor(rows[len(rows)-1].issue, sort, cursorHash)
		if err != nil {
			return Page[*PartialIssue]{}, errors.Join(ErrIssueRead, err)
		}
		nextPageToken = &token
	}

	return Page[*PartialIssue]{
		Items: items,
		PageInfo: PageInfo{
			HasMore:       hasMore,
			NextPageToken: nextPageToken,
		},
	}, nil
}

func (r *Neo4jIssueRepository) ListForIssue(ctx context.Context, query IssueListForIssueQuery) (Page[*Issue], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/ListForIssue")
	defer span.End()

	normalized, err := query.Page.Normalize()
	if err != nil {
		return Page[*Issue]{}, errors.Join(ErrIssueRead, err)
	}
	query.Page = normalized
	if query.Order == SortDirectionUnknown {
		query.Order = SortDirectionDesc
	}

	plan, err := CompileQuery(query)
	if err != nil {
		return Page[*Issue]{}, errors.Join(ErrIssueRead, err)
	}

	rows := make([]*issueDetailRow, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		rows, _, readErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scanDetail(query.Projection))
		if readErr != nil {
			return readErr
		}
		return r.applyIssueLoaders(ctx, tx, plan, rows)
	})
	if err != nil {
		return Page[*Issue]{}, errors.Join(ErrIssueRead, err)
	}

	pagedRows, err := PaginateSlice(rows, normalized.Size, func(row *issueDetailRow) model.ID {
		return row.issue.ID
	})
	if err != nil {
		return Page[*Issue]{}, errors.Join(ErrIssueRead, err)
	}

	items := make([]*Issue, 0, len(pagedRows.Items))
	for _, row := range pagedRows.Items {
		items = append(items, row.issue)
	}

	return Page[*Issue]{Items: items, PageInfo: pagedRows.PageInfo}, nil
}

func (r *Neo4jIssueRepository) AddWatcher(ctx context.Context, issue model.ID, user model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/AddWatcher")
	defer span.End()

	cypher := `
	MATCH (i:` + issue.Label() + `)
	WHERE i.id = $issue_id
	WITH i
	MATCH (u:` + user.Label() + `)
	WHERE u.id = $user_id
	CREATE (u)-[:` + EdgeKindWatches.String() + ` {id: $rel_id, created_at: datetime($created_at)}]->(i)`

	params := map[string]any{
		"issue_id":   issue.String(),
		"user_id":    user.String(),
		"rel_id":     model.NewRawID(),
		"created_at": time.Now().UTC().Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrIssueAddWatcher, err)
	}

	return nil
}

func (r *Neo4jIssueRepository) GetWatchers(ctx context.Context, issue model.ID) ([]*User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/GetWatchers")
	defer span.End()

	root, err := IssueWatchersQuery(issue)
	if err != nil {
		return nil, errors.Join(ErrIssueGetWatchers, err)
	}

	proj := UserProjection{}
	plan, err := compileUserRootQuery(userRootQueryInput{
		Name:       root.Name,
		Match:      root.Cypher,
		Params:     root.Params,
		Projection: proj,
	})
	if err != nil {
		return nil, errors.Join(ErrIssueGetWatchers, err)
	}

	userRepo := new(Neo4jUserRepository)
	users := make([]*User, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		users, _, readErr = Neo4jRunQuery(ctx, tx, plan.Root, userRepo.scan("u", proj))
		if readErr != nil {
			return readErr
		}
		return userRepo.applyUserLoaders(ctx, tx, plan, users)
	})
	if err != nil {
		return nil, errors.Join(ErrIssueGetWatchers, err)
	}

	return users, nil
}

func (r *Neo4jIssueRepository) RemoveWatcher(ctx context.Context, issue model.ID, user model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/RemoveWatcher")
	defer span.End()

	cypher := `
	MATCH (:` + issue.Label() + ` {id: $issue_id})<-[r:` + EdgeKindWatches.String() + `]-(:` + user.Label() + ` {id: $user_id})
	DELETE r`

	params := map[string]any{
		"issue_id": issue.String(),
		"user_id":  user.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrIssueRemoveWatcher, err)
	}

	return nil
}

func (r *Neo4jIssueRepository) AddRelation(ctx context.Context, opts CreateIssueRelationOpts) (*IssueRelation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/AddRelation")
	defer span.End()

	createdAt := convert.ToPointer(time.Now().UTC())

	relation := &IssueRelation{
		ID:        model.MustNewID(model.ResourceTypeIssueRelation),
		Source:    opts.Source,
		Target:    opts.Target,
		Kind:      opts.Kind,
		CreatedAt: createdAt,
		UpdatedAt: nil,
	}

	cypher := `
	MATCH (s:` + relation.Source.Label() + `)
	WHERE s.id = $source_id
	WITH s
	MATCH (t:` + relation.Target.Label() + `)
	WHERE t.id = $target_id
	MERGE (s)-[r:` + EdgeKindRelatedTo.String() + ` {kind: $kind}]->(t)
	ON CREATE SET r.id = $id, r.created_at = datetime($created_at)
	`

	params := map[string]any{
		"source_id":  relation.Source.String(),
		"target_id":  relation.Target.String(),
		"id":         relation.ID.String(),
		"kind":       relation.Kind.String(),
		"created_at": createdAt.Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrIssueAddRelation, err)
	}

	return relation, nil
}

func (r *Neo4jIssueRepository) GetRelations(ctx context.Context, issue model.ID) ([]*IssueRelation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/GetRelations")
	defer span.End()

	query, err := IssueRelationsQuery(issue)
	if err != nil {
		return nil, errors.Join(ErrIssueGetRelations, err)
	}

	var relations []*IssueRelation
	err = Neo4jExecuteReadPlan(ctx, r.db, QueryPlan{Root: query}, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		relations, _, readErr = Neo4jRunQuery(ctx, tx, query, r.scanRelation("issue_id", "r", "related_issue_id"))
		return readErr
	})
	if err != nil {
		return nil, errors.Join(ErrIssueGetRelations, err)
	}

	return relations, nil
}

func (r *Neo4jIssueRepository) GetRelation(ctx context.Context, relationID model.ID) (*IssueRelation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/GetRelation")
	defer span.End()

	query, err := IssueRelationByIDQuery(relationID)
	if err != nil {
		return nil, errors.Join(ErrIssueGetRelation, err)
	}

	var relation *IssueRelation
	err = Neo4jExecuteReadPlan(ctx, r.db, QueryPlan{Root: query}, func(tx neo4j.ManagedTransaction) error {
		relations, _, readErr := Neo4jRunQuery(ctx, tx, query, r.scanRelation("source_id", "r", "target_id"))
		if readErr != nil {
			return readErr
		}
		if len(relations) == 0 {
			return ErrNotFound
		}
		relation = relations[0]
		return nil
	})
	if err != nil {
		return nil, errors.Join(ErrIssueGetRelation, err)
	}

	return relation, nil
}

func (r *Neo4jIssueRepository) ListRelations(ctx context.Context, query IssueRelationListQuery) (Page[*IssueRelationItem], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/ListRelations")
	defer span.End()

	page, err := query.Page.Normalize()
	if err != nil {
		return Page[*IssueRelationItem]{}, errors.Join(ErrIssueGetRelations, err)
	}
	query.Page = page
	if query.Order == SortDirectionUnknown {
		query.Order = SortDirectionDesc
	}

	plan, err := CompileQuery(query)
	if err != nil {
		return Page[*IssueRelationItem]{}, errors.Join(ErrIssueGetRelations, err)
	}

	var items []*IssueRelationItem
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		items, _, readErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scanRelationItem())
		return readErr
	})
	if err != nil {
		return Page[*IssueRelationItem]{}, errors.Join(ErrIssueGetRelations, err)
	}

	paged, err := PaginateSlice(items, page.Size, func(item *IssueRelationItem) model.ID {
		return item.ID
	})
	if err != nil {
		return Page[*IssueRelationItem]{}, errors.Join(ErrIssueGetRelations, err)
	}

	return paged, nil
}

func (r *Neo4jIssueRepository) RemoveRelation(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/RemoveRelation")
	defer span.End()

	cypher := `
	MATCH (s:` + source.Label() + ` {id: $source_id})-[r:` + EdgeKindRelatedTo.String() + ` {kind: $kind}]->(t:` + target.Label() + ` {id: $target_id})
	DELETE r`

	params := map[string]any{
		"source_id": source.String(),
		"target_id": target.String(),
		"kind":      kind.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrIssueRemoveRelation, err)
	}

	return nil
}

func (r *Neo4jIssueRepository) RemoveRelationByID(ctx context.Context, relationID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/RemoveRelationByID")
	defer span.End()

	cypher := `
	MATCH ()-[r:` + EdgeKindRelatedTo.String() + ` {id: $id}]-()
	DELETE r`

	params := map[string]any{
		"id": relationID.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrIssueRemoveRelation, err)
	}

	return nil
}

func (r *Neo4jIssueRepository) Update(ctx context.Context, id model.ID, opts UpdateIssueOpts, proj IssueProjection) (*Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Update")
	defer span.End()

	cypher := `
	MATCH (i:` + id.Label() + ` {id: $id})
	SET i += $patch
	SET i.updated_at = datetime.statement()
	RETURN i.id AS id`

	params := map[string]any{
		"id":    id.String(),
		"patch": opts.patch(),
	}

	_, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	})
	if err != nil {
		return nil, errors.Join(ErrIssueUpdate, err)
	}

	return r.Get(ctx, id, proj)
}

func (r *Neo4jIssueRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Delete")
	defer span.End()

	cypher := `MATCH (i:` + id.Label() + ` {id: $id}) DETACH DELETE i`
	params := map[string]any{
		"id": id.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrIssueDelete, err)
	}

	return nil
}

// NewNeo4jIssueRepository creates a new issue neo4jBaseRepository.
func NewNeo4jIssueRepository(opts ...Neo4jRepositoryOption) (*Neo4jIssueRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jIssueRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

func clearIssuesPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeIssue.String(), pattern))
}

func clearIssuesKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return clearIssuesPattern(ctx, r, "Get", id.String(), "*")
}

func clearIssueForIssue(ctx context.Context, r *redisBaseRepository, issueID model.ID) error {
	return clearIssuesPattern(ctx, r, "*", "ListForIssue", issueID.String(), "*")
}

func clearIssueWatchers(ctx context.Context, r *redisBaseRepository, issueID model.ID) error {
	return clearIssuesPattern(ctx, r, "GetWatchers", issueID.String(), "*")
}

func clearIssueRelations(ctx context.Context, r *redisBaseRepository, issueID model.ID) error {
	if err := clearIssuesPattern(ctx, r, "GetRelations", issueID.String(), "*"); err != nil {
		return err
	}
	return clearIssuesPattern(ctx, r, "*", "ListRelations", issueID.String(), "*")
}

func clearIssueRelationPair(ctx context.Context, r *redisBaseRepository, source, target model.ID) error {
	for _, id := range []model.ID{source, target} {
		if err := clearIssuesKey(ctx, r, id); err != nil {
			return err
		}
		if err := clearIssueRelations(ctx, r, id); err != nil {
			return err
		}
	}
	return nil
}

func clearIssueGetByKey(ctx context.Context, r *redisBaseRepository, issueKey string) error {
	return clearIssuesPattern(ctx, r, "GetByKey", "*", issueKey, "*")
}

func clearIssueAllGetByKey(ctx context.Context, r *redisBaseRepository) error {
	return clearIssuesPattern(ctx, r, "GetByKey", "*")
}

func clearIssueAllCrossCache(ctx context.Context, r *redisBaseRepository) error {
	deleteFns := []func(context.Context, *redisBaseRepository, ...string) error{
		clearProjectsPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

const (
	issueListDefaultCacheTTL = 5 * time.Minute
	issueListGenPrefix       = "issue:list:gen"
)

func issueListProjectGenKey(projectID model.ID) string {
	return composeCacheKey(issueListGenPrefix, "project", projectID.String())
}

func issueListNamespaceGenKey(namespaceID model.ID) string {
	return composeCacheKey(issueListGenPrefix, "namespace", namespaceID.String())
}

func issueListUserGenKey(userID model.ID) string {
	return composeCacheKey(issueListGenPrefix, "user", userID.String())
}

func issueListAuthzEpochKey() string {
	return composeCacheKey(issueListGenPrefix, "authz_epoch")
}

func issueListProjectionEpochKey() string {
	return composeCacheKey(issueListGenPrefix, "projection_epoch")
}

func issueListReadGeneration(ctx context.Context, r *redisBaseRepository, key string) int64 {
	var gen int64
	_ = r.Get(ctx, key, &gen)
	return gen
}

func bumpIssueListGeneration(ctx context.Context, r *redisBaseRepository, key string) error {
	gen := issueListReadGeneration(ctx, r, key)
	return r.Set(ctx, key, gen+1)
}

func bumpIssueListProjectGeneration(ctx context.Context, r *redisBaseRepository, projectID model.ID) error {
	return bumpIssueListGeneration(ctx, r, issueListProjectGenKey(projectID))
}

func bumpIssueListNamespaceGeneration(ctx context.Context, r *redisBaseRepository, namespaceID model.ID) error {
	return bumpIssueListGeneration(ctx, r, issueListNamespaceGenKey(namespaceID))
}

func bumpIssueListUserGeneration(ctx context.Context, r *redisBaseRepository, userID model.ID) error {
	return bumpIssueListGeneration(ctx, r, issueListUserGenKey(userID))
}

func bumpIssueListAuthzEpoch(ctx context.Context, r *redisBaseRepository) error {
	return bumpIssueListGeneration(ctx, r, issueListAuthzEpochKey())
}

func bumpIssueListProjectionEpoch(ctx context.Context, r *redisBaseRepository) error {
	return bumpIssueListGeneration(ctx, r, issueListProjectionEpochKey())
}

func issueListCurrentEpochs(ctx context.Context, r *redisBaseRepository) (authz int64, projection int64) {
	authz = issueListReadGeneration(ctx, r, issueListAuthzEpochKey())
	projection = issueListReadGeneration(ctx, r, issueListProjectionEpochKey())
	return
}

func issueAssigneeIDs(issue *Issue) []model.ID {
	if issue == nil {
		return nil
	}
	ids := make([]model.ID, 0, len(issue.Assignments))
	seen := make(map[string]struct{}, len(issue.Assignments))
	for _, assignment := range issue.Assignments {
		if assignment.Kind != model.AssignmentKindAssignee {
			continue
		}
		key := assignment.ID.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		ids = append(ids, assignment.ID)
	}
	return ids
}

func uniqueIDs(ids []model.ID) []model.ID {
	if len(ids) <= 1 {
		return ids
	}
	seen := make(map[string]struct{}, len(ids))
	out := make([]model.ID, 0, len(ids))
	for _, id := range ids {
		key := id.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, id)
	}
	return out
}

// RedisCachedIssueRepository implements caching on the IssueRepository.
type RedisCachedIssueRepository struct {
	cacheRepo *redisBaseRepository
	issueRepo IssueRepository
}

func (r *RedisCachedIssueRepository) Create(ctx context.Context, opts CreateIssueOpts) (*Issue, error) {
	issue, err := r.issueRepo.Create(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := bumpIssueListProjectGeneration(ctx, r.cacheRepo, opts.ProjectID); err != nil {
		return nil, err
	}
	if issue.Namespace != nil {
		if err := bumpIssueListNamespaceGeneration(ctx, r.cacheRepo, issue.Namespace.ID); err != nil {
			return nil, err
		}
	}
	for _, assigneeID := range uniqueIDs(issueAssigneeIDs(issue)) {
		if err := bumpIssueListUserGeneration(ctx, r.cacheRepo, assigneeID); err != nil {
			return nil, err
		}
	}
	if err := clearIssueAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return issue, nil
}

func (r *RedisCachedIssueRepository) Get(ctx context.Context, id model.ID, proj IssueProjection) (*Issue, error) {
	var issue *Issue
	var err error

	key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(proj))
	if err = r.cacheRepo.Get(ctx, key, &issue); err != nil {
		return nil, err
	}

	if issue != nil {
		return issue, nil
	}

	if issue, err = r.issueRepo.Get(ctx, id, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, issue); err != nil {
		return nil, err
	}

	return issue, nil
}

func (r *RedisCachedIssueRepository) GetByKey(ctx context.Context, namespaceID model.ID, issueKey string, proj IssueProjection) (*Issue, error) {
	var issue *Issue
	var err error

	cacheKey := composeCacheKey(model.ResourceTypeIssue.String(), "GetByKey", namespaceID.String(), issueKey, projectionCacheValue(proj))
	if err = r.cacheRepo.Get(ctx, cacheKey, &issue); err != nil {
		return nil, err
	}

	if issue != nil {
		return issue, nil
	}

	if issue, err = r.issueRepo.GetByKey(ctx, namespaceID, issueKey, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, cacheKey, issue); err != nil {
		return nil, err
	}

	return issue, nil
}

func issueListForProjectScopedCacheKey(ctx context.Context, cacheRepo *redisBaseRepository, query IssueListQuery) (string, error) {
	plan, err := CompileQuery(query)
	if err != nil {
		return "", err
	}
	baseKey := plan.CacheKey(model.ResourceTypeIssue.String(), "ListForProject", query.ProjectID.String())
	authzEpoch, projectionEpoch := issueListCurrentEpochs(ctx, cacheRepo)
	projectGen := issueListReadGeneration(ctx, cacheRepo, issueListProjectGenKey(query.ProjectID))
	return composeCacheKey(
		baseKey,
		"g", projectGen,
		"ae", authzEpoch,
		"pe", projectionEpoch,
	), nil
}

func observeIssueListCall(scope string, started time.Time, hit bool, err error) {
	result := metrics.ResultMiss
	if err != nil {
		result = metrics.ResultError
	} else if hit {
		result = metrics.ResultHit
	}
	metrics.ObserveIssueList(scope, result, time.Since(started))
}

func (r *RedisCachedIssueRepository) ListForProject(ctx context.Context, query IssueListQuery) (issues Page[*PartialIssue], err error) {
	started := time.Now()
	hit := false
	defer func() {
		observeIssueListCall(metrics.IssueListScopeProject, started, hit, err)
	}()

	key, err := issueListForProjectScopedCacheKey(ctx, r.cacheRepo, query)
	if err != nil {
		return Page[*PartialIssue]{}, err
	}
	if err = r.cacheRepo.Get(ctx, key, &issues); err != nil {
		return Page[*PartialIssue]{}, err
	}

	if issues.Items != nil {
		hit = true
		return issues, nil
	}

	if issues, err = r.issueRepo.ListForProject(ctx, query); err != nil {
		return Page[*PartialIssue]{}, err
	}

	if err = r.cacheRepo.SetWithTTL(ctx, key, issues, issueListDefaultCacheTTL); err != nil {
		return Page[*PartialIssue]{}, err
	}

	return issues, nil
}

func issueListForNamespaceScopedCacheKey(ctx context.Context, cacheRepo *redisBaseRepository, query IssueListForNamespaceQuery) (string, error) {
	plan, err := CompileQuery(query)
	if err != nil {
		return "", err
	}
	baseKey := plan.CacheKey(model.ResourceTypeIssue.String(), "ListForNamespace", query.NamespaceID.String())
	authzEpoch, projectionEpoch := issueListCurrentEpochs(ctx, cacheRepo)
	namespaceGen := issueListReadGeneration(ctx, cacheRepo, issueListNamespaceGenKey(query.NamespaceID))
	return composeCacheKey(
		baseKey,
		"g", namespaceGen,
		"ae", authzEpoch,
		"pe", projectionEpoch,
	), nil
}

func (r *RedisCachedIssueRepository) ListForNamespace(ctx context.Context, query IssueListForNamespaceQuery) (issues Page[*PartialIssue], err error) {
	started := time.Now()
	hit := false
	defer func() {
		observeIssueListCall(metrics.IssueListScopeNamespace, started, hit, err)
	}()

	key, err := issueListForNamespaceScopedCacheKey(ctx, r.cacheRepo, query)
	if err != nil {
		return Page[*PartialIssue]{}, err
	}
	if err = r.cacheRepo.Get(ctx, key, &issues); err != nil {
		return Page[*PartialIssue]{}, err
	}

	if issues.Items != nil {
		hit = true
		return issues, nil
	}

	if issues, err = r.issueRepo.ListForNamespace(ctx, query); err != nil {
		return Page[*PartialIssue]{}, err
	}

	if err = r.cacheRepo.SetWithTTL(ctx, key, issues, issueListDefaultCacheTTL); err != nil {
		return Page[*PartialIssue]{}, err
	}

	return issues, nil
}

func issueListForUserScopedCacheKey(ctx context.Context, cacheRepo *redisBaseRepository, query IssueListForUserQuery) (string, error) {
	plan, err := CompileQuery(query)
	if err != nil {
		return "", err
	}
	baseKey := plan.CacheKey(model.ResourceTypeIssue.String(), "ListForUser", query.UserID.String())
	authzEpoch, projectionEpoch := issueListCurrentEpochs(ctx, cacheRepo)
	userGen := issueListReadGeneration(ctx, cacheRepo, issueListUserGenKey(query.UserID))
	return composeCacheKey(
		baseKey,
		"g", userGen,
		"ae", authzEpoch,
		"pe", projectionEpoch,
	), nil
}

func (r *RedisCachedIssueRepository) ListForUser(ctx context.Context, query IssueListForUserQuery) (issues Page[*PartialIssue], err error) {
	started := time.Now()
	hit := false
	defer func() {
		observeIssueListCall(metrics.IssueListScopeUser, started, hit, err)
	}()

	key, err := issueListForUserScopedCacheKey(ctx, r.cacheRepo, query)
	if err != nil {
		return Page[*PartialIssue]{}, err
	}
	if err = r.cacheRepo.Get(ctx, key, &issues); err != nil {
		return Page[*PartialIssue]{}, err
	}

	if issues.Items != nil {
		hit = true
		return issues, nil
	}

	if issues, err = r.issueRepo.ListForUser(ctx, query); err != nil {
		return Page[*PartialIssue]{}, err
	}

	if err = r.cacheRepo.SetWithTTL(ctx, key, issues, issueListDefaultCacheTTL); err != nil {
		return Page[*PartialIssue]{}, err
	}

	return issues, nil
}

func (r *RedisCachedIssueRepository) ListForIssue(ctx context.Context, query IssueListForIssueQuery) (Page[*Issue], error) {
	var issues Page[*Issue]
	var err error

	plan, err := CompileQuery(query)
	if err != nil {
		return Page[*Issue]{}, err
	}

	key := plan.CacheKey(model.ResourceTypeIssue.String(), "ListForIssue", query.IssueID.String())
	if err = r.cacheRepo.Get(ctx, key, &issues); err != nil {
		return Page[*Issue]{}, err
	}

	if issues.Items != nil {
		return issues, nil
	}

	if issues, err = r.issueRepo.ListForIssue(ctx, query); err != nil {
		return Page[*Issue]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, issues); err != nil {
		return Page[*Issue]{}, err
	}

	return issues, nil
}

func (r *RedisCachedIssueRepository) AddWatcher(ctx context.Context, issue model.ID, user model.ID) error {
	if err := clearIssuesKey(ctx, r.cacheRepo, issue); err != nil {
		return err
	}

	if err := clearIssueWatchers(ctx, r.cacheRepo, issue); err != nil {
		return err
	}

	return r.issueRepo.AddWatcher(ctx, issue, user)
}

func (r *RedisCachedIssueRepository) GetWatchers(ctx context.Context, issue model.ID) ([]*User, error) {
	var users []*User
	var err error

	key := composeCacheKey(model.ResourceTypeIssue.String(), "GetWatchers", issue.String())
	if err = r.cacheRepo.Get(ctx, key, &users); err != nil {
		return nil, err
	}

	if users != nil {
		return users, nil
	}

	if users, err = r.issueRepo.GetWatchers(ctx, issue); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *RedisCachedIssueRepository) RemoveWatcher(ctx context.Context, issue model.ID, user model.ID) error {
	if err := clearIssuesKey(ctx, r.cacheRepo, issue); err != nil {
		return err
	}

	if err := clearIssueWatchers(ctx, r.cacheRepo, issue); err != nil {
		return err
	}

	return r.issueRepo.RemoveWatcher(ctx, issue, user)
}

func (r *RedisCachedIssueRepository) AddRelation(ctx context.Context, opts CreateIssueRelationOpts) (*IssueRelation, error) {
	if err := clearIssueRelationPair(ctx, r.cacheRepo, opts.Source, opts.Target); err != nil {
		return nil, err
	}

	return r.issueRepo.AddRelation(ctx, opts)
}

func (r *RedisCachedIssueRepository) GetRelation(ctx context.Context, relationID model.ID) (*IssueRelation, error) {
	return r.issueRepo.GetRelation(ctx, relationID)
}

func (r *RedisCachedIssueRepository) ListRelations(ctx context.Context, query IssueRelationListQuery) (Page[*IssueRelationItem], error) {
	var relations Page[*IssueRelationItem]
	var err error

	plan, err := CompileQuery(query)
	if err != nil {
		return Page[*IssueRelationItem]{}, err
	}
	key := plan.CacheKey(model.ResourceTypeIssue.String(), "ListRelations", query.IssueID.String())
	if err = r.cacheRepo.Get(ctx, key, &relations); err != nil {
		return Page[*IssueRelationItem]{}, err
	}

	if relations.Items != nil {
		return relations, nil
	}

	if relations, err = r.issueRepo.ListRelations(ctx, query); err != nil {
		return Page[*IssueRelationItem]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, relations); err != nil {
		return Page[*IssueRelationItem]{}, err
	}

	return relations, nil
}

func (r *RedisCachedIssueRepository) GetRelations(ctx context.Context, issue model.ID) ([]*IssueRelation, error) {
	var relations []*IssueRelation
	var err error

	key := composeCacheKey(model.ResourceTypeIssue.String(), "GetRelations", issue.String())
	if err = r.cacheRepo.Get(ctx, key, &relations); err != nil {
		return nil, err
	}

	if relations != nil {
		return relations, nil
	}

	if relations, err = r.issueRepo.GetRelations(ctx, issue); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, relations); err != nil {
		return nil, err
	}

	return relations, nil
}

func (r *RedisCachedIssueRepository) RemoveRelation(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) error {
	if err := clearIssueRelationPair(ctx, r.cacheRepo, source, target); err != nil {
		return err
	}

	return r.issueRepo.RemoveRelation(ctx, source, target, kind)
}

func (r *RedisCachedIssueRepository) RemoveRelationByID(ctx context.Context, relationID model.ID) error {
	rel, err := r.issueRepo.GetRelation(ctx, relationID)
	if err != nil {
		return err
	}

	if err := clearIssueRelationPair(ctx, r.cacheRepo, rel.Source, rel.Target); err != nil {
		return err
	}

	return r.issueRepo.RemoveRelationByID(ctx, relationID)
}

func (r *RedisCachedIssueRepository) Update(ctx context.Context, id model.ID, opts UpdateIssueOpts, proj IssueProjection) (*Issue, error) {
	before, _ := r.issueRepo.Get(ctx, id, IssueProjection{
		Assignments: true,
	})

	issue, err := r.issueRepo.Update(ctx, id, opts, proj)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeIssue.String(), "Get", id.String(), projectionCacheValue(proj))
	if err = r.cacheRepo.Set(ctx, key, issue); err != nil {
		return nil, err
	}

	if issue.Parent != nil {
		if err := clearIssueForIssue(ctx, r.cacheRepo, issue.Parent.ID); err != nil {
			return nil, err
		}
	}

	if err := bumpIssueListProjectGeneration(ctx, r.cacheRepo, issue.Project.ID); err != nil {
		return nil, err
	}
	if issue.Namespace != nil {
		if err := bumpIssueListNamespaceGeneration(ctx, r.cacheRepo, issue.Namespace.ID); err != nil {
			return nil, err
		}
	}
	if before != nil && before.Project != nil && before.Project.ID != issue.Project.ID {
		if err := bumpIssueListProjectGeneration(ctx, r.cacheRepo, before.Project.ID); err != nil {
			return nil, err
		}
	}
	if before != nil && before.Namespace != nil && (issue.Namespace == nil || before.Namespace.ID != issue.Namespace.ID) {
		if err := bumpIssueListNamespaceGeneration(ctx, r.cacheRepo, before.Namespace.ID); err != nil {
			return nil, err
		}
	}
	assignees := uniqueIDs(append(issueAssigneeIDs(before), issueAssigneeIDs(issue)...))
	for _, assigneeID := range assignees {
		if err := bumpIssueListUserGeneration(ctx, r.cacheRepo, assigneeID); err != nil {
			return nil, err
		}
	}

	if err := clearIssueAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	if issue.Key != "" {
		if err := clearIssueGetByKey(ctx, r.cacheRepo, issue.Key); err != nil {
			return nil, err
		}
	}
	if before != nil && before.Key != "" && before.Key != issue.Key {
		if err := clearIssueGetByKey(ctx, r.cacheRepo, before.Key); err != nil {
			return nil, err
		}
	}

	return issue, nil
}

func (r *RedisCachedIssueRepository) Delete(ctx context.Context, id model.ID) error {
	issue, _ := r.issueRepo.Get(ctx, id, IssueProjection{
		Assignments: true,
	})

	if err := clearIssuesKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearIssueWatchers(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearIssueRelations(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearIssueAllGetByKey(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearIssueAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := r.issueRepo.Delete(ctx, id); err != nil {
		return err
	}

	if issue != nil && issue.Project != nil {
		if err := bumpIssueListProjectGeneration(ctx, r.cacheRepo, issue.Project.ID); err != nil {
			return err
		}
	}
	if issue != nil && issue.Namespace != nil {
		if err := bumpIssueListNamespaceGeneration(ctx, r.cacheRepo, issue.Namespace.ID); err != nil {
			return err
		}
	}
	for _, assigneeID := range uniqueIDs(issueAssigneeIDs(issue)) {
		if err := bumpIssueListUserGeneration(ctx, r.cacheRepo, assigneeID); err != nil {
			return err
		}
	}

	return nil
}

// NewCachedIssueRepository returns a new CachedIssueRepository.
func NewCachedIssueRepository(repo IssueRepository, opts ...RedisRepositoryOption) (*RedisCachedIssueRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedIssueRepository{
		cacheRepo: r,
		issueRepo: repo,
	}, nil
}
