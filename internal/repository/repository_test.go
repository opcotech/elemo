package repository

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/go-redis/cache/v9"
	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func testPageSize(limit int) int {
	if limit < MinPageSize {
		return DefaultPageSize
	}
	return limit
}

func issueListForIssueCacheKey(issueID model.ID, page CursorPage, proj IssueProjection) string {
	plan, err := CompileQuery(IssueListForIssueQuery{IssueID: issueID, Page: page, Projection: proj})
	if err != nil {
		panic(err)
	}
	return plan.CacheKey(model.ResourceTypeIssue.String(), "ListForIssue", issueID.String())
}

// redisCacheExpectingPatterns mocks Keys+DeletePattern for each pattern in order.
// If failIndex >= 0, that pattern's cache.Delete returns failErr and later patterns are not expected.
//
//nolint:revive // test cache factories take gomock.Controller first
func redisCacheExpectingPatterns(ctrl *gomock.Controller, ctx context.Context, patterns []string, failIndex int, failErr error) *redisBaseRepository {
	client := mock.NewUniversalClient(ctrl)
	backend := mock.NewCacheBackend(ctrl)
	span := mock.NewMockSpan(ctrl)
	tracer := mock.NewMockTracer(ctrl)

	count := 0
	for i, pattern := range patterns {
		count++
		cmd := new(redis.StringSliceCmd)
		cmd.SetVal([]string{pattern})
		client.EXPECT().Keys(ctx, pattern).Return(cmd)
		if i == failIndex {
			backend.EXPECT().Delete(ctx, pattern).Return(failErr)
			break
		}
		backend.EXPECT().Delete(ctx, pattern).Return(nil)
	}

	span.EXPECT().End(gomock.Len(0)).Times(count)
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(count)

	db, err := NewRedisDatabase(WithRedisClient(client))
	if err != nil {
		panic(err)
	}

	return &redisBaseRepository{
		db:     db,
		cache:  backend,
		tracer: tracer,
		logger: mock.NewMockLogger(ctrl),
	}
}

//nolint:revive // test cache factories take gomock.Controller first
func redisCacheExpectingSetThenPatterns(ctrl *gomock.Controller, ctx context.Context, setKey string, setValue any, patterns []string, failOnSet bool, failIndex int, failErr error) *redisBaseRepository {
	client := mock.NewUniversalClient(ctrl)
	backend := mock.NewCacheBackend(ctrl)
	span := mock.NewMockSpan(ctrl)
	tracer := mock.NewMockTracer(ctrl)

	count := 1
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
	if failOnSet {
		backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: setKey, Value: setValue}).Return(failErr)
		span.EXPECT().End(gomock.Len(0)).Times(1)
		db, err := NewRedisDatabase(WithRedisClient(client))
		if err != nil {
			panic(err)
		}
		return &redisBaseRepository{
			db:     db,
			cache:  backend,
			tracer: tracer,
			logger: mock.NewMockLogger(ctrl),
		}
	}

	backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: setKey, Value: setValue}).Return(nil)
	for i, pattern := range patterns {
		count++
		cmd := new(redis.StringSliceCmd)
		cmd.SetVal([]string{pattern})
		client.EXPECT().Keys(ctx, pattern).Return(cmd)
		if i == failIndex {
			backend.EXPECT().Delete(ctx, pattern).Return(failErr)
			break
		}
		backend.EXPECT().Delete(ctx, pattern).Return(nil)
	}

	span.EXPECT().End(gomock.Len(0)).Times(count)
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(count - 1)

	db, err := NewRedisDatabase(WithRedisClient(client))
	if err != nil {
		panic(err)
	}

	return &redisBaseRepository{
		db:     db,
		cache:  backend,
		tracer: tracer,
		logger: mock.NewMockLogger(ctrl),
	}
}

