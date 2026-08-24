package repository_test

import (
	"context"
	"testing"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"

	"github.com/go-redis/cache/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
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
		{name: "add new team with error", failIndex: -1, repoErr: repository.ErrNotFound, wantErr: repository.ErrNotFound},
		{name: "add new team with belongs to cache error", failIndex: 0, failErr: repository.ErrCacheDelete, wantErr: repository.ErrCacheDelete},
		{name: "add new team with organization cache error", failIndex: 1, failErr: repository.ErrCacheDelete, wantErr: repository.ErrCacheDelete},
		{name: "add new team with project cache error", failIndex: 2, failErr: repository.ErrCacheDelete, wantErr: repository.ErrCacheDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			opts := repository.CreateTeamOpts{
				Name:        "test team",
				Description: "test description",
				CreatedBy:   model.MustNewID(model.ResourceTypeUser),
				BelongsTo:   model.MustNewID(model.ResourceTypeOrganization),
			}
			repo := mockrepo.NewMockTeamRepository(ctrl)
			if tt.failIndex < 0 {
				if tt.repoErr != nil {
					repo.EXPECT().Create(ctx, opts).Return(nil, tt.repoErr)
				} else {
					repo.EXPECT().Create(ctx, opts).Return(&repository.Team{}, nil)
				}
			}
			bumpCount := 1
			if tt.repoErr != nil {
				bumpCount = 0
			}

			r := func() *repository.RedisCachedTeamRepository {
				r, err := repository.NewCachedTeamRepository(
					repo,
					redisCacheExpectingPatternsThenIssueAuthzEpochBump(ctrl, ctx, teamCreateCachePatterns(opts.BelongsTo), tt.failIndex, tt.failErr, bumpCount)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			_, err := r.Create(ctx, opts)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCachedTeamRepository_Get(t *testing.T) {
	type fields struct {
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, id model.ID, team *repository.Team) []repository.RedisRepositoryOption
		teamRepo  func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, team *repository.Team) repository.TeamRepository
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
		want    func(id model.ID) *repository.Team
		wantErr error
	}{
		{
			name: "get uncached team",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, team *repository.Team) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(repository.TeamDetailProjection()))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: team,
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, team *repository.Team) repository.TeamRepository {
					repo := mockrepo.NewMockTeamRepository(ctrl)
					repo.EXPECT().Get(ctx, id, belongsTo, repository.TeamDetailProjection()).Return(team, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeTeam),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func(id model.ID) *repository.Team {
				return &repository.Team{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, team *repository.Team) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(repository.TeamDetailProjection()))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if teamPtr, ok := dst.(**repository.Team); ok {
							*teamPtr = team
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *repository.Team) repository.TeamRepository {
					return mockrepo.NewMockTeamRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeTeam),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			want: func(_ model.ID) *repository.Team {
				return &repository.Team{
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
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Team) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(repository.TeamDetailProjection()))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, ctx context.Context, id, belongsTo model.ID, _ *repository.Team) repository.TeamRepository {
					repo := mockrepo.NewMockTeamRepository(ctrl)
					repo.EXPECT().Get(ctx, id, belongsTo, repository.TeamDetailProjection()).Return(nil, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeTeam),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "get cached team error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ *repository.Team) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(repository.TeamDetailProjection()))

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(assert.AnError)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, _ context.Context, _, _ model.ID, _ *repository.Team) repository.TeamRepository {
					return mockrepo.NewMockTeamRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				id:        model.MustNewID(model.ResourceTypeTeam),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
			},
			wantErr: repository.ErrCacheRead,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			team := &repository.Team{}
			if tt.want != nil {
				team = tt.want(tt.args.id)
			}

			r := func() *repository.RedisCachedTeamRepository {
				r, err := repository.NewCachedTeamRepository(
					tt.fields.teamRepo(ctrl, tt.args.ctx, tt.args.id, tt.args.belongsTo, team),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.id, team)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.Get(tt.args.ctx, tt.args.id, tt.args.belongsTo, repository.TeamDetailProjection())
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
		cacheRepo func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, teams []*repository.Team) []repository.RedisRepositoryOption
		teamRepo  func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, teams []*repository.Team) repository.TeamRepository
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
		want    []*repository.Team
		wantErr error
	}{
		{
			name: "get uncached teams",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, teams []*repository.Team) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(repository.TeamListProjection()), "", limit)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)
					cacheRepo.EXPECT().Set(&cache.Item{
						Ctx:   ctx,
						Key:   key,
						Value: repository.Page[*repository.Team]{Items: teams},
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, teams []*repository.Team) repository.TeamRepository {
					repo := mockrepo.NewMockTeamRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, repository.CursorPage{Size: limit}, repository.TeamListProjection()).Return(repository.Page[*repository.Team]{Items: teams}, nil)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				limit:     10,
			},
			want: []*repository.Team{
				{ID: model.MustNewID(model.ResourceTypeTeam), Name: "team-one"},
				{ID: model.MustNewID(model.ResourceTypeTeam), Name: "team-two"},
			},
		},
		{
			name: "get cached teams",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, teams []*repository.Team) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(repository.TeamListProjection()), "", limit)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Do(func(_ context.Context, _ string, dst any) {
						if ptr, ok := dst.(*repository.Page[*repository.Team]); ok {
							*ptr = repository.Page[*repository.Team]{Items: teams}
						}
					}).Return(nil)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, _ context.Context, _ model.ID, _ int, _ []*repository.Team) repository.TeamRepository {
					return mockrepo.NewMockTeamRepository(ctrl)
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				limit:     10,
			},
			want: []*repository.Team{
				{ID: model.MustNewID(model.ResourceTypeTeam), Name: "team-one"},
			},
		},
		{
			name: "get uncached teams error",
			fields: fields{
				cacheRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, _ []*repository.Team) []repository.RedisRepositoryOption {
					key := composeCacheKey(model.ResourceTypeTeam.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(repository.TeamListProjection()), "", limit)

					db, err := repository.NewRedisDatabase(
						repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)),
					)
					require.NoError(t, err)

					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(1)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)

					cacheRepo := mockrepo.NewMockCacheBackend(ctrl)
					cacheRepo.EXPECT().Get(ctx, key, gomock.Any()).Return(cache.ErrCacheMiss)

					return []repository.RedisRepositoryOption{
						repository.WithRedisDatabase(db),
						repository.WithCacheBackend(cacheRepo),
						repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
						repository.WithRedisRepositoryTracer(tracer),
					}
				},
				teamRepo: func(ctrl *gomock.Controller, ctx context.Context, belongsTo model.ID, limit int, _ []*repository.Team) repository.TeamRepository {
					repo := mockrepo.NewMockTeamRepository(ctrl)
					repo.EXPECT().ListBelongsTo(ctx, belongsTo, repository.CursorPage{Size: limit}, repository.TeamListProjection()).Return(repository.Page[*repository.Team]{}, repository.ErrNotFound)
					return repo
				},
			},
			args: args{
				ctx:       context.Background(),
				belongsTo: model.MustNewID(model.ResourceTypeOrganization),
				limit:     10,
			},
			wantErr: repository.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := func() *repository.RedisCachedTeamRepository {
				r, err := repository.NewCachedTeamRepository(
					tt.fields.teamRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.limit, tt.want),
					tt.fields.cacheRepo(ctrl, tt.args.ctx, tt.args.belongsTo, tt.args.limit, tt.want)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			got, err := r.ListBelongsTo(tt.args.ctx, tt.args.belongsTo, repository.CursorPage{Size: tt.args.limit}, repository.TeamListProjection())
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
	page := repository.CursorPage{Size: 10}
	members := repository.Page[*repository.User]{
		Items: []*repository.User{{ID: model.MustNewID(model.ResourceTypeUser)}},
	}

	t.Run("passthrough without cache", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mockrepo.NewMockTeamRepository(ctrl)
		repo.EXPECT().ListMembers(ctx, teamID, belongsTo, page).Return(members, nil)

		r := func() *repository.RedisCachedTeamRepository {
			r, err := repository.NewCachedTeamRepository(
				repo,
				redisRepoOptsNoop(ctrl)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		got, err := r.ListMembers(ctx, teamID, belongsTo, page)
		require.NoError(t, err)
		require.Equal(t, members, got)
	})

	t.Run("passthrough error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mockrepo.NewMockTeamRepository(ctrl)
		repo.EXPECT().ListMembers(ctx, teamID, belongsTo, page).Return(repository.Page[*repository.User]{}, repository.ErrNotFound)

		r := func() *repository.RedisCachedTeamRepository {
			r, err := repository.NewCachedTeamRepository(
				repo,
				redisRepoOptsNoop(ctrl)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		_, err := r.ListMembers(ctx, teamID, belongsTo, page)
		require.ErrorIs(t, err, repository.ErrNotFound)
	})
}

func TestCachedTeamRepository_Update(t *testing.T) {
	newTeam := func(id model.ID) *repository.Team {
		return &repository.Team{ID: id, Name: "test team", Description: "test description"}
	}
	opts := repository.UpdateTeamOpts{
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
		setKey := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(repository.TeamDetailProjection()))
		repo := mockrepo.NewMockTeamRepository(ctrl)
		repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(team, nil)
		r := func() *repository.RedisCachedTeamRepository {
			r, err := repository.NewCachedTeamRepository(
				repo,
				redisCacheExpectingSetThenPatterns(ctrl, ctx, setKey, team, teamUpdateInvalidatePatterns(), false, -1, nil)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
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
		db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
		require.NoError(t, err)
		repo := mockrepo.NewMockTeamRepository(ctrl)
		repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(nil, repository.ErrNotFound)
		r := func() *repository.RedisCachedTeamRepository {
			r, err := repository.NewCachedTeamRepository(
				repo,
				[]repository.RedisRepositoryOption{
					repository.WithRedisDatabase(db),
					repository.WithCacheBackend(mockrepo.NewMockCacheBackend(ctrl)),
					repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
					repository.WithRedisRepositoryTracer(mocktrace.NewMockTracer(ctrl)),
				}...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		_, err = r.Update(ctx, id, belongsTo, opts)
		require.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("update team set cache error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := context.Background()
		id := model.MustNewID(model.ResourceTypeTeam)
		belongsTo := model.MustNewID(model.ResourceTypeOrganization)
		team := newTeam(id)
		setKey := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(repository.TeamDetailProjection()))
		repo := mockrepo.NewMockTeamRepository(ctrl)
		repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(team, nil)
		r := func() *repository.RedisCachedTeamRepository {
			r, err := repository.NewCachedTeamRepository(
				repo,
				redisCacheExpectingSetThenPatterns(ctrl, ctx, setKey, team, nil, true, -1, assert.AnError)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		_, err := r.Update(ctx, id, belongsTo, opts)
		require.ErrorIs(t, err, repository.ErrCacheWrite)
	})

	t.Run("update team delete list cache error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := context.Background()
		id := model.MustNewID(model.ResourceTypeTeam)
		belongsTo := model.MustNewID(model.ResourceTypeOrganization)
		team := newTeam(id)
		setKey := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(repository.TeamDetailProjection()))
		repo := mockrepo.NewMockTeamRepository(ctrl)
		repo.EXPECT().Update(ctx, id, belongsTo, opts).Return(team, nil)
		r := func() *repository.RedisCachedTeamRepository {
			r, err := repository.NewCachedTeamRepository(
				repo,
				redisCacheExpectingSetThenPatterns(ctrl, ctx, setKey, team, teamUpdateInvalidatePatterns(), false, 0, assert.AnError)...,
			)
			if err != nil {
				panic(err)
			}
			return r
		}()
		_, err := r.Update(ctx, id, belongsTo, opts)
		require.ErrorIs(t, err, repository.ErrCacheDelete)
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
		{name: "add member with team error", failIndex: -1, repoErr: repository.ErrTeamAddMember, wantErr: repository.ErrTeamAddMember},
		{name: "add member with cache deletion error", failIndex: 0, failErr: repository.ErrCacheDelete, wantErr: repository.ErrCacheDelete},
		{name: "add member with related cache deletion error", failIndex: 2, failErr: repository.ErrCacheDelete, wantErr: repository.ErrCacheDelete},
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
			repo := mockrepo.NewMockTeamRepository(ctrl)
			if tt.failIndex < 0 {
				repo.EXPECT().AddMember(ctx, id, memberID, belongsToID).Return(tt.repoErr)
			}
			bumpCount := 1
			if tt.repoErr != nil {
				bumpCount = 0
			}
			r := func() *repository.RedisCachedTeamRepository {
				r, err := repository.NewCachedTeamRepository(
					repo,
					redisCacheExpectingPatternsThenIssueAuthzEpochBump(ctrl, ctx, teamMemberCachePatterns(id, belongsToID), tt.failIndex, tt.failErr, bumpCount)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
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
		{name: "remove member with team error", failIndex: -1, repoErr: repository.ErrTeamRemoveMember, wantErr: repository.ErrTeamRemoveMember},
		{name: "remove member with cache deletion error", failIndex: 0, failErr: repository.ErrCacheDelete, wantErr: repository.ErrCacheDelete},
		{name: "remove member with related cache deletion error", failIndex: 2, failErr: repository.ErrCacheDelete, wantErr: repository.ErrCacheDelete},
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
			repo := mockrepo.NewMockTeamRepository(ctrl)
			if tt.failIndex < 0 {
				repo.EXPECT().RemoveMember(ctx, id, memberID, belongsToID).Return(tt.repoErr)
			}
			bumpCount := 1
			if tt.repoErr != nil {
				bumpCount = 0
			}
			r := func() *repository.RedisCachedTeamRepository {
				r, err := repository.NewCachedTeamRepository(
					repo,
					redisCacheExpectingPatternsThenIssueAuthzEpochBump(ctrl, ctx, teamMemberCachePatterns(id, belongsToID), tt.failIndex, tt.failErr, bumpCount)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
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
		{name: "delete team with team deletion error", failIndex: -1, repoErr: repository.ErrTeamDelete, wantErr: repository.ErrTeamDelete},
		{name: "delete team with cache deletion error", failIndex: 0, failErr: repository.ErrCacheDelete, wantErr: repository.ErrCacheDelete},
		{name: "delete team with list cache deletion error", failIndex: 1, failErr: repository.ErrCacheDelete, wantErr: repository.ErrCacheDelete},
		{name: "delete team with organization cache deletion error", failIndex: 2, failErr: repository.ErrCacheDelete, wantErr: repository.ErrCacheDelete},
		{name: "delete team with project cache deletion error", failIndex: 3, failErr: repository.ErrCacheDelete, wantErr: repository.ErrCacheDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := context.Background()
			id := model.MustNewID(model.ResourceTypeTeam)
			belongsTo := model.MustNewID(model.ResourceTypeOrganization)
			repo := mockrepo.NewMockTeamRepository(ctrl)
			if tt.failIndex < 0 {
				repo.EXPECT().Delete(ctx, id, belongsTo).Return(tt.repoErr)
			}
			bumpCount := 1
			if tt.repoErr != nil {
				bumpCount = 0
			}
			r := func() *repository.RedisCachedTeamRepository {
				r, err := repository.NewCachedTeamRepository(
					repo,
					redisCacheExpectingPatternsThenIssueAuthzEpochBump(ctrl, ctx, teamDeleteCachePatterns(id), tt.failIndex, tt.failErr, bumpCount)...,
				)
				if err != nil {
					panic(err)
				}
				return r
			}()
			err := r.Delete(ctx, id, belongsTo)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
