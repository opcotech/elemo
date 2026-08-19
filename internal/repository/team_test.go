package repository

import (
	"context"
	"testing"

	"github.com/go-redis/cache/v9"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCachedTeamRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		failIndex int
		failErr   error
		repoErr   error
		wantErr   error
	}{
		{name: "add new team", failIndex: -1},
		{name: "add new team with error", failIndex: -1, repoErr: ErrNotFound, wantErr: ErrNotFound},
		{name: "add new team with belongs to cache error", failIndex: 0, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "add new team with organization cache error", failIndex: 1, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "add new team with project cache error", failIndex: 2, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			opts := CreateTeamOpts{
				Name:        "test team",
				Description: "test description",
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				BelongsTo:   model.MustNewID(model.ResourceTypeOrganization),
			}
			repo := NewMockTeamRepository(ctrl)
			if tt.failIndex < 0 {
				if tt.repoErr != nil {
					repo.EXPECT().Create(ctx, opts).Return(nil, tt.repoErr)
				} else {
					repo.EXPECT().Create(ctx, opts).Return(&Team{}, nil)
				}
			}

			r := &RedisCachedTeamRepository{
				cacheRepo: redisCacheExpectingPatterns(ctrl, ctx, teamCreateCachePatterns(opts.BelongsTo), tt.failIndex, tt.failErr),
				teamRepo:  repo,
			}
			_, err := r.Create(ctx, opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedTeamRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, team *Team) *redisBaseRepository
		teamRepo  func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, team *Team) TeamRepository
	}
	type args struct {
		ctx       context.Context
		id        model.ID
		belongsTo model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(id model.ID) *Team
		wantErr error
	}{
		{
			name: "get uncached team",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, team *Team) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(TeamDetailProjection()))

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: team,
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, team *Team) TeamRepository {
					repo := NewMockTeamRepository(ctrl)
					repo.EXPECT().Get(ctx, id, belongsTo, TeamDetailProjection()).Return(team, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeTeam),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func(id model.ID) *Team {
				return &Team{
					ID:          id,
					Name:        "test team",
					Description: "test description",
					MemberCount: convert.ToPointer(int64(0)),
				}
			},
		},
		{
			name: "get cached team",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, team *Team) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(TeamDetailProjection()))

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if teamPtr, ok := dst.(**Team); ok {
							*teamPtr = team
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *Team) TeamRepository {
					return NewMockTeamRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeTeam),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func(_ model.ID) *Team {
				return &Team{
					ID:          model.MustNewID(model.ResourceTypeTeam),
					Name:        "test team",
					Description: "test description",
					MemberCount: convert.ToPointer(int64(0)),
				}
			},
		},
		{
			name: "get uncached team error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Team) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(TeamDetailProjection()))

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, _ *Team) TeamRepository {
					repo := NewMockTeamRepository(ctrl)
					repo.EXPECT().Get(ctx, id, belongsTo, TeamDetailProjection()).Return(nil, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeTeam),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get cached team error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *Team) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(TeamDetailProjection()))

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *Team) TeamRepository {
					return NewMockTeamRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeTeam),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: ErrCacheRead,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			team := &Team{}
			if tt.want != nil {
				team = tt.want(tt.args.id)
			}

			r := &RedisCachedTeamRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, team),
				teamRepo:  tt.fields.teamRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.belongsTo, team),
			}
			got, err := r.Get(tt.args.ctx, tt.args.id, tt.args.belongsTo, TeamDetailProjection())
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			require.Equal(t, team, got)
		})
	}
}

func TestCachedTeamRepository_ListBelongsTo(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, teams []*Team) *redisBaseRepository
		teamRepo  func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, teams []*Team) TeamRepository
	}
	type args struct {
		ctx       context.Context
		belongsTo model.ID
		limit     int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Team
		wantErr error
	}{
		{
			name: "get uncached teams",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, teams []*Team) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(TeamListProjection()), "", limit)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: Page[*Team]{Items: teams},
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, teams []*Team) TeamRepository {
					repo := NewMockTeamRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, CursorPage{Size: limit}, TeamListProjection()).Return(Page[*Team]{Items: teams}, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				limit:     10,
			},
			want: []*Team{
				{ID: model.MustNewID(model.ResourceTypeTeam), Name: "team-one"},
				{ID: model.MustNewID(model.ResourceTypeTeam), Name: "team-two"},
			},
		},
		{
			name: "get cached teams",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, teams []*Team) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(TeamListProjection()), "", limit)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(*Page[*Team]); ok {
							*ptr = Page[*Team]{Items: teams}
						}
					}).Return(nil)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ int, _ []*Team) TeamRepository {
					return NewMockTeamRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				limit:     10,
			},
			want: []*Team{
				{ID: model.MustNewID(model.ResourceTypeTeam), Name: "team-one"},
			},
		},
		{
			name: "get uncached teams error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, _ []*Team) *redisBaseRepository {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(TeamListProjection()), "", limit)

					db, err := NewRedisDatabase(
						WithRedisClient(mock.NewUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mock.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mock.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mock.NewCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return &redisBaseRepository{
						db:     db,
						cache:  cacheRepo,
						tracer: tracer,
						logger: mock.NewMockLogger(ctrl),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, _ []*Team) TeamRepository {
					repo := NewMockTeamRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, CursorPage{Size: limit}, TeamListProjection()).Return(Page[*Team]{}, ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				limit:     10,
			},
			wantErr: ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := &RedisCachedTeamRepository{
				cacheRepo: tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.limit, tt.want),
				teamRepo:  tt.fields.teamRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.limit, tt.want),
			}
			got, err := r.ListBelongsTo(tt.args.ctx, tt.args.belongsTo, CursorPage{Size: tt.args.limit}, TeamListProjection())
			require.ErrorIs(t, err, tt.wantErr)
			require.ElementsMatch(t, tt.want, got.Items)
		})
	}
}

