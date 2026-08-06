package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
)

var (
	ErrIssueAddRelation    = errors.New("failed to add relation to issue")      // the relation could not be added to the issue
	ErrIssueAddWatcher     = errors.New("failed to add watcher to issue")       // the watcher could not be added to the issue
	ErrIssueCreate         = errors.New("failed to create issue")               // the issue could not be created
	ErrIssueDelete         = errors.New("failed to delete issue")               // the issue could not be deleted
	ErrIssueGetRelations   = errors.New("failed to get relations for issue")    // the relations could not be retrieved for the issue
	ErrIssueGetWatchers    = errors.New("failed to get watchers for issue")     // the watchers could not be retrieved for the issue
	ErrIssueRead           = errors.New("failed to read issue")                 // the issue could not be retrieved
	ErrIssueRemoveRelation = errors.New("failed to remove relation from issue") // the relation could not be removed from the issue
	ErrIssueRemoveWatcher  = errors.New("failed to remove watcher from issue")  // the watcher could not be removed from the issue
	ErrIssueUpdate         = errors.New("failed to update issue")               // the issue could not be updated
)

// Issue represents an issue persisted by the repository.
type Issue struct {
	ID          model.ID              `json:"id"`
	NumericID   uint                  `json:"numeric_id"`
	Parent      *model.ID             `json:"parent"`
	Kind        model.IssueKind       `json:"kind"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Status      model.IssueStatus     `json:"status"`
	Priority    model.IssuePriority   `json:"priority"`
	Resolution  model.IssueResolution `json:"resolution"`
	ReportedBy  model.ID              `json:"reported_by"`
	Assignees   []model.ID            `json:"assignees"`
	Labels      []model.ID            `json:"labels"`
	Comments    []model.ID            `json:"comments"`
	Attachments []model.ID            `json:"attachments"`
	Watchers    []model.ID            `json:"watchers"`
	Relations   []model.ID            `json:"relations"`
	Links       []string              `json:"links"`
	DueDate     *time.Time            `json:"due_date"`
	CreatedAt   *time.Time            `json:"created_at"`
	UpdatedAt   *time.Time            `json:"updated_at"`
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

// CreateIssueOpts holds the data required to create an issue.
type CreateIssueOpts struct {
	ProjectID   model.ID
	NumericID   uint
	Parent      *model.ID
	Kind        model.IssueKind
	Title       string
	Description string
	Status      model.IssueStatus
	Priority    model.IssuePriority
	Resolution  model.IssueResolution
	ReportedBy  model.ID
	Links       []string
	DueDate     *time.Time
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
	Links       optional.Optional[[]string]
	DueDate     optional.Optional[time.Time]
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
		p["description"] = *o.Description.Value
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
		if o.Links.Value == nil {
			p["links"] = []string{}
		} else {
			p["links"] = *o.Links.Value
		}
	}
	if o.DueDate.Defined {
		if o.DueDate.Value == nil {
			p["due_date"] = nil
		} else {
			p["due_date"] = o.DueDate.Value.Format(time.RFC3339Nano)
		}
	}

	return p
}

//go:generate mockgen -source=issue.go -destination=issue_mock_gen.go -package=repository -mock_names "IssueRepository=MockIssueRepository"
type IssueRepository interface {
	Create(ctx context.Context, opts CreateIssueOpts) (*Issue, error)
	Get(ctx context.Context, id model.ID) (*Issue, error)
	GetAllForProject(ctx context.Context, projectID model.ID, offset, limit int) ([]*Issue, error)
	GetAllForIssue(ctx context.Context, issueID model.ID, offset, limit int) ([]*Issue, error)
	AddWatcher(ctx context.Context, issue model.ID, user model.ID) error
	GetWatchers(ctx context.Context, issue model.ID) ([]*User, error)
	RemoveWatcher(ctx context.Context, issue model.ID, user model.ID) error
	AddRelation(ctx context.Context, opts CreateIssueRelationOpts) (*IssueRelation, error)
	GetRelations(ctx context.Context, issue model.ID) ([]*IssueRelation, error)
	RemoveRelation(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) error
	Update(ctx context.Context, id model.ID, opts UpdateIssueOpts) (*Issue, error)
	Delete(ctx context.Context, id model.ID) error
}

// issueScanParams is a struct for holding the cypher return parameter names
// for scanning an issue.
type issueScanParams struct {
	issue       string
	parent      string
	reportedBy  string
	assignees   string
	labels      string
	comments    string
	attachments string
	watchers    string
	relations   string
}

// Neo4jIssueRepository is a repository for managing user issues.
type Neo4jIssueRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jIssueRepository) scan(params *issueScanParams) func(rec *neo4j.Record) (*Issue, error) {
	return func(rec *neo4j.Record) (*Issue, error) {
		issue := new(Issue)
		issue.Links = make([]string, 0)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, params.issue)
		if err != nil {
			return nil, err
		}

		parent, err := Neo4jParseValueFromRecord[string](rec, params.parent)
		if err != nil {
			return nil, err
		}

		reportedBy, err := Neo4jParseValueFromRecord[string](rec, params.reportedBy)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &issue, []string{"id", "parent", "reported_by"}); err != nil {
			return nil, err
		}

		issue.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeIssue.String())
		issue.ReportedBy, _ = model.NewIDFromString(reportedBy, model.ResourceTypeUser.String())

		if parent != "" {
			parentID, _ := model.NewIDFromString(parent, model.ResourceTypeIssue.String())
			issue.Parent = &parentID
		}

		if issue.Assignees, err = Neo4jParseIDsFromRecord(rec, params.assignees, model.ResourceTypeUser.String()); err != nil {
			return nil, err
		}
		if issue.Labels, err = Neo4jParseIDsFromRecord(rec, params.labels, model.ResourceTypeLabel.String()); err != nil {
			return nil, err
		}
		if issue.Comments, err = Neo4jParseIDsFromRecord(rec, params.comments, model.ResourceTypeComment.String()); err != nil {
			return nil, err
		}
		if issue.Attachments, err = Neo4jParseIDsFromRecord(rec, params.attachments, model.ResourceTypeAttachment.String()); err != nil {
			return nil, err
		}
		if issue.Watchers, err = Neo4jParseIDsFromRecord(rec, params.watchers, model.ResourceTypeUser.String()); err != nil {
			return nil, err
		}
		if issue.Relations, err = Neo4jParseIDsFromRecord(rec, params.relations, model.ResourceTypeIssue.String()); err != nil {
			return nil, err
		}

		return issue, nil
	}
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

func (r *Neo4jIssueRepository) Create(ctx context.Context, opts CreateIssueOpts) (*Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Create")
	defer span.End()

	createdAt := convert.ToPointer(time.Now().UTC())

	links := opts.Links
	if links == nil {
		links = make([]string, 0)
	}

	issue := &Issue{
		ID:          model.MustNewID(model.ResourceTypeIssue),
		NumericID:   opts.NumericID,
		Parent:      opts.Parent,
		Kind:        opts.Kind,
		Title:       opts.Title,
		Description: opts.Description,
		Status:      opts.Status,
		Priority:    opts.Priority,
		Resolution:  opts.Resolution,
		ReportedBy:  opts.ReportedBy,
		Assignees:   make([]model.ID, 0),
		Labels:      make([]model.ID, 0),
		Comments:    make([]model.ID, 0),
		Attachments: make([]model.ID, 0),
		Watchers:    make([]model.ID, 0),
		Relations:   make([]model.ID, 0),
		Links:       links,
		DueDate:     opts.DueDate,
		CreatedAt:   createdAt,
		UpdatedAt:   nil,
	}

	cypher := `
	MATCH (p:` + opts.ProjectID.Label() + ` {id: $project_id})
	MATCH (u:` + issue.ReportedBy.Label() + ` {id: $reported_by_id})
	CREATE
		(i:` + issue.ID.Label() + ` {
			id: $id, numeric_id: $numeric_id, kind: $kind, title: $title, description: $description, status: $status,
			priority: $priority, resolution: $resolution, links: $links, due_date: datetime($due_date),
			created_at: datetime($created_at)
		}),
		(u)-[:` + EdgeKindCreated.String() + ` {id: $created_rel_id, created_at: datetime($created_at)}]->(i),
		(u)-[:` + EdgeKindWatches.String() + ` {id: $watches_rel_id, created_at: datetime($created_at)}]->(i),
		(i)-[:` + EdgeKindBelongsTo.String() + ` {id: $belongs_to_rel_id, created_at: datetime($created_at)}]->(p)`

	params := map[string]any{
		"project_id":        opts.ProjectID.String(),
		"reported_by_id":    issue.ReportedBy.String(),
		"id":                issue.ID.String(),
		"numeric_id":        issue.NumericID,
		"kind":              issue.Kind.String(),
		"title":             issue.Title,
		"description":       issue.Description,
		"status":            issue.Status.String(),
		"priority":          issue.Priority.String(),
		"resolution":        issue.Resolution.String(),
		"links":             issue.Links,
		"due_date":          nil,
		"created_at":        createdAt.Format(time.RFC3339Nano),
		"created_rel_id":    model.NewRawID(),
		"watches_rel_id":    model.NewRawID(),
		"belongs_to_rel_id": model.NewRawID(),
	}

	if issue.DueDate != nil {
		params["due_date"] = issue.DueDate.Format(time.RFC3339Nano)
	}

	if issue.Parent != nil {
		cypher += `
		WITH i
		MATCH (p:` + issue.Parent.Label() + ` {id: $parent_id})
		CREATE (i)-[:` + EdgeKindRelatedTo.String() + ` {id: $issue_rel_id, kind: $rel_kind, created_at: datetime($created_at)}]->(p)`

		params["parent_id"] = issue.Parent.String()
		params["issue_rel_id"] = model.NewRawID()
		params["rel_kind"] = model.IssueRelationKindSubtaskOf.String()
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrIssueCreate, err)
	}

	return issue, nil
}

func (r *Neo4jIssueRepository) Get(ctx context.Context, id model.ID) (*Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Read")
	defer span.End()

	cypher := `
	MATCH (i:` + id.Label() + ` {id: $id})
	OPTIONAL MATCH (i)-[:` + EdgeKindRelatedTo.String() + ` {kind: $parent_kind}]->(par:` + model.ResourceTypeIssue.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindRelatedTo.String() + `]->(rel:` + model.ResourceTypeIssue.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasComment.String() + `]->(comm:` + model.ResourceTypeComment.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasAttachment.String() + `]->(att:` + model.ResourceTypeAttachment.String() + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindWatches.String() + `]-(watch:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasLabel.String() + `]->(l:` + model.ResourceTypeLabel.String() + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindCreated.String() + `]-(cr:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindAssignedTo.String() + `]-(assignees:` + model.ResourceTypeUser.String() + `)
	RETURN i, par.id AS par, collect(DISTINCT rel.id) AS rel, collect(DISTINCT comm.id) AS comm,
		collect(DISTINCT att.id) AS att, collect(DISTINCT watch.id) AS watch, collect(DISTINCT l.id) AS l, cr.id as cr,
		collect(DISTINCT assignees.id) AS assignees`

	params := map[string]any{
		"id":          id.String(),
		"parent_kind": model.IssueRelationKindSubtaskOf.String(),
	}

	scanParams := &issueScanParams{
		issue:       "i",
		parent:      "par",
		reportedBy:  "cr",
		assignees:   "assignees",
		labels:      "l",
		comments:    "comm",
		attachments: "att",
		watchers:    "watch",
		relations:   "rel",
	}

	issue, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan(scanParams))
	if err != nil {
		return nil, errors.Join(ErrIssueRead, err)
	}

	return issue, nil
}

func (r *Neo4jIssueRepository) GetAllForProject(ctx context.Context, projectID model.ID, offset, limit int) ([]*Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/GetAllForProject")
	defer span.End()

	cypher := `
	MATCH (i:` + model.ResourceTypeIssue.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(:` + projectID.Label() + ` {id: $id})
	OPTIONAL MATCH (i)-[:` + EdgeKindRelatedTo.String() + ` {kind: $parent_kind}]->(par:` + model.ResourceTypeIssue.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindRelatedTo.String() + `]->(rel:` + model.ResourceTypeIssue.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasComment.String() + `]->(comm:` + model.ResourceTypeComment.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasAttachment.String() + `]->(att:` + model.ResourceTypeAttachment.String() + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindWatches.String() + `]-(watch:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasLabel.String() + `]->(l:` + model.ResourceTypeLabel.String() + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindCreated.String() + `]-(cr:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindAssignedTo.String() + `]-(assignees:` + model.ResourceTypeUser.String() + `)
	RETURN i, par.id AS par, collect(DISTINCT rel.id) AS rel, collect(DISTINCT comm.id) AS comm,
		collect(DISTINCT att.id) AS att, collect(DISTINCT watch.id) AS watch, collect(DISTINCT l.id) AS l, cr.id as cr,
		collect(DISTINCT assignees.id) AS assignees
	ORDER BY i.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"id":          projectID.String(),
		"parent_kind": model.IssueRelationKindSubtaskOf.String(),
		"offset":      offset,
		"limit":       limit,
	}

	scanParams := &issueScanParams{
		issue:       "i",
		parent:      "par",
		reportedBy:  "cr",
		assignees:   "assignees",
		labels:      "l",
		comments:    "comm",
		attachments: "att",
		watchers:    "watch",
		relations:   "rel",
	}

	issues, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan(scanParams))
	if err != nil {
		return nil, errors.Join(ErrIssueRead, err)
	}

	return issues, nil
}

func (r *Neo4jIssueRepository) GetAllForIssue(ctx context.Context, issueID model.ID, offset, limit int) ([]*Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/GetAllForProject")
	defer span.End()

	cypher := `
	MATCH (i:` + model.ResourceTypeIssue.String() + `)-[:` + EdgeKindRelatedTo.String() + `]-(:` + issueID.Label() + ` {id: $id})
	OPTIONAL MATCH (i)-[:` + EdgeKindRelatedTo.String() + ` {kind: $parent_kind}]->(par:` + model.ResourceTypeIssue.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindRelatedTo.String() + `]->(rel:` + model.ResourceTypeIssue.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasComment.String() + `]->(comm:` + model.ResourceTypeComment.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasAttachment.String() + `]->(att:` + model.ResourceTypeAttachment.String() + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindWatches.String() + `]-(watch:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasLabel.String() + `]->(l:` + model.ResourceTypeLabel.String() + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindCreated.String() + `]-(cr:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindAssignedTo.String() + `]-(assignees:` + model.ResourceTypeUser.String() + `)
	RETURN i, par.id AS par, collect(DISTINCT rel.id) AS rel, collect(DISTINCT comm.id) AS comm,
		collect(DISTINCT att.id) AS att, collect(DISTINCT watch.id) AS watch, collect(DISTINCT l.id) AS l, cr.id as cr,
		collect(DISTINCT assignees.id) AS assignees
	ORDER BY i.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"id":          issueID.String(),
		"parent_kind": model.IssueRelationKindSubtaskOf.String(),
		"offset":      offset,
		"limit":       limit,
	}

	scanParams := &issueScanParams{
		issue:       "i",
		parent:      "par",
		reportedBy:  "cr",
		assignees:   "assignees",
		labels:      "l",
		comments:    "comm",
		attachments: "att",
		watchers:    "watch",
		relations:   "rel",
	}

	issues, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan(scanParams))
	if err != nil {
		return nil, errors.Join(ErrIssueRead, err)
	}

	return issues, nil
}

