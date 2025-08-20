package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type ProjectRepositoryOld struct {
	mock.Mock
}

func (p *ProjectRepositoryOld) Create(ctx context.Context, namespaceID model.ID, project *model.Project) error {
	args := p.Called(ctx, namespaceID, project)
	return args.Error(0)
}

func (p *ProjectRepositoryOld) Get(ctx context.Context, id model.ID) (*model.Project, error) {
	args := p.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Project), args.Error(1)
}

func (p *ProjectRepositoryOld) GetByKey(ctx context.Context, key string) (*model.Project, error) {
	args := p.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Project), args.Error(1)
}

func (p *ProjectRepositoryOld) GetAll(ctx context.Context, namespaceID model.ID, offset, limit int) ([]*model.Project, error) {
	args := p.Called(ctx, namespaceID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Project), args.Error(1)
}

func (p *ProjectRepositoryOld) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Project, error) {
	args := p.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Project), args.Error(1)
}

func (p *ProjectRepositoryOld) Delete(ctx context.Context, id model.ID) error {
	args := p.Called(ctx, id)
	return args.Error(0)
}
