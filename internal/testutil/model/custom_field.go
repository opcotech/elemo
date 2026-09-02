package model

import (
	"strings"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
)

func fieldKey(prefix string) string {
	return prefix + strings.ToLower(pkg.GenerateRandomStringAlpha(8))
}

// NewCustomFieldDefinition creates a valid project-scoped Issue text field.
func NewCustomFieldDefinition(scope, owner model.ID) *model.CustomFieldDefinition {
	def, err := model.NewCustomFieldDefinition(
		fieldKey("f_"),
		"Custom field",
		model.CustomFieldKindText,
		scope,
		owner,
		model.ResourceTypeIssue,
		model.CustomFieldSchema{},
	)
	if err != nil {
		panic(err)
	}
	return def
}

// NewIntegerCustomFieldDefinition creates an indexed integer Issue field.
func NewIntegerCustomFieldDefinition(scope, owner model.ID) *model.CustomFieldDefinition {
	def, err := model.NewCustomFieldDefinition(
		fieldKey("n_"),
		"Numeric field",
		model.CustomFieldKindInteger,
		scope,
		owner,
		model.ResourceTypeIssue,
		model.CustomFieldSchema{},
	)
	if err != nil {
		panic(err)
	}
	def.IndexExact = true
	def.IndexRange = true
	return def
}

// NewSelectCustomFieldDefinition creates a single-select Issue field.
func NewSelectCustomFieldDefinition(scope, owner model.ID) *model.CustomFieldDefinition {
	def, err := model.NewCustomFieldDefinition(
		fieldKey("s_"),
		"Select field",
		model.CustomFieldKindSingleSelect,
		scope,
		owner,
		model.ResourceTypeIssue,
		model.CustomFieldSchema{Select: &model.CustomFieldSelectSchema{
			Options: []model.CustomFieldOption{
				{Key: "alpha", Label: "Alpha"},
				{Key: "beta", Label: "Beta"},
			},
		}},
	)
	if err != nil {
		panic(err)
	}
	return def
}

// NewDecimalCustomFieldDefinition creates an indexed decimal Issue field.
func NewDecimalCustomFieldDefinition(scope, owner model.ID) *model.CustomFieldDefinition {
	def, err := model.NewCustomFieldDefinition(
		fieldKey("d_"),
		"Decimal field",
		model.CustomFieldKindDecimal,
		scope,
		owner,
		model.ResourceTypeIssue,
		model.CustomFieldSchema{},
	)
	if err != nil {
		panic(err)
	}
	def.IndexExact = true
	def.IndexRange = true
	return def
}

// NewUserReferenceCustomFieldDefinition creates a single-user Issue field.
func NewUserReferenceCustomFieldDefinition(scope, owner model.ID) *model.CustomFieldDefinition {
	def, err := model.NewCustomFieldDefinition(
		fieldKey("u_"),
		"User field",
		model.CustomFieldKindUserReference,
		scope,
		owner,
		model.ResourceTypeIssue,
		model.CustomFieldSchema{},
	)
	if err != nil {
		panic(err)
	}
	return def
}

// NewResourceReferenceCustomFieldDefinition creates a single-issue Issue field.
func NewResourceReferenceCustomFieldDefinition(scope, owner model.ID) *model.CustomFieldDefinition {
	def, err := model.NewCustomFieldDefinition(
		fieldKey("r_"),
		"Resource field",
		model.CustomFieldKindResourceReference,
		scope,
		owner,
		model.ResourceTypeIssue,
		model.CustomFieldSchema{ResourceReference: &model.CustomFieldResourceReferenceSchema{
			AllowedTypes: []model.ResourceType{model.ResourceTypeIssue},
		}},
	)
	if err != nil {
		panic(err)
	}
	return def
}
