package repository_test

import (
	"encoding/base64"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func TestCursorPageNormalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		page    repository.CursorPage
		want    repository.CursorPage
		wantErr error
	}{
		{
			name: "applies default size",
			page: repository.CursorPage{},
			want: repository.CursorPage{Size: repository.DefaultPageSize},
		},
		{
			name: "keeps explicit size",
			page: repository.CursorPage{Size: 25},
			want: repository.CursorPage{Size: 25},
		},
		{
			name:    "rejects negative",
			page:    repository.CursorPage{Size: -1},
			wantErr: repository.ErrInvalidPageSize,
		},
		{
			name:    "rejects too large",
			page:    repository.CursorPage{Size: repository.MaxPageSize + 1},
			wantErr: repository.ErrInvalidPageSize,
		},
		{
			name: "preserves token",
			page: repository.CursorPage{Size: 10, Token: ptr("abc")},
			want: repository.CursorPage{Size: 10, Token: ptr("abc")},
		},
		{
			name: "coerces empty token to nil",
			page: repository.CursorPage{Size: 10, Token: ptr("")},
			want: repository.CursorPage{Size: 10, Token: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.page.Normalize()
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want.Size, got.Size)
			assert.Equal(t, tt.want.Token, got.Token)
			assert.Equal(t, got.Size+1, got.FetchLimit())
		})
	}
}

func TestEncodeDecodeCursor(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeIssue)

	tests := []struct {
		name    string
		id      model.ID
		wantErr error
	}{
		{
			name: "round trip",
			id:   id,
		},
		{
			name:    "rejects invalid id type",
			id:      model.ID{},
			wantErr: repository.ErrInvalidCursor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			token, err := repository.EncodeCursor(tt.id)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, token)

			gotID, err := repository.DecodeCursor(token)
			require.NoError(t, err)
			assert.Equal(t, tt.id, gotID)
		})
	}
}