//nolint:revive // test cache factories take gomock.Controller first
func redisCacheExpectingBumpThenPatterns(ctrl *gomock.Controller, ctx context.Context, principal model.ID, patterns []string) *redisBaseRepository {
	client := mock.NewUniversalClient(ctrl)
	backend := mock.NewCacheBackend(ctrl)
	span := mock.NewMockSpan(ctrl)
	tracer := mock.NewMockTracer(ctrl)

	genKey := authzGenKey(principal)
	count := 2 + len(patterns)
	span.EXPECT().End(gomock.Len(0)).Times(count)
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(len(patterns))

	backend.EXPECT().Get(ctx, genKey, gomock.Any()).Return(cache.ErrCacheMiss)
	backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: genKey, Value: int64(1)}).Return(nil)
	for _, pattern := range patterns {
		cmd := new(redis.StringSliceCmd)
		cmd.SetVal([]string{pattern})
		client.EXPECT().Keys(ctx, pattern).Return(cmd)
		backend.EXPECT().Delete(ctx, pattern).Return(nil)
	}

	db, err := NewRedisDatabase(WithRedisClient(client))
	if err != nil {
		panic(err)
	}
	return &redisBaseRepository{
		db:     db,
		cache:  backend,
		tracer: tracer,
		logger: mock.NewMockLogger(ctrl),
	}
}

func teamCreateCachePatterns(belongsTo model.ID) []string {
	return []string{
		composeCacheKey(model.ResourceTypeTeam.String(), "ListBelongsTo", belongsTo.String(), "*", "*", "*"),
		composeCacheKey(model.ResourceTypeOrganization.String(), "*"),
		composeCacheKey(model.ResourceTypeProject.String(), "*"),
	}
}

func teamDeleteCachePatterns(id model.ID) []string {
	return []string{
		composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), "*"),
		composeCacheKey(model.ResourceTypeTeam.String(), "ListBelongsTo", "*"),
		composeCacheKey(model.ResourceTypeOrganization.String(), "*"),
		composeCacheKey(model.ResourceTypeProject.String(), "*"),
	}
}

func teamMemberCachePatterns(id, belongsToID model.ID) []string {
	return []string{
		composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), "*"),
		composeCacheKey(model.ResourceTypeTeam.String(), "ListBelongsTo", "*"),
		composeCacheKey(model.ResourceTypeOrganization.String(), "Get", belongsToID.String(), "*"),
		composeCacheKey(model.ResourceTypeProject.String(), "*", "Get", belongsToID.String(), "*"),
	}
}

func teamUpdateInvalidatePatterns() []string {
	return []string{
		composeCacheKey(model.ResourceTypeTeam.String(), "ListBelongsTo", "*"),
	}
}

func roleCreateCachePatterns(belongsTo model.ID) []string {
	return []string{
		composeCacheKey(model.ResourceTypeRole.String(), "ListBelongsTo", belongsTo.String(), "*", "*", "*"),
		composeCacheKey(model.ResourceTypeRole.String(), "GetByKey", belongsTo.String(), "*"),
		composeCacheKey(model.ResourceTypeOrganization.String(), "*"),
		composeCacheKey(model.ResourceTypeProject.String(), "*"),
	}
}

func roleDeleteCachePatterns(id, belongsTo model.ID) []string {
	return []string{
		composeCacheKey(model.ResourceTypeRole.String(), "Get", id.String(), "*"),
		composeCacheKey(model.ResourceTypeRole.String(), "GetByID", id.String()),
		composeCacheKey(model.ResourceTypeRole.String(), "GetByKey", belongsTo.String(), "*"),
		composeCacheKey(model.ResourceTypeRole.String(), "ListBelongsTo", "*"),
		composeCacheKey(model.ResourceTypeOrganization.String(), "*"),
		composeCacheKey(model.ResourceTypeProject.String(), "*"),
	}
}

func roleMemberCachePatterns(id, belongsToID model.ID) []string {
	return []string{
		composeCacheKey(model.ResourceTypeRole.String(), "Get", id.String(), "*"),
		composeCacheKey(model.ResourceTypeRole.String(), "GetByID", id.String()),
		composeCacheKey(model.ResourceTypeRole.String(), "ListBelongsTo", "*"),
		composeCacheKey(model.ResourceTypeOrganization.String(), "Get", belongsToID.String(), "*"),
	}
}

