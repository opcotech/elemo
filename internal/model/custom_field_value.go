package model

import (
	"errors"
	"regexp"
	"time"
	"unicode/utf8"
)

// CustomFieldAtomicValue is one typed representation of a custom-field
// value. Exactly one payload field is set. Multi-valued fields use one row
// per element with Ordinal.
type CustomFieldAtomicValue struct {
	Ordinal    int
	Text       *string
	Integer    *int64
	Decimal    *string
	Boolean    *bool
	Date       *time.Time
	DateTime   *time.Time
	URL        *string
	OptionKey  *string
	UserID     *ID
	ResourceID *ID
}

// CustomFieldTypedValue is the API-facing value for a definition. Multi-valued
// kinds use the slice fields; scalars use the matching pointer.
type CustomFieldTypedValue struct {
	Kind        CustomFieldKind
	Text        *string
	Integer     *int64
	Decimal     *string
	Boolean     *bool
	Date        *time.Time
	DateTime    *time.Time
	URL         *string
	OptionKey   *string
	OptionKeys  []string
	UserID      *ID
	UserIDs     []ID
	ResourceID  *ID
	ResourceIDs []ID
}

func (v CustomFieldAtomicValue) populatedCount() int {
	n := 0
	if v.Text != nil {
		n++
	}
	if v.Integer != nil {
		n++
	}
	if v.Decimal != nil {
		n++
	}
	if v.Boolean != nil {
		n++
	}
	if v.Date != nil {
		n++
	}
	if v.DateTime != nil {
		n++
	}
	if v.URL != nil {
		n++
	}
	if v.OptionKey != nil {
		n++
	}
	if v.UserID != nil {
		n++
	}
	if v.ResourceID != nil {
		n++
	}
	return n
}

// Atomics converts a typed value into ordered atomic rows.
func (v CustomFieldTypedValue) Atomics() ([]CustomFieldAtomicValue, error) {
	switch v.Kind {
	case CustomFieldKindText:
		if v.Text == nil {
			return nil, ErrInvalidCustomFieldValue
		}
		return []CustomFieldAtomicValue{{Text: v.Text}}, nil
	case CustomFieldKindInteger:
		if v.Integer == nil {
			return nil, ErrInvalidCustomFieldValue
		}
		return []CustomFieldAtomicValue{{Integer: v.Integer}}, nil
	case CustomFieldKindDecimal:
		if v.Decimal == nil {
			return nil, ErrInvalidCustomFieldValue
		}
		return []CustomFieldAtomicValue{{Decimal: v.Decimal}}, nil
	case CustomFieldKindBoolean:
		if v.Boolean == nil {
			return nil, ErrInvalidCustomFieldValue
		}
		return []CustomFieldAtomicValue{{Boolean: v.Boolean}}, nil
	case CustomFieldKindDate:
		if v.Date == nil {
			return nil, ErrInvalidCustomFieldValue
		}
		return []CustomFieldAtomicValue{{Date: v.Date}}, nil
	case CustomFieldKindDatetime:
		if v.DateTime == nil {
			return nil, ErrInvalidCustomFieldValue
		}
		return []CustomFieldAtomicValue{{DateTime: v.DateTime}}, nil
	case CustomFieldKindURL:
		if v.URL == nil {
			return nil, ErrInvalidCustomFieldValue
		}
		return []CustomFieldAtomicValue{{URL: v.URL}}, nil
	case CustomFieldKindSingleSelect:
		if v.OptionKey == nil {
			return nil, ErrInvalidCustomFieldValue
		}
		return []CustomFieldAtomicValue{{OptionKey: v.OptionKey}}, nil
	case CustomFieldKindMultiSelect:
		if len(v.OptionKeys) == 0 {
			return nil, ErrInvalidCustomFieldValue
		}
		out := make([]CustomFieldAtomicValue, len(v.OptionKeys))
		for i := range v.OptionKeys {
			key := v.OptionKeys[i]
			out[i] = CustomFieldAtomicValue{Ordinal: i, OptionKey: &key}
		}
		return out, nil
	case CustomFieldKindUserReference:
		if v.UserID != nil {
			return []CustomFieldAtomicValue{{UserID: v.UserID}}, nil
		}
		if len(v.UserIDs) == 0 {
			return nil, ErrInvalidCustomFieldValue
		}
		out := make([]CustomFieldAtomicValue, len(v.UserIDs))
		for i := range v.UserIDs {
			id := v.UserIDs[i]
			out[i] = CustomFieldAtomicValue{Ordinal: i, UserID: &id}
		}
		return out, nil
	case CustomFieldKindResourceReference:
		if v.ResourceID != nil {
			return []CustomFieldAtomicValue{{ResourceID: v.ResourceID}}, nil
		}
		if len(v.ResourceIDs) == 0 {
			return nil, ErrInvalidCustomFieldValue
		}
		out := make([]CustomFieldAtomicValue, len(v.ResourceIDs))
		for i := range v.ResourceIDs {
			id := v.ResourceIDs[i]
			out[i] = CustomFieldAtomicValue{Ordinal: i, ResourceID: &id}
		}
		return out, nil
	default:
		return nil, ErrInvalidCustomFieldKind
	}
}

