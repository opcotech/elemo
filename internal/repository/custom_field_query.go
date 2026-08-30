package repository

import (
	"time"

	"github.com/opcotech/elemo/internal/model"
)

const (
	CustomFieldOpStageValues    = "stage_values"
	CustomFieldOpDeleteResource = "delete_resource"
	CustomFieldOpPending        = "pending"
	CustomFieldOpCommitted      = "committed"
	CustomFieldOpAborted        = "aborted"
)

type CustomFieldPredicateOp string

const (
	CustomFieldPredEq    CustomFieldPredicateOp = "eq"
	CustomFieldPredGt    CustomFieldPredicateOp = "gt"
	CustomFieldPredGte   CustomFieldPredicateOp = "gte"
	CustomFieldPredLt    CustomFieldPredicateOp = "lt"
	CustomFieldPredLte   CustomFieldPredicateOp = "lte"
	CustomFieldPredMatch CustomFieldPredicateOp = "match"
)

// CustomFieldPredicate is a typed indexed lookup against one definition.
type CustomFieldPredicate struct {
	Op         CustomFieldPredicateOp
	Text       *string
	Integer    *int64
	Decimal    *string
	Boolean    *bool
	Date       *time.Time
	DateTime   *time.Time
	URL        *string
	OptionKey  *string
	UserID     *model.ID
	ResourceID *model.ID
}

// CustomFieldStoredValue is one committed or staged atomic value row.
type CustomFieldStoredValue struct {
	DefinitionID model.ID
	ResourceID   model.ID
	Committed    bool
	Value        model.CustomFieldAtomicValue
}

// CustomFieldOperation records a hybrid create/delete that spans stores.
type CustomFieldOperation struct {
	ID         string
	Kind       string
	Status     string
	ResourceID model.ID
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

// CustomFieldStagedWrite is one definition's uncommitted atomic rows.
type CustomFieldStagedWrite struct {
	Definition *model.CustomFieldDefinition
	Values     []model.CustomFieldAtomicValue
}

const customFieldDefinitionColumns = `
	id, field_key, name, description, kind, scope_id, target_type,
	required, archived, index_exact, index_range, index_fulltext,
	sort_order, owner_user_id, registrar_client_id, created_at, updated_at`

const customFieldInsertDefinition = `
INSERT INTO custom_field_definitions (
	id, field_key, name, description, kind, scope_id, target_type,
	required, archived, index_exact, index_range, index_fulltext,
	sort_order, owner_user_id, registrar_client_id, created_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7,
	$8, $9, $10, $11, $12,
	$13, $14, $15, $16
)`

const customFieldSelectDefinition = `
SELECT ` + customFieldDefinitionColumns + `
FROM custom_field_definitions
WHERE id = $1`

const customFieldListDefinitions = `
SELECT ` + customFieldDefinitionColumns + `
FROM custom_field_definitions
WHERE scope_id = ANY($1)
	AND target_type = $2
	AND ($3::BOOLEAN OR archived = FALSE)
ORDER BY array_position($1, scope_id) DESC, sort_order ASC, created_at ASC, id ASC`

const customFieldNextSortOrder = `
SELECT COALESCE(MAX(sort_order), -1) + 1
FROM custom_field_definitions
WHERE scope_id = $1 AND target_type = $2`

const customFieldUpdateDefinition = `
UPDATE custom_field_definitions SET
	name = $2,
	description = $3,
	required = $4,
	archived = $5,
	index_exact = $6,
	index_range = $7,
	index_fulltext = $8,
	sort_order = $9,
	updated_at = $10
WHERE id = $1`

const customFieldDeleteDefinition = `
DELETE FROM custom_field_definitions WHERE id = $1`

const customFieldCountValues = `
SELECT COUNT(*) FROM custom_field_values WHERE definition_id = $1`

const customFieldDeleteValuesForDefinitionResource = `
DELETE FROM custom_field_values WHERE definition_id = $1 AND resource_id = $2`

const customFieldDeleteValuesForResource = `
DELETE FROM custom_field_values WHERE resource_id = $1`

const customFieldListValuesForResource = `
SELECT
	id, definition_id, resource_id, resource_type, ordinal, committed,
	text_value, integer_value, decimal_value::TEXT, boolean_value,
	date_value, datetime_value, url_value, option_key, user_id, ref_resource_id,
	created_at, updated_at
FROM custom_field_values
WHERE resource_id = $1 AND ($2::BOOLEAN OR committed = TRUE)
ORDER BY definition_id, ordinal`

const customFieldInsertValue = `
INSERT INTO custom_field_values (
	id, definition_id, resource_id, resource_type, ordinal, committed,
	text_value, integer_value, decimal_value, boolean_value,
	date_value, datetime_value, url_value, option_key, user_id, ref_resource_id,
	index_exact, index_range, index_fulltext, created_at
) VALUES (
	$1, $2, $3, $4, $5, $6,
	$7, $8, $9, $10,
	$11, $12, $13, $14, $15, $16,
	$17, $18, $19, $20
)`

const customFieldCommitValues = `
UPDATE custom_field_values
SET committed = TRUE, updated_at = $2
WHERE resource_id = $1 AND committed = FALSE`

const customFieldAbortValues = `
DELETE FROM custom_field_values
WHERE resource_id = $1 AND committed = FALSE`

const customFieldInsertOperation = `
INSERT INTO custom_field_operations (id, kind, status, resource_id, created_at)
VALUES ($1, $2, $3, $4, $5)`

const customFieldUpdateOperation = `
UPDATE custom_field_operations SET status = $2, updated_at = $3 WHERE id = $1`

const customFieldUpdatePendingOperations = `
UPDATE custom_field_operations SET status = $2, updated_at = $3
WHERE resource_id = $1 AND status = 'pending'`

const customFieldListPendingOperations = `
SELECT id, kind, status, resource_id, created_at, updated_at
FROM custom_field_operations
WHERE status = 'pending' AND created_at < $1
ORDER BY created_at ASC
LIMIT $2`

const customFieldInsertOption = `
INSERT INTO custom_field_options (
	id, definition_id, option_key, label, color, sort_order, disabled, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

const customFieldListOptions = `
SELECT option_key, label, color, sort_order, disabled
FROM custom_field_options
WHERE definition_id = $1
ORDER BY sort_order ASC, option_key ASC`