func roleUpdateInvalidatePatterns(id, belongsTo model.ID) []string {
	return []string{
		composeCacheKey(model.ResourceTypeRole.String(), "GetByID", id.String()),
		composeCacheKey(model.ResourceTypeRole.String(), "GetByKey", belongsTo.String(), "*"),
		composeCacheKey(model.ResourceTypeRole.String(), "ListBelongsTo", "*"),
	}
}

func permissionCrossCachePatterns() []string {
	return []string{
		composeCacheKey(model.ResourceTypeRole.String(), "*"),
		composeCacheKey(model.ResourceTypeUser.String(), "*"),
		composeCacheKey(model.ResourceTypeOrganization.String(), "List", "*", "*", "*", "*"),
		composeCacheKey(model.ResourceTypeNamespace.String(), "List", "*", "*", "*", "*"),
		composeCacheKey(model.ResourceTypeNamespace.String(), "ListAccessible", "*", "*", "*", "*"),
		composeCacheKey(model.ResourceTypeProject.String(), "*", "List", "*"),
		composeCacheKey(model.ResourceTypeIssue.String(), "*", "ListFor*", "*"),
		composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", "*"),
		composeCacheKey(model.ResourceTypeDocument.String(), "ListRelated", "*"),
		composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*"),
		composeCacheKey(model.ResourceTypeFolder.String(), "List", "*"),
	}
}

