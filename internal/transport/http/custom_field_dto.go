package http

import (
	"errors"

	oapiTypes "github.com/oapi-codegen/runtime/types"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func resourceTypeFromAPI(rt api.ResourceType) (model.ResourceType, error) {
	return model.ResourceTypeString(string(rt))
}

func customFieldKindFromAPI(kind api.CustomFieldKind) (model.CustomFieldKind, error) {
	return model.CustomFieldKindString(string(kind))
}

func definitionIDFromAPI(id string) (model.ID, error) {
	return model.NewIDFromString(id, model.ResourceTypeCustomFieldDefinition.String())
}

func resourceIDFromAPI(resourceType api.ResourceType, id string) (model.ID, error) {
	rt, err := resourceTypeFromAPI(resourceType)
	if err != nil {
		return model.ID{}, err
	}
	return model.NewIDFromString(id, rt.String())
}

func parseUserRefID(raw string) (model.ID, error) {
	if id, err := model.ParseCompositeID(raw); err == nil {
		if id.Type != model.ResourceTypeUser {
			return model.ID{}, model.ErrInvalidID
		}
		return id, nil
	}
	return model.NewIDFromString(raw, model.ResourceTypeUser.String())
}

func parseResourceRefID(raw string) (model.ID, error) {
	if id, err := model.ParseCompositeID(raw); err == nil {
		return id, nil
	}
	return model.ID{}, errors.Join(model.ErrInvalidID, errors.New("resource reference requires Type:xid"))
}

func derefBool(v *bool) bool {
	return v != nil && *v
}

func customFieldSchemaFromAPI(schema api.CustomFieldSchema) (model.CustomFieldSchema, error) {
	raw, err := schema.ValueByDiscriminator()
	if err != nil {
		return model.CustomFieldSchema{}, errors.Join(model.ErrInvalidCustomFieldDetails, err)
	}

	switch s := raw.(type) {
	case api.CustomFieldTextSchema:
		out := model.CustomFieldTextSchema{MinLength: s.MinLength, MaxLength: s.MaxLength}
		if s.Pattern != nil {
			out.Pattern = *s.Pattern
		}
		return model.CustomFieldSchema{Text: &out}, nil
	case api.CustomFieldIntegerSchema:
		return model.CustomFieldSchema{Integer: &model.CustomFieldIntegerSchema{Min: s.Min, Max: s.Max}}, nil
	case api.CustomFieldDecimalSchema:
		out := model.CustomFieldDecimalSchema{Scale: s.Scale}
		if s.Min != nil {
			out.Min = *s.Min
		}
		if s.Max != nil {
			out.Max = *s.Max
		}
		return model.CustomFieldSchema{Decimal: &out}, nil
	case api.CustomFieldBooleanSchema:
		return model.CustomFieldSchema{Boolean: &model.CustomFieldBooleanSchema{}}, nil
	case api.CustomFieldDateSchema:
		out := model.CustomFieldDateSchema{}
		if s.Min != nil {
			t := s.Min.Time
			out.Min = &t
		}
		if s.Max != nil {
			t := s.Max.Time
			out.Max = &t
		}
		return model.CustomFieldSchema{Date: &out}, nil
	case api.CustomFieldDateTimeSchema:
		return model.CustomFieldSchema{DateTime: &model.CustomFieldDateTimeSchema{Min: s.Min, Max: s.Max}}, nil
	case api.CustomFieldURLSchema:
		return model.CustomFieldSchema{URL: &model.CustomFieldURLSchema{AllowedSchemes: s.AllowedSchemes}}, nil
	case api.CustomFieldSelectSchema:
		options := make([]model.CustomFieldOption, len(s.Options))
		for i, opt := range s.Options {
			options[i] = customFieldOptionFromAPI(opt)
		}
		return model.CustomFieldSchema{Select: &model.CustomFieldSelectSchema{Options: options}}, nil
	case api.CustomFieldUserReferenceSchema:
		return model.CustomFieldSchema{UserReference: &model.CustomFieldUserReferenceSchema{
			Multiple: derefBool(s.Multiple),
		}}, nil
	case api.CustomFieldResourceReferenceSchema:
		types := make([]model.ResourceType, len(s.AllowedTypes))
		for i, rt := range s.AllowedTypes {
			parsed, err := resourceTypeFromAPI(rt)
			if err != nil {
				return model.CustomFieldSchema{}, errors.Join(model.ErrInvalidCustomFieldDetails, err)
			}
			types[i] = parsed
		}
		return model.CustomFieldSchema{ResourceReference: &model.CustomFieldResourceReferenceSchema{
			AllowedTypes: types,
			Multiple:     derefBool(s.Multiple),
		}}, nil
	default:
		return model.CustomFieldSchema{}, model.ErrInvalidCustomFieldKind
	}
}

func customFieldOptionFromAPI(opt api.CustomFieldOption) model.CustomFieldOption {
	out := model.CustomFieldOption{
		Key:      opt.Key,
		Label:    opt.Label,
		Disabled: opt.Disabled,
	}
	if opt.Color != nil {
		out.Color = *opt.Color
	}
	if opt.Order != nil {
		out.Order = *opt.Order
	}
	return out
}

func customFieldSchemaToAPI(kind model.CustomFieldKind, schema model.CustomFieldSchema) (api.CustomFieldSchema, error) {
	var dto api.CustomFieldSchema
	switch kind {
	case model.CustomFieldKindText:
		s := schema.Text
		if s == nil {
			s = &model.CustomFieldTextSchema{}
		}
		var pattern *string
		if s.Pattern != "" {
			pattern = &s.Pattern
		}
		if err := dto.FromCustomFieldTextSchema(api.CustomFieldTextSchema{
			Kind:      api.CustomFieldTextSchemaKindText,
			MinLength: s.MinLength,
			MaxLength: s.MaxLength,
			Pattern:   pattern,
		}); err != nil {
			return api.CustomFieldSchema{}, err
		}
	case model.CustomFieldKindInteger:
		s := schema.Integer
		if s == nil {
			s = &model.CustomFieldIntegerSchema{}
		}
		if err := dto.FromCustomFieldIntegerSchema(api.CustomFieldIntegerSchema{
			Kind: api.CustomFieldIntegerSchemaKindInteger,
			Min:  s.Min,
			Max:  s.Max,
		}); err != nil {
			return api.CustomFieldSchema{}, err
		}
	case model.CustomFieldKindDecimal:
		s := schema.Decimal
		if s == nil {
			s = &model.CustomFieldDecimalSchema{}
		}
		var minVal, maxVal *string
		if s.Min != "" {
			minVal = &s.Min
		}
		if s.Max != "" {
			maxVal = &s.Max
		}
		if err := dto.FromCustomFieldDecimalSchema(api.CustomFieldDecimalSchema{
			Kind:  api.CustomFieldDecimalSchemaKindDecimal,
			Min:   minVal,
			Max:   maxVal,
			Scale: s.Scale,
		}); err != nil {
			return api.CustomFieldSchema{}, err
		}
	case model.CustomFieldKindBoolean:
		if err := dto.FromCustomFieldBooleanSchema(api.CustomFieldBooleanSchema{
			Kind: api.CustomFieldBooleanSchemaKindBoolean,
		}); err != nil {
			return api.CustomFieldSchema{}, err
		}
	case model.CustomFieldKindDate:
		s := schema.Date
		if s == nil {
			s = &model.CustomFieldDateSchema{}
		}
		out := api.CustomFieldDateSchema{Kind: api.CustomFieldDateSchemaKindDate}
		if s.Min != nil {
			d := oapiTypes.Date{Time: *s.Min}
			out.Min = &d
		}
		if s.Max != nil {
			d := oapiTypes.Date{Time: *s.Max}
			out.Max = &d
		}
		if err := dto.FromCustomFieldDateSchema(out); err != nil {
			return api.CustomFieldSchema{}, err
		}
	case model.CustomFieldKindDatetime:
		s := schema.DateTime
		if s == nil {
			s = &model.CustomFieldDateTimeSchema{}
		}
		if err := dto.FromCustomFieldDateTimeSchema(api.CustomFieldDateTimeSchema{
			Kind: api.CustomFieldDateTimeSchemaKindDatetime,
			Min:  s.Min,
			Max:  s.Max,
		}); err != nil {
			return api.CustomFieldSchema{}, err
		}
	case model.CustomFieldKindURL:
		s := schema.URL
		if s == nil {
			s = &model.CustomFieldURLSchema{}
		}
		if err := dto.FromCustomFieldURLSchema(api.CustomFieldURLSchema{
			Kind:           api.CustomFieldURLSchemaKindUrl,
			AllowedSchemes: s.AllowedSchemes,
		}); err != nil {
			return api.CustomFieldSchema{}, err
		}
	case model.CustomFieldKindSingleSelect, model.CustomFieldKindMultiSelect:
		s := schema.Select
		if s == nil {
			s = &model.CustomFieldSelectSchema{}
		}
		kindDTO := api.CustomFieldSelectSchemaKindSingleSelect
		if kind == model.CustomFieldKindMultiSelect {
			kindDTO = api.CustomFieldSelectSchemaKindMultiSelect
		}
		options := make([]api.CustomFieldOption, len(s.Options))
		for i, opt := range s.Options {
			options[i] = customFieldOptionToAPI(opt)
		}
		if err := dto.FromCustomFieldSelectSchema(api.CustomFieldSelectSchema{
			Kind:    kindDTO,
			Options: options,
		}); err != nil {
			return api.CustomFieldSchema{}, err
		}
	case model.CustomFieldKindUserReference:
		s := schema.UserReference
		if s == nil {
			s = &model.CustomFieldUserReferenceSchema{}
		}
		if err := dto.FromCustomFieldUserReferenceSchema(api.CustomFieldUserReferenceSchema{
			Kind:     api.CustomFieldUserReferenceSchemaKindUserReference,
			Multiple: &s.Multiple,
		}); err != nil {
			return api.CustomFieldSchema{}, err
		}
	case model.CustomFieldKindResourceReference:
		s := schema.ResourceReference
		if s == nil {
			s = &model.CustomFieldResourceReferenceSchema{}
		}
		types := make([]api.ResourceType, len(s.AllowedTypes))
		for i, rt := range s.AllowedTypes {
			types[i] = api.ResourceType(rt.String())
		}
		if err := dto.FromCustomFieldResourceReferenceSchema(api.CustomFieldResourceReferenceSchema{
			Kind:         api.CustomFieldResourceReferenceSchemaKindResourceReference,
			AllowedTypes: types,
			Multiple:     &s.Multiple,
		}); err != nil {
			return api.CustomFieldSchema{}, err
		}
	default:
		return api.CustomFieldSchema{}, model.ErrInvalidCustomFieldKind
	}
	return dto, nil
}

func customFieldOptionToAPI(opt model.CustomFieldOption) api.CustomFieldOption {
	out := api.CustomFieldOption{
		Key:      opt.Key,
		Label:    opt.Label,
		Disabled: opt.Disabled,
		Order:    &opt.Order,
	}
	if opt.Color != "" {
		out.Color = &opt.Color
	}
	return out
}

func customFieldDefinitionToDTO(def *model.CustomFieldDefinition) (api.CustomFieldDefinition, error) {
	schema, err := customFieldSchemaToAPI(def.Kind, def.Schema)
	if err != nil {
		return api.CustomFieldDefinition{}, err
	}
	out := api.CustomFieldDefinition{
		Id:            def.ID.String(),
		Key:           def.Key,
		Name:          def.Name,
		Kind:          api.CustomFieldKind(def.Kind.String()),
		ScopeId:       def.Scope.String(),
		ScopeType:     api.ResourceType(def.Scope.Type.String()),
		TargetType:    api.ResourceType(def.TargetType.String()),
		Required:      def.Required,
		Archived:      def.Archived,
		IndexExact:    def.IndexExact,
		IndexRange:    def.IndexRange,
		IndexFulltext: def.IndexFullText,
		Order:         def.SortOrder,
		OwnerUserId:   def.OwnerUserID.String(),
		Schema:        schema,
		CreatedAt:     def.CreatedAt,
		UpdatedAt:     def.UpdatedAt,
	}
	if def.Description != "" {
		out.Description = &def.Description
	}
	if def.RegistrarClientID != "" {
		out.RegistrarClientId = &def.RegistrarClientID
	}
	return out, nil
}

func customFieldValueFromAPI(value api.CustomFieldValue) (model.CustomFieldTypedValue, error) {
	raw, err := value.ValueByDiscriminator()
	if err != nil {
		return model.CustomFieldTypedValue{}, errors.Join(model.ErrInvalidCustomFieldValue, err)
	}

	switch v := raw.(type) {
	case api.CustomFieldTextValue:
		text := v.Text
		return model.CustomFieldTypedValue{Kind: model.CustomFieldKindText, Text: &text}, nil
	case api.CustomFieldIntegerValue:
		n := v.Integer
		return model.CustomFieldTypedValue{Kind: model.CustomFieldKindInteger, Integer: &n}, nil
	case api.CustomFieldDecimalValue:
		d := v.Decimal
		return model.CustomFieldTypedValue{Kind: model.CustomFieldKindDecimal, Decimal: &d}, nil
	case api.CustomFieldBooleanValue:
		b := v.Boolean
		return model.CustomFieldTypedValue{Kind: model.CustomFieldKindBoolean, Boolean: &b}, nil
	case api.CustomFieldDateValue:
		t := v.Date.Time
		return model.CustomFieldTypedValue{Kind: model.CustomFieldKindDate, Date: &t}, nil
	case api.CustomFieldDateTimeValue:
		t := v.Datetime
		return model.CustomFieldTypedValue{Kind: model.CustomFieldKindDatetime, DateTime: &t}, nil
	case api.CustomFieldURLValue:
		u := v.Url
		return model.CustomFieldTypedValue{Kind: model.CustomFieldKindURL, URL: &u}, nil
	case api.CustomFieldSingleSelectValue:
		key := v.OptionKey
		return model.CustomFieldTypedValue{Kind: model.CustomFieldKindSingleSelect, OptionKey: &key}, nil
	case api.CustomFieldMultiSelectValue:
		return model.CustomFieldTypedValue{Kind: model.CustomFieldKindMultiSelect, OptionKeys: v.OptionKeys}, nil
	case api.CustomFieldUserReferenceValue:
		out := model.CustomFieldTypedValue{Kind: model.CustomFieldKindUserReference}
		if v.UserId != nil && *v.UserId != "" {
			id, err := parseUserRefID(*v.UserId)
			if err != nil {
				return model.CustomFieldTypedValue{}, err
			}
			out.UserID = &id
		}
		if v.UserIds != nil {
			ids := make([]model.ID, len(*v.UserIds))
			for i, raw := range *v.UserIds {
				id, err := parseUserRefID(raw)
				if err != nil {
					return model.CustomFieldTypedValue{}, err
				}
				ids[i] = id
			}
			out.UserIDs = ids
		}
		return out, nil
	case api.CustomFieldResourceReferenceValue:
		out := model.CustomFieldTypedValue{Kind: model.CustomFieldKindResourceReference}
		if v.ResourceId != nil && *v.ResourceId != "" {
			id, err := parseResourceRefID(*v.ResourceId)
			if err != nil {
				return model.CustomFieldTypedValue{}, err
			}
			out.ResourceID = &id
		}
		if v.ResourceIds != nil {
			ids := make([]model.ID, len(*v.ResourceIds))
			for i, raw := range *v.ResourceIds {
				id, err := parseResourceRefID(raw)
				if err != nil {
					return model.CustomFieldTypedValue{}, err
				}
				ids[i] = id
			}
			out.ResourceIDs = ids
		}
		return out, nil
	default:
		return model.CustomFieldTypedValue{}, model.ErrInvalidCustomFieldValue
	}
}

func customFieldValueToAPI(value model.CustomFieldTypedValue) (api.CustomFieldValue, error) {
	var dto api.CustomFieldValue
	switch value.Kind {
	case model.CustomFieldKindText:
		if value.Text == nil {
			return api.CustomFieldValue{}, model.ErrInvalidCustomFieldValue
		}
		if err := dto.FromCustomFieldTextValue(api.CustomFieldTextValue{
			Kind: api.CustomFieldTextValueKindText,
			Text: *value.Text,
		}); err != nil {
			return api.CustomFieldValue{}, err
		}
	case model.CustomFieldKindInteger:
		if value.Integer == nil {
			return api.CustomFieldValue{}, model.ErrInvalidCustomFieldValue
		}
		if err := dto.FromCustomFieldIntegerValue(api.CustomFieldIntegerValue{
			Kind:    api.CustomFieldIntegerValueKindInteger,
			Integer: *value.Integer,
		}); err != nil {
			return api.CustomFieldValue{}, err
		}
	case model.CustomFieldKindDecimal:
		if value.Decimal == nil {
			return api.CustomFieldValue{}, model.ErrInvalidCustomFieldValue
		}
		if err := dto.FromCustomFieldDecimalValue(api.CustomFieldDecimalValue{
			Kind:    api.CustomFieldDecimalValueKindDecimal,
			Decimal: *value.Decimal,
		}); err != nil {
			return api.CustomFieldValue{}, err
		}
	case model.CustomFieldKindBoolean:
		if value.Boolean == nil {
			return api.CustomFieldValue{}, model.ErrInvalidCustomFieldValue
		}
		if err := dto.FromCustomFieldBooleanValue(api.CustomFieldBooleanValue{
			Kind:    api.CustomFieldBooleanValueKindBoolean,
			Boolean: *value.Boolean,
		}); err != nil {
			return api.CustomFieldValue{}, err
		}
	case model.CustomFieldKindDate:
		if value.Date == nil {
			return api.CustomFieldValue{}, model.ErrInvalidCustomFieldValue
		}
		if err := dto.FromCustomFieldDateValue(api.CustomFieldDateValue{
			Kind: api.CustomFieldDateValueKindDate,
			Date: oapiTypes.Date{Time: *value.Date},
		}); err != nil {
			return api.CustomFieldValue{}, err
		}
	case model.CustomFieldKindDatetime:
		if value.DateTime == nil {
			return api.CustomFieldValue{}, model.ErrInvalidCustomFieldValue
		}
		if err := dto.FromCustomFieldDateTimeValue(api.CustomFieldDateTimeValue{
			Kind:     api.CustomFieldDateTimeValueKindDatetime,
			Datetime: *value.DateTime,
		}); err != nil {
			return api.CustomFieldValue{}, err
		}
	case model.CustomFieldKindURL:
		if value.URL == nil {
			return api.CustomFieldValue{}, model.ErrInvalidCustomFieldValue
		}
		if err := dto.FromCustomFieldURLValue(api.CustomFieldURLValue{
			Kind: api.CustomFieldURLValueKindUrl,
			Url:  *value.URL,
		}); err != nil {
			return api.CustomFieldValue{}, err
		}
	case model.CustomFieldKindSingleSelect:
		if value.OptionKey == nil {
			return api.CustomFieldValue{}, model.ErrInvalidCustomFieldValue
		}
		if err := dto.FromCustomFieldSingleSelectValue(api.CustomFieldSingleSelectValue{
			Kind:      api.CustomFieldSingleSelectValueKindSingleSelect,
			OptionKey: *value.OptionKey,
		}); err != nil {
			return api.CustomFieldValue{}, err
		}
	case model.CustomFieldKindMultiSelect:
		if err := dto.FromCustomFieldMultiSelectValue(api.CustomFieldMultiSelectValue{
			Kind:       api.CustomFieldMultiSelectValueKindMultiSelect,
			OptionKeys: value.OptionKeys,
		}); err != nil {
			return api.CustomFieldValue{}, err
		}
	case model.CustomFieldKindUserReference:
		out := api.CustomFieldUserReferenceValue{Kind: api.CustomFieldUserReferenceValueKindUserReference}
		if value.UserID != nil {
			id := value.UserID.String()
			out.UserId = &id
		}
		if len(value.UserIDs) > 0 {
			ids := make([]string, len(value.UserIDs))
			for i, id := range value.UserIDs {
				ids[i] = id.String()
			}
			out.UserIds = &ids
		}
		if err := dto.FromCustomFieldUserReferenceValue(out); err != nil {
			return api.CustomFieldValue{}, err
		}
	case model.CustomFieldKindResourceReference:
		out := api.CustomFieldResourceReferenceValue{Kind: api.CustomFieldResourceReferenceValueKindResourceReference}
		if value.ResourceID != nil {
			id := value.ResourceID.Composite()
			out.ResourceId = &id
		}
		if len(value.ResourceIDs) > 0 {
			ids := make([]string, len(value.ResourceIDs))
			for i, id := range value.ResourceIDs {
				ids[i] = id.Composite()
			}
			out.ResourceIds = &ids
		}
		if err := dto.FromCustomFieldResourceReferenceValue(out); err != nil {
			return api.CustomFieldValue{}, err
		}
	default:
		return api.CustomFieldValue{}, model.ErrInvalidCustomFieldValue
	}
	return dto, nil
}

func customFieldEntryToDTO(entry service.CustomFieldEntry) (api.CustomFieldEntry, error) {
	def, err := customFieldDefinitionToDTO(entry.Definition)
	if err != nil {
		return api.CustomFieldEntry{}, err
	}
	out := api.CustomFieldEntry{Definition: def}
	if entry.Value != nil {
		value, err := customFieldValueToAPI(*entry.Value)
		if err != nil {
			return api.CustomFieldEntry{}, err
		}
		out.Value = &value
	}
	return out, nil
}

func customFieldWritesFromAPI(writes []api.CustomFieldWrite) ([]service.CustomFieldWrite, error) {
	out := make([]service.CustomFieldWrite, len(writes))
	for i, write := range writes {
		definitionID, err := definitionIDFromAPI(write.DefinitionId)
		if err != nil {
			return nil, err
		}
		value, err := customFieldValueFromAPI(write.Value)
		if err != nil {
			return nil, err
		}
		out[i] = service.CustomFieldWrite{DefinitionID: definitionID, Value: value}
	}
	return out, nil
}

func createCustomFieldOptsFromAPI(body *api.V1CustomFieldsCreateJSONRequestBody) (service.CreateCustomFieldOpts, error) {
	if body == nil {
		return service.CreateCustomFieldOpts{}, errors.New("request body is required")
	}
	kind, err := customFieldKindFromAPI(body.Kind)
	if err != nil {
		return service.CreateCustomFieldOpts{}, err
	}
	target, err := resourceTypeFromAPI(body.TargetType)
	if err != nil {
		return service.CreateCustomFieldOpts{}, err
	}
	schema, err := customFieldSchemaFromAPI(body.Schema)
	if err != nil {
		return service.CreateCustomFieldOpts{}, err
	}
	opts := service.CreateCustomFieldOpts{
		Key:           body.Key,
		Name:          body.Name,
		Kind:          kind,
		TargetType:    target,
		Required:      derefBool(body.Required),
		IndexExact:    derefBool(body.IndexExact),
		IndexRange:    derefBool(body.IndexRange),
		IndexFullText: derefBool(body.IndexFulltext),
		Schema:        schema,
	}
	if body.Order != nil {
		opts.SortOrder = optional.Some(*body.Order)
	}
	if body.Description != nil {
		opts.Description = *body.Description
	}
	scopeType, err := resourceTypeFromAPI(body.ScopeType)
	if err != nil {
		return service.CreateCustomFieldOpts{}, err
	}
	scope, err := model.NewIDFromString(body.ScopeId, scopeType.String())
	if err != nil {
		return service.CreateCustomFieldOpts{}, err
	}
	opts.Scope = scope
	return opts, nil
}

func updateCustomFieldOptsFromAPI(body *api.V1CustomFieldUpdateJSONRequestBody) (service.UpdateCustomFieldOpts, error) {
	if body == nil {
		return service.UpdateCustomFieldOpts{}, errors.New("request body is required")
	}
	opts := service.UpdateCustomFieldOpts{
		Name:          body.Name,
		Description:   body.Description,
		Required:      body.Required,
		Archived:      body.Archived,
		IndexExact:    body.IndexExact,
		IndexRange:    body.IndexRange,
		IndexFullText: body.IndexFulltext,
		SortOrder:     body.Order,
	}
	if body.Schema != nil {
		schema, err := customFieldSchemaFromAPI(*body.Schema)
		if err != nil {
			return service.UpdateCustomFieldOpts{}, err
		}
		opts.Schema = optional.Some(schema)
	}
	return opts, nil
}

func customFieldSearchQueryFromAPI(body *api.V1CustomFieldsSearchJSONRequestBody) (service.CustomFieldSearchQuery, error) {
	if body == nil {
		return service.CustomFieldSearchQuery{}, errors.New("request body is required")
	}
	definitionID, err := definitionIDFromAPI(body.DefinitionId)
	if err != nil {
		return service.CustomFieldSearchQuery{}, err
	}
	pred := repository.CustomFieldPredicate{
		Op:        repository.CustomFieldPredicateOp(body.Op),
		Text:      body.Text,
		Integer:   body.Integer,
		Decimal:   body.Decimal,
		Boolean:   body.Boolean,
		URL:       body.Url,
		OptionKey: body.OptionKey,
	}
	if body.Date != nil {
		t := body.Date.Time
		pred.Date = &t
	}
	if body.Datetime != nil {
		pred.DateTime = body.Datetime
	}
	if body.UserId != nil && *body.UserId != "" {
		id, err := parseUserRefID(*body.UserId)
		if err != nil {
			return service.CustomFieldSearchQuery{}, err
		}
		pred.UserID = &id
	}
	if body.ResourceId != nil && *body.ResourceId != "" {
		id, err := parseResourceRefID(*body.ResourceId)
		if err != nil {
			return service.CustomFieldSearchQuery{}, err
		}
		pred.ResourceID = &id
	}
	limit := 100
	if body.Limit != nil {
		limit = *body.Limit
	}
	return service.CustomFieldSearchQuery{
		DefinitionID: definitionID,
		Predicate:    pred,
		Limit:        limit,
	}, nil
}
