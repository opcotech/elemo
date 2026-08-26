package repository

import (
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j/dbtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
)

func testNode(labels []string, props map[string]any) neo4j.Node {
	if props == nil {
		props = map[string]any{}
	}
	return neo4j.Node{Labels: labels, Props: props}
}

func testRecord(key string, val any) *neo4j.Record {
	return &neo4j.Record{Keys: []string{key}, Values: []any{val}}
}

func TestNeo4jNodeProperty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		node    neo4j.Node
		want    string
		wantErr bool
	}{
		{
			name: "returns typed property",
			node: testNode(nil, map[string]any{"name": "acme"}),
			want: "acme",
		},
		{
			name:    "missing key",
			node:    testNode(nil, map[string]any{}),
			wantErr: true,
		},
		{
			name:    "wrong type",
			node:    testNode(nil, map[string]any{"name": int64(1)}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Neo4jNodeProperty[string](tt.node, "name")
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNeo4jOptionalNodeProperty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		node    neo4j.Node
		want    string
		wantOK  bool
		wantErr bool
	}{
		{
			name:   "returns typed property",
			node:   testNode(nil, map[string]any{"name": "acme"}),
			want:   "acme",
			wantOK: true,
		},
		{
			name: "missing key",
			node: testNode(nil, map[string]any{}),
		},
		{
			name:    "wrong type",
			node:    testNode(nil, map[string]any{"name": int64(1)}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok, err := Neo4jOptionalNodeProperty[string](tt.node, "name")
			if tt.wantErr {
				require.Error(t, err)
				assert.False(t, ok)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNeo4jDecodeID(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeIssue)

	tests := []struct {
		name    string
		node    neo4j.Node
		want    model.ID
		wantErr error
	}{
		{
			name: "valid id",
			node: testNode(nil, map[string]any{"id": id.String()}),
			want: id,
		},
		{
			name:    "missing id",
			node:    testNode(nil, map[string]any{}),
			wantErr: ErrMalformedResult,
		},
		{
			name:    "invalid xid",
			node:    testNode(nil, map[string]any{"id": "not-a-xid"}),
			wantErr: ErrMalformedResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Neo4jDecodeID(tt.node, model.ResourceTypeIssue)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNeo4jDecodeIDFromLabel(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeUser)
	organizationID := model.MustNewID(model.ResourceTypeOrganization)

	tests := []struct {
		name    string
		node    neo4j.Node
		want    model.ID
		wantErr error
	}{
		{
			name: "uses first label",
			node: testNode(
				[]string{model.ResourceTypeUser.String(), "Extra"},
				map[string]any{"id": id.String()},
			),
			want: id,
		},
		{
			name: "skips principal when it is first",
			node: testNode(
				[]string{model.LabelPrincipal, model.ResourceTypeOrganization.String()},
				map[string]any{"id": organizationID.String()},
			),
			want: organizationID,
		},
		{
			name:    "no labels",
			node:    testNode(nil, map[string]any{"id": id.String()}),
			wantErr: ErrMalformedResult,
		},
		{
			name:    "missing id",
			node:    testNode([]string{model.ResourceTypeUser.String()}, map[string]any{}),
			wantErr: ErrMalformedResult,
		},
		{
			name:    "unknown label",
			node:    testNode([]string{"NotAType"}, map[string]any{"id": id.String()}),
			wantErr: ErrMalformedResult,
		},
		{
			name: "principal only",
			node: testNode(
				[]string{model.LabelPrincipal},
				map[string]any{"id": id.String()},
			),
			wantErr: ErrMalformedResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Neo4jDecodeIDFromLabel(tt.node)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNeo4jDecodeTime(t *testing.T) {
	t.Parallel()

	ts := time.Date(2024, 6, 1, 12, 30, 0, 123456789, time.UTC)

	tests := []struct {
		name    string
		val     any
		want    *time.Time
		wantErr error
	}{
		{
			name: "nil",
			val:  nil,
		},
		{
			name: "time.Time",
			val:  ts,
			want: &ts,
		},
		{
			name: "local datetime",
			val:  dbtype.LocalDateTime(ts),
			want: &ts,
		},
		{
			name: "date",
			val:  neo4j.DateOf(ts),
			want: func() *time.Time {
				d := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
				return &d
			}(),
		},
		{
			name: "rfc3339nano string",
			val:  ts.Format(time.RFC3339Nano),
			want: &ts,
		},
		{
			name: "rfc3339 string",
			val:  "2024-06-01T12:30:00Z",
			want: func() *time.Time {
				parsed := time.Date(2024, 6, 1, 12, 30, 0, 0, time.UTC)
				return &parsed
			}(),
		},
		{
			name:    "invalid string",
			val:     "not-a-time",
			wantErr: ErrMalformedResult,
		},
		{
			name:    "unsupported type",
			val:     int64(1),
			wantErr: ErrMalformedResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Neo4jDecodeTime(tt.val)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.True(t, tt.want.Equal(*got))
			assert.Equal(t, time.UTC, got.Location())
		})
	}
}

func TestNeo4jNodeTime(t *testing.T) {
	t.Parallel()

	ts := time.Date(2024, 2, 3, 4, 5, 6, 0, time.UTC)

	tests := []struct {
		name string
		node neo4j.Node
		want *time.Time
	}{
		{
			name: "missing key",
			node: testNode(nil, map[string]any{}),
		},
		{
			name: "nil value",
			node: testNode(nil, map[string]any{"created_at": nil}),
		},
		{
			name: "valid time",
			node: testNode(nil, map[string]any{"created_at": ts}),
			want: &ts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Neo4jNodeTime(tt.node, "created_at")
			require.NoError(t, err)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.True(t, tt.want.Equal(*got))
		})
	}

	t.Run("unsupported type", func(t *testing.T) {
		t.Parallel()
		_, err := Neo4jNodeTime(testNode(nil, map[string]any{"created_at": int64(1)}), "created_at")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrMalformedResult)
	})
}

func TestNeo4jRecordNode(t *testing.T) {
	t.Parallel()

	node := testNode([]string{model.ResourceTypeProject.String()}, map[string]any{"name": "p"})

	tests := []struct {
		name    string
		record  *neo4j.Record
		want    neo4j.Node
		wantErr error
	}{
		{
			name:   "returns node",
			record: testRecord("p", node),
			want:   node,
		},
		{
			name:    "missing key",
			record:  &neo4j.Record{Keys: []string{}, Values: []any{}},
			wantErr: ErrMalformedResult,
		},
		{
			name:    "wrong type",
			record:  testRecord("p", "not-a-node"),
			wantErr: ErrMalformedResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Neo4jRecordNode(tt.record, "p")
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNeo4jRecordOptionalNode(t *testing.T) {
	t.Parallel()

	node := testNode([]string{model.ResourceTypeNamespace.String()}, map[string]any{"name": "n"})

	tests := []struct {
		name    string
		record  *neo4j.Record
		want    *neo4j.Node
		wantErr error
	}{
		{
			name:   "returns node",
			record: testRecord("n", node),
			want:   &node,
		},
		{
			name:   "missing key",
			record: &neo4j.Record{Keys: []string{}, Values: []any{}},
		},
		{
			name:   "nil value",
			record: testRecord("n", nil),
		},
		{
			name:    "wrong type",
			record:  testRecord("n", "not-a-node"),
			wantErr: ErrMalformedResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Neo4jRecordOptionalNode(tt.record, "n")
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		m    map[string]any
		want string
	}{
		{
			name: "string value",
			m:    map[string]any{"key": "value"},
			want: "value",
		},
		{
			name: "missing key",
			m:    map[string]any{},
		},
		{
			name: "non-string value",
			m:    map[string]any{"key": int64(1)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, mapString(tt.m, "key"))
		})
	}
}

func TestNeo4jRecordIDs(t *testing.T) {
	t.Parallel()

	id1 := model.MustNewID(model.ResourceTypeLabel)
	id2 := model.MustNewID(model.ResourceTypeLabel)

	tests := []struct {
		name    string
		record  *neo4j.Record
		want    []model.ID
		wantErr error
	}{
		{
			name:   "decodes string ids",
			record: testRecord("ids", []any{id1.String(), id2.String()}),
			want:   []model.ID{id1, id2},
		},
		{
			name:   "skips nil items",
			record: testRecord("ids", []any{id1.String(), nil, id2.String()}),
			want:   []model.ID{id1, id2},
		},
		{
			name:   "empty list",
			record: testRecord("ids", []any{}),
			want:   []model.ID{},
		},
		{
			name:    "missing key",
			record:  &neo4j.Record{Keys: []string{}, Values: []any{}},
			wantErr: ErrMalformedResult,
		},
		{
			name:    "non-string item",
			record:  testRecord("ids", []any{int64(1)}),
			wantErr: ErrMalformedResult,
		},
		{
			name:    "invalid xid",
			record:  testRecord("ids", []any{"not-a-xid"}),
			wantErr: ErrMalformedResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Neo4jRecordIDs(tt.record, "ids", model.ResourceTypeLabel)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNeo4jScanNodeScalars(t *testing.T) {
	t.Parallel()

	type dst struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}

	node := testNode(nil, map[string]any{
		"name": "acme",
		"id":   "skip-me",
	})

	var out dst
	require.NoError(t, Neo4jScanNodeScalars(node, &out, []string{"id"}))
	assert.Equal(t, "acme", out.Name)
	assert.Empty(t, out.ID)
}

func TestNeo4jScanIntoStruct(t *testing.T) {
	t.Parallel()

	type dst struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}

	node := testNode(nil, map[string]any{
		"name": "acme",
		"id":   "skip-me",
	})

	var out dst
	require.NoError(t, Neo4jScanIntoStruct(&node, &out, []string{"id"}))
	assert.Equal(t, "acme", out.Name)
	assert.Empty(t, out.ID)
}

func TestNeo4jParseValueFromRecord(t *testing.T) {
	t.Parallel()

	t.Run("string value", func(t *testing.T) {
		t.Parallel()
		got, err := Neo4jParseValueFromRecord[string](testRecord("k", "v"), "k")
		require.NoError(t, err)
		assert.Equal(t, "v", got)
	})

	t.Run("missing key", func(t *testing.T) {
		t.Parallel()
		_, err := Neo4jParseValueFromRecord[string](&neo4j.Record{Keys: []string{}, Values: []any{}}, "k")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrMalformedResult)
	})
}

func TestNeo4jParseIDsFromRecord(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeUser)
	got, err := Neo4jParseIDsFromRecord(testRecord("ids", []any{id.String()}), "ids", model.ResourceTypeUser.String())
	require.NoError(t, err)
	assert.Equal(t, []model.ID{id}, got)
}

func TestNeo4jRecordPartialProject(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeProject)
	node := testNode([]string{model.ResourceTypeProject.String()}, map[string]any{
		"id":          id.String(),
		"key":         "ENG",
		"name":        "Engineering",
		"description": "desc",
		"logo":        "",
		"status":      "active",
	})

	got, err := Neo4jRecordPartialProject(testRecord("p", node), "p")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "ENG", got.Key)
	assert.Equal(t, "Engineering", got.Name)
	assert.Equal(t, model.ProjectStatusActive, got.Status)
}

func TestNeo4jRecordOptionalPartialNamespace(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeNamespace)
	node := testNode([]string{model.ResourceTypeNamespace.String()}, map[string]any{
		"id":   id.String(),
		"slug": "acme",
		"name": "Acme",
	})

	t.Run("present", func(t *testing.T) {
		t.Parallel()
		got, err := Neo4jRecordOptionalPartialNamespace(testRecord("n", node), "n")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "acme", got.Slug)
		assert.Equal(t, "Acme", got.Name)
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		got, err := Neo4jRecordOptionalPartialNamespace(testRecord("n", nil), "n")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestNeo4jRecordPartialLabels(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeLabel)
	node := testNode([]string{model.ResourceTypeLabel.String()}, map[string]any{
		"id":   id.String(),
		"name": "bug",
	})

	got, err := Neo4jRecordPartialLabels(testRecord("labels", []any{node}), "labels")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, id, got[0].ID)
	assert.Equal(t, "bug", got[0].Name)
}

func TestNeo4jRecordPartialUser(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeUser)
	node := testNode([]string{model.ResourceTypeUser.String()}, map[string]any{
		"id":         id.String(),
		"first_name": "Ada",
		"last_name":  "Lovelace",
		"picture":    "https://example.com/a.png",
	})

	t.Run("present", func(t *testing.T) {
		t.Parallel()
		got, err := Neo4jRecordPartialUser(testRecord("u", node), "u")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "Ada", got.FirstName)
		assert.Equal(t, "Lovelace", got.LastName)
		assert.Equal(t, "https://example.com/a.png", got.Picture)
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		got, err := Neo4jRecordPartialUser(testRecord("u", nil), "u")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