func TestEdgeKind_String(t *testing.T) {
	tests := []struct {
		name string
		s    EdgeKind
		want string
	}{
		{"ASSIGNED_TO", EdgeKindAssignedTo, "ASSIGNED_TO"},
		{"BELONGS_TO", EdgeKindBelongsTo, "BELONGS_TO"},
		{"COMMENTED", EdgeKindCommented, "COMMENTED"},
		{"CREATED", EdgeKindCreated, "CREATED"},
		{"HAS_ATTACHMENT", EdgeKindHasAttachment, "HAS_ATTACHMENT"},
		{"HAS_COMMENT", EdgeKindHasComment, "HAS_COMMENT"},
		{"HAS_LABEL", EdgeKindHasLabel, "HAS_LABEL"},
		{"HAS_NAMESPACE", EdgeKindHasNamespace, "HAS_NAMESPACE"},
		{"HAS_PERMISSION", EdgeKindHasPermission, "HAS_PERMISSION"},
		{"HAS_PROJECT", EdgeKindHasProject, "HAS_PROJECT"},
		{"HAS_TEAM", EdgeKindHasTeam, "HAS_TEAM"},
		{"INVITED", EdgeKindInvited, "INVITED"},
		{"INVITED_TO", EdgeKindInvitedTo, "INVITED_TO"},
		{"KIND_OF", EdgeKindKindOf, "KIND_OF"},
		{"MEMBER_OF", EdgeKindMemberOf, "MEMBER_OF"},
		{"RELATED_TO", EdgeKindRelatedTo, "RELATED_TO"},
		{"SPEAKS", EdgeKindSpeaks, "SPEAKS"},
		{"WATCHES", EdgeKindWatches, "WATCHES"},
		{"SCOPED_TO", EdgeKindScopedTo, "SCOPED_TO"},
		{"LOCATED_IN", EdgeKindLocatedIn, "LOCATED_IN"},
		{"IN_SCOPE_OF", EdgeKindInScopeOf, "IN_SCOPE_OF"},
		{"GRANTED", EdgeKindGranted, "GRANTED"},
		{"DEFINES_ROLE", EdgeKindDefinesRole, "DEFINES_ROLE"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}

func TestNewPGPool(t *testing.T) {
	type args struct {
		ctx  context.Context
		conf *config.RelationalDatabaseConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create new PG pool",
			args: args{
				ctx: context.Background(),
				conf: &config.RelationalDatabaseConfig{
					Host:           "localhost",
					Port:           5432,
					Username:       "postgres",
					Password:       "postgres",
					Database:       "postgres",
					MaxConnections: 10,
				},
			},
		},
		{
			name: "create new PG pool with invalid mac connections",
			args: args{
				ctx: context.Background(),
				conf: &config.RelationalDatabaseConfig{
					Host:           "localhost",
					Port:           5432,
					Username:       "postgres",
					Password:       "postgres",
					Database:       "postgres",
					MaxConnections: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "create new PG pool with invalid config",
			args: args{
				ctx:  context.Background(),
				conf: &config.RelationalDatabaseConfig{},
			},
			wantErr: true,
		},
		{
			name: "create new PG pool with nil config",
			args: args{
				ctx:  context.Background(),
				conf: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewPool(tt.args.ctx, tt.args.conf)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewPGPoolPreservesDefaultLifetimes(t *testing.T) {
	t.Parallel()

	pool, err := NewPool(context.Background(), &config.RelationalDatabaseConfig{
		Host:           "localhost",
		Port:           5432,
		Username:       "postgres",
		Password:       "postgres",
		Database:       "postgres",
		MaxConnections: 10,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	conf := pool.Config()
	require.NotNil(t, conf)
	assert.Positive(t, conf.MaxConnLifetime)
	assert.Positive(t, conf.MaxConnIdleTime)
}

func TestNewPGPoolAppliesConfiguredLifetimes(t *testing.T) {
	t.Parallel()

	pool, err := NewPool(context.Background(), &config.RelationalDatabaseConfig{
		Host:                  "localhost",
		Port:                  5432,
		Username:              "postgres",
		Password:              "postgres",
		Database:              "postgres",
		MaxConnections:        10,
		MaxConnectionLifetime: 300,
		MaxConnectionIdleTime: 10,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	conf := pool.Config()
	require.NotNil(t, conf)
	assert.Equal(t, 300*time.Second, conf.MaxConnLifetime)
	assert.Equal(t, 10*time.Second, conf.MaxConnIdleTime)
}

func TestWithDatabasePool(t *testing.T) {
	type args struct {
		pool PGPool
	}
	tests := []struct {
		name    string
		args    args
		want    PGPool
		wantErr error
	}{
		{
			name: "create new option with pool",
			args: args{
				pool: mock.NewPGPool(nil),
			},
			want: mock.NewPGPool(nil),
		},
		{
			name: "create new option with nil pool",
			args: args{
				pool: nil,
			},
			wantErr: ErrNoPool,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := new(PGDatabase)
			err := WithDatabasePool(tt.args.pool)(db)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db.pool)
		})
	}
}

func TestWithPGDatabaseLogger(t *testing.T) {
	type args struct {
		logger log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    log.Logger
		wantErr error
	}{
		{
			name: "create new option with logger",
			args: args{
				logger: mock.NewMockLogger(nil),
			},
			want: mock.NewMockLogger(nil),
		},
		{
			name: "create new option with nil logger",
			args: args{
				logger: nil,
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := new(PGDatabase)
			err := WithPGDatabaseLogger(tt.args.logger)(db)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db.logger)
		})
	}
}

func TestWithPGDatabaseTracer(t *testing.T) {
	type args struct {
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    tracing.Tracer
		wantErr error
	}{
		{
			name: "create new option with tracer",
			args: args{
				tracer: mock.NewMockTracer(nil),
			},
			want: mock.NewMockTracer(nil),
		},
		{
			name: "create new option with nil tracer",
			args: args{
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := new(PGDatabase)
			err := WithPGDatabaseTracer(tt.args.tracer)(db)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db.tracer)
		})
	}
}

func TestNewPGDatabase(t *testing.T) {
	type args struct {
		pool   PGPool
		logger log.Logger
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    *PGDatabase
		wantErr error
	}{
		{
			name: "create new database",
			args: args{
				pool:   mock.NewPGPool(nil),
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
			want: &PGDatabase{
				pool:   mock.NewPGPool(nil),
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
		},
		{
			name: "create new database with nil pool",
			args: args{
				pool:   nil,
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
			wantErr: ErrNoPool,
		},
		{
			name: "create new database with nil logger",
			args: args{
				pool:   mock.NewPGPool(nil),
				logger: nil,
				tracer: mock.NewMockTracer(nil),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new database with nil tracer",
			args: args{
				pool:   mock.NewPGPool(nil),
				logger: mock.NewMockLogger(nil),
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, err := NewPGDatabase(
				WithDatabasePool(tt.args.pool),
				WithPGDatabaseLogger(tt.args.logger),
				WithPGDatabaseTracer(tt.args.tracer),
			)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db)
		})
	}
}

func TestPGDatabase_Close(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	pool := mock.NewPGPool(ctrl)
	pool.EXPECT().Close()

	db := &PGDatabase{
		pool: pool,
	}

	require.NoError(t, db.Close())
}

func TestDatabase_Pool(t *testing.T) {
	t.Parallel()

	pool := mock.NewPGPool(nil)

	db := &PGDatabase{
		pool: pool,
	}

	require.Equal(t, pool, db.Pool())
}

func TestPGDatabase_Ping(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type fields struct {
		pool func(ctx context.Context, ctrl *gomock.Controller) PGPool
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name: "ping database",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				pool: func(ctx context.Context, ctrl *gomock.Controller) PGPool {
					p := mock.NewPGPool(ctrl)
					p.EXPECT().Ping(ctx).Return(nil)
					return p
				},
			},
		},
		{
			name: "ping database with error",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				pool: func(ctx context.Context, ctrl *gomock.Controller) PGPool {
					p := mock.NewPGPool(ctrl)
					p.EXPECT().Ping(ctx).Return(assert.AnError)
					return p
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			db := &PGDatabase{
				pool: tt.fields.pool(tt.args.ctx, ctrl),
			}
			err := db.Ping(tt.args.ctx)
			if !tt.wantErr && err != nil {
				require.Error(t, err)
			}
		})
	}
}

func TestNewRedisClient(t *testing.T) {
	type args struct {
		conf *config.CacheDatabaseConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "create new redis client",
			args: args{
				conf: &config.CacheDatabaseConfig{
					RedisConfig: config.RedisConfig{
						Host:         "localhost",
						Port:         6379,
						Username:     "default",
						Password:     "redisSecret",
						Database:     0,
						IsSecure:     false,
						DialTimeout:  10,
						ReadTimeout:  10,
						WriteTimeout: 10,
						PoolSize:     10,
					},
					MaxIdleConnections:    10,
					MinIdleConnections:    10,
					ConnectionMaxIdleTime: 10,
					ConnectionMaxLifetime: 10,
				},
			},
		},
		{
			name: "create new redis client with no config",
			args: args{
				conf: nil,
			},
			wantErr: config.ErrNoConfig,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewRedisClient(tt.args.conf)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestWithDatabaseClient(t *testing.T) {
	type args struct {
		client redis.UniversalClient
	}
	tests := []struct {
		name    string
		args    args
		want    redis.UniversalClient
		wantErr error
	}{
		{
			name: "create new option with client",
			args: args{
				client: func() redis.UniversalClient {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mock.NewUniversalClient(ctrl)
				}(),
			},
			want: func() redis.UniversalClient {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				return mock.NewUniversalClient(ctrl)
			}(),
		},
		{
			name: "create new option with nil client",
			args: args{
				client: nil,
			},
			wantErr: ErrNoClient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := new(RedisDatabase)
			err := WithRedisClient(tt.args.client)(db)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db.client)
		})
	}
}

func TestWithRedisDatabaseLogger(t *testing.T) {
	type args struct {
		logger log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    log.Logger
		wantErr error
	}{
		{
			name: "create new option with logger",
			args: args{
				logger: mock.NewMockLogger(nil),
			},
			want: mock.NewMockLogger(nil),
		},
		{
			name: "create new option with nil logger",
			args: args{
				logger: nil,
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := new(RedisDatabase)
			err := WithRedisDatabaseLogger(tt.args.logger)(db)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db.logger)
		})
	}
}

func TestWithRedisDatabaseTracer(t *testing.T) {
	type args struct {
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    tracing.Tracer
		wantErr error
	}{
		{
			name: "create new option with tracer",
			args: args{
				tracer: mock.NewMockTracer(nil),
			},
			want: mock.NewMockTracer(nil),
		},
		{
			name: "create new option with nil tracer",
			args: args{
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := new(RedisDatabase)
			err := WithRedisDatabaseTracer(tt.args.tracer)(db)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db.tracer)
		})
	}
}

func TestNewRedisDatabase(t *testing.T) {
	type args struct {
		client redis.UniversalClient
		logger log.Logger
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    *RedisDatabase
		wantErr error
	}{
		{
			name: "create new database",
			args: args{
				client: func() redis.UniversalClient {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mock.NewUniversalClient(ctrl)
				}(),
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
			want: &RedisDatabase{
				client: func() redis.UniversalClient {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mock.NewUniversalClient(ctrl)
				}(),
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
		},
		{
			name: "create new database with nil client",
			args: args{
				client: nil,
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
			wantErr: ErrNoClient,
		},
		{
			name: "create new database with nil logger",
			args: args{
				client: func() redis.UniversalClient {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mock.NewUniversalClient(ctrl)
				}(),
				logger: nil,
				tracer: mock.NewMockTracer(nil),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new database with nil tracer",
			args: args{
				client: func() redis.UniversalClient {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					return mock.NewUniversalClient(ctrl)
				}(),
				logger: mock.NewMockLogger(nil),
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, err := NewRedisDatabase(
				WithRedisClient(tt.args.client),
				WithRedisDatabaseLogger(tt.args.logger),
				WithRedisDatabaseTracer(tt.args.tracer),
			)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, db)
		})
	}
}

func TestDatabase_Client(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewUniversalClient(ctrl)

	db := &RedisDatabase{
		client: client,
	}

	require.Equal(t, client, db.Client())
}

func TestRedisDatabase_Close(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewUniversalClient(ctrl)
	client.EXPECT().Close().Return(nil)

	db := &RedisDatabase{
		client: client,
	}

	require.NoError(t, db.Close())
}

func TestRedisDatabase_Ping(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type fields struct {
		client func(ctrl *gomock.Controller, ctx context.Context) redis.UniversalClient
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name: "ping database",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				client: func(ctrl *gomock.Controller, ctx context.Context) redis.UniversalClient {
					p := mock.NewUniversalClient(ctrl)
					p.EXPECT().Ping(ctx).Return(&redis.StatusCmd{})
					return p
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			db := &RedisDatabase{
				client: tt.fields.client(ctrl, tt.args.ctx),
			}
			err := db.Ping(tt.args.ctx)
			if !tt.wantErr && err != nil {
				require.Error(t, err)
			}
		})
	}
}

func TestNewS3Client(t *testing.T) {
	type args struct {
		ctx  context.Context
		conf *config.S3StorageConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "create new S3 client",
			args: args{
				ctx: context.Background(),
				conf: &config.S3StorageConfig{
					Region:          "us-east-1",
					AccessKeyID:     "test-access-key",
					SecretAccessKey: "test-secret-key",
					BaseEndpoint:    "http://localhost:9000",
				},
			},
		},
		{
			name: "create new S3 client with no config",
			args: args{
				ctx:  context.Background(),
				conf: nil,
			},
			wantErr: config.ErrNoConfig,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewS3Client(tt.args.ctx, tt.args.conf)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestWithStorageClient(t *testing.T) {
	type args struct {
		client S3Client
	}
	tests := []struct {
		name    string
		args    args
		want    S3Client
		wantErr error
	}{
		{
			name: "create new option with client",
			args: args{
				client: mock.NewS3Client(nil),
			},
			want: mock.NewS3Client(nil),
		},
		{
			name: "create new option with nil client",
			args: args{
				client: nil,
			},
			wantErr: ErrNoClient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage := new(S3Storage)
			err := WithStorageClient(tt.args.client)(storage)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, storage.client)
		})
	}
}

func TestWithStorageBucket(t *testing.T) {
	type args struct {
		bucket string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			name: "create new option with bucket",
			args: args{
				bucket: "test-bucket",
			},
			want: "test-bucket",
		},
		{
			name: "create new option with empty bucket",
			args: args{
				bucket: "",
			},
			wantErr: ErrNoBucket,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage := new(S3Storage)
			err := WithStorageBucket(tt.args.bucket)(storage)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, storage.bucket)
		})
	}
}

func TestWithStorageLogger(t *testing.T) {
	type args struct {
		logger log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    log.Logger
		wantErr error
	}{
		{
			name: "create new option with logger",
			args: args{
				logger: mock.NewMockLogger(nil),
			},
			want: mock.NewMockLogger(nil),
		},
		{
			name: "create new option with nil logger",
			args: args{
				logger: nil,
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage := new(S3Storage)
			err := WithStorageLogger(tt.args.logger)(storage)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, storage.logger)
		})
	}
}

func TestWithStorageTracer(t *testing.T) {
	type args struct {
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    tracing.Tracer
		wantErr error
	}{
		{
			name: "create new option with tracer",
			args: args{
				tracer: mock.NewMockTracer(nil),
			},
			want: mock.NewMockTracer(nil),
		},
		{
			name: "create new option with nil tracer",
			args: args{
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage := new(S3Storage)
			err := WithStorageTracer(tt.args.tracer)(storage)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, storage.tracer)
		})
	}
}

func TestNewStorage(t *testing.T) {
	type args struct {
		client S3Client
		bucket string
		logger log.Logger
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    *S3Storage
		wantErr error
	}{
		{
			name: "create new storage",
			args: args{
				client: mock.NewS3Client(nil),
				bucket: "test-bucket",
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
			want: &S3Storage{
				client: mock.NewS3Client(nil),
				bucket: "test-bucket",
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
		},
		{
			name: "create new storage with nil client",
			args: args{
				client: nil,
				bucket: "test-bucket",
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
			wantErr: ErrNoClient,
		},
		{
			name: "create new storage with empty bucket",
			args: args{
				client: mock.NewS3Client(nil),
				bucket: "",
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
			wantErr: ErrNoBucket,
		},
		{
			name: "create new storage with nil logger",
			args: args{
				client: mock.NewS3Client(nil),
				bucket: "test-bucket",
				logger: nil,
				tracer: mock.NewMockTracer(nil),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new storage with nil tracer",
			args: args{
				client: mock.NewS3Client(nil),
				bucket: "test-bucket",
				logger: mock.NewMockLogger(nil),
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage, err := NewStorage(
				WithStorageClient(tt.args.client),
				WithStorageBucket(tt.args.bucket),
				WithStorageLogger(tt.args.logger),
				WithStorageTracer(tt.args.tracer),
			)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, storage)
		})
	}
}

func TestStorage_Client(t *testing.T) {
	t.Parallel()

	client := mock.NewS3Client(nil)

	storage := &S3Storage{
		client: client,
	}

	require.Equal(t, client, storage.Client())
}

func TestStorage_Ping(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type fields struct {
		client func(ctx context.Context, ctrl *gomock.Controller) S3Client
		bucket string
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name: "ping storage",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				client: func(ctx context.Context, ctrl *gomock.Controller) S3Client {
					c := mock.NewS3Client(ctrl)
					c.EXPECT().HeadBucket(ctx, &awsS3.HeadBucketInput{Bucket: aws.String("test-bucket")}, gomock.Any()).Return(&awsS3.HeadBucketOutput{}, nil)
					return c
				},
				bucket: "test-bucket",
			},
		},
		{
			name: "ping storage with error",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				client: func(ctx context.Context, ctrl *gomock.Controller) S3Client {
					c := mock.NewS3Client(ctrl)
					c.EXPECT().HeadBucket(ctx, &awsS3.HeadBucketInput{Bucket: aws.String("test-bucket")}, gomock.Any()).Return(&awsS3.HeadBucketOutput{}, assert.AnError)
					return c
				},
				bucket: "test-bucket",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storage := &S3Storage{
				client: tt.fields.client(tt.args.ctx, ctrl),
				bucket: tt.fields.bucket,
			}
			err := storage.Ping(tt.args.ctx)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWithStorage(t *testing.T) {
	type args struct {
		storage *S3Storage
	}
	tests := []struct {
		name    string
		args    args
		want    *S3Storage
		wantErr error
	}{
		{
			name: "create new option with storage",
			args: args{
				storage: &S3Storage{
					client: mock.NewS3Client(nil),
					bucket: "test-bucket",
					logger: mock.NewMockLogger(nil),
					tracer: mock.NewMockTracer(nil),
				},
			},
			want: &S3Storage{
				client: mock.NewS3Client(nil),
				bucket: "test-bucket",
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
		},
		{
			name: "create new option with nil storage",
			args: args{
				storage: nil,
			},
			wantErr: ErrNoDriver,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := new(s3BaseRepository)
			err := WithS3Storage(tt.args.storage)(repo)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, repo.storage)
		})
	}
}

func TestWithRepositoryLogger(t *testing.T) {
	type args struct {
		logger log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    log.Logger
		wantErr error
	}{
		{
			name: "create new option with logger",
			args: args{
				logger: mock.NewMockLogger(nil),
			},
			want: mock.NewMockLogger(nil),
		},
		{
			name: "create new option with nil logger",
			args: args{
				logger: nil,
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := new(s3BaseRepository)
			err := WithS3RepositoryLogger(tt.args.logger)(repo)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, repo.logger)
		})
	}
}

func TestWithRepositoryTracer(t *testing.T) {
	type args struct {
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    tracing.Tracer
		wantErr error
	}{
		{
			name: "create new option with tracer",
			args: args{
				tracer: mock.NewMockTracer(nil),
			},
			want: mock.NewMockTracer(nil),
		},
		{
			name: "create new option with nil tracer",
			args: args{
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := new(s3BaseRepository)
			err := WithS3RepositoryTracer(tt.args.tracer)(repo)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, repo.tracer)
		})
	}
}

func TestNewBaseRepository(t *testing.T) {
	type args struct {
		storage *S3Storage
		logger  log.Logger
		tracer  tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    *s3BaseRepository
		wantErr error
	}{
		{
			name: "create new base repository",
			args: args{
				storage: &S3Storage{
					client: mock.NewS3Client(nil),
					bucket: "test-bucket",
					logger: mock.NewMockLogger(nil),
					tracer: mock.NewMockTracer(nil),
				},
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
			want: &s3BaseRepository{
				storage: &S3Storage{
					client: mock.NewS3Client(nil),
					bucket: "test-bucket",
					logger: mock.NewMockLogger(nil),
					tracer: mock.NewMockTracer(nil),
				},
				logger: mock.NewMockLogger(nil),
				tracer: mock.NewMockTracer(nil),
			},
		},
		{
			name: "create new base repository with nil storage",
			args: args{
				storage: nil,
				logger:  mock.NewMockLogger(nil),
				tracer:  mock.NewMockTracer(nil),
			},
			wantErr: ErrNoDriver,
		},
		{
			name: "create new base repository with nil logger",
			args: args{
				storage: &S3Storage{
					client: mock.NewS3Client(nil),
					bucket: "test-bucket",
					logger: mock.NewMockLogger(nil),
					tracer: mock.NewMockTracer(nil),
				},
				logger: nil,
				tracer: mock.NewMockTracer(nil),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new base repository with nil tracer",
			args: args{
				storage: &S3Storage{
					client: mock.NewS3Client(nil),
					bucket: "test-bucket",
					logger: mock.NewMockLogger(nil),
					tracer: mock.NewMockTracer(nil),
				},
				logger: mock.NewMockLogger(nil),
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo, err := newS3BaseRepository(
				WithS3Storage(tt.args.storage),
				WithS3RepositoryLogger(tt.args.logger),
				WithS3RepositoryTracer(tt.args.tracer),
			)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, repo)
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "not found error",
			err:  &mockAPIError{errorCode: "NoSuchKey"},
			want: true,
		},
		{
			name: "other error",
			err:  assert.AnError,
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isNotFoundError(tt.err)
			require.Equal(t, tt.want, got)
		})
	}
}

// Mock APIError for testing isNotFoundError
type mockAPIError struct {
	errorCode string
}

func (m *mockAPIError) Error() string {
	return "mock API error"
}

func (m *mockAPIError) ErrorCode() string {
	return m.errorCode
}

func (m *mockAPIError) ErrorMessage() string {
	return "mock error message"
}

func (m *mockAPIError) ErrorFault() smithy.ErrorFault {
	return smithy.FaultUnknown
}
