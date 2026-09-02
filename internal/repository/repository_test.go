package repository_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

func testPageSize(limit int) int {
	if limit < repository.MinPageSize {
		return repository.DefaultPageSize
	}
	return limit
}

func mustPlanCacheKey(t *testing.T, q repository.QueryCompiler, prefix string, extra ...any) string {
	t.Helper()
	plan, err := repository.CompileQuery(q)
	require.NoError(t, err)
	return plan.CacheKey(prefix, extra...)
}

// Test-local copies of unexported production helpers used to assert exact cache keys.
func composeCacheKey(params ...any) string {
	sep := ":"

	key := make([]string, len(params))
	for i, param := range params {
		if param != nil {
			switch p := param.(type) {
			case []string:
				key[i] = strings.Join(p, sep)
			default:
				key[i] = fmt.Sprintf("%v", param)
			}
		}
	}

	return strings.Join(key, sep)
}

func projectionCacheValue(proj any) string {
	return fmt.Sprintf("%+v", proj)
}

func authzGenKey(principal model.ID) string {
	return composeCacheKey("authz", "gen", principal.String())
}

const issueListGenPrefix = "issue:list:gen"

func issueListProjectGenKey(projectID model.ID) string {
	return composeCacheKey(issueListGenPrefix, "project", projectID.String())
}

func issueListNamespaceGenKey(namespaceID model.ID) string {
	return composeCacheKey(issueListGenPrefix, "namespace", namespaceID.String())
}

func issueListUserGenKey(userID model.ID) string {
	return composeCacheKey(issueListGenPrefix, "user", userID.String())
}

func issueListAuthzEpochKey() string {
	return composeCacheKey(issueListGenPrefix, "authz_epoch")
}

func issueListProjectionEpochKey() string {
	return composeCacheKey(issueListGenPrefix, "projection_epoch")
}

func mustPGDatabase(t *testing.T, pool repository.PGPool) *repository.PGDatabase {
	t.Helper()
	db, err := repository.NewPGDatabase(repository.WithDatabasePool(pool))
	require.NoError(t, err)
	return db
}

func mustRedisDatabase(t *testing.T, client redis.UniversalClient) *repository.RedisDatabase {
	t.Helper()
	db, err := repository.NewRedisDatabase(repository.WithRedisClient(client))
	require.NoError(t, err)
	return db
}

func mustS3Storage(t *testing.T, client repository.S3Client, bucket string) *repository.S3Storage {
	t.Helper()
	opts := []repository.S3StorageOption{repository.WithStorageClient(client)}
	if bucket != "" {
		opts = append(opts, repository.WithStorageBucket(bucket))
	}
	storage, err := repository.NewStorage(opts...)
	require.NoError(t, err)
	return storage
}

func redisRepoOptsNoop(ctrl *gomock.Controller) []repository.RedisRepositoryOption {
	db, err := repository.NewRedisDatabase(repository.WithRedisClient(mockrepo.NewMockUniversalClient(ctrl)))
	if err != nil {
		panic(err)
	}
	return []repository.RedisRepositoryOption{
		repository.WithRedisDatabase(db),
		repository.WithCacheBackend(mockrepo.NewMockCacheBackend(ctrl)),
	}
}

func issueListForIssueCacheKey(issueID model.ID, page repository.CursorPage, proj repository.IssueProjection) string {
	plan, err := repository.CompileQuery(repository.IssueListForIssueQuery{IssueID: issueID, Page: page, Projection: proj})
	if err != nil {
		panic(err)
	}
	return plan.CacheKey(model.ResourceTypeIssue.String(), "ListForIssue", issueID.String())
}

