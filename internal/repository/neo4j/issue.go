package neo4j

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
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
	*baseRepository
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

func (r *IssueRepository) scanRelation(ip, rp, tp string) func(rec *neo4j.Record) (*model.IssueRelation, error) {
	return func(rec *neo4j.Record) (*model.IssueRelation, error) {
		rel := new(model.IssueRelation)

		val, _, err := neo4j.GetRecordValue[neo4j.Relationship](rec, rp)
		if err != nil {
			return nil, err
		}

		source, err := ParseValueFromRecord[string](rec, ip)
		if err != nil {
			return nil, err
		}

		target, err := ParseValueFromRecord[string](rec, tp)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &rel, []string{"id", "source", "target"}); err != nil {
			return nil, err
		}

		rel.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), EdgeKindRelatedTo.String())
		rel.Source, _ = model.NewIDFromString(source, model.IssueIDType)
		rel.Target, _ = model.NewIDFromString(target, model.IssueIDType)

		if err := rel.Validate(); err != nil {
			return nil, err
		}

		return rel, nil
	}
}

func (r *IssueRepository) Create(ctx context.Context, project model.ID, issue *model.Issue) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Create")
	defer span.End()

	if err := project.Validate(); err != nil {
		return errors.Join(repository.ErrIssueCreate, err)
	}

	if err := issue.Validate(); err != nil {
		return errors.Join(repository.ErrIssueCreate, err)
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
		return errors.Join(repository.ErrIssueCreate, err)
	}

	return nil
}

func (r *IssueRepository) Get(ctx context.Context, id model.ID) (*model.Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Read")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(repository.ErrIssueRead, err)
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
		return nil, errors.Join(repository.ErrIssueRead, err)
	}

	return issue, nil
}

func (r *IssueRepository) AddWatcher(ctx context.Context, issue model.ID, user model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/AddWatcher")
	defer span.End()

	if err := issue.Validate(); err != nil {
		return errors.Join(repository.ErrIssueAddWatcher, err)
	}

	if err := user.Validate(); err != nil {
		return errors.Join(repository.ErrIssueAddWatcher, err)
	}

	watchesRelID := model.MustNewID(EdgeKindWatches.String())

	cypher := `
	MATCH (i:` + issue.Label() + ` {id: $issue_id}), (u:` + user.Label() + ` {id: $user_id})
	CREATE (u)-[:` + watchesRelID.Label() + ` {id: $rel_id, created_at: datetime($created_at)}]->(i)`

	params := map[string]any{
		"issue_id":   issue.String(),
		"user_id":    user.String(),
		"rel_id":     watchesRelID.String(),
		"created_at": time.Now().Format(time.RFC3339Nano),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrIssueAddWatcher, err)
	}

	return nil
}

func (r *IssueRepository) GetWatchers(ctx context.Context, issue model.ID) ([]*model.User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/GetWatchers")
	defer span.End()

	if err := issue.Validate(); err != nil {
		return nil, errors.Join(repository.ErrIssueGetWatchers, err)
	}

	cypher := `
	MATCH (i:` + issue.Label() + ` {id: $issue_id})<-[:` + EdgeKindWatches.String() + `]-(u:` + model.UserIDType + `)
	OPTIONAL MATCH (u)-[:` + EdgeKindSpeaks.String() + `]->(l:` + languageIDType + `)
	OPTIONAL MATCH (u)-[p:` + EdgeKindHasPermission.String() + `]->()
	OPTIONAL MATCH (u)<-[r:` + EdgeKindBelongsTo.String() + `]-(d:` + model.DocumentIDType + `)
	RETURN u, collect(DISTINCT l.code) AS l, collect(DISTINCT p.id) AS p, collect(DISTINCT d.id) AS d`

	params := map[string]any{
		"issue_id": issue.String(),
	}

	users, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, new(UserRepository).scan("u", "p", "d"))
	if err != nil {
		return nil, errors.Join(repository.ErrIssueGetWatchers, err)
	}

	return users, nil
}

