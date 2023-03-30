package neo4j

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

var (
	ErrIssueCreate = errors.New("failed to create issue") // the issue could not be created
	ErrIssueRead   = errors.New("failed to read issue")   // the issue could not be retrieved
	ErrIssueUpdate = errors.New("failed to update issue") // the issue could not be updated
	ErrIssueDelete = errors.New("failed to delete issue") // the issue could not be deleted
)

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

// IssueRepository is a repository for managing user issues.
type IssueRepository struct {
	*repository
}

func (r *IssueRepository) scan(params *issueScanParams) func(rec *neo4j.Record) (*model.Issue, error) {
	return func(rec *neo4j.Record) (*model.Issue, error) {
		issue := new(model.Issue)
		issue.Links = make([]string, 0)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, params.issue)
		if err != nil {
			return nil, err
		}

		parent, err := ParseValueFromRecord[string](rec, params.parent)
		if err != nil {
			return nil, err
		}

		reportedBy, err := ParseValueFromRecord[string](rec, params.reportedBy)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &issue, []string{"id", "parent", "reported_by"}); err != nil {
			return nil, err
		}

		issue.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.IssueIDType)
		issue.ReportedBy, _ = model.NewIDFromString(reportedBy, model.UserIDType)

		if parent != "" {
			parentID, _ := model.NewIDFromString(parent, model.IssueIDType)
			issue.Parent = &parentID
		}

		if issue.Assignees, err = ParseIDsFromRecord(rec, params.assignees, model.UserIDType); err != nil {
			return nil, err
		}
		if issue.Labels, err = ParseIDsFromRecord(rec, params.labels, model.LabelIDType); err != nil {
			return nil, err
		}
		if issue.Comments, err = ParseIDsFromRecord(rec, params.comments, model.CommentIDType); err != nil {
			return nil, err
		}
		if issue.Attachments, err = ParseIDsFromRecord(rec, params.attachments, model.AttachmentIDType); err != nil {
			return nil, err
		}
		if issue.Watchers, err = ParseIDsFromRecord(rec, params.watchers, model.UserIDType); err != nil {
			return nil, err
		}
		if issue.Relations, err = ParseIDsFromRecord(rec, params.relations, model.IssueIDType); err != nil {
			return nil, err
		}

		if err := issue.Validate(); err != nil {
			return nil, err
		}

		return issue, nil
	}
}

func (r *IssueRepository) Create(ctx context.Context, project model.ID, issue *model.Issue) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Create")
	defer span.End()

	if err := project.Validate(); err != nil {
		return errors.Join(ErrIssueCreate, err)
	}

	if err := issue.Validate(); err != nil {
		return errors.Join(ErrIssueCreate, err)
	}

	createdRelID := model.MustNewID(EdgeKindCreated.String())
	watchesRelID := model.MustNewID(EdgeKindWatches.String())
	belongsToRelID := model.MustNewID(EdgeKindBelongsTo.String())

	createdAt := time.Now()

	issue.ID = model.MustNewID(model.IssueIDType)
	issue.CreatedAt = convert.ToPointer(createdAt)
	issue.UpdatedAt = nil

	cypher := `
	MATCH (p:` + project.Label() + ` {id: $project_id}), (u:` + issue.ReportedBy.Label() + ` {id: $reported_by_id})
	CREATE
		(i:` + issue.ID.Label() + ` {
			id: $id, numeric_id: $numeric_id, kind: $kind, title: $title, description: $description, status: $status,
			priority: $priority, resolution: $resolution, links: $links, due_date: datetime($due_date),
			created_at: datetime($created_at)
		}),
		(u)-[:` + createdRelID.Label() + ` {id: $created_rel_id, created_at: datetime($created_at)}]->(i),
		(u)-[:` + watchesRelID.Label() + ` {id: $watches_rel_id, created_at: datetime($created_at)}]->(i),
		(i)-[:` + belongsToRelID.Label() + ` {id: $belongs_to_rel_id, created_at: datetime($created_at)}]->(p)`

	params := map[string]any{
		"project_id":        project.String(),
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
		"created_rel_id":    createdRelID.String(),
		"watches_rel_id":    watchesRelID.String(),
		"belongs_to_rel_id": belongsToRelID.String(),
	}

	if issue.DueDate != nil {
		params["due_date"] = issue.DueDate.Format(time.RFC3339Nano)
	}

	if issue.Parent != nil {
		issueRelID := model.MustNewID(EdgeKindRelatedTo.String())

		cypher += `
		WITH i
		MATCH (p:` + issue.Parent.Label() + ` {id: $parent_id})
		CREATE (i)-[:` + issueRelID.Label() + ` {id: $issue_rel_id, kind: $rel_kind, created_at: datetime($created_at)}]->(p)`

		params["parent_id"] = issue.Parent.String()
		params["issue_rel_id"] = issueRelID.String()
		params["rel_kind"] = model.IssueRelationKindSubtaskOf.String()
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrIssueCreate, err)
	}

	return nil
}