// redisCacheExpectingPatterns mocks Keys+DeletePattern for each pattern in order.
// If failIndex >= 0, that pattern's cache.Delete returns failErr and later patterns are not expected.
//
//nolint:revive // test cache factories take gomock.Controller first
func redisCacheExpectingPatterns(ctrl *gomock.Controller, ctx context.Context, patterns []string, failIndex int, failErr error) []repository.RedisRepositoryOption {
	client := mockrepo.NewMockUniversalClient(ctrl)
	backend := mockrepo.NewMockCacheBackend(ctrl)
	span := mocktrace.NewMockSpan(ctrl)
	tracer := mocktrace.NewMockTracer(ctrl)

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

	db, err := repository.NewRedisDatabase(repository.WithRedisClient(client))
	if err != nil {
		panic(err)
	}

	return []repository.RedisRepositoryOption{
		repository.WithRedisDatabase(db),
		repository.WithCacheBackend(backend),
		repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
		repository.WithRedisRepositoryTracer(tracer),
	}
}

//nolint:revive // test cache factories take gomock.Controller first
func redisCacheExpectingSetThenPatterns(ctrl *gomock.Controller, ctx context.Context, setKey string, setValue any, patterns []string, failOnSet bool, failIndex int, failErr error) []repository.RedisRepositoryOption {
	client := mockrepo.NewMockUniversalClient(ctrl)
	backend := mockrepo.NewMockCacheBackend(ctrl)
	span := mocktrace.NewMockSpan(ctrl)
	tracer := mocktrace.NewMockTracer(ctrl)

	count := 1
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
	if failOnSet {
		backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: setKey, Value: setValue}).Return(failErr)
		span.EXPECT().End(gomock.Len(0)).Times(1)
		db, err := repository.NewRedisDatabase(repository.WithRedisClient(client))
		if err != nil {
			panic(err)
		}
		return []repository.RedisRepositoryOption{
			repository.WithRedisDatabase(db),
			repository.WithCacheBackend(backend),
			repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
			repository.WithRedisRepositoryTracer(tracer),
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

	db, err := repository.NewRedisDatabase(repository.WithRedisClient(client))
	if err != nil {
		panic(err)
	}

	return []repository.RedisRepositoryOption{
		repository.WithRedisDatabase(db),
		repository.WithCacheBackend(backend),
		repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
		repository.WithRedisRepositoryTracer(tracer),
	}
}

//nolint:revive // test cache factories take gomock.Controller first
func redisCacheExpectingPatternsThenIssueAuthzEpochBump(ctrl *gomock.Controller, ctx context.Context, patterns []string, failIndex int, failErr error, bumpCount int) []repository.RedisRepositoryOption {
	client := mockrepo.NewMockUniversalClient(ctrl)
	backend := mockrepo.NewMockCacheBackend(ctrl)
	span := mocktrace.NewMockSpan(ctrl)
	tracer := mocktrace.NewMockTracer(ctrl)

	count := 0
	for i, pattern := range patterns {
		count++
		cmd := new(redis.StringSliceCmd)
		cmd.SetVal([]string{pattern})
		client.EXPECT().Keys(ctx, pattern).Return(cmd)
		if i == failIndex {
			backend.EXPECT().Delete(ctx, pattern).Return(failErr)
			bumpCount = 0
			break
		}
		backend.EXPECT().Delete(ctx, pattern).Return(nil)
	}

	authzEpochKey := issueListAuthzEpochKey()
	for i := 0; i < bumpCount; i++ {
		count += 2
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
		backend.EXPECT().Get(ctx, authzEpochKey, gomock.Any()).Return(cache.ErrCacheMiss)
		backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: authzEpochKey, Value: int64(1)}).Return(nil)
	}

	span.EXPECT().End(gomock.Len(0)).Times(count)
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(count - (2 * bumpCount))

	db, err := repository.NewRedisDatabase(repository.WithRedisClient(client))
	if err != nil {
		panic(err)
	}

	return []repository.RedisRepositoryOption{
		repository.WithRedisDatabase(db),
		repository.WithCacheBackend(backend),
		repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
		repository.WithRedisRepositoryTracer(tracer),
	}
}