func (r *IssueRepository) RemoveWatcher(ctx context.Context, issue model.ID, user model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/RemoveWatcher")
	defer span.End()

	if err := issue.Validate(); err != nil {
		return errors.Join(repository.ErrIssueRemoveWatcher, err)
	}

	if err := user.Validate(); err != nil {
		return errors.Join(repository.ErrIssueRemoveWatcher, err)
	}

	cypher := `
	MATCH (:` + issue.Label() + ` {id: $issue_id})<-[r:` + EdgeKindWatches.String() + `]-(:` + user.Label() + ` {id: $user_id})
	DELETE r`

	params := map[string]any{
		"issue_id": issue.String(),
		"user_id":  user.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrIssueRemoveWatcher, err)
	}

	return nil
}

func (r *IssueRepository) AddRelation(ctx context.Context, relation *model.IssueRelation) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/AddRelation")
	defer span.End()

	if err := relation.Validate(); err != nil {
		return errors.Join(repository.ErrIssueAddRelation, err)
	}

	createdAt := time.Now()
	relation.ID = model.MustNewID(EdgeKindRelatedTo.String())
	relation.CreatedAt = convert.ToPointer(createdAt)
	relation.UpdatedAt = nil

	cypher := `
	MATCH (s:` + relation.Source.Label() + ` {id: $source_id}), (t:` + relation.Target.Label() + ` {id: $target_id})
	MERGE (s)-[r:` + relation.ID.Label() + ` {kind: $kind}]->(t)
	ON CREATE SET r.id = $id, r.created_at = datetime($created_at)
	`

	params := map[string]any{
		"source_id":  relation.Source.String(),
		"target_id":  relation.Target.String(),
		"id":         relation.ID.String(),
		"kind":       relation.Kind.String(),
		"created_at": createdAt.Format(time.RFC3339Nano),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrIssueAddRelation, err)
	}

	return nil
}

func (r *IssueRepository) GetRelations(ctx context.Context, issue model.ID) ([]*model.IssueRelation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/GetRelations")
	defer span.End()

	if err := issue.Validate(); err != nil {
		return nil, errors.Join(repository.ErrIssueGetRelations, err)
	}

	cypher := `
	MATCH (i:` + issue.Label() + ` {id: $issue_id})-[r:` + EdgeKindRelatedTo.String() + `]-(t)
	RETURN i.id as i, r, t.id as t`

	params := map[string]any{
		"issue_id": issue.String(),
	}

	relations, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scanRelation("i", "r", "t"))
	if err != nil {
		return nil, errors.Join(repository.ErrIssueGetRelations, err)
	}

	return relations, nil
}

func (r *IssueRepository) RemoveRelation(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/RemoveRelation")
	defer span.End()

	if err := source.Validate(); err != nil {
		return errors.Join(repository.ErrIssueRemoveRelation, err)
	}

	if err := target.Validate(); err != nil {
		return errors.Join(repository.ErrIssueRemoveRelation, err)
	}

	cypher := `
	MATCH (s:` + source.Label() + ` {id: $source_id})-[r:` + EdgeKindRelatedTo.String() + ` {kind: $kind}]->(t:` + target.Label() + ` {id: $target_id})
	DELETE r`

	params := map[string]any{
		"source_id": source.String(),
		"target_id": target.String(),
		"kind":      kind.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrIssueRemoveRelation, err)
	}

	return nil
}

func (r *IssueRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Issue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Update")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(repository.ErrIssueUpdate, err)
	}

	cypher := `
	MATCH (i:` + id.Label() + ` {id: $id})
	SET i += $patch, i.updated_at = datetime($updated_at)
	WITH i
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
		"patch":       patch,
		"parent_kind": model.IssueRelationKindSubtaskOf.String(),
		"updated_at":  time.Now().Format(time.RFC3339Nano),
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

	issue, err := ExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan(scanParams))
	if err != nil {
		return nil, errors.Join(repository.ErrIssueRead, err)
	}

	return issue, nil

}

func (r *IssueRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.IssueRepository/Delete")
	defer span.End()

	if err := id.Validate(); err != nil {
		return errors.Join(repository.ErrIssueDelete, err)
	}

	cypher := `MATCH (i:` + id.Label() + ` {id: $id}) DETACH DELETE i`
	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrIssueDelete, err)
	}

	return nil
}

// NewIssueRepository creates a new issue baseRepository.
func NewIssueRepository(opts ...RepositoryOption) (*IssueRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &IssueRepository{
		baseRepository: baseRepo,
	}, nil
}