func (r *IssueRepository) Get(ctx context.Context, id model.ID) (*model.Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Read")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrIssueRead, err)
	}

	cypher := `
	MATCH (i:` + id.Label() + ` {id: $id})
	OPTIONAL MATCH (i)-[:` + EdgeKindRelatedTo.String() + ` {kind: $parent_kind}]->(par:` + model.IssueIDType + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindRelatedTo.String() + `]->(rel:` + model.IssueIDType + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasComment.String() + `]->(comm:` + model.CommentIDType + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasAttachment.String() + `]->(att:` + model.AttachmentIDType + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindWatches.String() + `]-(watch:` + model.UserIDType + `)
	OPTIONAL MATCH (i)-[:` + EdgeKindHasLabel.String() + `]->(l:` + model.LabelIDType + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindCreated.String() + `]-(cr:` + model.UserIDType + `)
	OPTIONAL MATCH (i)<-[:` + EdgeKindAssignedTo.String() + `]-(assignees:` + model.UserIDType + `)
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

	issue, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan(scanParams))
	if err != nil {
		return nil, errors.Join(ErrIssueRead, err)
	}

	return issue, nil
}

func (r *IssueRepository) AddWatcher(ctx context.Context, issue model.ID, user model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/AddWatcher")
	defer span.End()

	panic("implement me")
}

func (r *IssueRepository) GetWatchers(ctx context.Context, issue model.ID) ([]*model.User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/GetWatchers")
	defer span.End()

	panic("implement me")
}

func (r *IssueRepository) RemoveWatcher(ctx context.Context, issue model.ID, user model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/RemoveWatcher")
	defer span.End()

	panic("implement me")
}

func (r *IssueRepository) AddRelation(ctx context.Context, issue model.ID, relation *model.IssueRelation) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/AddRelation")
	defer span.End()

	panic("implement me")
}

func (r *IssueRepository) GetRelations(ctx context.Context, issue model.ID) ([]*model.Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/GetRelations")
	defer span.End()

	panic("implement me")
}

func (r *IssueRepository) RemoveRelation(ctx context.Context, issue model.ID, relation model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/RemoveRelation")
	defer span.End()

	panic("implement me")
}

func (r *IssueRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Update")
	defer span.End()

	panic("implement me")
}

func (r *IssueRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Delete")
	defer span.End()

	if err := id.Validate(); err != nil {
		return errors.Join(ErrIssueDelete, err)
	}

	cypher := `MATCH (i:` + id.Label() + ` {id: $id}) DETACH DELETE i`
	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrIssueDelete, err)
	}

	return nil
}

// NewIssueRepository creates a new issue repository.
func NewIssueRepository(opts ...RepositoryOption) (*IssueRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &IssueRepository{
		repository: baseRepo,
	}, nil
}