//nolint:revive // test cache factories take gomock.Controller first
func redisCacheExpectingSetThenPatternsThenIssueAuthzEpochBump(ctrl *gomock.Controller, ctx context.Context, setKey string, setValue any, patterns []string, failOnSet bool, failIndex int, failErr error, bumpCount int) []repository.RedisRepositoryOption {
	client := mockrepo.NewMockUniversalClient(ctrl)
	backend := mockrepo.NewMockCacheBackend(ctrl)
	span := mocktrace.NewMockSpan(ctrl)
	tracer := mocktrace.NewMockTracer(ctrl)

	count := 1
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
	if failOnSet {
		backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: setKey, Value: setValue}).Return(failErr)
		span.EXPECT().End(gomock.Len(0)).Times(1)
		db, err := repository.NewRedisDatabase(repository.WithRedisClient(client))
		if err != nil {
			panic(err)
		}
		return []repository.RedisRepositoryOption{
			repository.WithRedisDatabase(db),
			repository.WithCacheBackend(backend),
			repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
			repository.WithRedisRepositoryTracer(tracer),
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
			bumpCount = 0
			break
		}
		backend.EXPECT().Delete(ctx, pattern).Return(nil)
	}

	authzEpochKey := issueListAuthzEpochKey()
	for i := 0; i < bumpCount; i++ {
		count += 2
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span)
		tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span)
		backend.EXPECT().Get(ctx, authzEpochKey, gomock.Any()).Return(cache.ErrCacheMiss)
		backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: authzEpochKey, Value: int64(1)}).Return(nil)
	}

	span.EXPECT().End(gomock.Len(0)).Times(count)
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(count - 1 - (2 * bumpCount))

	db, err := repository.NewRedisDatabase(repository.WithRedisClient(client))
	if err != nil {
		panic(err)
	}

	return []repository.RedisRepositoryOption{
		repository.WithRedisDatabase(db),
		repository.WithCacheBackend(backend),
		repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
		repository.WithRedisRepositoryTracer(tracer),
	}
}

//nolint:revive // test cache factories take gomock.Controller first
func redisCacheExpectingBumpThenPatternsAndIssueAuthzEpoch(ctrl *gomock.Controller, ctx context.Context, principal model.ID, patterns []string) []repository.RedisRepositoryOption {
	client := mockrepo.NewMockUniversalClient(ctrl)
	backend := mockrepo.NewMockCacheBackend(ctrl)
	span := mocktrace.NewMockSpan(ctrl)
	tracer := mocktrace.NewMockTracer(ctrl)

	genKey := authzGenKey(principal)
	issueEpochKey := issueListAuthzEpochKey()
	count := 4 + len(patterns)
	span.EXPECT().End(gomock.Len(0)).Times(count)
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Get", gomock.Len(0)).Return(ctx, span).Times(2)
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/Set", gomock.Len(0)).Return(ctx, span).Times(2)
	tracer.EXPECT().Start(ctx, "repository.redisBaseRepository/DeletePattern", gomock.Len(0)).Return(ctx, span).Times(len(patterns))

	backend.EXPECT().Get(ctx, genKey, gomock.Any()).Return(cache.ErrCacheMiss)
	backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: genKey, Value: int64(1)}).Return(nil)
	for _, pattern := range patterns {
		cmd := new(redis.StringSliceCmd)
		cmd.SetVal([]string{pattern})
		client.EXPECT().Keys(ctx, pattern).Return(cmd)
		backend.EXPECT().Delete(ctx, pattern).Return(nil)
	}
	backend.EXPECT().Get(ctx, issueEpochKey, gomock.Any()).Return(cache.ErrCacheMiss)
	backend.EXPECT().Set(&cache.Item{Ctx: ctx, Key: issueEpochKey, Value: int64(1)}).Return(nil)

	db, err := repository.NewRedisDatabase(repository.WithRedisClient(client))
	if err != nil {
		panic(err)
	}
	return []repository.RedisRepositoryOption{
		repository.WithRedisDatabase(db),
		repository.WithCacheBackend(backend),
		repository.WithRedisRepositoryLogger(mocklog.NewMockLogger(ctrl)),
		repository.WithRedisRepositoryTracer(tracer),
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
		composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*"),
		composeCacheKey(model.ResourceTypeNamespace.String(), "*", "ListForOrganization", "*"),
		composeCacheKey(model.ResourceTypeNamespace.String(), "*", "ListAccessible", "*"),
		composeCacheKey(model.ResourceTypeProject.String(), "*", "ListForNamespace", "*"),
		composeCacheKey(model.ResourceTypeDocument.String(), "ListLibrary", "*"),
		composeCacheKey(model.ResourceTypeDocument.String(), "ListRelated", "*"),
		composeCacheKey(model.ResourceTypeDocument.String(), "ListByCreator", "*"),
		composeCacheKey(model.ResourceTypeFolder.String(), "*", "ListForLibrary", "*"),
	}
}