// TypedValueFromAtomics rebuilds an API value from atomic rows.
func TypedValueFromAtomics(kind CustomFieldKind, rows []CustomFieldAtomicValue) (CustomFieldTypedValue, error) {
	out := CustomFieldTypedValue{Kind: kind}
	if len(rows) == 0 {
		return out, ErrInvalidCustomFieldValue
	}
	switch kind {
	case CustomFieldKindMultiSelect:
		out.OptionKeys = make([]string, len(rows))
		for i, row := range rows {
			if row.OptionKey == nil {
				return CustomFieldTypedValue{}, ErrInvalidCustomFieldValue
			}
			out.OptionKeys[i] = *row.OptionKey
		}
		return out, nil
	case CustomFieldKindUserReference:
		if len(rows) == 1 {
			out.UserID = rows[0].UserID
			return out, nil
		}
		out.UserIDs = make([]ID, len(rows))
		for i, row := range rows {
			if row.UserID == nil {
				return CustomFieldTypedValue{}, ErrInvalidCustomFieldValue
			}
			out.UserIDs[i] = *row.UserID
		}
		return out, nil
	case CustomFieldKindResourceReference:
		if len(rows) == 1 {
			out.ResourceID = rows[0].ResourceID
			return out, nil
		}
		out.ResourceIDs = make([]ID, len(rows))
		for i, row := range rows {
			if row.ResourceID == nil {
				return CustomFieldTypedValue{}, ErrInvalidCustomFieldValue
			}
			out.ResourceIDs[i] = *row.ResourceID
		}
		return out, nil
	default:
		if len(rows) != 1 {
			return CustomFieldTypedValue{}, ErrInvalidCustomFieldValue
		}
		row := rows[0]
		out.Text = row.Text
		out.Integer = row.Integer
		out.Decimal = row.Decimal
		out.Boolean = row.Boolean
		out.Date = row.Date
		out.DateTime = row.DateTime
		out.URL = row.URL
		out.OptionKey = row.OptionKey
		out.UserID = row.UserID
		out.ResourceID = row.ResourceID
		return out, nil
	}
}

