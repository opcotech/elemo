package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func TestCustomFieldSchemaRoundTrip_RemainingKinds(t *testing.T) {
	t.Parallel()

	t.Run("text", func(t *testing.T) {
		t.Parallel()
		minLen, maxLen := 1, 80
		pattern := "^[a-z]+$"
		schema := model.CustomFieldSchema{Text: &model.CustomFieldTextSchema{
			MinLength: &minLen,
			MaxLength: &maxLen,
			Pattern:   pattern,
		}}
		dto, err := customFieldSchemaToAPI(model.CustomFieldKindText, schema)
		require.NoError(t, err)
		got, err := customFieldSchemaFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.Text)
		assert.Equal(t, minLen, *got.Text.MinLength)
		assert.Equal(t, maxLen, *got.Text.MaxLength)
		assert.Equal(t, pattern, got.Text.Pattern)
	})

	t.Run("decimal", func(t *testing.T) {
		t.Parallel()
		scale := 2
		schema := model.CustomFieldSchema{Decimal: &model.CustomFieldDecimalSchema{
			Min:   "0.01",
			Max:   "100.00",
			Scale: &scale,
		}}
		dto, err := customFieldSchemaToAPI(model.CustomFieldKindDecimal, schema)
		require.NoError(t, err)
		got, err := customFieldSchemaFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.Decimal)
		assert.Equal(t, "0.01", got.Decimal.Min)
		assert.Equal(t, "100.00", got.Decimal.Max)
		require.NotNil(t, got.Decimal.Scale)
		assert.Equal(t, 2, *got.Decimal.Scale)
	})

	t.Run("boolean", func(t *testing.T) {
		t.Parallel()
		schema := model.CustomFieldSchema{Boolean: &model.CustomFieldBooleanSchema{}}
		dto, err := customFieldSchemaToAPI(model.CustomFieldKindBoolean, schema)
		require.NoError(t, err)
		got, err := customFieldSchemaFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.Boolean)
	})

	t.Run("date", func(t *testing.T) {
		t.Parallel()
		minDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		schema := model.CustomFieldSchema{Date: &model.CustomFieldDateSchema{Min: &minDate}}
		dto, err := customFieldSchemaToAPI(model.CustomFieldKindDate, schema)
		require.NoError(t, err)
		got, err := customFieldSchemaFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.Date)
		require.NotNil(t, got.Date.Min)
		assert.True(t, got.Date.Min.Equal(minDate))
	})

	t.Run("datetime", func(t *testing.T) {
		t.Parallel()
		minDate := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
		schema := model.CustomFieldSchema{DateTime: &model.CustomFieldDateTimeSchema{Min: &minDate}}
		dto, err := customFieldSchemaToAPI(model.CustomFieldKindDatetime, schema)
		require.NoError(t, err)
		got, err := customFieldSchemaFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.DateTime)
		require.NotNil(t, got.DateTime.Min)
		assert.True(t, got.DateTime.Min.Equal(minDate))
	})

	t.Run("url", func(t *testing.T) {
		t.Parallel()
		schema := model.CustomFieldSchema{URL: &model.CustomFieldURLSchema{AllowedSchemes: []string{"https"}}}
		dto, err := customFieldSchemaToAPI(model.CustomFieldKindURL, schema)
		require.NoError(t, err)
		got, err := customFieldSchemaFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.URL)
		assert.Equal(t, []string{"https"}, got.URL.AllowedSchemes)
	})

	t.Run("multi select", func(t *testing.T) {
		t.Parallel()
		schema := model.CustomFieldSchema{Select: &model.CustomFieldSelectSchema{
			Options: []model.CustomFieldOption{{Key: "alpha", Label: "Alpha"}},
		}}
		dto, err := customFieldSchemaToAPI(model.CustomFieldKindMultiSelect, schema)
		require.NoError(t, err)
		got, err := customFieldSchemaFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.Select)
		require.Len(t, got.Select.Options, 1)
		assert.Equal(t, "alpha", got.Select.Options[0].Key)
	})

	t.Run("user reference", func(t *testing.T) {
		t.Parallel()
		schema := model.CustomFieldSchema{UserReference: &model.CustomFieldUserReferenceSchema{Multiple: true}}
		dto, err := customFieldSchemaToAPI(model.CustomFieldKindUserReference, schema)
		require.NoError(t, err)
		got, err := customFieldSchemaFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.UserReference)
		assert.True(t, got.UserReference.Multiple)
	})

	t.Run("resource reference", func(t *testing.T) {
		t.Parallel()
		schema := model.CustomFieldSchema{ResourceReference: &model.CustomFieldResourceReferenceSchema{
			AllowedTypes: []model.ResourceType{model.ResourceTypeIssue},
			Multiple:     true,
		}}
		dto, err := customFieldSchemaToAPI(model.CustomFieldKindResourceReference, schema)
		require.NoError(t, err)
		got, err := customFieldSchemaFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.ResourceReference)
		assert.Equal(t, []model.ResourceType{model.ResourceTypeIssue}, got.ResourceReference.AllowedTypes)
		assert.True(t, got.ResourceReference.Multiple)
	})
}

