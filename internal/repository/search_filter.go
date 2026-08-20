package repository

import (
	"strings"
)

func quoteFilterValue(v string) string {
	escaped := strings.ReplaceAll(v, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

// buildSearchFilter compiles a Meilisearch filter from structured fields.
// Empty type filters fail closed so callers cannot run an unfiltered query.
func buildSearchFilter(q SearchQuery) (string, error) {
	if len(q.TypeFilters) == 0 {
		return "", ErrSearchFilter
	}

	branches := make([]string, 0, len(q.TypeFilters))
	for _, tf := range q.TypeFilters {
		if tf.Type == "" || len(tf.ScopeIDs) == 0 {
			continue
		}
		ids := make([]string, len(tf.ScopeIDs))
		for i, id := range tf.ScopeIDs {
			ids[i] = quoteFilterValue(id)
		}
		branches = append(branches, "(type = "+quoteFilterValue(tf.Type)+
			" AND scope_ids IN ["+strings.Join(ids, ", ")+"])")
	}
	if len(branches) == 0 {
		return "", ErrSearchFilter
	}

	filter := strings.Join(branches, " OR ")
	if len(branches) > 1 {
		filter = "(" + filter + ")"
	}
	if q.OrganizationID != "" {
		filter += " AND organization_id = " + quoteFilterValue(q.OrganizationID)
	}
	if q.NamespaceID != "" {
		filter += " AND namespace_id = " + quoteFilterValue(q.NamespaceID)
	}
	if q.ProjectID != "" {
		filter += " AND project_id = " + quoteFilterValue(q.ProjectID)
	}
	return filter, nil
}