func TestEdgeKind_String(t *testing.T) {
	tests := []struct {
		name string
		s    repository.EdgeKind
		want string
	}{
		{"ASSIGNED_TO", repository.EdgeKindAssignedTo, "ASSIGNED_TO"},
		{"BELONGS_TO", repository.EdgeKindBelongsTo, "BELONGS_TO"},
		{"COMMENTED", repository.EdgeKindCommented, "COMMENTED"},
		{"CREATED", repository.EdgeKindCreated, "CREATED"},
		{"HAS_ATTACHMENT", repository.EdgeKindHasAttachment, "HAS_ATTACHMENT"},
		{"HAS_COMMENT", repository.EdgeKindHasComment, "HAS_COMMENT"},
		{"HAS_LABEL", repository.EdgeKindHasLabel, "HAS_LABEL"},
		{"HAS_NAMESPACE", repository.EdgeKindHasNamespace, "HAS_NAMESPACE"},
		{"HAS_PERMISSION", repository.EdgeKindHasPermission, "HAS_PERMISSION"},
		{"HAS_PROJECT", repository.EdgeKindHasProject, "HAS_PROJECT"},
		{"HAS_TEAM", repository.EdgeKindHasTeam, "HAS_TEAM"},
		{"INVITED", repository.EdgeKindInvited, "INVITED"},
		{"INVITED_TO", repository.EdgeKindInvitedTo, "INVITED_TO"},
		{"KIND_OF", repository.EdgeKindKindOf, "KIND_OF"},
		{"MEMBER_OF", repository.EdgeKindMemberOf, "MEMBER_OF"},
		{"RELATED_TO", repository.EdgeKindRelatedTo, "RELATED_TO"},
		{"SPEAKS", repository.EdgeKindSpeaks, "SPEAKS"},
		{"WATCHES", repository.EdgeKindWatches, "WATCHES"},
		{"SCOPED_TO", repository.EdgeKindScopedTo, "SCOPED_TO"},
		{"LOCATED_IN", repository.EdgeKindLocatedIn, "LOCATED_IN"},
		{"IN_SCOPE_OF", repository.EdgeKindInScopeOf, "IN_SCOPE_OF"},
		{"GRANTED", repository.EdgeKindGranted, "GRANTED"},
		{"DEFINES_ROLE", repository.EdgeKindDefinesRole, "DEFINES_ROLE"},
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
			_, err := repository.NewPool(tt.args.ctx, tt.args.conf)
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

	pool, err := repository.NewPool(context.Background(), &config.RelationalDatabaseConfig{
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

	pool, err := repository.NewPool(context.Background(), &config.RelationalDatabaseConfig{
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
		pool repository.PGPool
	}
	tests := []struct {
		name    string
		args    args
		want    repository.PGPool
		wantErr error
	}{
		{
			name: "create new option with pool",
			args: args{
				pool: mockrepo.NewMockPGPool(nil),
			},
			want: mockrepo.NewMockPGPool(nil),
		},
		{
			name: "create new option with nil pool",
			args: args{
				pool: nil,
			},
			wantErr: repository.ErrNoPool,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := new(repository.PGDatabase)
			err := repository.WithDatabasePool(tt.args.pool)(db)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.args.pool, db.Pool())
			}
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
				logger: mocklog.NewMockLogger(nil),
			},
			want: mocklog.NewMockLogger(nil),
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
			db := new(repository.PGDatabase)
			err := repository.WithPGDatabaseLogger(tt.args.logger)(db)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, tt.args.logger)
			}
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
				tracer: mocktrace.NewMockTracer(nil),
			},
			want: mocktrace.NewMockTracer(nil),
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
			db := new(repository.PGDatabase)
			err := repository.WithPGDatabaseTracer(tt.args.tracer)(db)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, tt.args.tracer)
			}
		})
	}
}