func TestDecodeCursorMalformed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		token string
	}{
		{name: "empty", token: ""},
		{name: "not base64", token: "%%%"},
		{name: "not json", token: "YWJj"}, // "abc"
		{name: "missing fields", token: mustB64(`{"created_at":"2024-01-01T00:00:00Z"}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := repository.DecodeCursor(tt.token)
			require.Error(t, err)
			assert.ErrorIs(t, err, repository.ErrInvalidCursor)
		})
	}
}

func TestPaginateSlice(t *testing.T) {
	t.Parallel()

	type item struct {
		id model.ID
	}

	mk := func(n int) []item {
		out := make([]item, n)
		for i := range out {
			out[i] = item{id: model.MustNewID(model.ResourceTypeProject)}
		}
		return out
	}

	tests := []struct {
		name     string
		items    []item
		pageSize int
		wantLen  int
		wantMore bool
		wantErr  error
	}{
		{
			name:     "exact page no more",
			items:    mk(3),
			pageSize: 3,
			wantLen:  3,
			wantMore: false,
		},
		{
			name:     "has more trims and token",
			items:    mk(4),
			pageSize: 3,
			wantLen:  3,
			wantMore: true,
		},
		{
			name:     "empty",
			items:    nil,
			pageSize: 10,
			wantLen:  0,
			wantMore: false,
		},
		{
			name:     "invalid page size",
			items:    mk(1),
			pageSize: 0,
			wantErr:  repository.ErrInvalidPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			page, err := repository.PaginateSlice(
				tt.items,
				tt.pageSize,
				func(i item) model.ID { return i.id },
			)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, page.Items, tt.wantLen)
			assert.Equal(t, tt.wantMore, page.PageInfo.HasMore)
			if tt.wantMore {
				require.NotNil(t, page.PageInfo.NextPageToken)
				gotID, err := repository.DecodeCursor(*page.PageInfo.NextPageToken)
				require.NoError(t, err)
				assert.Equal(t, tt.items[tt.wantLen-1].id, gotID)
			} else {
				assert.Nil(t, page.PageInfo.NextPageToken)
			}
		})
	}
}

func TestSortDirection(t *testing.T) {
	t.Parallel()
	assert.True(t, repository.SortDirectionAsc.Valid())
	assert.True(t, repository.SortDirectionDesc.Valid())
	assert.False(t, repository.SortDirection(100).Valid())
	assert.False(t, repository.SortDirectionUnknown.Valid())
	assert.Equal(t, "ASC", repository.SortDirectionAsc.Cypher())
}

func TestCursorWhereCypher(t *testing.T) {
	t.Parallel()

	asc, err := repository.CursorWhereCypher("p", repository.SortDirectionAsc)
	require.NoError(t, err)
	assert.Equal(t, "p.id > $cursor_id", asc)

	desc, err := repository.CursorWhereCypher("i", repository.SortDirectionDesc)
	require.NoError(t, err)
	assert.Equal(t, "i.id < $cursor_id", desc)

	_, err = repository.CursorWhereCypher("p", repository.SortDirection(100))
	assert.ErrorIs(t, err, repository.ErrUnsupportedOrder)
}

func TestCompiledQueryFingerprintStable(t *testing.T) {
	t.Parallel()

	q1 := repository.CompiledQuery{Name: "project.get", Cypher: "MATCH (p) RETURN p", Params: map[string]any{"id": "a"}}
	q2 := repository.CompiledQuery{Name: "project.get", Cypher: "MATCH (p) RETURN p", Params: map[string]any{"id": "b"}}
	q3 := repository.CompiledQuery{Name: "project.get", Cypher: "MATCH (x) RETURN x", Params: map[string]any{"id": "a"}}

	assert.Equal(t, q1.Fingerprint(), q2.Fingerprint())
	assert.NotEqual(t, q1.Fingerprint(), q3.Fingerprint())
	assert.Contains(t, q1.CacheKey("Project", "Get"), q1.Fingerprint())
}

func TestPlanCacheKeyMatchesListInvalidationGlob(t *testing.T) {
	t.Parallel()

	libraryID := model.MustNewID(model.ResourceTypeNamespace)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	actorID := model.MustNewID(model.ResourceTypeUser)

	tests := []struct {
		name    string
		query   repository.QueryCompiler
		prefix  string
		extra   []any
		pattern string
	}{
		{
			name: "folder list",
			query: repository.FolderListQuery{
				LibraryID: libraryID,
				ActorID:   actorID,
				Page:      repository.CursorPage{Size: 10},
				Order:     repository.SortDirectionDesc,
			},
			prefix:  model.ResourceTypeFolder.String(),
			extra:   []any{"ListForLibrary", libraryID.String()},
			pattern: composeCacheKey(model.ResourceTypeFolder.String(), "*", "ListForLibrary", "*"),
		},
		{
			name: "namespace organization list",
			query: repository.NamespaceListQuery{
				OrgID:      orgID,
				ActorID:    actorID,
				Page:       repository.CursorPage{Size: 10},
				Order:      repository.SortDirectionDesc,
				Projection: repository.NamespaceListProjection(),
			},
			prefix:  model.ResourceTypeNamespace.String(),
			extra:   []any{"ListForOrganization", orgID.String()},
			pattern: composeCacheKey(model.ResourceTypeNamespace.String(), "*", "ListForOrganization", "*"),
		},
		{
			name: "namespace accessible list",
			query: repository.NamespaceListAccessibleQuery{
				ActorID:    actorID,
				Page:       repository.CursorPage{Size: 10},
				Order:      repository.SortDirectionDesc,
				Projection: repository.NamespaceListProjection(),
			},
			prefix:  model.ResourceTypeNamespace.String(),
			extra:   []any{"ListAccessible"},
			pattern: composeCacheKey(model.ResourceTypeNamespace.String(), "*", "ListAccessible", "*"),
		},
		{
			name: "organization user list",
			query: repository.OrganizationListQuery{
				UserID:     actorID,
				Action:     model.ActionOrganizationRead,
				Page:       repository.CursorPage{Size: 10},
				Order:      repository.SortDirectionDesc,
				Projection: repository.OrganizationListProjection(),
			},
			prefix:  model.ResourceTypeOrganization.String(),
			extra:   []any{"ListForUser"},
			pattern: composeCacheKey(model.ResourceTypeOrganization.String(), "*", "ListForUser", "*"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plan, err := repository.CompileQuery(tt.query)
			require.NoError(t, err)

			key := plan.CacheKey(tt.prefix, tt.extra...)
			matched, err := path.Match(tt.pattern, key)
			require.NoError(t, err)
			assert.True(t, matched, "pattern %q should match key %q", tt.pattern, key)
		})
	}
}

func TestQueryPlanValidate(t *testing.T) {
	t.Parallel()

	err := (repository.QueryPlan{}).Validate()
	assert.ErrorIs(t, err, repository.ErrQueryCompile)

	err = (repository.QueryPlan{
		Root: repository.CompiledQuery{Name: "root", Cypher: "RETURN 1"},
		Loaders: []repository.CompiledQuery{
			{Name: "", Cypher: "RETURN 1"},
		},
	}).Validate()
	assert.ErrorIs(t, err, repository.ErrQueryCompile)

	err = (repository.QueryPlan{
		Root: repository.CompiledQuery{Name: "root", Cypher: "RETURN 1"},
	}).Validate()
	assert.NoError(t, err)
}

func TestApplyCursorParams(t *testing.T) {
	t.Parallel()

	params := map[string]any{}
	require.NoError(t, repository.ApplyCursorParams(params, nil))
	assert.Empty(t, params)

	id := model.MustNewID(model.ResourceTypeIssue)
	token, err := repository.EncodeCursor(id)
	require.NoError(t, err)

	require.NoError(t, repository.ApplyCursorParams(params, &token))
	assert.Equal(t, id.String(), params["cursor_id"])
	assert.NotContains(t, params, "cursor_created_at")

	bad := "not-a-token"
	err = repository.ApplyCursorParams(map[string]any{}, &bad)
	assert.ErrorIs(t, err, repository.ErrInvalidCursor)
}

func ptr[T any](v T) *T { return &v }

func mustB64(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}
