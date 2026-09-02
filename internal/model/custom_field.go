package model

import (
	"errors"
	"math/big"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	CustomFieldKindText              CustomFieldKind = iota + 1 // text
	CustomFieldKindInteger                                      // integer
	CustomFieldKindDecimal                                      // decimal
	CustomFieldKindBoolean                                      // boolean
	CustomFieldKindDate                                         // date
	CustomFieldKindDatetime                                     // datetime
	CustomFieldKindURL                                          // url
	CustomFieldKindSingleSelect                                 // single_select
	CustomFieldKindMultiSelect                                  // multi_select
	CustomFieldKindUserReference                                // user_reference
	CustomFieldKindResourceReference                            // resource_reference
)

// CustomFieldKind is a closed set of custom-field value types.
//
//go:generate go tool enumer -type=CustomFieldKind -text -sql -transform=noop -linecomment -output=custom_field_kind_gen.go
type CustomFieldKind uint8

const (
	customFieldKeyMinLen    = 2
	customFieldKeyMaxLen    = 63
	customFieldNameMinLen   = 3
	customFieldNameMaxLen   = 120
	customFieldDescMaxLen   = 500
	customFieldOptionMinLen = 1
	customFieldTextMaxLen   = 8000
	customFieldDecimalMax   = 76
)

var customFieldKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,62}$`)

// CustomFieldOption is a stable select choice. Keys are never reused.
type CustomFieldOption struct {
	Key      string `json:"key" validate:"required,min=2,max=63"`
	Label    string `json:"label" validate:"required,min=1,max=120"`
	Color    string `json:"color" validate:"omitempty,max=32"`
	Disabled bool   `json:"disabled"`
	Order    int    `json:"order"`
}

// CustomFieldTextSchema constrains text values.
type CustomFieldTextSchema struct {
	MinLength *int   `json:"min_length,omitempty"`
	MaxLength *int   `json:"max_length,omitempty"`
	Pattern   string `json:"pattern,omitempty"`
}

// CustomFieldIntegerSchema constrains integer values.
type CustomFieldIntegerSchema struct {
	Min *int64 `json:"min,omitempty"`
	Max *int64 `json:"max,omitempty"`
}

// CustomFieldDecimalSchema constrains decimal values stored as canonical
// decimal strings (no binary float).
type CustomFieldDecimalSchema struct {
	Min   string `json:"min,omitempty"`
	Max   string `json:"max,omitempty"`
	Scale *int   `json:"scale,omitempty"`
}

// CustomFieldBooleanSchema is an empty marker for boolean fields.
type CustomFieldBooleanSchema struct{}

// CustomFieldDateSchema constrains calendar dates.
type CustomFieldDateSchema struct {
	Min *time.Time `json:"min,omitempty"`
	Max *time.Time `json:"max,omitempty"`
}

// CustomFieldDateTimeSchema constrains timestamps.
type CustomFieldDateTimeSchema struct {
	Min *time.Time `json:"min,omitempty"`
	Max *time.Time `json:"max,omitempty"`
}

// CustomFieldURLSchema constrains URL values.
type CustomFieldURLSchema struct {
	AllowedSchemes []string `json:"allowed_schemes,omitempty"`
}

// CustomFieldSelectSchema holds stable options for single- and multi-select.
type CustomFieldSelectSchema struct {
	Options []CustomFieldOption `json:"options" validate:"required,min=1,dive"`
}

// CustomFieldUserReferenceSchema constrains user references.
type CustomFieldUserReferenceSchema struct {
	Multiple bool `json:"multiple"`
}

// CustomFieldResourceReferenceSchema constrains resource references.
type CustomFieldResourceReferenceSchema struct {
	AllowedTypes []ResourceType `json:"allowed_types" validate:"required,min=1"`
	Multiple     bool           `json:"multiple"`
}

// CustomFieldSchema is a discriminated union: exactly one kind-specific
// member is set, matching CustomFieldDefinition.Kind.
type CustomFieldSchema struct {
	Text              *CustomFieldTextSchema              `json:"text,omitempty"`
	Integer           *CustomFieldIntegerSchema           `json:"integer,omitempty"`
	Decimal           *CustomFieldDecimalSchema           `json:"decimal,omitempty"`
	Boolean           *CustomFieldBooleanSchema           `json:"boolean,omitempty"`
	Date              *CustomFieldDateSchema              `json:"date,omitempty"`
	DateTime          *CustomFieldDateTimeSchema          `json:"datetime,omitempty"`
	URL               *CustomFieldURLSchema               `json:"url,omitempty"`
	Select            *CustomFieldSelectSchema            `json:"select,omitempty"`
	UserReference     *CustomFieldUserReferenceSchema     `json:"user_reference,omitempty"`
	ResourceReference *CustomFieldResourceReferenceSchema `json:"resource_reference,omitempty"`
}

// CustomFieldDefinition is a typed attribute attached to a scope and target
// resource type. Identity (key, kind, scope, owner, target) is immutable.
// Display order is mutable.
type CustomFieldDefinition struct {
	ID                ID                `json:"id" validate:"required"`
	Key               string            `json:"key" validate:"required,min=2,max=63"`
	Name              string            `json:"name" validate:"required,min=3,max=120"`
	Description       string            `json:"description" validate:"omitempty,max=500"`
	Kind              CustomFieldKind   `json:"kind" validate:"required,min=1,max=11"`
	Scope             ID                `json:"scope" validate:"required"`
	TargetType        ResourceType      `json:"target_type" validate:"required"`
	Required          bool              `json:"required"`
	Archived          bool              `json:"archived"`
	IndexExact        bool              `json:"index_exact"`
	IndexRange        bool              `json:"index_range"`
	IndexFullText     bool              `json:"index_fulltext"`
	SortOrder         int               `json:"order"`
	OwnerUserID       ID                `json:"owner_user_id" validate:"required"`
	RegistrarClientID string            `json:"registrar_client_id" validate:"omitempty,max=128"`
	Schema            CustomFieldSchema `json:"schema"`
	CreatedAt         *time.Time        `json:"created_at" validate:"omitempty"`
	UpdatedAt         *time.Time        `json:"updated_at" validate:"omitempty"`
}

// AllowsExact reports whether kind supports exact-match indexes.
func (k CustomFieldKind) AllowsExact() bool {
	return k.IsACustomFieldKind()
}

// AllowsRange reports whether kind supports range indexes.
func (k CustomFieldKind) AllowsRange() bool {
	switch k {
	case CustomFieldKindInteger, CustomFieldKindDecimal, CustomFieldKindDate, CustomFieldKindDatetime:
		return true
	default:
		return false
	}
}

// AllowsFullText reports whether kind supports full-text indexes.
func (k CustomFieldKind) AllowsFullText() bool {
	return k == CustomFieldKindText
}

// IsMultiValued reports whether a kind stores more than one atomic value.
func (k CustomFieldKind) IsMultiValued(schema CustomFieldSchema) bool {
	switch k {
	case CustomFieldKindMultiSelect:
		return true
	case CustomFieldKindUserReference:
		return schema.UserReference != nil && schema.UserReference.Multiple
	case CustomFieldKindResourceReference:
		return schema.ResourceReference != nil && schema.ResourceReference.Multiple
	default:
		return false
	}
}

func (s CustomFieldSchema) populatedCount() int {
	n := 0
	if s.Text != nil {
		n++
	}
	if s.Integer != nil {
		n++
	}
	if s.Decimal != nil {
		n++
	}
	if s.Boolean != nil {
		n++
	}
	if s.Date != nil {
		n++
	}
	if s.DateTime != nil {
		n++
	}
	if s.URL != nil {
		n++
	}
	if s.Select != nil {
		n++
	}
	if s.UserReference != nil {
		n++
	}
	if s.ResourceReference != nil {
		n++
	}
	return n
}

func (s CustomFieldSchema) matches(kind CustomFieldKind) bool {
	switch kind {
	case CustomFieldKindText:
		return s.Text != nil
	case CustomFieldKindInteger:
		return s.Integer != nil
	case CustomFieldKindDecimal:
		return s.Decimal != nil
	case CustomFieldKindBoolean:
		return s.Boolean != nil
	case CustomFieldKindDate:
		return s.Date != nil
	case CustomFieldKindDatetime:
		return s.DateTime != nil
	case CustomFieldKindURL:
		return s.URL != nil
	case CustomFieldKindSingleSelect, CustomFieldKindMultiSelect:
		return s.Select != nil
	case CustomFieldKindUserReference:
		return s.UserReference != nil
	case CustomFieldKindResourceReference:
		return s.ResourceReference != nil
	default:
		return false
	}
}

func defaultSchema(kind CustomFieldKind) CustomFieldSchema {
	switch kind {
	case CustomFieldKindText:
		return CustomFieldSchema{Text: &CustomFieldTextSchema{}}
	case CustomFieldKindInteger:
		return CustomFieldSchema{Integer: &CustomFieldIntegerSchema{}}
	case CustomFieldKindDecimal:
		return CustomFieldSchema{Decimal: &CustomFieldDecimalSchema{}}
	case CustomFieldKindBoolean:
		return CustomFieldSchema{Boolean: &CustomFieldBooleanSchema{}}
	case CustomFieldKindDate:
		return CustomFieldSchema{Date: &CustomFieldDateSchema{}}
	case CustomFieldKindDatetime:
		return CustomFieldSchema{DateTime: &CustomFieldDateTimeSchema{}}
	case CustomFieldKindURL:
		return CustomFieldSchema{URL: &CustomFieldURLSchema{
			AllowedSchemes: []string{"https", "http"},
		}}
	case CustomFieldKindSingleSelect, CustomFieldKindMultiSelect:
		return CustomFieldSchema{Select: &CustomFieldSelectSchema{Options: []CustomFieldOption{}}}
	case CustomFieldKindUserReference:
		return CustomFieldSchema{UserReference: &CustomFieldUserReferenceSchema{}}
	case CustomFieldKindResourceReference:
		return CustomFieldSchema{ResourceReference: &CustomFieldResourceReferenceSchema{}}
	default:
		return CustomFieldSchema{}
	}
}

func (d *CustomFieldDefinition) Validate() error {
	if err := validate.Struct(d); err != nil {
		return errors.Join(ErrInvalidCustomFieldDetails, err)
	}
	if err := d.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidCustomFieldDetails, err)
	}
	if d.ID.Type != ResourceTypeCustomFieldDefinition {
		return errors.Join(ErrInvalidCustomFieldDetails, ErrInvalidID)
	}
	if !customFieldKeyPattern.MatchString(d.Key) {
		return errors.Join(ErrInvalidCustomFieldDetails, errors.New("invalid custom field key"))
	}
	if !d.Kind.IsACustomFieldKind() {
		return errors.Join(ErrInvalidCustomFieldDetails, ErrInvalidCustomFieldKind)
	}
	if err := d.Scope.Validate(); err != nil {
		return errors.Join(ErrInvalidCustomFieldDetails, err)
	}
	if !IsCustomFieldScopeType(d.Scope.Type) {
		return errors.Join(ErrInvalidCustomFieldDetails, ErrInvalidResourceType)
	}
	if !IsCustomFieldTargetType(d.TargetType) {
		return errors.Join(ErrInvalidCustomFieldDetails, ErrInvalidResourceType)
	}
	if err := d.OwnerUserID.Validate(); err != nil {
		return errors.Join(ErrInvalidCustomFieldDetails, err)
	}
	if d.OwnerUserID.Type != ResourceTypeUser {
		return errors.Join(ErrInvalidCustomFieldDetails, ErrInvalidID)
	}
	if d.IndexExact && !d.Kind.AllowsExact() {
		return errors.Join(ErrInvalidCustomFieldDetails, errors.New("exact index is not allowed"))
	}
	if d.IndexRange && !d.Kind.AllowsRange() {
		return errors.Join(ErrInvalidCustomFieldDetails, errors.New("range index is not allowed"))
	}
	if d.IndexFullText && !d.Kind.AllowsFullText() {
		return errors.Join(ErrInvalidCustomFieldDetails, errors.New("full-text index is not allowed"))
	}
	if err := d.Schema.validate(d.Kind); err != nil {
		return errors.Join(ErrInvalidCustomFieldDetails, err)
	}
	return nil
}

func (s CustomFieldSchema) validate(kind CustomFieldKind) error {
	if s.populatedCount() != 1 || !s.matches(kind) {
		return errors.New("schema does not match field kind")
	}
	switch kind {
	case CustomFieldKindText:
		return s.Text.validate()
	case CustomFieldKindInteger:
		return s.Integer.validate()
	case CustomFieldKindDecimal:
		return s.Decimal.validate()
	case CustomFieldKindDate:
		return s.Date.validate()
	case CustomFieldKindDatetime:
		return s.DateTime.validate()
	case CustomFieldKindURL:
		return s.URL.validate()
	case CustomFieldKindSingleSelect, CustomFieldKindMultiSelect:
		return s.Select.validate()
	case CustomFieldKindResourceReference:
		return s.ResourceReference.validate()
	default:
		return nil
	}
}

func (s *CustomFieldTextSchema) validate() error {
	if s.MinLength != nil && *s.MinLength < 0 {
		return errors.New("text min_length is negative")
	}
	if s.MaxLength != nil && (*s.MaxLength < 0 || *s.MaxLength > customFieldTextMaxLen) {
		return errors.New("text max_length is invalid")
	}
	if s.MinLength != nil && s.MaxLength != nil && *s.MinLength > *s.MaxLength {
		return errors.New("text min_length exceeds max_length")
	}
	if s.Pattern != "" {
		if _, err := regexp.Compile(s.Pattern); err != nil {
			return errors.New("text pattern is invalid")
		}
	}
	return nil
}

func (s *CustomFieldIntegerSchema) validate() error {
	if s.Min != nil && s.Max != nil && *s.Min > *s.Max {
		return errors.New("integer min exceeds max")
	}
	return nil
}

func (s *CustomFieldDecimalSchema) validate() error {
	if s.Scale != nil && (*s.Scale < 0 || *s.Scale > 38) {
		return errors.New("decimal scale is invalid")
	}
	if s.Min != "" {
		if _, err := parseDecimal(s.Min); err != nil {
			return err
		}
	}
	if s.Max != "" {
		if _, err := parseDecimal(s.Max); err != nil {
			return err
		}
	}
	if s.Min != "" && s.Max != "" {
		minRat, _ := parseDecimal(s.Min)
		maxRat, _ := parseDecimal(s.Max)
		if minRat.Cmp(maxRat) > 0 {
			return errors.New("decimal min exceeds max")
		}
	}
	return nil
}

func (s *CustomFieldDateSchema) validate() error {
	if s.Min != nil && s.Max != nil && s.Min.After(*s.Max) {
		return errors.New("date min exceeds max")
	}
	return nil
}

func (s *CustomFieldDateTimeSchema) validate() error {
	if s.Min != nil && s.Max != nil && s.Min.After(*s.Max) {
		return errors.New("datetime min exceeds max")
	}
	return nil
}

func (s *CustomFieldURLSchema) validate() error {
	if len(s.AllowedSchemes) == 0 {
		return errors.New("url allowed_schemes is empty")
	}
	for _, scheme := range s.AllowedSchemes {
		if scheme == "" || strings.Contains(scheme, ":") {
			return errors.New("url allowed scheme is invalid")
		}
	}
	return nil
}

func (s *CustomFieldSelectSchema) validate() error {
	if len(s.Options) < 1 {
		return errors.New("select options are required")
	}
	seen := make(map[string]struct{}, len(s.Options))
	enabled := 0
	for _, opt := range s.Options {
		if !customFieldKeyPattern.MatchString(opt.Key) {
			return errors.New("select option key is invalid")
		}
		if strings.TrimSpace(opt.Label) == "" || utf8.RuneCountInString(opt.Label) > customFieldNameMaxLen {
			return errors.New("select option label is invalid")
		}
		if _, exists := seen[opt.Key]; exists {
			return errors.New("select option key is duplicated")
		}
		seen[opt.Key] = struct{}{}
		if !opt.Disabled {
			enabled++
		}
	}
	if enabled < 1 {
		return errors.New("select requires an enabled option")
	}
	return nil
}

func (s *CustomFieldResourceReferenceSchema) validate() error {
	if len(s.AllowedTypes) < 1 {
		return errors.New("resource reference allowed_types is empty")
	}
	seen := make(map[ResourceType]struct{}, len(s.AllowedTypes))
	for _, rt := range s.AllowedTypes {
		if !IsCustomFieldReferenceType(rt) {
			return errors.Join(ErrInvalidResourceType)
		}
		if _, exists := seen[rt]; exists {
			return errors.New("resource reference allowed_types is duplicated")
		}
		seen[rt] = struct{}{}
	}
	return nil
}

// AssertWritable returns an error when values must not be written.
func (d *CustomFieldDefinition) AssertWritable() error {
	if d.Archived {
		return ErrCustomFieldArchived
	}
	return nil
}

// CanHardDelete reports whether the definition may be deleted. The caller
// supplies whether any values exist.
func (d *CustomFieldDefinition) CanHardDelete(hasValues bool) error {
	if hasValues {
		return ErrCustomFieldInUse
	}
	return nil
}

// IdentityEquals reports whether frozen identity fields match other.
func (d *CustomFieldDefinition) IdentityEquals(other *CustomFieldDefinition) bool {
	if other == nil {
		return false
	}
	return d.Key == other.Key &&
		d.Kind == other.Kind &&
		d.Scope == other.Scope &&
		d.TargetType == other.TargetType &&
		d.OwnerUserID == other.OwnerUserID &&
		d.RegistrarClientID == other.RegistrarClientID
}

// OptionByKey returns the option with the given key.
func (d *CustomFieldDefinition) OptionByKey(key string) (CustomFieldOption, bool) {
	if d.Schema.Select == nil {
		return CustomFieldOption{}, false
	}
	for _, opt := range d.Schema.Select.Options {
		if opt.Key == key {
			return opt, true
		}
	}
	return CustomFieldOption{}, false
}

// NewCustomFieldDefinition creates a definition with a nil ID. The repository
// assigns a real ID before persist. A zero schema is replaced with the kind's
// default constraints.
func NewCustomFieldDefinition(
	key, name string,
	kind CustomFieldKind,
	scope, owner ID,
	target ResourceType,
	schema CustomFieldSchema,
) (*CustomFieldDefinition, error) {
	if schema.populatedCount() == 0 {
		schema = defaultSchema(kind)
	}
	def := &CustomFieldDefinition{
		ID:          MustNewNilID(ResourceTypeCustomFieldDefinition),
		Key:         key,
		Name:        name,
		Kind:        kind,
		Scope:       scope,
		TargetType:  target,
		OwnerUserID: owner,
		Schema:      schema,
	}
	if err := def.Validate(); err != nil {
		return nil, err
	}
	return def, nil
}

func parseDecimal(value string) (*big.Rat, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || len(trimmed) > customFieldDecimalMax {
		return nil, errors.New("decimal value is invalid")
	}
	if strings.ContainsAny(trimmed, "/eEiI") {
		return nil, errors.New("decimal value is invalid")
	}
	rat := new(big.Rat)
	if _, ok := rat.SetString(trimmed); !ok {
		return nil, errors.New("decimal value is invalid")
	}
	return rat, nil
}

func decimalScale(value string) int {
	trimmed := strings.TrimSpace(value)
	dot := strings.IndexByte(trimmed, '.')
	if dot < 0 {
		return 0
	}
	return len(trimmed) - dot - 1
}

func parseURLValue(raw string, allowed []string) error {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return errors.New("url value is invalid")
	}
	scheme := strings.ToLower(parsed.Scheme)
	for _, allowedScheme := range allowed {
		if scheme == strings.ToLower(allowedScheme) {
			return nil
		}
	}
	return errors.New("url scheme is not allowed")
}
