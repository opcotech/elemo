package repository

import (
	"context"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
)

func TestDecodeIssueLinks(t *testing.T) {
	t.Parallel()

	t.Run("splits encoded url and label", func(t *testing.T) {
		t.Parallel()
		got := decodeIssueLinks(map[string]any{
			"links": []any{
				"https://example.com/a" + issueLinkLabelSep + "Spec",
				"https://example.com/b" + issueLinkLabelSep + "Runbook",
			},
		})
		assert.Equal(t, []model.IssueLink{
			{URL: "https://example.com/a", Label: "Spec"},
			{URL: "https://example.com/b", Label: "Runbook"},
		}, got)
	})

	t.Run("defaults missing labels to the url", func(t *testing.T) {
		t.Parallel()
		got := decodeIssueLinks(map[string]any{
			"links": []string{"https://example.com/legacy"},
		})
		assert.Equal(t, []model.IssueLink{
			{URL: "https://example.com/legacy", Label: "https://example.com/legacy"},
		}, got)
	})

	t.Run("zips leftover parallel label lists", func(t *testing.T) {
		t.Parallel()
		got := decodeIssueLinks(map[string]any{
			"links":       []any{"https://example.com/a", "https://example.com/b"},
			"link_labels": []any{"Spec", "Runbook"},
		})
		assert.Equal(t, []model.IssueLink{
			{URL: "https://example.com/a", Label: "Spec"},
			{URL: "https://example.com/b", Label: "Runbook"},
		}, got)
	})

	t.Run("returns an empty slice when links are absent", func(t *testing.T) {
		t.Parallel()
		got := decodeIssueLinks(map[string]any{})
		assert.Equal(t, []model.IssueLink{}, got)
	})
}

func TestEncodeIssueLinks(t *testing.T) {
	t.Parallel()

	got := encodeIssueLinks([]model.IssueLink{
		{URL: "https://example.com/a", Label: "Spec"},
		{URL: "https://example.com/b", Label: "https://example.com/b"},
		{URL: "https://example.com/c", Label: ""},
	})
	assert.Equal(t, []string{
		"https://example.com/a" + issueLinkLabelSep + "Spec",
		"https://example.com/b",
		"https://example.com/c",
	}, got)

	assert.Equal(t, []string{}, encodeIssueLinks(nil))
}

func TestUpdateIssueOpts_patch(t *testing.T) {
	t.Parallel()

	t.Run("nil description clears the field", func(t *testing.T) {
		t.Parallel()

		got := UpdateIssueOpts{
			Description: optional.Null[string](),
		}.patch()
		require.Contains(t, got, "description")
		assert.Nil(t, got["description"])
	})

	t.Run("set description", func(t *testing.T) {
		t.Parallel()

		got := UpdateIssueOpts{
			Description: optional.Some("updated description"),
		}.patch()
		assert.Equal(t, "updated description", got["description"])
	})

	t.Run("undefined description is omitted", func(t *testing.T) {
		t.Parallel()

		got := UpdateIssueOpts{}.patch()
		_, ok := got["description"]
		assert.False(t, ok)
	})
}

func TestParentProjectKeyFromRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		record   *neo4j.Record
		fallback string
		want     string
	}{
		{
			name:     "missing field uses fallback",
			record:   &neo4j.Record{Keys: []string{}, Values: []any{}},
			fallback: "ENG",
			want:     "ENG",
		},
		{
			name:     "nil value uses fallback",
			record:   &neo4j.Record{Keys: []string{"parent_project_key"}, Values: []any{nil}},
			fallback: "ENG",
			want:     "ENG",
		},
		{
			name:     "empty string uses fallback",
			record:   &neo4j.Record{Keys: []string{"parent_project_key"}, Values: []any{""}},
			fallback: "ENG",
			want:     "ENG",
		},
		{
			name:     "parent project key wins",
			record:   &neo4j.Record{Keys: []string{"parent_project_key"}, Values: []any{"PLAT"}},
			fallback: "ENG",
			want:     "PLAT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, parentProjectKeyFromRecord(tt.record, tt.fallback))
		})
	}
}

func TestRelationCountAfterCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		hasParent bool
		want      int64
	}{
		{
			name:      "without parent",
			hasParent: false,
			want:      0,
		},
		{
			name:      "with parent",
			hasParent: true,
			want:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := relationCountAfterCreate(tt.hasParent)
			require.NotNil(t, got)
			assert.Equal(t, tt.want, *got)
		})
	}
}

func TestNeo4jIssueRepository_applyIssueLoadersUnknown(t *testing.T) {
	t.Parallel()

	r := new(Neo4jIssueRepository)
	err := r.applyIssueLoaders(context.Background(), nil, QueryPlan{
		Loaders: []CompiledQuery{{Name: "issue.load_unknown", Params: map[string]any{}}},
	}, []*issueDetailRow{{
		projectKey: "ENG",
		issue:      &Issue{ID: model.MustNewID(model.ResourceTypeIssue)},
	}})
	assert.ErrorIs(t, err, ErrQueryCompile)
}