// ValidateAgainst checks atomic rows against a definition's kind and schema.
func ValidateAgainst(def *CustomFieldDefinition, values []CustomFieldAtomicValue) error {
	if def == nil {
		return ErrInvalidCustomFieldDetails
	}
	if err := def.AssertWritable(); err != nil {
		return err
	}
	if def.Required && len(values) == 0 {
		return ErrCustomFieldRequired
	}
	if len(values) == 0 {
		return nil
	}
	multi := def.Kind.IsMultiValued(def.Schema)
	if !multi && len(values) != 1 {
		return errors.Join(ErrInvalidCustomFieldValue, errors.New("single-valued field has multiple atoms"))
	}
	seen := make(map[string]struct{}, len(values))
	for i, value := range values {
		if value.Ordinal != i {
			return errors.Join(ErrInvalidCustomFieldValue, errors.New("value ordinal is not contiguous"))
		}
		if err := validateAtomic(def, value); err != nil {
			return err
		}
		dupKey := atomicDupKey(value)
		if dupKey != "" {
			if _, exists := seen[dupKey]; exists {
				return errors.Join(ErrInvalidCustomFieldValue, errors.New("duplicate multi-value entry"))
			}
			seen[dupKey] = struct{}{}
		}
	}
	return nil
}

func atomicDupKey(v CustomFieldAtomicValue) string {
	switch {
	case v.OptionKey != nil:
		return "opt:" + *v.OptionKey
	case v.UserID != nil:
		return "user:" + v.UserID.Composite()
	case v.ResourceID != nil:
		return "res:" + v.ResourceID.Composite()
	default:
		return ""
	}
}

func validateAtomic(def *CustomFieldDefinition, value CustomFieldAtomicValue) error {
	if value.populatedCount() != 1 {
		return errors.Join(ErrInvalidCustomFieldValue, errors.New("value must have exactly one representation"))
	}
	switch def.Kind {
	case CustomFieldKindText:
		return validateTextAtomic(def.Schema.Text, value)
	case CustomFieldKindInteger:
		return validateIntegerAtomic(def.Schema.Integer, value)
	case CustomFieldKindDecimal:
		return validateDecimalAtomic(def.Schema.Decimal, value)
	case CustomFieldKindBoolean:
		if value.Boolean == nil {
			return ErrInvalidCustomFieldValue
		}
		return nil
	case CustomFieldKindDate:
		return validateDateAtomic(def.Schema.Date, value)
	case CustomFieldKindDatetime:
		return validateDateTimeAtomic(def.Schema.DateTime, value)
	case CustomFieldKindURL:
		return validateURLAtomic(def.Schema.URL, value)
	case CustomFieldKindSingleSelect, CustomFieldKindMultiSelect:
		return validateOptionAtomic(def, value)
	case CustomFieldKindUserReference:
		return validateUserAtomic(value)
	case CustomFieldKindResourceReference:
		return validateResourceAtomic(def.Schema.ResourceReference, value)
	default:
		return ErrInvalidCustomFieldKind
	}
}

func validateTextAtomic(schema *CustomFieldTextSchema, value CustomFieldAtomicValue) error {
	if value.Text == nil {
		return ErrInvalidCustomFieldValue
	}
	n := utf8.RuneCountInString(*value.Text)
	if n > customFieldTextMaxLen {
		return ErrInvalidCustomFieldValue
	}
	if schema != nil {
		if schema.MinLength != nil && n < *schema.MinLength {
			return ErrInvalidCustomFieldValue
		}
		if schema.MaxLength != nil && n > *schema.MaxLength {
			return ErrInvalidCustomFieldValue
		}
		if schema.Pattern != "" {
			matched, err := regexpMatch(schema.Pattern, *value.Text)
			if err != nil || !matched {
				return ErrInvalidCustomFieldValue
			}
		}
	}
	return nil
}

func validateIntegerAtomic(schema *CustomFieldIntegerSchema, value CustomFieldAtomicValue) error {
	if value.Integer == nil {
		return ErrInvalidCustomFieldValue
	}
	if schema != nil {
		if schema.Min != nil && *value.Integer < *schema.Min {
			return ErrInvalidCustomFieldValue
		}
		if schema.Max != nil && *value.Integer > *schema.Max {
			return ErrInvalidCustomFieldValue
		}
	}
	return nil
}

