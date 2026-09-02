package http

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
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

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(nil, service.ErrNoPermission)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldGet(context.Background(), api.V1CustomFieldGetRequestObject{
			Id: api.Id(def.ID.String()),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldGet403JSONResponse)
		assert.True(t, ok)
	})
}

func testIntegerCustomFieldDefinition(t *testing.T) *model.CustomFieldDefinition {
	t.Helper()
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
	return def
}

func integerCustomFieldSchema(t *testing.T) api.CustomFieldSchema {
	t.Helper()
	var schema api.CustomFieldSchema
	require.NoError(t, schema.FromCustomFieldIntegerSchema(api.CustomFieldIntegerSchema{
		Kind: api.CustomFieldIntegerSchemaKindInteger,
	}))
	return schema
}

func integerCustomFieldValue(t *testing.T, n int64) api.CustomFieldValue {
	t.Helper()
	var value api.CustomFieldValue
	require.NoError(t, value.FromCustomFieldIntegerValue(api.CustomFieldIntegerValue{
		Kind:    api.CustomFieldIntegerValueKindInteger,
		Integer: n,
	}))
	return value
}

func TestCustomFieldController_V1CustomFieldsGet(t *testing.T) {
	t.Parallel()

	def := testIntegerCustomFieldDefinition(t)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().ListDefinitions(gomock.Any(), def.Scope, model.ResourceTypeIssue, false).
			Return([]*model.CustomFieldDefinition{def}, nil)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldsGet(context.Background(), api.V1CustomFieldsGetRequestObject{
			Params: api.V1CustomFieldsGetParams{
				ScopeId:    def.Scope.String(),
				ScopeType:  api.ResourceTypeProject,
				TargetType: api.ResourceTypeIssue,
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1CustomFieldsGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got, 1)
		assert.Equal(t, def.ID.String(), got[0].Id)
	})

	t.Run("bad scope type", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestCustomFieldController(t, mocksvc.NewMockCustomFieldService(ctrl))
		resp, err := c.V1CustomFieldsGet(context.Background(), api.V1CustomFieldsGetRequestObject{
			Params: api.V1CustomFieldsGetParams{
				ScopeId:    def.Scope.String(),
				ScopeType:  api.ResourceType("bogus"),
				TargetType: api.ResourceTypeIssue,
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldsGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().ListDefinitions(gomock.Any(), def.Scope, model.ResourceTypeIssue, false).
			Return(nil, service.ErrNoPermission)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldsGet(context.Background(), api.V1CustomFieldsGetRequestObject{
			Params: api.V1CustomFieldsGetParams{
				ScopeId:    def.Scope.String(),
				ScopeType:  api.ResourceTypeProject,
				TargetType: api.ResourceTypeIssue,
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldsGet403JSONResponse)
		assert.True(t, ok)
	})
}

func TestCustomFieldController_V1CustomFieldsCreate(t *testing.T) {
	t.Parallel()

	def := testIntegerCustomFieldDefinition(t)
	body := &api.V1CustomFieldsCreateJSONRequestBody{
		Key:        "story_points",
		Name:       "Story points",
		Kind:       api.CustomFieldKindInteger,
		ScopeId:    def.Scope.String(),
		ScopeType:  api.ResourceTypeProject,
		TargetType: api.ResourceTypeIssue,
		Schema:     integerCustomFieldSchema(t),
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().CreateDefinition(gomock.Any(), gomock.Any()).Return(def, nil)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldsCreate(context.Background(), api.V1CustomFieldsCreateRequestObject{Body: body})
		require.NoError(t, err)
		got, ok := resp.(api.V1CustomFieldsCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, def.ID.String(), got.Id)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestCustomFieldController(t, mocksvc.NewMockCustomFieldService(ctrl))
		resp, err := c.V1CustomFieldsCreate(context.Background(), api.V1CustomFieldsCreateRequestObject{Body: nil})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().CreateDefinition(gomock.Any(), gomock.Any()).Return(nil, service.ErrNoPermission)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldsCreate(context.Background(), api.V1CustomFieldsCreateRequestObject{Body: body})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldsCreate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("key conflict", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().CreateDefinition(gomock.Any(), gomock.Any()).Return(nil, repository.ErrCustomFieldKeyConflict)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldsCreate(context.Background(), api.V1CustomFieldsCreateRequestObject{Body: body})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldsCreate409JSONResponse)
		assert.True(t, ok)
	})
}

func TestCustomFieldController_V1CustomFieldsSearch(t *testing.T) {
	t.Parallel()

	def := testIntegerCustomFieldDefinition(t)
	issueID := model.MustNewID(model.ResourceTypeIssue)
	n := int64(8)
	body := &api.V1CustomFieldsSearchJSONRequestBody{
		DefinitionId: def.ID.String(),
		Op:           api.CustomFieldPredicateOpEq,
		Integer:      &n,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().Search(gomock.Any(), gomock.Any()).Return([]model.ID{issueID}, nil)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldsSearch(context.Background(), api.V1CustomFieldsSearchRequestObject{Body: body})
		require.NoError(t, err)
		got, ok := resp.(api.V1CustomFieldsSearch200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, []string{issueID.Composite()}, got.ResourceIds)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestCustomFieldController(t, mocksvc.NewMockCustomFieldService(ctrl))
		resp, err := c.V1CustomFieldsSearch(context.Background(), api.V1CustomFieldsSearchRequestObject{Body: nil})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldsSearch400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().Search(gomock.Any(), gomock.Any()).Return(nil, service.ErrNoPermission)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldsSearch(context.Background(), api.V1CustomFieldsSearchRequestObject{Body: body})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldsSearch403JSONResponse)
		assert.True(t, ok)
	})
}

func TestCustomFieldController_V1CustomFieldUpdate(t *testing.T) {
	t.Parallel()

	def := testIntegerCustomFieldDefinition(t)
	name := "Points"
	body := &api.V1CustomFieldUpdateJSONRequestBody{Name: optional.Some(name)}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().UpdateDefinition(gomock.Any(), def.ID, gomock.Any()).Return(def, nil)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldUpdate(context.Background(), api.V1CustomFieldUpdateRequestObject{
			Id:   api.Id(def.ID.String()),
			Body: body,
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1CustomFieldUpdate200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, def.ID.String(), got.Id)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestCustomFieldController(t, mocksvc.NewMockCustomFieldService(ctrl))
		resp, err := c.V1CustomFieldUpdate(context.Background(), api.V1CustomFieldUpdateRequestObject{
			Id:   api.Id(def.ID.String()),
			Body: nil,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldUpdate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().UpdateDefinition(gomock.Any(), def.ID, gomock.Any()).Return(nil, repository.ErrNotFound)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldUpdate(context.Background(), api.V1CustomFieldUpdateRequestObject{
			Id:   api.Id(def.ID.String()),
			Body: body,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldUpdate404JSONResponse)
		assert.True(t, ok)
	})
}

func TestCustomFieldController_V1CustomFieldDelete(t *testing.T) {
	t.Parallel()

	def := testIntegerCustomFieldDefinition(t)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().DeleteDefinition(gomock.Any(), def.ID).Return(nil)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldDelete(context.Background(), api.V1CustomFieldDeleteRequestObject{
			Id: api.Id(def.ID.String()),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldDelete204Response)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().DeleteDefinition(gomock.Any(), def.ID).Return(repository.ErrNotFound)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldDelete(context.Background(), api.V1CustomFieldDeleteRequestObject{
			Id: api.Id(def.ID.String()),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldDelete404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("in use", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().DeleteDefinition(gomock.Any(), def.ID).Return(model.ErrCustomFieldInUse)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldDelete(context.Background(), api.V1CustomFieldDeleteRequestObject{
			Id: api.Id(def.ID.String()),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldDelete400JSONResponse)
		assert.True(t, ok)
	})
}

func TestCustomFieldController_V1CustomFieldArchive(t *testing.T) {
	t.Parallel()

	def := testIntegerCustomFieldDefinition(t)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().ArchiveDefinition(gomock.Any(), def.ID).Return(def, nil)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldArchive(context.Background(), api.V1CustomFieldArchiveRequestObject{
			Id: api.Id(def.ID.String()),
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1CustomFieldArchive200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, def.ID.String(), got.Id)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().ArchiveDefinition(gomock.Any(), def.ID).Return(nil, repository.ErrNotFound)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1CustomFieldArchive(context.Background(), api.V1CustomFieldArchiveRequestObject{
			Id: api.Id(def.ID.String()),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1CustomFieldArchive404JSONResponse)
		assert.True(t, ok)
	})
}

func TestCustomFieldController_V1ResourceCustomFieldsGet(t *testing.T) {
	t.Parallel()

	def := testIntegerCustomFieldDefinition(t)
	issueID := model.MustNewID(model.ResourceTypeIssue)
	n := int64(8)
	value := model.CustomFieldTypedValue{Kind: model.CustomFieldKindInteger, Integer: &n}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().ListEffective(gomock.Any(), issueID).Return([]service.CustomFieldEntry{
			{Definition: def, Value: &value},
		}, nil)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1ResourceCustomFieldsGet(context.Background(), api.V1ResourceCustomFieldsGetRequestObject{
			ResourceType: api.ResourceTypeIssue,
			Id:           api.Id(issueID.String()),
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1ResourceCustomFieldsGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got, 1)
		assert.Equal(t, def.ID.String(), got[0].Definition.Id)
		require.NotNil(t, got[0].Value)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().ListEffective(gomock.Any(), issueID).Return(nil, repository.ErrNotFound)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1ResourceCustomFieldsGet(context.Background(), api.V1ResourceCustomFieldsGetRequestObject{
			ResourceType: api.ResourceTypeIssue,
			Id:           api.Id(issueID.String()),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ResourceCustomFieldsGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestCustomFieldController_V1ResourceCustomFieldValuePut(t *testing.T) {
	t.Parallel()

	def := testIntegerCustomFieldDefinition(t)
	issueID := model.MustNewID(model.ResourceTypeIssue)
	value := integerCustomFieldValue(t, 8)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().SetValue(gomock.Any(), issueID, def.ID, gomock.Any()).Return(nil)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1ResourceCustomFieldValuePut(context.Background(), api.V1ResourceCustomFieldValuePutRequestObject{
			ResourceType: api.ResourceTypeIssue,
			Id:           api.Id(issueID.String()),
			DefinitionId: def.ID.String(),
			Body:         &value,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ResourceCustomFieldValuePut204Response)
		assert.True(t, ok)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestCustomFieldController(t, mocksvc.NewMockCustomFieldService(ctrl))
		resp, err := c.V1ResourceCustomFieldValuePut(context.Background(), api.V1ResourceCustomFieldValuePutRequestObject{
			ResourceType: api.ResourceTypeIssue,
			Id:           api.Id(issueID.String()),
			DefinitionId: def.ID.String(),
			Body:         nil,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ResourceCustomFieldValuePut400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().SetValue(gomock.Any(), issueID, def.ID, gomock.Any()).Return(repository.ErrNotFound)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1ResourceCustomFieldValuePut(context.Background(), api.V1ResourceCustomFieldValuePutRequestObject{
			ResourceType: api.ResourceTypeIssue,
			Id:           api.Id(issueID.String()),
			DefinitionId: def.ID.String(),
			Body:         &value,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ResourceCustomFieldValuePut404JSONResponse)
		assert.True(t, ok)
	})
}

func TestCustomFieldController_V1ResourceCustomFieldValueDelete(t *testing.T) {
	t.Parallel()

	def := testIntegerCustomFieldDefinition(t)
	issueID := model.MustNewID(model.ResourceTypeIssue)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().DeleteValue(gomock.Any(), issueID, def.ID).Return(nil)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1ResourceCustomFieldValueDelete(context.Background(), api.V1ResourceCustomFieldValueDeleteRequestObject{
			ResourceType: api.ResourceTypeIssue,
			Id:           api.Id(issueID.String()),
			DefinitionId: def.ID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ResourceCustomFieldValueDelete204Response)
		assert.True(t, ok)
	})

	t.Run("required", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().DeleteValue(gomock.Any(), issueID, def.ID).Return(model.ErrCustomFieldRequired)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1ResourceCustomFieldValueDelete(context.Background(), api.V1ResourceCustomFieldValueDeleteRequestObject{
			ResourceType: api.ResourceTypeIssue,
			Id:           api.Id(issueID.String()),
			DefinitionId: def.ID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ResourceCustomFieldValueDelete400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocksvc.NewMockCustomFieldService(ctrl)
		svc.EXPECT().DeleteValue(gomock.Any(), issueID, def.ID).Return(repository.ErrNotFound)

		c := newTestCustomFieldController(t, svc)
		resp, err := c.V1ResourceCustomFieldValueDelete(context.Background(), api.V1ResourceCustomFieldValueDeleteRequestObject{
			ResourceType: api.ResourceTypeIssue,
			Id:           api.Id(issueID.String()),
			DefinitionId: def.ID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ResourceCustomFieldValueDelete404JSONResponse)
		assert.True(t, ok)
	})
}
