package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type IssueRepository struct {
	mock.Mock
}

func (i *IssueRepository) Create(ctx context.Context, project model.ID, issue *model.Issue) error {
	args := i.Called(ctx, project, issue)
	return args.Error(0)
}

func (i *IssueRepository) Get(ctx context.Context, id model.ID) (*model.Issue, error) {
	args := i.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Issue), args.Error(1)
}

func (i *IssueRepository) FindAllForProject(ctx context.Context, projectID model.ID, offset, limit int) ([]*model.Issue, error) {
	args := i.Called(ctx, projectID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Issue), args.Error(1)
}

func (i *IssueRepository) FindAllForIssue(ctx context.Context, issueID model.ID, offset, limit int) ([]*model.Issue, error) {
	args := i.Called(ctx, issueID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Issue), args.Error(1)
}

func (i *IssueRepository) AddWatcher(ctx context.Context, issue model.ID, user model.ID) error {
	args := i.Called(ctx, issue, user)
	return args.Error(0)
}

func (i *IssueRepository) GetWatchers(ctx context.Context, issue model.ID) ([]*model.User, error) {
	args := i.Called(ctx, issue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (i *IssueRepository) RemoveWatcher(ctx context.Context, issue model.ID, user model.ID) error {
	args := i.Called(ctx, issue, user)
	return args.Error(0)
}

func (i *IssueRepository) AddRelation(ctx context.Context, relation *model.IssueRelation) error {
	args := i.Called(ctx, relation)
	return args.Error(0)
}

func (i *IssueRepository) GetRelations(ctx context.Context, issue model.ID) ([]*model.IssueRelation, error) {
	args := i.Called(ctx, issue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.IssueRelation), args.Error(1)
}

func (i *IssueRepository) RemoveRelation(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) error {
	args := i.Called(ctx, source, target, kind)
	return args.Error(0)
}

func (i *IssueRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Issue, error) {
	args := i.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Issue), args.Error(1)
}

func (i *IssueRepository) Delete(ctx context.Context, id model.ID) error {
	args := i.Called(ctx, id)
	return args.Error(0)
}