func (r *Neo4jIssueRepository) AddWatcher(ctx context.Context, issue model.ID, user model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/AddWatcher")
	defer span.End()

	cypher := `
	MATCH (i:` + issue.Label() + ` {id: $issue_id})
	MATCH (u:` + user.Label() + ` {id: $user_id})
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

	cypher := `
	MATCH (i:` + issue.Label() + ` {id: $issue_id})<-[:` + EdgeKindWatches.String() + `]-(u:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (u)-[:` + EdgeKindSpeaks.String() + `]->(l:` + languageIDType + `)
	OPTIONAL MATCH (u)-[p:` + EdgeKindHasPermission.String() + `]->()
	OPTIONAL MATCH (u)<-[r:` + EdgeKindBelongsTo.String() + `]-(d:` + model.ResourceTypeDocument.String() + `)
	RETURN u, collect(DISTINCT p.id) AS p, collect(DISTINCT d.id) AS d`

	params := map[string]any{
		"issue_id": issue.String(),
	}

	users, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, new(Neo4jUserRepository).scan("u", "p", "d"))
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
	MATCH (s:` + relation.Source.Label() + ` {id: $source_id})
	MATCH (t:` + relation.Target.Label() + ` {id: $target_id})
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

	cypher := `
	MATCH (i:` + issue.Label() + ` {id: $issue_id})-[r:` + EdgeKindRelatedTo.String() + `]-(t)
	RETURN i.id as i, r, t.id as t`

	params := map[string]any{
		"issue_id": issue.String(),
	}

	relations, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scanRelation("i", "r", "t"))
	if err != nil {
		return nil, errors.Join(ErrIssueGetRelations, err)
	}

	return relations, nil
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

func (r *Neo4jIssueRepository) Update(ctx context.Context, id model.ID, opts UpdateIssueOpts) (*Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Update")
	defer span.End()

	cypher := `
	MATCH (i:` + id.Label() + ` {id: $id})
	SET i += $patch, i.updated_at = datetime()
	WITH i
	OPTIONAL MATCH (i)-[:` + EdgeKindRelatedTo.String() + ` {kind: $parent_kind}]->(par:` + model.ResourceTypeIssue.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindRelatedTo.String() + `]->(rel:` + model.ResourceTypeIssue.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasComment.String() + `]->(comm:` + model.ResourceTypeComment.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasAttachment.String() + `]->(att:` + model.ResourceTypeAttachment.String() + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindWatches.String() + `]-(watch:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasLabel.String() + `]->(l:` + model.ResourceTypeLabel.String() + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindCreated.String() + `]-(cr:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindAssignedTo.String() + `]-(assignees:` + model.ResourceTypeUser.String() + `)
	RETURN i, par.id AS par, collect(DISTINCT rel.id) AS rel, collect(DISTINCT comm.id) AS comm,
		collect(DISTINCT att.id) AS att, collect(DISTINCT watch.id) AS watch, collect(DISTINCT l.id) AS l, cr.id as cr,
		collect(DISTINCT assignees.id) AS assignees`

	params := map[string]any{
		"id":          id.String(),
		"patch":       opts.patch(),
		"parent_kind": model.IssueRelationKindSubtaskOf.String(),
	}

	scanParams := &issueScanParams{
		issue:       "i",
		parent:      "par",
		reportedBy:  "cr",
		assignees:   "assignees",
		labels:      "l",
		comments:    "comm",
		attachments: "att",
		watchers:    "watch",
		relations:   "rel",
	}

	issue, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan(scanParams))
	if err != nil {
		return nil, errors.Join(ErrIssueUpdate, err)
	}

	return issue, nil
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
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeIssue.String(), id.String()))
}

func clearIssueForProject(ctx context.Context, r *redisBaseRepository, projectID model.ID) error {
	return clearIssuesPattern(ctx, r, "GetAllForProject", projectID.String(), "*")
}

func clearIssueAllForProject(ctx context.Context, r *redisBaseRepository) error {
	return clearIssuesPattern(ctx, r, "GetAllForProject", "*")
}

func clearIssueForIssue(ctx context.Context, r *redisBaseRepository, issueID model.ID) error {
	return clearIssuesPattern(ctx, r, "GetAllForIssue", issueID.String(), "*")
}

func clearIssueAllForIssue(ctx context.Context, r *redisBaseRepository) error {
	return clearIssuesPattern(ctx, r, "GetAllForIssue", "*")
}

func clearIssueWatchers(ctx context.Context, r *redisBaseRepository, issueID model.ID) error {
	return clearIssuesPattern(ctx, r, "GetWatchers", issueID.String(), "*")
}

func clearIssueRelations(ctx context.Context, r *redisBaseRepository, issueID model.ID) error {
	return clearIssuesPattern(ctx, r, "GetRelations", issueID.String(), "*")
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

// RedisCachedIssueRepository implements caching on the IssueRepository.
type RedisCachedIssueRepository struct {
	cacheRepo *redisBaseRepository
	issueRepo IssueRepository
}

func (r *RedisCachedIssueRepository) Create(ctx context.Context, opts CreateIssueOpts) (*Issue, error) {
	if err := clearIssueForProject(ctx, r.cacheRepo, opts.ProjectID); err != nil {
		return nil, err
	}

	if opts.Parent != nil {
		if err := clearIssueForIssue(ctx, r.cacheRepo, *opts.Parent); err != nil {
			return nil, err
		}
	}

	if err := clearIssueAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return r.issueRepo.Create(ctx, opts)
}

func (r *RedisCachedIssueRepository) Get(ctx context.Context, id model.ID) (*Issue, error) {
	var issue *Issue
	var err error

	key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &issue); err != nil {
		return nil, err
	}

	if issue != nil {
		return issue, nil
	}

	if issue, err = r.issueRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, issue); err != nil {
		return nil, err
	}

	return issue, nil
}

func (r *RedisCachedIssueRepository) GetAllForProject(ctx context.Context, projectID model.ID, offset, limit int) ([]*Issue, error) {
	var issues []*Issue
	var err error

	key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForProject", projectID.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &issues); err != nil {
		return nil, err
	}

	if issues != nil {
		return issues, nil
	}

	if issues, err = r.issueRepo.GetAllForProject(ctx, projectID, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, issues); err != nil {
		return nil, err
	}

	return issues, nil
}

func (r *RedisCachedIssueRepository) GetAllForIssue(ctx context.Context, issueID model.ID, offset, limit int) ([]*Issue, error) {
	var issues []*Issue
	var err error

	key := composeCacheKey(model.ResourceTypeIssue.String(), "GetAllForIssue", issueID.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &issues); err != nil {
		return nil, err
	}

	if issues != nil {
		return issues, nil
	}

	if issues, err = r.issueRepo.GetAllForIssue(ctx, issueID, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, issues); err != nil {
		return nil, err
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

	if err := clearIssueAllForIssue(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearIssueAllForProject(ctx, r.cacheRepo); err != nil {
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

	if err := clearIssueAllForIssue(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearIssueAllForProject(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.issueRepo.RemoveWatcher(ctx, issue, user)
}

func (r *RedisCachedIssueRepository) AddRelation(ctx context.Context, opts CreateIssueRelationOpts) (*IssueRelation, error) {
	var issueID model.ID
	if opts.Source.Type == model.ResourceTypeIssue {
		issueID = opts.Source
	} else {
		issueID = opts.Target
	}

	if err := clearIssuesKey(ctx, r.cacheRepo, issueID); err != nil {
		return nil, err
	}

	if err := clearIssueRelations(ctx, r.cacheRepo, issueID); err != nil {
		return nil, err
	}

	if err := clearIssueAllForIssue(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	if err := clearIssueAllForProject(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return r.issueRepo.AddRelation(ctx, opts)
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
	var issueID model.ID
	if source.Type == model.ResourceTypeIssue {
		issueID = source
	} else {
		issueID = target
	}

	if err := clearIssuesKey(ctx, r.cacheRepo, issueID); err != nil {
		return err
	}

	if err := clearIssueRelations(ctx, r.cacheRepo, issueID); err != nil {
		return err
	}

	if err := clearIssueAllForIssue(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearIssueAllForProject(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.issueRepo.RemoveRelation(ctx, source, target, kind)
}

func (r *RedisCachedIssueRepository) Update(ctx context.Context, id model.ID, opts UpdateIssueOpts) (*Issue, error) {
	issue, err := r.issueRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeIssue.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, issue); err != nil {
		return nil, err
	}

	if issue.Parent != nil {
		if err := clearIssueForIssue(ctx, r.cacheRepo, *issue.Parent); err != nil {
			return nil, err
		}
	}

	if err := clearIssueAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return issue, nil
}

func (r *RedisCachedIssueRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearIssuesKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearIssueWatchers(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearIssueRelations(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearIssueAllForIssue(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearIssueAllForProject(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearIssueAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.issueRepo.Delete(ctx, id)
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