func TestCustomFieldValueRoundTrip_RemainingKinds(t *testing.T) {
	t.Parallel()

	t.Run("decimal", func(t *testing.T) {
		t.Parallel()
		dec := "12.50"
		dto, err := customFieldValueToAPI(model.CustomFieldTypedValue{Kind: model.CustomFieldKindDecimal, Decimal: &dec})
		require.NoError(t, err)
		got, err := customFieldValueFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.Decimal)
		assert.Equal(t, dec, *got.Decimal)
	})

	t.Run("boolean", func(t *testing.T) {
		t.Parallel()
		flag := true
		dto, err := customFieldValueToAPI(model.CustomFieldTypedValue{Kind: model.CustomFieldKindBoolean, Boolean: &flag})
		require.NoError(t, err)
		got, err := customFieldValueFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.Boolean)
		assert.True(t, *got.Boolean)
	})

	t.Run("date", func(t *testing.T) {
		t.Parallel()
		day := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		dto, err := customFieldValueToAPI(model.CustomFieldTypedValue{Kind: model.CustomFieldKindDate, Date: &day})
		require.NoError(t, err)
		got, err := customFieldValueFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.Date)
		assert.True(t, got.Date.Equal(day))
	})

	t.Run("datetime", func(t *testing.T) {
		t.Parallel()
		ts := time.Date(2024, 3, 15, 13, 45, 0, 0, time.UTC)
		dto, err := customFieldValueToAPI(model.CustomFieldTypedValue{Kind: model.CustomFieldKindDatetime, DateTime: &ts})
		require.NoError(t, err)
		got, err := customFieldValueFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.DateTime)
		assert.True(t, got.DateTime.Equal(ts))
	})

	t.Run("url", func(t *testing.T) {
		t.Parallel()
		raw := "https://example.com/docs"
		dto, err := customFieldValueToAPI(model.CustomFieldTypedValue{Kind: model.CustomFieldKindURL, URL: &raw})
		require.NoError(t, err)
		got, err := customFieldValueFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.URL)
		assert.Equal(t, raw, *got.URL)
	})

	t.Run("single select", func(t *testing.T) {
		t.Parallel()
		key := "prod"
		dto, err := customFieldValueToAPI(model.CustomFieldTypedValue{Kind: model.CustomFieldKindSingleSelect, OptionKey: &key})
		require.NoError(t, err)
		got, err := customFieldValueFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.OptionKey)
		assert.Equal(t, key, *got.OptionKey)
	})

	t.Run("multi select", func(t *testing.T) {
		t.Parallel()
		keys := []string{"prod", "staging"}
		dto, err := customFieldValueToAPI(model.CustomFieldTypedValue{Kind: model.CustomFieldKindMultiSelect, OptionKeys: keys})
		require.NoError(t, err)
		got, err := customFieldValueFromAPI(dto)
		require.NoError(t, err)
		assert.Equal(t, keys, got.OptionKeys)
	})

	t.Run("user reference", func(t *testing.T) {
		t.Parallel()
		userID := model.MustNewID(model.ResourceTypeUser)
		dto, err := customFieldValueToAPI(model.CustomFieldTypedValue{Kind: model.CustomFieldKindUserReference, UserID: &userID})
		require.NoError(t, err)
		got, err := customFieldValueFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.UserID)
		assert.Equal(t, userID, *got.UserID)
	})

	t.Run("resource reference", func(t *testing.T) {
		t.Parallel()
		issueID := model.MustNewID(model.ResourceTypeIssue)
		dto, err := customFieldValueToAPI(model.CustomFieldTypedValue{Kind: model.CustomFieldKindResourceReference, ResourceID: &issueID})
		require.NoError(t, err)
		got, err := customFieldValueFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.ResourceID)
		assert.Equal(t, issueID, *got.ResourceID)
	})
}