func TestNewPGDatabase(t *testing.T) {
	type args struct {
		pool   repository.PGPool
		logger log.Logger
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    *repository.PGDatabase
		wantErr error
	}{
		{
			name: "create new database",
			args: args{
				pool:   mockrepo.NewMockPGPool(nil),
				logger: mocklog.NewMockLogger(nil),
				tracer: mocktrace.NewMockTracer(nil),
			},
		},
		{
			name: "create new database with nil pool",
			args: args{
				pool:   nil,
				logger: mocklog.NewMockLogger(nil),
				tracer: mocktrace.NewMockTracer(nil),
			},
			wantErr: repository.ErrNoPool,
		},
		{
			name: "create new database with nil logger",
			args: args{
				pool:   mockrepo.NewMockPGPool(nil),
				logger: nil,
				tracer: mocktrace.NewMockTracer(nil),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new database with nil tracer",
			args: args{
				pool:   mockrepo.NewMockPGPool(nil),
				logger: mocklog.NewMockLogger(nil),
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, err := repository.NewPGDatabase(
				repository.WithDatabasePool(tt.args.pool),
				repository.WithPGDatabaseLogger(tt.args.logger),
				repository.WithPGDatabaseTracer(tt.args.tracer),
			)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, db)
			}
		})
	}
}

func TestPGDatabase_Close(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	pool := mockrepo.NewMockPGPool(ctrl)
	pool.EXPECT().Close()

	db := mustPGDatabase(t, pool)

	require.NoError(t, db.Close())
}

func TestDatabase_Pool(t *testing.T) {
	t.Parallel()

	pool := mockrepo.NewMockPGPool(nil)

	db := mustPGDatabase(t, pool)

	require.Equal(t, pool, db.Pool())
}

