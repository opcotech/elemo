package repository

import (
	"testing"

	"github.com/opcotech/elemo/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileDocumentGetQuery(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeDocument)
	tests := []struct {
		name        string
		proj        DocumentProjection
		loaderNames []string
	}{
		{
			name: "summary has no relation loaders",
			proj: DocumentSummaryProjection(),
		},
		{
			name: "list matches summary",
			proj: DocumentListProjection(),
		},
		{
			name: "detail loads labels and high-cardinality counts",
			proj: DocumentDetailProjection(),
			loaderNames: []string{
				"document.load_labels",
				"document.load_comment_count",
				"document.load_attachment_count",
				"document.load_folder",
				"document.load_relations",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plan, err := CompileQuery(DocumentGetQuery{ID: id, Projection: tt.proj})
			require.NoError(t, err)
			require.Equal(t, "document.get", plan.Root.Name)

			got := make([]string, 0, len(plan.Loaders))
			for _, loader := range plan.Loaders {
				got = append(got, loader.Name)
			}
			if tt.loaderNames == nil {
				assert.Empty(t, got)
				return
			}
			assert.Equal(t, tt.loaderNames, got)
		})
	}
}

func TestCompileDocumentListLibraryQuery(t *testing.T) {
	t.Parallel()

	libraryID := model.MustNewID(model.ResourceTypeNamespace)
	folderID := model.MustNewID(model.ResourceTypeFolder)

	tests := []struct {
		name   string
		filter LibraryListFilter
		want   string
	}{
		{name: "root unfiled", filter: LibraryListFilter{}, want: "document.list_library"},
		{name: "all", filter: LibraryListFilter{All: true}, want: "document.list_library"},
		{name: "folder", filter: LibraryListFilter{FolderID: &folderID}, want: "document.list_library"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			plan, err := CompileQuery(DocumentListLibraryQuery{
				LibraryID: libraryID,
				Filter:    tt.filter,
				Page:      CursorPage{Size: 10},
			})
			require.NoError(t, err)
			assert.Equal(t, tt.want, plan.Root.Name)
			if tt.filter.All {
				assert.NotContains(t, plan.Root.Cypher, "LOCATED_IN")
			}
			if tt.filter.FolderID != nil {
				assert.Contains(t, plan.Root.Cypher, "LOCATED_IN")
			}
			if !tt.filter.All && tt.filter.FolderID == nil {
				assert.Contains(t, plan.Root.Cypher, "NOT (d)-[:LOCATED_IN]->(:Folder)")
			}
		})
	}
}
