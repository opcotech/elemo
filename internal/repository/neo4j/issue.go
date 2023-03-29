package neo4j

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrIssueCreate = errors.New("failed to create issue") // the issue could not be created
	ErrIssueRead   = errors.New("failed to read issue")   // the issue could not be retrieved
	ErrIssueUpdate = errors.New("failed to update issue") // the issue could not be updated
	ErrIssueDelete = errors.New("failed to delete issue") // the issue could not be deleted
)

// IssueRepository is a repository for managing user issues.
type IssueRepository struct {
	*repository
}

func (r *IssueRepository) Create(ctx context.Context, issue *model.Issue) error {
	return errors.New("not implemented")
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