func TestPGDatabase_Ping(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type fields struct {
		pool func(ctx context.Context, ctrl *gomock.Controller) repository.PGPool
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
				pool: func(ctx context.Context, ctrl *gomock.Controller) repository.PGPool {
					p := mockrepo.NewMockPGPool(ctrl)
					p.EXPECT().Ping(ctx).Return(nil).Times(1)
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
				pool: func(ctx context.Context, ctrl *gomock.Controller) repository.PGPool {
					p := mockrepo.NewMockPGPool(ctrl)
					p.EXPECT().Ping(ctx).Return(assert.AnError).Times(1)
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
			db := mustPGDatabase(t, tt.fields.pool(tt.args.ctx, ctrl))
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
			_, err := repository.NewRedisClient(tt.args.conf)
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
					return mockrepo.NewMockUniversalClient(ctrl)
				}(),
			},
			want: func() redis.UniversalClient {
				ctrl := gomock.NewController(t)
				return mockrepo.NewMockUniversalClient(ctrl)
			}(),
		},
		{
			name: "create new option with nil client",
			args: args{
				client: nil,
			},
			wantErr: repository.ErrNoClient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := new(repository.RedisDatabase)
			err := repository.WithRedisClient(tt.args.client)(db)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.args.client, db.Client())
			}
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
				logger: mocklog.NewMockLogger(nil),
			},
			want: mocklog.NewMockLogger(nil),
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
			db := new(repository.RedisDatabase)
			err := repository.WithRedisDatabaseLogger(tt.args.logger)(db)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, tt.args.logger)
			}
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
				tracer: mocktrace.NewMockTracer(nil),
			},
			want: mocktrace.NewMockTracer(nil),
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
			db := new(repository.RedisDatabase)
			err := repository.WithRedisDatabaseTracer(tt.args.tracer)(db)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, tt.args.tracer)
			}
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
		want    *repository.RedisDatabase
		wantErr error
	}{
		{
			name: "create new database",
			args: args{
				client: func() redis.UniversalClient {
					ctrl := gomock.NewController(t)
					return mockrepo.NewMockUniversalClient(ctrl)
				}(),
				logger: mocklog.NewMockLogger(nil),
				tracer: mocktrace.NewMockTracer(nil),
			},
		},
		{
			name: "create new database with nil client",
			args: args{
				client: nil,
				logger: mocklog.NewMockLogger(nil),
				tracer: mocktrace.NewMockTracer(nil),
			},
			wantErr: repository.ErrNoClient,
		},
		{
			name: "create new database with nil logger",
			args: args{
				client: func() redis.UniversalClient {
					ctrl := gomock.NewController(t)
					return mockrepo.NewMockUniversalClient(ctrl)
				}(),
				logger: nil,
				tracer: mocktrace.NewMockTracer(nil),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new database with nil tracer",
			args: args{
				client: func() redis.UniversalClient {
					ctrl := gomock.NewController(t)
					return mockrepo.NewMockUniversalClient(ctrl)
				}(),
				logger: mocklog.NewMockLogger(nil),
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, err := repository.NewRedisDatabase(
				repository.WithRedisClient(tt.args.client),
				repository.WithRedisDatabaseLogger(tt.args.logger),
				repository.WithRedisDatabaseTracer(tt.args.tracer),
			)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, db)
			}
		})
	}
}

func TestDatabase_Client(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	client := mockrepo.NewMockUniversalClient(ctrl)

	db := mustRedisDatabase(t, client)

	require.Equal(t, client, db.Client())
}

func TestRedisDatabase_Close(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	client := mockrepo.NewMockUniversalClient(ctrl)
	client.EXPECT().Close().Return(nil)

	db := mustRedisDatabase(t, client)

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
					p := mockrepo.NewMockUniversalClient(ctrl)
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
			db := mustRedisDatabase(t, tt.fields.client(ctrl, tt.args.ctx))
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
			_, err := repository.NewS3Client(tt.args.ctx, tt.args.conf)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestWithStorageClient(t *testing.T) {
	type args struct {
		client repository.S3Client
	}
	tests := []struct {
		name    string
		args    args
		want    repository.S3Client
		wantErr error
	}{
		{
			name: "create new option with client",
			args: args{
				client: mockrepo.NewMockS3Client(nil),
			},
			want: mockrepo.NewMockS3Client(nil),
		},
		{
			name: "create new option with nil client",
			args: args{
				client: nil,
			},
			wantErr: repository.ErrNoClient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage := new(repository.S3Storage)
			err := repository.WithStorageClient(tt.args.client)(storage)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.args.client, storage.Client())
			}
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
			wantErr: repository.ErrNoBucket,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage := new(repository.S3Storage)
			err := repository.WithStorageBucket(tt.args.bucket)(storage)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				ctrl := gomock.NewController(t)
				client := mockrepo.NewMockS3Client(ctrl)
				client.EXPECT().HeadBucket(gomock.Any(), &awsS3.HeadBucketInput{Bucket: aws.String(tt.args.bucket)}, gomock.Any()).
					Return(&awsS3.HeadBucketOutput{}, nil)
				require.NoError(t, repository.WithStorageClient(client)(storage))
				require.NoError(t, storage.Ping(context.Background()))
			}
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
				logger: mocklog.NewMockLogger(nil),
			},
			want: mocklog.NewMockLogger(nil),
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
			storage := new(repository.S3Storage)
			err := repository.WithStorageLogger(tt.args.logger)(storage)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, tt.args.logger)
			}
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
				tracer: mocktrace.NewMockTracer(nil),
			},
			want: mocktrace.NewMockTracer(nil),
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
			storage := new(repository.S3Storage)
			err := repository.WithStorageTracer(tt.args.tracer)(storage)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, tt.args.tracer)
			}
		})
	}
}

