package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomFieldTypedValue_AtomicsRemainingKinds(t *testing.T) {
	t.Parallel()

	t.Run("integer", func(t *testing.T) {
		t.Parallel()
		n := int64(8)
		atoms, err := (CustomFieldTypedValue{Kind: CustomFieldKindInteger, Integer: &n}).Atomics()
		require.NoError(t, err)
		require.Len(t, atoms, 1)
		assert.Equal(t, &n, atoms[0].Integer)

		got, err := TypedValueFromAtomics(CustomFieldKindInteger, atoms)
		require.NoError(t, err)
		require.NotNil(t, got.Integer)
		assert.Equal(t, n, *got.Integer)
	})

	t.Run("decimal", func(t *testing.T) {
		t.Parallel()
		dec := "12.50"
		atoms, err := (CustomFieldTypedValue{Kind: CustomFieldKindDecimal, Decimal: &dec}).Atomics()
		require.NoError(t, err)
		got, err := TypedValueFromAtomics(CustomFieldKindDecimal, atoms)
		require.NoError(t, err)
		require.NotNil(t, got.Decimal)
		assert.Equal(t, dec, *got.Decimal)
	})

	t.Run("boolean", func(t *testing.T) {
		t.Parallel()
		flag := true
		atoms, err := (CustomFieldTypedValue{Kind: CustomFieldKindBoolean, Boolean: &flag}).Atomics()
		require.NoError(t, err)
		got, err := TypedValueFromAtomics(CustomFieldKindBoolean, atoms)
		require.NoError(t, err)
		require.NotNil(t, got.Boolean)
		assert.True(t, *got.Boolean)
	})

	t.Run("date", func(t *testing.T) {
		t.Parallel()
		day := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		atoms, err := (CustomFieldTypedValue{Kind: CustomFieldKindDate, Date: &day}).Atomics()
		require.NoError(t, err)
		got, err := TypedValueFromAtomics(CustomFieldKindDate, atoms)
		require.NoError(t, err)
		require.NotNil(t, got.Date)
		assert.True(t, got.Date.Equal(day))
	})

	t.Run("datetime", func(t *testing.T) {
		t.Parallel()
		ts := time.Date(2024, 3, 15, 13, 0, 0, 0, time.UTC)
		atoms, err := (CustomFieldTypedValue{Kind: CustomFieldKindDatetime, DateTime: &ts}).Atomics()
		require.NoError(t, err)
		got, err := TypedValueFromAtomics(CustomFieldKindDatetime, atoms)
		require.NoError(t, err)
		require.NotNil(t, got.DateTime)
		assert.True(t, got.DateTime.Equal(ts))
	})

	t.Run("url", func(t *testing.T) {
		t.Parallel()
		raw := "https://example.com"
		atoms, err := (CustomFieldTypedValue{Kind: CustomFieldKindURL, URL: &raw}).Atomics()
		require.NoError(t, err)
		got, err := TypedValueFromAtomics(CustomFieldKindURL, atoms)
		require.NoError(t, err)
		require.NotNil(t, got.URL)
		assert.Equal(t, raw, *got.URL)
	})

	t.Run("single select", func(t *testing.T) {
		t.Parallel()
		key := "prod"
		atoms, err := (CustomFieldTypedValue{Kind: CustomFieldKindSingleSelect, OptionKey: &key}).Atomics()
		require.NoError(t, err)
		got, err := TypedValueFromAtomics(CustomFieldKindSingleSelect, atoms)
		require.NoError(t, err)
		require.NotNil(t, got.OptionKey)
		assert.Equal(t, key, *got.OptionKey)
	})

	t.Run("user reference", func(t *testing.T) {
		t.Parallel()
		userID := MustNewID(ResourceTypeUser)
		atoms, err := (CustomFieldTypedValue{Kind: CustomFieldKindUserReference, UserID: &userID}).Atomics()
		require.NoError(t, err)
		got, err := TypedValueFromAtomics(CustomFieldKindUserReference, atoms)
		require.NoError(t, err)
		require.NotNil(t, got.UserID)
		assert.Equal(t, userID, *got.UserID)
	})

	t.Run("user references multi", func(t *testing.T) {
		t.Parallel()
		ids := []ID{MustNewID(ResourceTypeUser), MustNewID(ResourceTypeUser)}
		atoms, err := (CustomFieldTypedValue{Kind: CustomFieldKindUserReference, UserIDs: ids}).Atomics()
		require.NoError(t, err)
		require.Len(t, atoms, 2)
		got, err := TypedValueFromAtomics(CustomFieldKindUserReference, atoms)
		require.NoError(t, err)
		assert.Equal(t, ids, got.UserIDs)
	})

	t.Run("resource reference", func(t *testing.T) {
		t.Parallel()
		issueID := MustNewID(ResourceTypeIssue)
		atoms, err := (CustomFieldTypedValue{Kind: CustomFieldKindResourceReference, ResourceID: &issueID}).Atomics()
		require.NoError(t, err)
		got, err := TypedValueFromAtomics(CustomFieldKindResourceReference, atoms)
		require.NoError(t, err)
		require.NotNil(t, got.ResourceID)
		assert.Equal(t, issueID, *got.ResourceID)
	})

	t.Run("empty rows", func(t *testing.T) {
		t.Parallel()
		_, err := TypedValueFromAtomics(CustomFieldKindText, nil)
		require.ErrorIs(t, err, ErrInvalidCustomFieldValue)
	})

	t.Run("nil scalar payload", func(t *testing.T) {
		t.Parallel()
		_, err := (CustomFieldTypedValue{Kind: CustomFieldKindInteger}).Atomics()
		require.ErrorIs(t, err, ErrInvalidCustomFieldValue)
	})
}