func TestCachedTeamRepository_ListMembers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	teamID := model.MustNewID(model.ResourceTypeTeam)
	belongsTo := model.MustNewID(model.ResourceTypeOrganization)
	page := CursorPage{Size: 10}
	members := Page[*User]{
		Items: []*User{{ID: model.MustNewID(model.ResourceTypeUser)}},
	}

	t.Run("passthrough without cache", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := NewMockTeamRepository(ctrl)
		repo.EXPECT().ListMembers(ctx, teamID, belongsTo, page).Return(members, nil)

		r := &RedisCachedTeamRepository{
			cacheRepo: &redisBaseRepository{},
			teamRepo:  repo,
		}
		got, err := r.ListMembers(ctx, teamID, belongsTo, page)
		require.NoError(t, err)
		require.Equal(t, members, got)
	})

	t.Run("passthrough error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := NewMockTeamRepository(ctrl)
		repo.EXPECT().ListMembers(ctx, teamID, belongsTo, page).Return(Page[*User]{}, ErrNotFound)

		r := &RedisCachedTeamRepository{
			cacheRepo: &redisBaseRepository{},
			teamRepo:  repo,
		}
		_, err := r.ListMembers(ctx, teamID, belongsTo, page)
		require.ErrorIs(t, err, ErrNotFound)
	})
}

func TestCachedTeamRepository_Update(t *testing.T) {
	newTeam := func(id model.ID) *Team {
		return &Team{ID: id, Name: "test team", Description: "test description"}
	}
	opts := UpdateTeamOpts{
		Name:        optional.Some("updated team"),
		Description: optional.Some("updated description"),
	}

	t.Run("update team", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := context.Background()
		id := model.MustNewID(model.ResourceTypeTeam)
		belongsTo := model.MustNewID(model.ResourceTypeOrganization)
		team := newTeam(id)
		setKey := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(TeamDetailProjection()))
		repo := NewMockTeamRepository(ctrl)
		repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(team, nil)
		r := &RedisCachedTeamRepository{
			cacheRepo: redisCacheExpectingSetThenPatterns(ctrl, ctx, setKey, team, teamUpdateInvalidatePatterns(), false, -1, nil),
			teamRepo:  repo,
		}
		got, err := r.Update(ctx, id, belongsTo, opts)
		require.NoError(t, err)
		require.Equal(t, team, got)
	})

	t.Run("update team with error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := context.Background()
		id := model.MustNewID(model.ResourceTypeTeam)
		belongsTo := model.MustNewID(model.ResourceTypeOrganization)
		db, err := NewRedisDatabase(WithRedisClient(mock.NewUniversalClient(ctrl)))
		require.NoError(t, err)
		repo := NewMockTeamRepository(ctrl)
		repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(nil, ErrNotFound)
		r := &RedisCachedTeamRepository{
			cacheRepo: &redisBaseRepository{db: db, cache: mock.NewCacheBackend(ctrl), tracer: mock.NewMockTracer(ctrl), logger: mock.NewMockLogger(ctrl)},
			teamRepo:  repo,
		}
		_, err = r.Update(ctx, id, belongsTo, opts)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("update team set cache error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := context.Background()
		id := model.MustNewID(model.ResourceTypeTeam)
		belongsTo := model.MustNewID(model.ResourceTypeOrganization)
		team := newTeam(id)
		setKey := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(TeamDetailProjection()))
		repo := NewMockTeamRepository(ctrl)
		repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(team, nil)
		r := &RedisCachedTeamRepository{
			cacheRepo: redisCacheExpectingSetThenPatterns(ctrl, ctx, setKey, team, nil, true, -1, assert.AnError),
			teamRepo:  repo,
		}
		_, err := r.Update(ctx, id, belongsTo, opts)
		require.ErrorIs(t, err, ErrCacheWrite)
	})

	t.Run("update team delete list cache error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := context.Background()
		id := model.MustNewID(model.ResourceTypeTeam)
		belongsTo := model.MustNewID(model.ResourceTypeOrganization)
		team := newTeam(id)
		setKey := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(TeamDetailProjection()))
		repo := NewMockTeamRepository(ctrl)
		repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(team, nil)
		r := &RedisCachedTeamRepository{
			cacheRepo: redisCacheExpectingSetThenPatterns(ctrl, ctx, setKey, team, teamUpdateInvalidatePatterns(), false, 0, assert.AnError),
			teamRepo:  repo,
		}
		_, err := r.Update(ctx, id, belongsTo, opts)
		require.ErrorIs(t, err, ErrCacheDelete)
	})
}