func TestNewStorage(t *testing.T) {
	type args struct {
		client repository.S3Client
		bucket string
		logger log.Logger
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    *repository.S3Storage
		wantErr error
	}{
		{
			name: "create new storage",
			args: args{
				client: mockrepo.NewMockS3Client(nil),
				bucket: "test-bucket",
				logger: mocklog.NewMockLogger(nil),
				tracer: mocktrace.NewMockTracer(nil),
			},
		},
		{
			name: "create new storage with nil client",
			args: args{
				client: nil,
				bucket: "test-bucket",
				logger: mocklog.NewMockLogger(nil),
				tracer: mocktrace.NewMockTracer(nil),
			},
			wantErr: repository.ErrNoClient,
		},
		{
			name: "create new storage with empty bucket",
			args: args{
				client: mockrepo.NewMockS3Client(nil),
				bucket: "",
				logger: mocklog.NewMockLogger(nil),
				tracer: mocktrace.NewMockTracer(nil),
			},
			wantErr: repository.ErrNoBucket,
		},
		{
			name: "create new storage with nil logger",
			args: args{
				client: mockrepo.NewMockS3Client(nil),
				bucket: "test-bucket",
				logger: nil,
				tracer: mocktrace.NewMockTracer(nil),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new storage with nil tracer",
			args: args{
				client: mockrepo.NewMockS3Client(nil),
				bucket: "test-bucket",
				logger: mocklog.NewMockLogger(nil),
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage, err := repository.NewStorage(
				repository.WithStorageClient(tt.args.client),
				repository.WithStorageBucket(tt.args.bucket),
				repository.WithStorageLogger(tt.args.logger),
				repository.WithStorageTracer(tt.args.tracer),
			)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, storage)
			}
		})
	}
}

func TestStorage_Client(t *testing.T) {
	t.Parallel()

	client := mockrepo.NewMockS3Client(nil)

	storage := mustS3Storage(t, client, "")

	require.Equal(t, client, storage.Client())
}

func TestStorage_Ping(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type fields struct {
		client func(ctx context.Context, ctrl *gomock.Controller) repository.S3Client
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
				client: func(ctx context.Context, ctrl *gomock.Controller) repository.S3Client {
					c := mockrepo.NewMockS3Client(ctrl)
					c.EXPECT().HeadBucket(ctx, &awsS3.HeadBucketInput{Bucket: aws.String("test-bucket")}, gomock.Any()).Return(&awsS3.HeadBucketOutput{}, nil).Times(1)
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
				client: func(ctx context.Context, ctrl *gomock.Controller) repository.S3Client {
					c := mockrepo.NewMockS3Client(ctrl)
					c.EXPECT().HeadBucket(ctx, &awsS3.HeadBucketInput{Bucket: aws.String("test-bucket")}, gomock.Any()).Return(&awsS3.HeadBucketOutput{}, assert.AnError).Times(1)
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
			storage := mustS3Storage(t, tt.fields.client(tt.args.ctx, ctrl), tt.fields.bucket)
			err := storage.Ping(tt.args.ctx)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
