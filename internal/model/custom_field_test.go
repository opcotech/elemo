package model

import (
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testScope() ID {
	return ID{Inner: xid.NilID(), Type: ResourceTypeProject}
}

func testOwner() ID {
	return ID{Inner: xid.NilID(), Type: ResourceTypeUser}
}

func TestNewCustomFieldDefinition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		fname   string
		kind    CustomFieldKind
		schema  CustomFieldSchema
		wantErr error
	}{
		{
			name:  "text field",
			key:   "severity_note",
			fname: "Severity note",
			kind:  CustomFieldKindText,
		},
		{
			name:  "integer field",
			key:   "story_points",
			fname: "Story points",
			kind:  CustomFieldKindInteger,
		},
		{
			name:  "select with options",
			key:   "env",
			fname: "Environment",
			kind:  CustomFieldKindSingleSelect,
			schema: CustomFieldSchema{Select: &CustomFieldSelectSchema{
				Options: []CustomFieldOption{
					{Key: "prod", Label: "Production"},
					{Key: "staging", Label: "Staging"},
				},
			}},
		},
		{
			name:    "invalid key",
			key:     "1bad",
			fname:   "Bad key",
			kind:    CustomFieldKindBoolean,
			wantErr: ErrInvalidCustomFieldDetails,
		},
		{
			name:    "short name",
			key:     "ok_key",
			fname:   "ab",
			kind:    CustomFieldKindBoolean,
			wantErr: ErrInvalidCustomFieldDetails,
		},
		{
			name:    "select without options",
			key:     "env",
			fname:   "Environment",
			kind:    CustomFieldKindSingleSelect,
			wantErr: ErrInvalidCustomFieldDetails,
		},
		{
			name:  "resource reference to team",
			key:   "sprint_team",
			fname: "Sprint team",
			kind:  CustomFieldKindResourceReference,
			schema: CustomFieldSchema{ResourceReference: &CustomFieldResourceReferenceSchema{
				AllowedTypes: []ResourceType{ResourceTypeTeam},
			}},
		},
		{
			name:  "resource reference to permission",
			key:   "grant_ref",
			fname: "Grant ref",
			kind:  CustomFieldKindResourceReference,
			schema: CustomFieldSchema{ResourceReference: &CustomFieldResourceReferenceSchema{
				AllowedTypes: []ResourceType{ResourceTypePermission},
			}},
			wantErr: ErrInvalidCustomFieldDetails,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewCustomFieldDefinition(
				tt.key,
				tt.fname,
				tt.kind,
				testScope(),
				testOwner(),
				ResourceTypeIssue,
				tt.schema,
			)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, ResourceTypeCustomFieldDefinition, got.ID.Type)
			assert.Equal(t, tt.key, got.Key)
			assert.Equal(t, tt.kind, got.Kind)
			assert.False(t, got.Archived)
		})
	}
}

func TestCustomFieldDefinition_Validate(t *testing.T) {
	t.Parallel()

	valid := func() *CustomFieldDefinition {
		def, err := NewCustomFieldDefinition(
			"story_points",
			"Story points",
			CustomFieldKindInteger,
			testScope(),
			testOwner(),
			ResourceTypeIssue,
			CustomFieldSchema{},
		)
		require.NoError(t, err)
		return def
	}

	t.Run("range index rejected for text", func(t *testing.T) {
		t.Parallel()
		def, err := NewCustomFieldDefinition(
			"notes",
			"Notes",
			CustomFieldKindText,
			testScope(),
			testOwner(),
			ResourceTypeIssue,
			CustomFieldSchema{},
		)
		require.NoError(t, err)
		def.IndexRange = true
		require.ErrorIs(t, def.Validate(), ErrInvalidCustomFieldDetails)
	})

	t.Run("full-text allowed for text", func(t *testing.T) {
		t.Parallel()
		def, err := NewCustomFieldDefinition(
			"notes",
			"Notes",
			CustomFieldKindText,
			testScope(),
			testOwner(),
			ResourceTypeIssue,
			CustomFieldSchema{},
		)
		require.NoError(t, err)
		def.IndexFullText = true
		require.NoError(t, def.Validate())
	})

	t.Run("invalid scope type", func(t *testing.T) {
		t.Parallel()
		def := valid()
		def.Scope = ID{Inner: xid.NilID(), Type: ResourceTypeIssue}
		require.ErrorIs(t, def.Validate(), ErrInvalidCustomFieldDetails)
	})

	t.Run("invalid target type", func(t *testing.T) {
		t.Parallel()
		def := valid()
		def.TargetType = ResourceTypeLabel
		require.ErrorIs(t, def.Validate(), ErrInvalidCustomFieldDetails)
	})
}

