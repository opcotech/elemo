package http

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func TestCustomFieldSchemaRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("integer", func(t *testing.T) {
		t.Parallel()
		minVal := int64(1)
		maxVal := int64(10)
		schema := model.CustomFieldSchema{Integer: &model.CustomFieldIntegerSchema{Min: &minVal, Max: &maxVal}}
		dto, err := customFieldSchemaToAPI(model.CustomFieldKindInteger, schema)
		require.NoError(t, err)
		got, err := customFieldSchemaFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.Integer)
		assert.Equal(t, minVal, *got.Integer.Min)
		assert.Equal(t, maxVal, *got.Integer.Max)
	})

	t.Run("select", func(t *testing.T) {
		t.Parallel()
		schema := model.CustomFieldSchema{Select: &model.CustomFieldSelectSchema{
			Options: []model.CustomFieldOption{{Key: "alpha", Label: "Alpha"}},
		}}
		dto, err := customFieldSchemaToAPI(model.CustomFieldKindSingleSelect, schema)
		require.NoError(t, err)
		got, err := customFieldSchemaFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.Select)
		require.Len(t, got.Select.Options, 1)
		assert.Equal(t, "alpha", got.Select.Options[0].Key)
	})
}

func TestCustomFieldValueRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("text", func(t *testing.T) {
		t.Parallel()
		text := "hello"
		value := model.CustomFieldTypedValue{Kind: model.CustomFieldKindText, Text: &text}
		dto, err := customFieldValueToAPI(value)
		require.NoError(t, err)
		got, err := customFieldValueFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.Text)
		assert.Equal(t, text, *got.Text)
		assert.Equal(t, model.CustomFieldKindText, got.Kind)
	})

	t.Run("integer", func(t *testing.T) {
		t.Parallel()
		n := int64(8)
		value := model.CustomFieldTypedValue{Kind: model.CustomFieldKindInteger, Integer: &n}
		dto, err := customFieldValueToAPI(value)
		require.NoError(t, err)
		got, err := customFieldValueFromAPI(dto)
		require.NoError(t, err)
		require.NotNil(t, got.Integer)
		assert.Equal(t, n, *got.Integer)
	})
}

func TestCustomFieldDefinitionToDTO(t *testing.T) {
	t.Parallel()

	scope := model.MustNewID(model.ResourceTypeProject)
	owner := model.MustNewID(model.ResourceTypeUser)
	def, err := model.NewCustomFieldDefinition(
		"story_points",
		"Story points",
		model.CustomFieldKindInteger,
		scope,
		owner,
		model.ResourceTypeIssue,
		model.CustomFieldSchema{Integer: &model.CustomFieldIntegerSchema{}},
	)
	require.NoError(t, err)
	def.ID = model.MustNewID(model.ResourceTypeCustomFieldDefinition)

	dto, err := customFieldDefinitionToDTO(def)
	require.NoError(t, err)
	assert.Equal(t, def.ID.String(), dto.Id)
	assert.Equal(t, "story_points", dto.Key)
	assert.Equal(t, api.CustomFieldKindInteger, dto.Kind)
	assert.Equal(t, scope.String(), dto.ScopeId)
	assert.Equal(t, api.ResourceTypeProject, dto.ScopeType)
	assert.Equal(t, api.ResourceTypeIssue, dto.TargetType)
	assert.Equal(t, def.SortOrder, dto.Order)
}

func TestNewCustomFieldController(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, err := NewCustomFieldController(mocksvc.NewMockCustomFieldService(ctrl))
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("no service", func(t *testing.T) {
		t.Parallel()
		_, err := NewCustomFieldController(nil)
		assert.ErrorIs(t, err, ErrNoCustomFieldService)
	})
}

func TestCustomFieldController_V1CustomFieldGet(t *testing.T) {
	t.Parallel()

	scope := model.MustNewID(model.ResourceTypeProject)
	owner := model.MustNewID(model.ResourceTypeUser)
	def, err := model.NewCustomFieldDefinition(
		"story_points",
		"Story points",
		model.CustomFieldKindInteger,
		scope,
		owner,
		model.ResourceTypeIssue,
		model.CustomFieldSchema{Integer: &model.CustomFieldIntegerSchema{}},
	)
	require.NoError(t, err)
	def.ID = model.MustNewID(model.ResourceTypeCustomFieldDefinition)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(def, nil)

		c, err := NewCustomFieldController(svc)
		require.NoError(t, err)

		resp, err := c.V1CustomFieldGet(context.Background(), api.V1CustomFieldGetRequestObject{
			Id: api.Id(def.ID.String()),
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1CustomFieldGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, def.ID.String(), got.Id)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(nil, repository.ErrNotFound)

		c, err := NewCustomFieldController(svc)
		require.NoError(t, err)

		resp, err := c.V1CustomFieldGet(context.Background(), api.V1CustomFieldGetRequestObject{
			Id: api.Id(def.ID.String()),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldGet404JSONResponse)
		assert.True(t, ok)
	})
}
