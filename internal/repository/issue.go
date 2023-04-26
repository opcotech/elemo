package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// IssueRepository is a repository for managing user issues.
type IssueRepository interface {
	Create(ctx context.Context, project model.ID, issue *model.Issue) error
	Get(ctx context.Context, id model.ID) (*model.Issue, error)
	GetAllForProject(ctx context.Context, projectID model.ID, offset, limit int) ([]*model.Issue, error)
	GetAllForIssue(ctx context.Context, issueID model.ID, offset, limit int) ([]*model.Issue, error)
	AddWatcher(ctx context.Context, issue model.ID, user model.ID) error
	GetWatchers(ctx context.Context, issue model.ID) ([]*model.User, error)
	RemoveWatcher(ctx context.Context, issue model.ID, user model.ID) error
	AddRelation(ctx context.Context, relation *model.IssueRelation) error
	GetRelations(ctx context.Context, issue model.ID) ([]*model.IssueRelation, error)
	RemoveRelation(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) error
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Issue, error)
	Delete(ctx context.Context, id model.ID) error
}
