package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
)

func TestCompileCustomFieldSearch(t *testing.T) {
	t.Parallel()

	definitionID := model.MustNewID(model.ResourceTypeCustomFieldDefinition)

	t.Run("eq text", func(t *testing.T) {
		t.Parallel()
		text := "ready"
		query, args, err := compileCustomFieldSearch(definitionID, CustomFieldPredicate{
			Op:   CustomFieldPredEq,
			Text: &text,
		}, 10)
		require.NoError(t, err)
		assert.Contains(t, query, "text_value = $2")
		assert.Contains(t, query, "index_exact")
		assert.Equal(t, []any{definitionID, text, 10}, args)
	})

	t.Run("match text", func(t *testing.T) {
		t.Parallel()
		text := "search me"
		query, args, err := compileCustomFieldSearch(definitionID, CustomFieldPredicate{
			Op:   CustomFieldPredMatch,
			Text: &text,
		}, 5)
		require.NoError(t, err)
		assert.Contains(t, query, "to_tsvector")
		assert.Contains(t, query, "index_fulltext")
		assert.Equal(t, []any{definitionID, text, 5}, args)
	})

	t.Run("gt integer", func(t *testing.T) {
		t.Parallel()
		n := int64(8)
		query, args, err := compileCustomFieldSearch(definitionID, CustomFieldPredicate{
			Op:      CustomFieldPredGt,
			Integer: &n,
		}, 10)
		require.NoError(t, err)
		assert.Contains(t, query, "integer_value > $2")
		assert.Contains(t, query, "index_range")
		assert.Equal(t, []any{definitionID, n, 10}, args)
	})

	t.Run("gte decimal", func(t *testing.T) {
		t.Parallel()
		dec := "10.00"
		query, _, err := compileCustomFieldSearch(definitionID, CustomFieldPredicate{
			Op:      CustomFieldPredGte,
			Decimal: &dec,
		}, 10)
		require.NoError(t, err)
		assert.Contains(t, query, "decimal_value >= $2")
	})

	t.Run("lt date", func(t *testing.T) {
		t.Parallel()
		day := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		query, _, err := compileCustomFieldSearch(definitionID, CustomFieldPredicate{
			Op:   CustomFieldPredLt,
			Date: &day,
		}, 10)
		require.NoError(t, err)
		assert.Contains(t, query, "date_value < $2")
	})

	t.Run("lte datetime", func(t *testing.T) {
		t.Parallel()
		ts := time.Now().UTC()
		query, _, err := compileCustomFieldSearch(definitionID, CustomFieldPredicate{
			Op:       CustomFieldPredLte,
			DateTime: &ts,
		}, 10)
		require.NoError(t, err)
		assert.Contains(t, query, "datetime_value <= $2")
	})

	t.Run("unknown operator", func(t *testing.T) {
		t.Parallel()
		text := "x"
		_, _, err := compileCustomFieldSearch(definitionID, CustomFieldPredicate{
			Op:   CustomFieldPredicateOp("bogus"),
			Text: &text,
		}, 10)
		require.Error(t, err)
	})

	t.Run("empty predicate", func(t *testing.T) {
		t.Parallel()
		_, _, err := compileCustomFieldSearch(definitionID, CustomFieldPredicate{Op: CustomFieldPredEq}, 10)
		require.Error(t, err)
	})
}