func TestCustomFieldDefinition_ArchiveAndDelete(t *testing.T) {
	t.Parallel()

	def, err := NewCustomFieldDefinition(
		"story_points",
		"Story points",
		CustomFieldKindInteger,
		testScope(),
		testOwner(),
		ResourceTypeIssue,
		CustomFieldSchema{},
	)
	require.NoError(t, err)

	require.NoError(t, def.AssertWritable())
	def.Archived = true
	require.ErrorIs(t, def.AssertWritable(), ErrCustomFieldArchived)
	require.ErrorIs(t, def.CanHardDelete(true), ErrCustomFieldInUse)
	require.NoError(t, def.CanHardDelete(false))
}

func TestCustomFieldKind_IndexCapabilities(t *testing.T) {
	t.Parallel()

	assert.True(t, CustomFieldKindText.AllowsExact())
	assert.True(t, CustomFieldKindText.AllowsFullText())
	assert.False(t, CustomFieldKindText.AllowsRange())
	assert.True(t, CustomFieldKindInteger.AllowsRange())
	assert.False(t, CustomFieldKindBoolean.AllowsRange())
	assert.False(t, CustomFieldKindBoolean.AllowsFullText())
}

func TestCustomFieldTypedValue_Atomics(t *testing.T) {
	t.Parallel()

	text := "hello"
	atoms, err := (CustomFieldTypedValue{Kind: CustomFieldKindText, Text: &text}).Atomics()
	require.NoError(t, err)
	require.Len(t, atoms, 1)
	assert.Equal(t, &text, atoms[0].Text)

	keys := []string{"prod", "staging"}
	atoms, err = (CustomFieldTypedValue{Kind: CustomFieldKindMultiSelect, OptionKeys: keys}).Atomics()
	require.NoError(t, err)
	require.Len(t, atoms, 2)
	assert.Equal(t, 1, atoms[1].Ordinal)
	assert.Equal(t, "staging", *atoms[1].OptionKey)

	roundTrip, err := TypedValueFromAtomics(CustomFieldKindMultiSelect, atoms)
	require.NoError(t, err)
	assert.Equal(t, keys, roundTrip.OptionKeys)
}