func TestCachedTeamRepository_AddMember(t *testing.T) {
	tests := []struct {
		name      string
		failIndex int
		failErr   error
		repoErr   error
		wantErr   error
	}{
		{name: "add member success", failIndex: -1},
		{name: "add member with team error", failIndex: -1, repoErr: ErrTeamAddMember, wantErr: ErrTeamAddMember},
		{name: "add member with cache deletion error", failIndex: 0, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "add member with related cache deletion error", failIndex: 2, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := context.Background()
			id := model.MustNewID(model.ResourceTypeTeam)
			memberID := model.MustNewID(model.ResourceTypeUser)
			belongsToID := model.MustNewID(model.ResourceTypeOrganization)
			repo := NewMockTeamRepository(ctrl)
			if tt.failIndex < 0 {
				repo.EXPECT().AddMember(ctx, id, memberID, belongsToID).Return(tt.repoErr)
			}
			r := &RedisCachedTeamRepository{
				cacheRepo: redisCacheExpectingPatterns(ctrl, ctx, teamMemberCachePatterns(id, belongsToID), tt.failIndex, tt.failErr),
				teamRepo:  repo,
			}
			err := r.AddMember(ctx, id, memberID, belongsToID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedTeamRepository_RemoveMember(t *testing.T) {
	tests := []struct {
		name      string
		failIndex int
		failErr   error
		repoErr   error
		wantErr   error
	}{
		{name: "remove member success", failIndex: -1},
		{name: "remove member with team error", failIndex: -1, repoErr: ErrTeamRemoveMember, wantErr: ErrTeamRemoveMember},
		{name: "remove member with cache deletion error", failIndex: 0, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "remove member with related cache deletion error", failIndex: 2, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := context.Background()
			id := model.MustNewID(model.ResourceTypeTeam)
			memberID := model.MustNewID(model.ResourceTypeUser)
			belongsToID := model.MustNewID(model.ResourceTypeOrganization)
			repo := NewMockTeamRepository(ctrl)
			if tt.failIndex < 0 {
				repo.EXPECT().RemoveMember(ctx, id, memberID, belongsToID).Return(tt.repoErr)
			}
			r := &RedisCachedTeamRepository{
				cacheRepo: redisCacheExpectingPatterns(ctrl, ctx, teamMemberCachePatterns(id, belongsToID), tt.failIndex, tt.failErr),
				teamRepo:  repo,
			}
			err := r.RemoveMember(ctx, id, memberID, belongsToID)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedTeamRepository_Delete(t *testing.T) {
	tests := []struct {
		name      string
		failIndex int
		failErr   error
		repoErr   error
		wantErr   error
	}{
		{name: "delete team success", failIndex: -1},
		{name: "delete team with team deletion error", failIndex: -1, repoErr: ErrTeamDelete, wantErr: ErrTeamDelete},
		{name: "delete team with cache deletion error", failIndex: 0, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "delete team with list cache deletion error", failIndex: 1, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "delete team with organization cache deletion error", failIndex: 2, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
		{name: "delete team with project cache deletion error", failIndex: 3, failErr: ErrCacheDelete, wantErr: ErrCacheDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := context.Background()
			id := model.MustNewID(model.ResourceTypeTeam)
			belongsTo := model.MustNewID(model.ResourceTypeOrganization)
			repo := NewMockTeamRepository(ctrl)
			if tt.failIndex < 0 {
				repo.EXPECT().Delete(ctx, id, belongsTo).Return(tt.repoErr)
			}
			r := &RedisCachedTeamRepository{
				cacheRepo: redisCacheExpectingPatterns(ctrl, ctx, teamDeleteCachePatterns(id), tt.failIndex, tt.failErr),
				teamRepo:  repo,
			}
			err := r.Delete(ctx, id, belongsTo)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
