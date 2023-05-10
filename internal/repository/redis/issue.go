package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func clearIssuesPattern(ctx context.Context, r *baseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeIssue.String(), pattern))
}

func clearIssuesKey(ctx context.Context, r *baseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeIssue.String(), id.String()))
}

func clearIssueForProject(ctx context.Context, r *baseRepository, projectID model.ID) error {
	return clearIssuesPattern(ctx, r, "GetAllForProject", projectID.String(), "*")
}

func clearIssueAllForProject(ctx context.Context, r *baseRepository) error {
	return clearIssuesPattern(ctx, r, "GetAllForProject", "*")
}

func clearIssueForIssue(ctx context.Context, r *baseRepository, issueID model.ID) error {
	return clearIssuesPattern(ctx, r, "GetAllForIssue", issueID.String(), "*")
}

func clearIssueAllForIssue(ctx context.Context, r *baseRepository) error {
	return clearIssuesPattern(ctx, r, "GetAllForIssue", "*")
}

func clearIssueWatchers(ctx context.Context, r *baseRepository, issueID model.ID) error {
	return clearIssuesPattern(ctx, r, "GetWatchers", issueID.String(), "*")
}

func clearIssueRelations(ctx context.Context, r *baseRepository, issueID model.ID) error {
	return clearIssuesPattern(ctx, r, "GetRelations", issueID.String(), "*")
}

func clearIssueAllCrossCache(ctx context.Context, r *baseRepository) error {
	deleteFns := []func(context.Context, *baseRepository, ...string) error{
		clearProjectsPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// CachedIssueRepository implements caching on the
// repository.IssueRepository.
type CachedIssueRepository struct {
	cacheRepo *baseRepository
	issueRepo repository.IssueRepository
}

func (r *CachedIssueRepository) Create(ctx context.Context, project model.ID, issue *model.Issue) error {
	if err := clearIssueForProject(ctx, r.cacheRepo, project); err != nil {
		return err
	}

	if issue.Parent != nil {
		if err := clearIssueForIssue(ctx, r.cacheRepo, *issue.Parent); err != nil {
			return err
		}
	}

	if err := clearIssueAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.issueRepo.Create(ctx, project, issue)
}

func (r *CachedIssueRepository) Get(ctx context.Context, id model.ID) (*model.Issue, error) {
	var issue *model.Issue
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

func (r *CachedIssueRepository) GetAllForProject(ctx context.Context, projectID model.ID, offset, limit int) ([]*model.Issue, error) {
	var issues []*model.Issue
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

func (r *CachedIssueRepository) GetAllForIssue(ctx context.Context, issueID model.ID, offset, limit int) ([]*model.Issue, error) {
	var issues []*model.Issue
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

func (r *CachedIssueRepository) AddWatcher(ctx context.Context, issue model.ID, user model.ID) error {
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

func (r *CachedIssueRepository) GetWatchers(ctx context.Context, issue model.ID) ([]*model.User, error) {
	var users []*model.User
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

func (r *CachedIssueRepository) RemoveWatcher(ctx context.Context, issue model.ID, user model.ID) error {
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

func (r *CachedIssueRepository) AddRelation(ctx context.Context, relation *model.IssueRelation) error {
	var issueID model.ID
	if relation.Source.Type == model.ResourceTypeIssue {
		issueID = relation.Source
	} else {
		issueID = relation.Target
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

	return r.issueRepo.AddRelation(ctx, relation)
}

func (r *CachedIssueRepository) GetRelations(ctx context.Context, issue model.ID) ([]*model.IssueRelation, error) {
	var relations []*model.IssueRelation
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

func (r *CachedIssueRepository) RemoveRelation(ctx context.Context, source, target model.ID, kind model.IssueRelationKind) error {
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

func (r *CachedIssueRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Issue, error) {
	var issue *model.Issue
	var err error

	issue, err = r.issueRepo.Update(ctx, id, patch)
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

func (r *CachedIssueRepository) Delete(ctx context.Context, id model.ID) error {
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
func NewCachedIssueRepository(repo repository.IssueRepository, opts ...RepositoryOption) (*CachedIssueRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedIssueRepository{
		cacheRepo: r,
		issueRepo: repo,
	}, nil
}