func TestValidateAgainst_RemainingKinds(t *testing.T) {
	t.Parallel()

	t.Run("text length and pattern", func(t *testing.T) {
		t.Parallel()
		minLen, maxLen := 3, 8
		def, err := NewCustomFieldDefinition(
			"code",
			"Code",
			CustomFieldKindText,
			testScope(),
			testOwner(),
			ResourceTypeIssue,
			CustomFieldSchema{Text: &CustomFieldTextSchema{
				MinLength: &minLen,
				MaxLength: &maxLen,
				Pattern:   "^[A-Z]+$",
			}},
		)
		require.NoError(t, err)
		ok := "ABCD"
		require.NoError(t, ValidateAgainst(def, []CustomFieldAtomicValue{{Text: &ok}}))
		tooShort := "AB"
		require.ErrorIs(t, ValidateAgainst(def, []CustomFieldAtomicValue{{Text: &tooShort}}), ErrInvalidCustomFieldValue)
		badPattern := "abcd"
		require.ErrorIs(t, ValidateAgainst(def, []CustomFieldAtomicValue{{Text: &badPattern}}), ErrInvalidCustomFieldValue)
	})

	t.Run("date bounds", func(t *testing.T) {
		t.Parallel()
		minDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		maxDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		def, err := NewCustomFieldDefinition(
			"due_date",
			"Due date",
			CustomFieldKindDate,
			testScope(),
			testOwner(),
			ResourceTypeIssue,
			CustomFieldSchema{Date: &CustomFieldDateSchema{Min: &minDate, Max: &maxDate}},
		)
		require.NoError(t, err)
		ok := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		require.NoError(t, ValidateAgainst(def, []CustomFieldAtomicValue{{Date: &ok}}))
		tooEarly := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
		require.ErrorIs(t, ValidateAgainst(def, []CustomFieldAtomicValue{{Date: &tooEarly}}), ErrInvalidCustomFieldValue)
	})

	t.Run("user type", func(t *testing.T) {
		t.Parallel()
		def, err := NewCustomFieldDefinition(
			"reviewer",
			"Reviewer",
			CustomFieldKindUserReference,
			testScope(),
			testOwner(),
			ResourceTypeIssue,
			CustomFieldSchema{},
		)
		require.NoError(t, err)
		userID := MustNewID(ResourceTypeUser)
		require.NoError(t, ValidateAgainst(def, []CustomFieldAtomicValue{{UserID: &userID}}))
		issueID := MustNewID(ResourceTypeIssue)
		require.ErrorIs(t, ValidateAgainst(def, []CustomFieldAtomicValue{{UserID: &issueID}}), ErrInvalidCustomFieldValue)
	})

	t.Run("resource allowed types", func(t *testing.T) {
		t.Parallel()
		def, err := NewCustomFieldDefinition(
			"linked_issue",
			"Linked issue",
			CustomFieldKindResourceReference,
			testScope(),
			testOwner(),
			ResourceTypeIssue,
			CustomFieldSchema{ResourceReference: &CustomFieldResourceReferenceSchema{
				AllowedTypes: []ResourceType{ResourceTypeIssue},
			}},
		)
		require.NoError(t, err)
		issueID := MustNewID(ResourceTypeIssue)
		require.NoError(t, ValidateAgainst(def, []CustomFieldAtomicValue{{ResourceID: &issueID}}))
		docID := MustNewID(ResourceTypeDocument)
		require.ErrorIs(t, ValidateAgainst(def, []CustomFieldAtomicValue{{ResourceID: &docID}}), ErrInvalidCustomFieldValue)
	})

	t.Run("duplicate multi values", func(t *testing.T) {
		t.Parallel()
		def, err := NewCustomFieldDefinition(
			"labels",
			"Labels",
			CustomFieldKindMultiSelect,
			testScope(),
			testOwner(),
			ResourceTypeIssue,
			CustomFieldSchema{Select: &CustomFieldSelectSchema{
				Options: []CustomFieldOption{
					{Key: "prod", Label: "Production"},
					{Key: "staging", Label: "Staging"},
				},
			}},
		)
		require.NoError(t, err)
		prod := "prod"
		require.ErrorIs(t, ValidateAgainst(def, []CustomFieldAtomicValue{
			{Ordinal: 0, OptionKey: &prod},
			{Ordinal: 1, OptionKey: &prod},
		}), ErrInvalidCustomFieldValue)
	})

	t.Run("wrong representation", func(t *testing.T) {
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
		text := "8"
		n := int64(8)
		require.ErrorIs(t, ValidateAgainst(def, []CustomFieldAtomicValue{
			{Text: &text, Integer: &n},
		}), ErrInvalidCustomFieldValue)
	})

	t.Run("non contiguous ordinals", func(t *testing.T) {
		t.Parallel()
		def, err := NewCustomFieldDefinition(
			"labels",
			"Labels",
			CustomFieldKindMultiSelect,
			testScope(),
			testOwner(),
			ResourceTypeIssue,
			CustomFieldSchema{Select: &CustomFieldSelectSchema{
				Options: []CustomFieldOption{
					{Key: "prod", Label: "Production"},
					{Key: "staging", Label: "Staging"},
				},
			}},
		)
		require.NoError(t, err)
		prod := "prod"
		staging := "staging"
		require.ErrorIs(t, ValidateAgainst(def, []CustomFieldAtomicValue{
			{Ordinal: 0, OptionKey: &prod},
			{Ordinal: 2, OptionKey: &staging},
		}), ErrInvalidCustomFieldValue)
	})
}