func TestValidateAgainst(t *testing.T) {
	t.Parallel()

	intDef, err := NewCustomFieldDefinition(
		"story_points",
		"Story points",
		CustomFieldKindInteger,
		testScope(),
		testOwner(),
		ResourceTypeIssue,
		CustomFieldSchema{Integer: &CustomFieldIntegerSchema{
			Min: ptr[int64](1),
			Max: ptr[int64](8),
		}},
	)
	require.NoError(t, err)

	n := int64(5)
	require.NoError(t, ValidateAgainst(intDef, []CustomFieldAtomicValue{{Integer: &n}}))

	tooHigh := int64(99)
	require.ErrorIs(t, ValidateAgainst(intDef, []CustomFieldAtomicValue{{Integer: &tooHigh}}), ErrInvalidCustomFieldValue)

	intDef.Required = true
	require.ErrorIs(t, ValidateAgainst(intDef, nil), ErrCustomFieldRequired)

	intDef.Archived = true
	require.ErrorIs(t, ValidateAgainst(intDef, []CustomFieldAtomicValue{{Integer: &n}}), ErrCustomFieldArchived)

	selectDef, err := NewCustomFieldDefinition(
		"env",
		"Environment",
		CustomFieldKindSingleSelect,
		testScope(),
		testOwner(),
		ResourceTypeIssue,
		CustomFieldSchema{Select: &CustomFieldSelectSchema{
			Options: []CustomFieldOption{
				{Key: "prod", Label: "Production"},
				{Key: "legacy", Label: "Legacy", Disabled: true},
			},
		}},
	)
	require.NoError(t, err)
	prod := "prod"
	require.NoError(t, ValidateAgainst(selectDef, []CustomFieldAtomicValue{{OptionKey: &prod}}))
	legacy := "legacy"
	require.ErrorIs(t, ValidateAgainst(selectDef, []CustomFieldAtomicValue{{OptionKey: &legacy}}), ErrInvalidCustomFieldValue)

	urlDef, err := NewCustomFieldDefinition(
		"docs_url",
		"Docs URL",
		CustomFieldKindURL,
		testScope(),
		testOwner(),
		ResourceTypeIssue,
		CustomFieldSchema{},
	)
	require.NoError(t, err)
	good := "https://example.com/docs"
	require.NoError(t, ValidateAgainst(urlDef, []CustomFieldAtomicValue{{URL: &good}}))
	bad := "ftp://example.com/docs"
	require.ErrorIs(t, ValidateAgainst(urlDef, []CustomFieldAtomicValue{{URL: &bad}}), ErrInvalidCustomFieldValue)

	decDef, err := NewCustomFieldDefinition(
		"estimate",
		"Estimate",
		CustomFieldKindDecimal,
		testScope(),
		testOwner(),
		ResourceTypeIssue,
		CustomFieldSchema{Decimal: &CustomFieldDecimalSchema{Min: "0.01", Max: "100.00"}},
	)
	require.NoError(t, err)
	dec := "12.50"
	require.NoError(t, ValidateAgainst(decDef, []CustomFieldAtomicValue{{Decimal: &dec}}))
	under := "0.001"
	require.ErrorIs(t, ValidateAgainst(decDef, []CustomFieldAtomicValue{{Decimal: &under}}), ErrInvalidCustomFieldValue)
}

func TestCustomFieldDefinition_IdentityEquals(t *testing.T) {
	t.Parallel()

	a, err := NewCustomFieldDefinition(
		"story_points",
		"Story points",
		CustomFieldKindInteger,
		testScope(),
		testOwner(),
		ResourceTypeIssue,
		CustomFieldSchema{},
	)
	require.NoError(t, err)
	b := *a
	b.Name = "Points"
	assert.True(t, a.IdentityEquals(&b))
	b.Key = "points"
	assert.False(t, a.IdentityEquals(&b))
}

func TestUpdateActionFor(t *testing.T) {
	t.Parallel()

	got, ok := UpdateActionFor(ResourceTypeIssue)
	assert.True(t, ok)
	assert.Equal(t, ActionIssueUpdate, got)

	_, ok = UpdateActionFor(ResourceTypeCustomFieldDefinition)
	assert.False(t, ok)
}

func TestIsCustomFieldScopeAndTarget(t *testing.T) {
	t.Parallel()

	assert.True(t, IsCustomFieldScopeType(ResourceTypeProject))
	assert.False(t, IsCustomFieldScopeType(ResourceTypeIssue))
	assert.True(t, IsCustomFieldTargetType(ResourceTypeIssue))
	assert.False(t, IsCustomFieldTargetType(ResourceTypeLabel))
	assert.True(t, IsCustomFieldReferenceType(ResourceTypeIssue))
	assert.True(t, IsCustomFieldReferenceType(ResourceTypeTeam))
	assert.True(t, IsCustomFieldReferenceType(ResourceTypeLabel))
	assert.False(t, IsCustomFieldReferenceType(ResourceTypePermission))
	assert.False(t, IsCustomFieldReferenceType(ResourceTypeKind))
	assert.False(t, IsCustomFieldReferenceType(0))
}

func TestDecimalExactness(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	_ = now
	rat, err := parseDecimal("0.10")
	require.NoError(t, err)
	other, err := parseDecimal("0.1")
	require.NoError(t, err)
	assert.Equal(t, 0, rat.Cmp(other))
	_, err = parseDecimal("not-a-number")
	require.Error(t, err)
	_, err = parseDecimal("1/2")
	require.Error(t, err)
}

func ptr[T any](v T) *T {
	return &v
}