func validateDecimalAtomic(schema *CustomFieldDecimalSchema, value CustomFieldAtomicValue) error {
	if value.Decimal == nil {
		return ErrInvalidCustomFieldValue
	}
	rat, err := parseDecimal(*value.Decimal)
	if err != nil {
		return errors.Join(ErrInvalidCustomFieldValue, err)
	}
	if schema == nil {
		return nil
	}
	if schema.Scale != nil && decimalScale(*value.Decimal) > *schema.Scale {
		return ErrInvalidCustomFieldValue
	}
	if schema.Min != "" {
		minRat, err := parseDecimal(schema.Min)
		if err != nil {
			return errors.Join(ErrInvalidCustomFieldValue, err)
		}
		if rat.Cmp(minRat) < 0 {
			return ErrInvalidCustomFieldValue
		}
	}
	if schema.Max != "" {
		maxRat, err := parseDecimal(schema.Max)
		if err != nil {
			return errors.Join(ErrInvalidCustomFieldValue, err)
		}
		if rat.Cmp(maxRat) > 0 {
			return ErrInvalidCustomFieldValue
		}
	}
	return nil
}

func validateDateAtomic(schema *CustomFieldDateSchema, value CustomFieldAtomicValue) error {
	if value.Date == nil {
		return ErrInvalidCustomFieldValue
	}
	if schema != nil {
		if schema.Min != nil && value.Date.Before(*schema.Min) {
			return ErrInvalidCustomFieldValue
		}
		if schema.Max != nil && value.Date.After(*schema.Max) {
			return ErrInvalidCustomFieldValue
		}
	}
	return nil
}

func validateDateTimeAtomic(schema *CustomFieldDateTimeSchema, value CustomFieldAtomicValue) error {
	if value.DateTime == nil {
		return ErrInvalidCustomFieldValue
	}
	if schema != nil {
		if schema.Min != nil && value.DateTime.Before(*schema.Min) {
			return ErrInvalidCustomFieldValue
		}
		if schema.Max != nil && value.DateTime.After(*schema.Max) {
			return ErrInvalidCustomFieldValue
		}
	}
	return nil
}

func validateURLAtomic(schema *CustomFieldURLSchema, value CustomFieldAtomicValue) error {
	if value.URL == nil {
		return ErrInvalidCustomFieldValue
	}
	allowed := []string{"https", "http"}
	if schema != nil && len(schema.AllowedSchemes) > 0 {
		allowed = schema.AllowedSchemes
	}
	if err := parseURLValue(*value.URL, allowed); err != nil {
		return errors.Join(ErrInvalidCustomFieldValue, err)
	}
	return nil
}

func validateOptionAtomic(def *CustomFieldDefinition, value CustomFieldAtomicValue) error {
	if value.OptionKey == nil {
		return ErrInvalidCustomFieldValue
	}
	opt, ok := def.OptionByKey(*value.OptionKey)
	if !ok {
		return ErrInvalidCustomFieldValue
	}
	if opt.Disabled {
		return ErrInvalidCustomFieldValue
	}
	return nil
}

func validateUserAtomic(value CustomFieldAtomicValue) error {
	if value.UserID == nil {
		return ErrInvalidCustomFieldValue
	}
	if err := value.UserID.Validate(); err != nil {
		return errors.Join(ErrInvalidCustomFieldValue, err)
	}
	if value.UserID.Type != ResourceTypeUser {
		return ErrInvalidCustomFieldValue
	}
	return nil
}

func validateResourceAtomic(schema *CustomFieldResourceReferenceSchema, value CustomFieldAtomicValue) error {
	if value.ResourceID == nil {
		return ErrInvalidCustomFieldValue
	}
	if err := value.ResourceID.Validate(); err != nil {
		return errors.Join(ErrInvalidCustomFieldValue, err)
	}
	if schema == nil {
		return nil
	}
	for _, allowed := range schema.AllowedTypes {
		if value.ResourceID.Type == allowed {
			return nil
		}
	}
	return ErrInvalidCustomFieldValue
}

func regexpMatch(pattern, value string) (bool, error) {
	return regexp.MatchString(pattern, value)
}