func TestCreateCustomFieldOptsFromAPI(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		scope := model.MustNewID(model.ResourceTypeProject)
		order := 3
		desc := "Points for a story"
		body := &api.V1CustomFieldsCreateJSONRequestBody{
			Key:         "story_points",
			Name:        "Story points",
			Description: &desc,
			Kind:        api.CustomFieldKindInteger,
			ScopeId:     scope.String(),
			ScopeType:   api.ResourceTypeProject,
			TargetType:  api.ResourceTypeIssue,
			Order:       &order,
			Schema:      integerCustomFieldSchema(t),
		}
		opts, err := createCustomFieldOptsFromAPI(body)
		require.NoError(t, err)
		assert.Equal(t, "story_points", opts.Key)
		assert.Equal(t, "Story points", opts.Name)
		assert.Equal(t, desc, opts.Description)
		assert.Equal(t, model.CustomFieldKindInteger, opts.Kind)
		assert.Equal(t, scope, opts.Scope)
		assert.Equal(t, model.ResourceTypeIssue, opts.TargetType)
		require.True(t, opts.SortOrder.Defined)
		assert.Equal(t, 3, *opts.SortOrder.Value)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		_, err := createCustomFieldOptsFromAPI(nil)
		require.Error(t, err)
	})
}

func TestUpdateCustomFieldOptsFromAPI(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		name := "Points"
		opts, err := updateCustomFieldOptsFromAPI(&api.V1CustomFieldUpdateJSONRequestBody{
			Name: optional.Some(name),
		})
		require.NoError(t, err)
		require.True(t, opts.Name.Defined)
		assert.Equal(t, name, *opts.Name.Value)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		_, err := updateCustomFieldOptsFromAPI(nil)
		require.Error(t, err)
	})
}

func TestCustomFieldSearchQueryFromAPI(t *testing.T) {
	t.Parallel()

	definitionID := model.MustNewID(model.ResourceTypeCustomFieldDefinition)
	n := int64(8)
	limit := 25
	query, err := customFieldSearchQueryFromAPI(&api.V1CustomFieldsSearchJSONRequestBody{
		DefinitionId: definitionID.String(),
		Op:           api.CustomFieldPredicateOpGte,
		Integer:      &n,
		Limit:        &limit,
	})
	require.NoError(t, err)
	assert.Equal(t, definitionID, query.DefinitionID)
	assert.Equal(t, repository.CustomFieldPredGte, query.Predicate.Op)
	require.NotNil(t, query.Predicate.Integer)
	assert.Equal(t, n, *query.Predicate.Integer)
	assert.Equal(t, 25, query.Limit)

	_, err = customFieldSearchQueryFromAPI(nil)
	require.Error(t, err)
}

func TestCustomFieldWritesFromAPI(t *testing.T) {
	t.Parallel()

	definitionID := model.MustNewID(model.ResourceTypeCustomFieldDefinition)
	value := integerCustomFieldValue(t, 5)
	got, err := customFieldWritesFromAPI([]api.CustomFieldWrite{{
		DefinitionId: definitionID.String(),
		Value:        value,
	}})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, definitionID, got[0].DefinitionID)
	require.NotNil(t, got[0].Value.Integer)
	assert.Equal(t, int64(5), *got[0].Value.Integer)

	_, err = customFieldWritesFromAPI([]api.CustomFieldWrite{{
		DefinitionId: "not-an-id",
		Value:        value,
	}})
	require.Error(t, err)
}

func TestCustomFieldEntryToDTO(t *testing.T) {
	t.Parallel()

	def := testIntegerCustomFieldDefinition(t)
	n := int64(8)
	value := model.CustomFieldTypedValue{Kind: model.CustomFieldKindInteger, Integer: &n}
	dto, err := customFieldEntryToDTO(service.CustomFieldEntry{Definition: def, Value: &value})
	require.NoError(t, err)
	assert.Equal(t, def.ID.String(), dto.Definition.Id)
	require.NotNil(t, dto.Value)

	dto, err = customFieldEntryToDTO(service.CustomFieldEntry{Definition: def})
	require.NoError(t, err)
	assert.Nil(t, dto.Value)
}

func TestParseUserRefID(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	got, err := parseUserRefID(userID.String())
	require.NoError(t, err)
	assert.Equal(t, userID, got)

	got, err = parseUserRefID(userID.Composite())
	require.NoError(t, err)
	assert.Equal(t, userID, got)

	_, err = parseUserRefID("not-an-id")
	require.Error(t, err)

	issueID := model.MustNewID(model.ResourceTypeIssue)
	_, err = parseUserRefID(issueID.Composite())
	require.ErrorIs(t, err, model.ErrInvalidID)
}

func TestParseResourceRefID(t *testing.T) {
	t.Parallel()

	issueID := model.MustNewID(model.ResourceTypeIssue)
	got, err := parseResourceRefID(issueID.Composite())
	require.NoError(t, err)
	assert.Equal(t, issueID, got)

	_, err = parseResourceRefID(issueID.String())
	require.ErrorIs(t, err, model.ErrInvalidID)
}
