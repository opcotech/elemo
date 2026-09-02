package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

var (
	ErrCustomFieldCreate      = errors.New("failed to create custom field")
	ErrCustomFieldRead        = errors.New("failed to read custom field")
	ErrCustomFieldUpdate      = errors.New("failed to update custom field")
	ErrCustomFieldDelete      = errors.New("failed to delete custom field")
	ErrCustomFieldKeyConflict = errors.New("custom field key already in use")
	ErrCustomFieldValueWrite  = errors.New("failed to write custom field value")
	ErrCustomFieldSearch      = errors.New("failed to search custom fields")
)

type pgQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

//go:generate go tool mockgen -source=custom_field.go -destination=mock/mock_custom_field_gen.go -package=mockrepo
type CustomFieldRepository interface {
	CreateDefinition(ctx context.Context, def *model.CustomFieldDefinition) (*model.CustomFieldDefinition, error)
	GetDefinition(ctx context.Context, id model.ID) (*model.CustomFieldDefinition, error)
	ListDefinitions(ctx context.Context, scopes []model.ID, target model.ResourceType, includeArchived bool) ([]*model.CustomFieldDefinition, error)
	NextSortOrder(ctx context.Context, scope model.ID, target model.ResourceType) (int, error)
	UpdateDefinition(ctx context.Context, def *model.CustomFieldDefinition) (*model.CustomFieldDefinition, error)
	DeleteDefinition(ctx context.Context, id model.ID) error
	CountValues(ctx context.Context, definitionID model.ID) (int64, error)

	ReplaceValues(ctx context.Context, def *model.CustomFieldDefinition, resourceID model.ID, values []model.CustomFieldAtomicValue, committed bool) error
	ListValues(ctx context.Context, resourceID model.ID, includeUncommitted bool) ([]CustomFieldStoredValue, error)
	DeleteValues(ctx context.Context, definitionID, resourceID model.ID) error
	DeleteForResource(ctx context.Context, resourceID model.ID) error
	CommitValues(ctx context.Context, resourceID model.ID) error
	AbortValues(ctx context.Context, resourceID model.ID) error

	StageValues(ctx context.Context, resourceID model.ID, writes []CustomFieldStagedWrite, op CustomFieldOperation) error
	CreateOperation(ctx context.Context, op CustomFieldOperation) (*CustomFieldOperation, error)
	UpdateOperationStatus(ctx context.Context, id, status string) error
	UpdatePendingOperations(ctx context.Context, resourceID model.ID, status string) error
	ListPendingOperations(ctx context.Context, olderThan time.Time, limit int) ([]CustomFieldOperation, error)

	Search(ctx context.Context, definitionID model.ID, pred CustomFieldPredicate, limit int) ([]model.ID, error)
}

// PGCustomFieldRepository persists custom fields in PostgreSQL.
type PGCustomFieldRepository struct {
	*pgBaseRepository
}

func (r *PGCustomFieldRepository) withTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PGCustomFieldRepository) CreateDefinition(
	ctx context.Context,
	def *model.CustomFieldDefinition,
) (*model.CustomFieldDefinition, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/CreateDefinition")
	defer span.End()

	if def == nil {
		return nil, errors.Join(ErrCustomFieldCreate, model.ErrInvalidCustomFieldDetails)
	}

	created := *def
	created.ID = model.MustNewID(model.ResourceTypeCustomFieldDefinition)
	now := time.Now().UTC().Round(time.Microsecond)
	created.CreatedAt = convert.ToPointer(now)
	created.UpdatedAt = nil
	if err := created.Validate(); err != nil {
		return nil, errors.Join(ErrCustomFieldCreate, err)
	}

	err := r.withTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, customFieldInsertDefinition,
			created.ID,
			created.Key,
			created.Name,
			created.Description,
			created.Kind.String(),
			created.Scope,
			created.TargetType.String(),
			created.Required,
			created.Archived,
			created.IndexExact,
			created.IndexRange,
			created.IndexFullText,
			created.SortOrder,
			created.OwnerUserID,
			created.RegistrarClientID,
			now,
		)
		if err != nil {
			return mapCustomFieldWriteError(err, ErrCustomFieldCreate)
		}
		return insertCustomFieldSchema(ctx, tx, &created)
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *PGCustomFieldRepository) GetDefinition(
	ctx context.Context,
	id model.ID,
) (*model.CustomFieldDefinition, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/GetDefinition")
	defer span.End()

	def, err := scanCustomFieldDefinition(r.db.pool.QueryRow(ctx, customFieldSelectDefinition, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(ErrCustomFieldRead, err)
	}
	if err := loadCustomFieldSchema(ctx, r.db.pool, def); err != nil {
		return nil, errors.Join(ErrCustomFieldRead, err)
	}
	return def, nil
}

func (r *PGCustomFieldRepository) ListDefinitions(
	ctx context.Context,
	scopes []model.ID,
	target model.ResourceType,
	includeArchived bool,
) ([]*model.CustomFieldDefinition, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/ListDefinitions")
	defer span.End()

	scopeKeys := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scopeKeys = append(scopeKeys, scope.Composite())
	}

	rows, err := r.db.pool.Query(ctx, customFieldListDefinitions, scopeKeys, target.String(), includeArchived)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldRead, err)
	}
	defer rows.Close()

	var defs []*model.CustomFieldDefinition
	for rows.Next() {
		def, err := scanCustomFieldDefinition(rows)
		if err != nil {
			return nil, errors.Join(ErrCustomFieldRead, err)
		}
		defs = append(defs, def)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Join(ErrCustomFieldRead, err)
	}
	for _, def := range defs {
		if err := loadCustomFieldSchema(ctx, r.db.pool, def); err != nil {
			return nil, errors.Join(ErrCustomFieldRead, err)
		}
	}
	return defs, nil
}

func (r *PGCustomFieldRepository) NextSortOrder(
	ctx context.Context,
	scope model.ID,
	target model.ResourceType,
) (int, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/NextSortOrder")
	defer span.End()

	var next int
	err := r.db.pool.QueryRow(ctx, customFieldNextSortOrder, scope, target.String()).Scan(&next)
	if err != nil {
		return 0, errors.Join(ErrCustomFieldRead, err)
	}
	return next, nil
}

func (r *PGCustomFieldRepository) UpdateDefinition(
	ctx context.Context,
	def *model.CustomFieldDefinition,
) (*model.CustomFieldDefinition, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/UpdateDefinition")
	defer span.End()

	if def == nil {
		return nil, errors.Join(ErrCustomFieldUpdate, model.ErrInvalidCustomFieldDetails)
	}
	now := time.Now().UTC().Round(time.Microsecond)
	def.UpdatedAt = convert.ToPointer(now)
	if err := def.Validate(); err != nil {
		return nil, errors.Join(ErrCustomFieldUpdate, err)
	}

	err := r.withTx(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, customFieldUpdateDefinition,
			def.ID,
			def.Name,
			def.Description,
			def.Required,
			def.Archived,
			def.IndexExact,
			def.IndexRange,
			def.IndexFullText,
			def.SortOrder,
			now,
		)
		if err != nil {
			return errors.Join(ErrCustomFieldUpdate, err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		if err := replaceCustomFieldSchema(ctx, tx, def); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return def, nil
}

func (r *PGCustomFieldRepository) DeleteDefinition(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/DeleteDefinition")
	defer span.End()

	tag, err := r.db.pool.Exec(ctx, customFieldDeleteDefinition, id)
	if err != nil {
		return errors.Join(ErrCustomFieldDelete, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PGCustomFieldRepository) CountValues(ctx context.Context, definitionID model.ID) (int64, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/CountValues")
	defer span.End()

	var count int64
	err := r.db.pool.QueryRow(ctx, customFieldCountValues, definitionID).Scan(&count)
	if err != nil {
		return 0, errors.Join(ErrCustomFieldRead, err)
	}
	return count, nil
}

func (r *PGCustomFieldRepository) ReplaceValues(
	ctx context.Context,
	def *model.CustomFieldDefinition,
	resourceID model.ID,
	values []model.CustomFieldAtomicValue,
	committed bool,
) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/ReplaceValues")
	defer span.End()

	if def == nil {
		return errors.Join(ErrCustomFieldValueWrite, model.ErrInvalidCustomFieldDetails)
	}

	return r.withTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, customFieldDeleteValuesForDefinitionResource, def.ID, resourceID); err != nil {
			return errors.Join(ErrCustomFieldValueWrite, err)
		}
		now := time.Now().UTC().Round(time.Microsecond)
		for _, value := range values {
			if err := insertCustomFieldValue(ctx, tx, def, resourceID, value, committed, now); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PGCustomFieldRepository) ListValues(
	ctx context.Context,
	resourceID model.ID,
	includeUncommitted bool,
) ([]CustomFieldStoredValue, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/ListValues")
	defer span.End()

	rows, err := r.db.pool.Query(ctx, customFieldListValuesForResource, resourceID, includeUncommitted)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldRead, err)
	}
	defer rows.Close()

	var out []CustomFieldStoredValue
	for rows.Next() {
		stored, err := scanCustomFieldStoredValue(rows)
		if err != nil {
			return nil, errors.Join(ErrCustomFieldRead, err)
		}
		out = append(out, stored)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Join(ErrCustomFieldRead, err)
	}
	return out, nil
}

func (r *PGCustomFieldRepository) DeleteValues(ctx context.Context, definitionID, resourceID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/DeleteValues")
	defer span.End()

	_, err := r.db.pool.Exec(ctx, customFieldDeleteValuesForDefinitionResource, definitionID, resourceID)
	if err != nil {
		return errors.Join(ErrCustomFieldDelete, err)
	}
	return nil
}

func (r *PGCustomFieldRepository) DeleteForResource(ctx context.Context, resourceID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/DeleteForResource")
	defer span.End()

	_, err := r.db.pool.Exec(ctx, customFieldDeleteValuesForResource, resourceID)
	if err != nil {
		return errors.Join(ErrCustomFieldDelete, err)
	}
	return nil
}

func (r *PGCustomFieldRepository) CommitValues(ctx context.Context, resourceID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/CommitValues")
	defer span.End()

	_, err := r.db.pool.Exec(ctx, customFieldCommitValues, resourceID, time.Now().UTC())
	if err != nil {
		return errors.Join(ErrCustomFieldValueWrite, err)
	}
	return nil
}

func (r *PGCustomFieldRepository) AbortValues(ctx context.Context, resourceID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/AbortValues")
	defer span.End()

	_, err := r.db.pool.Exec(ctx, customFieldAbortValues, resourceID)
	if err != nil {
		return errors.Join(ErrCustomFieldValueWrite, err)
	}
	return nil
}

func (r *PGCustomFieldRepository) StageValues(
	ctx context.Context,
	resourceID model.ID,
	writes []CustomFieldStagedWrite,
	op CustomFieldOperation,
) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/StageValues")
	defer span.End()

	return r.withTx(ctx, func(tx pgx.Tx) error {
		now := time.Now().UTC().Round(time.Microsecond)
		for _, write := range writes {
			if write.Definition == nil {
				return errors.Join(ErrCustomFieldValueWrite, model.ErrInvalidCustomFieldDetails)
			}
			if _, err := tx.Exec(ctx, customFieldDeleteValuesForDefinitionResource, write.Definition.ID, resourceID); err != nil {
				return errors.Join(ErrCustomFieldValueWrite, err)
			}
			for _, value := range write.Values {
				if err := insertCustomFieldValue(ctx, tx, write.Definition, resourceID, value, false, now); err != nil {
					return err
				}
			}
		}
		if op.ID == "" {
			op.ID = model.NewRawID()
		}
		if op.Kind == "" {
			op.Kind = CustomFieldOpStageValues
		}
		if op.Status == "" {
			op.Status = CustomFieldOpPending
		}
		_, err := tx.Exec(ctx, customFieldInsertOperation, op.ID, op.Kind, op.Status, resourceID, now)
		if err != nil {
			return errors.Join(ErrCustomFieldValueWrite, err)
		}
		return nil
	})
}

func (r *PGCustomFieldRepository) CreateOperation(
	ctx context.Context,
	op CustomFieldOperation,
) (*CustomFieldOperation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/CreateOperation")
	defer span.End()

	if op.ID == "" {
		op.ID = model.NewRawID()
	}
	if op.Status == "" {
		op.Status = CustomFieldOpPending
	}
	op.CreatedAt = time.Now().UTC().Round(time.Microsecond)
	_, err := r.db.pool.Exec(ctx, customFieldInsertOperation, op.ID, op.Kind, op.Status, op.ResourceID, op.CreatedAt)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldValueWrite, err)
	}
	return &op, nil
}

func (r *PGCustomFieldRepository) UpdateOperationStatus(ctx context.Context, id, status string) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/UpdateOperationStatus")
	defer span.End()

	tag, err := r.db.pool.Exec(ctx, customFieldUpdateOperation, id, status, time.Now().UTC())
	if err != nil {
		return errors.Join(ErrCustomFieldValueWrite, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PGCustomFieldRepository) UpdatePendingOperations(ctx context.Context, resourceID model.ID, status string) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/UpdatePendingOperations")
	defer span.End()

	_, err := r.db.pool.Exec(ctx, customFieldUpdatePendingOperations, resourceID, status, time.Now().UTC())
	if err != nil {
		return errors.Join(ErrCustomFieldValueWrite, err)
	}
	return nil
}

func (r *PGCustomFieldRepository) ListPendingOperations(
	ctx context.Context,
	olderThan time.Time,
	limit int,
) ([]CustomFieldOperation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/ListPendingOperations")
	defer span.End()

	rows, err := r.db.pool.Query(ctx, customFieldListPendingOperations, olderThan, limit)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldRead, err)
	}
	defer rows.Close()

	var ops []CustomFieldOperation
	for rows.Next() {
		var op CustomFieldOperation
		if err := rows.Scan(&op.ID, &op.Kind, &op.Status, &op.ResourceID, &op.CreatedAt, &op.UpdatedAt); err != nil {
			return nil, errors.Join(ErrCustomFieldRead, err)
		}
		ops = append(ops, op)
	}
	return ops, rows.Err()
}

func (r *PGCustomFieldRepository) Search(
	ctx context.Context,
	definitionID model.ID,
	pred CustomFieldPredicate,
	limit int,
) ([]model.ID, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.CustomFieldRepository/Search")
	defer span.End()

	if limit <= 0 {
		limit = 100
	}

	query, args, err := compileCustomFieldSearch(definitionID, pred, limit)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldSearch, err)
	}

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Join(ErrCustomFieldSearch, err)
	}
	defer rows.Close()

	var ids []model.ID
	for rows.Next() {
		var id model.ID
		if err := rows.Scan(&id); err != nil {
			return nil, errors.Join(ErrCustomFieldSearch, err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Join(ErrCustomFieldSearch, err)
	}
	return ids, nil
}

// NewCustomFieldRepository creates a PostgreSQL custom-field repository.
func NewCustomFieldRepository(opts ...PGRepositoryOption) (*PGCustomFieldRepository, error) {
	baseRepo, err := newPGRepository(opts...)
	if err != nil {
		return nil, err
	}
	return &PGCustomFieldRepository{pgBaseRepository: baseRepo}, nil
}

func mapCustomFieldWriteError(err, wrap error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return errors.Join(wrap, ErrCustomFieldKeyConflict, err)
	}
	return errors.Join(wrap, err)
}

func insertCustomFieldValue(
	ctx context.Context,
	q pgQuerier,
	def *model.CustomFieldDefinition,
	resourceID model.ID,
	value model.CustomFieldAtomicValue,
	committed bool,
	now time.Time,
) error {
	_, err := q.Exec(ctx, customFieldInsertValue,
		model.NewRawID(),
		def.ID,
		resourceID,
		resourceID.Type.String(),
		value.Ordinal,
		committed,
		value.Text,
		value.Integer,
		value.Decimal,
		value.Boolean,
		value.Date,
		value.DateTime,
		value.URL,
		value.OptionKey,
		compositeIDArg(value.UserID),
		compositeIDArg(value.ResourceID),
		def.IndexExact,
		def.IndexRange,
		def.IndexFullText,
		now,
	)
	if err != nil {
		return errors.Join(ErrCustomFieldValueWrite, err)
	}
	return nil
}

func insertCustomFieldSchema(ctx context.Context, q pgQuerier, def *model.CustomFieldDefinition) error {
	switch def.Kind {
	case model.CustomFieldKindText:
		s := def.Schema.Text
		_, err := q.Exec(ctx,
			`INSERT INTO custom_field_schema_text (definition_id, min_length, max_length, pattern) VALUES ($1, $2, $3, $4)`,
			def.ID, s.MinLength, s.MaxLength, s.Pattern,
		)
		return err
	case model.CustomFieldKindInteger:
		s := def.Schema.Integer
		_, err := q.Exec(ctx,
			`INSERT INTO custom_field_schema_integer (definition_id, min_value, max_value) VALUES ($1, $2, $3)`,
			def.ID, s.Min, s.Max,
		)
		return err
	case model.CustomFieldKindDecimal:
		s := def.Schema.Decimal
		var minVal, maxVal any
		if s.Min != "" {
			minVal = s.Min
		}
		if s.Max != "" {
			maxVal = s.Max
		}
		_, err := q.Exec(ctx,
			`INSERT INTO custom_field_schema_decimal (definition_id, min_value, max_value, scale) VALUES ($1, $2, $3, $4)`,
			def.ID, minVal, maxVal, s.Scale,
		)
		return err
	case model.CustomFieldKindBoolean:
		_, err := q.Exec(ctx,
			`INSERT INTO custom_field_schema_boolean (definition_id) VALUES ($1)`,
			def.ID,
		)
		return err
	case model.CustomFieldKindDate:
		s := def.Schema.Date
		_, err := q.Exec(ctx,
			`INSERT INTO custom_field_schema_date (definition_id, min_value, max_value) VALUES ($1, $2, $3)`,
			def.ID, s.Min, s.Max,
		)
		return err
	case model.CustomFieldKindDatetime:
		s := def.Schema.DateTime
		_, err := q.Exec(ctx,
			`INSERT INTO custom_field_schema_datetime (definition_id, min_value, max_value) VALUES ($1, $2, $3)`,
			def.ID, s.Min, s.Max,
		)
		return err
	case model.CustomFieldKindURL:
		s := def.Schema.URL
		_, err := q.Exec(ctx,
			`INSERT INTO custom_field_schema_url (definition_id, allowed_schemes) VALUES ($1, $2)`,
			def.ID, s.AllowedSchemes,
		)
		return err
	case model.CustomFieldKindSingleSelect, model.CustomFieldKindMultiSelect:
		if _, err := q.Exec(ctx, `INSERT INTO custom_field_schema_select (definition_id) VALUES ($1)`, def.ID); err != nil {
			return err
		}
		return insertCustomFieldOptions(ctx, q, def)
	case model.CustomFieldKindUserReference:
		s := def.Schema.UserReference
		_, err := q.Exec(ctx,
			`INSERT INTO custom_field_schema_user_reference (definition_id, multiple) VALUES ($1, $2)`,
			def.ID, s.Multiple,
		)
		return err
	case model.CustomFieldKindResourceReference:
		s := def.Schema.ResourceReference
		types := make([]string, len(s.AllowedTypes))
		for i, rt := range s.AllowedTypes {
			types[i] = rt.String()
		}
		_, err := q.Exec(ctx,
			`INSERT INTO custom_field_schema_resource_reference (definition_id, allowed_types, multiple) VALUES ($1, $2, $3)`,
			def.ID, types, s.Multiple,
		)
		return err
	default:
		return model.ErrInvalidCustomFieldKind
	}
}

func replaceCustomFieldSchema(ctx context.Context, q pgQuerier, def *model.CustomFieldDefinition) error {
	tables := []string{
		"custom_field_options",
		"custom_field_schema_text",
		"custom_field_schema_integer",
		"custom_field_schema_decimal",
		"custom_field_schema_boolean",
		"custom_field_schema_date",
		"custom_field_schema_datetime",
		"custom_field_schema_url",
		"custom_field_schema_select",
		"custom_field_schema_user_reference",
		"custom_field_schema_resource_reference",
	}
	for _, table := range tables {
		if _, err := q.Exec(ctx, "DELETE FROM "+table+" WHERE definition_id = $1", def.ID); err != nil {
			return errors.Join(ErrCustomFieldUpdate, err)
		}
	}
	if err := insertCustomFieldSchema(ctx, q, def); err != nil {
		return errors.Join(ErrCustomFieldUpdate, err)
	}
	return nil
}

func insertCustomFieldOptions(ctx context.Context, q pgQuerier, def *model.CustomFieldDefinition) error {
	if def.Schema.Select == nil {
		return nil
	}
	now := time.Now().UTC()
	for i, opt := range def.Schema.Select.Options {
		order := opt.Order
		if order == 0 {
			order = i
		}
		if _, err := q.Exec(ctx, customFieldInsertOption,
			model.NewRawID(), def.ID, opt.Key, opt.Label, opt.Color, order, opt.Disabled, now,
		); err != nil {
			return err
		}
	}
	return nil
}

func loadCustomFieldSchema(ctx context.Context, q pgQuerier, def *model.CustomFieldDefinition) error {
	switch def.Kind {
	case model.CustomFieldKindText:
		var schema model.CustomFieldTextSchema
		err := q.QueryRow(ctx,
			`SELECT min_length, max_length, pattern FROM custom_field_schema_text WHERE definition_id = $1`,
			def.ID,
		).Scan(&schema.MinLength, &schema.MaxLength, &schema.Pattern)
		if err != nil {
			return err
		}
		def.Schema.Text = &schema
	case model.CustomFieldKindInteger:
		var schema model.CustomFieldIntegerSchema
		err := q.QueryRow(ctx,
			`SELECT min_value, max_value FROM custom_field_schema_integer WHERE definition_id = $1`,
			def.ID,
		).Scan(&schema.Min, &schema.Max)
		if err != nil {
			return err
		}
		def.Schema.Integer = &schema
	case model.CustomFieldKindDecimal:
		var schema model.CustomFieldDecimalSchema
		var minVal, maxVal *string
		err := q.QueryRow(ctx,
			`SELECT min_value::TEXT, max_value::TEXT, scale FROM custom_field_schema_decimal WHERE definition_id = $1`,
			def.ID,
		).Scan(&minVal, &maxVal, &schema.Scale)
		if err != nil {
			return err
		}
		if minVal != nil {
			schema.Min = *minVal
		}
		if maxVal != nil {
			schema.Max = *maxVal
		}
		def.Schema.Decimal = &schema
	case model.CustomFieldKindBoolean:
		var exists int
		if err := q.QueryRow(ctx, `SELECT 1 FROM custom_field_schema_boolean WHERE definition_id = $1`, def.ID).Scan(&exists); err != nil {
			return err
		}
		def.Schema.Boolean = &model.CustomFieldBooleanSchema{}
	case model.CustomFieldKindDate:
		var schema model.CustomFieldDateSchema
		err := q.QueryRow(ctx,
			`SELECT min_value, max_value FROM custom_field_schema_date WHERE definition_id = $1`,
			def.ID,
		).Scan(&schema.Min, &schema.Max)
		if err != nil {
			return err
		}
		def.Schema.Date = &schema
	case model.CustomFieldKindDatetime:
		var schema model.CustomFieldDateTimeSchema
		err := q.QueryRow(ctx,
			`SELECT min_value, max_value FROM custom_field_schema_datetime WHERE definition_id = $1`,
			def.ID,
		).Scan(&schema.Min, &schema.Max)
		if err != nil {
			return err
		}
		def.Schema.DateTime = &schema
	case model.CustomFieldKindURL:
		var schema model.CustomFieldURLSchema
		err := q.QueryRow(ctx,
			`SELECT allowed_schemes FROM custom_field_schema_url WHERE definition_id = $1`,
			def.ID,
		).Scan(&schema.AllowedSchemes)
		if err != nil {
			return err
		}
		def.Schema.URL = &schema
	case model.CustomFieldKindSingleSelect, model.CustomFieldKindMultiSelect:
		options, err := loadCustomFieldOptions(ctx, q, def.ID)
		if err != nil {
			return err
		}
		def.Schema.Select = &model.CustomFieldSelectSchema{Options: options}
	case model.CustomFieldKindUserReference:
		var schema model.CustomFieldUserReferenceSchema
		err := q.QueryRow(ctx,
			`SELECT multiple FROM custom_field_schema_user_reference WHERE definition_id = $1`,
			def.ID,
		).Scan(&schema.Multiple)
		if err != nil {
			return err
		}
		def.Schema.UserReference = &schema
	case model.CustomFieldKindResourceReference:
		var schema model.CustomFieldResourceReferenceSchema
		var types []string
		err := q.QueryRow(ctx,
			`SELECT allowed_types, multiple FROM custom_field_schema_resource_reference WHERE definition_id = $1`,
			def.ID,
		).Scan(&types, &schema.Multiple)
		if err != nil {
			return err
		}
		schema.AllowedTypes = make([]model.ResourceType, len(types))
		for i, raw := range types {
			if err := schema.AllowedTypes[i].UnmarshalText([]byte(raw)); err != nil {
				return err
			}
		}
		def.Schema.ResourceReference = &schema
	}
	return nil
}

func loadCustomFieldOptions(ctx context.Context, q pgQuerier, definitionID model.ID) ([]model.CustomFieldOption, error) {
	rows, err := q.Query(ctx, customFieldListOptions, definitionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []model.CustomFieldOption
	for rows.Next() {
		var opt model.CustomFieldOption
		if err := rows.Scan(&opt.Key, &opt.Label, &opt.Color, &opt.Order, &opt.Disabled); err != nil {
			return nil, err
		}
		options = append(options, opt)
	}
	return options, rows.Err()
}

func scanCustomFieldDefinition(row interface{ Scan(dest ...any) error }) (*model.CustomFieldDefinition, error) {
	var (
		def        model.CustomFieldDefinition
		kind       string
		targetType string
	)
	err := row.Scan(
		&def.ID,
		&def.Key,
		&def.Name,
		&def.Description,
		&kind,
		&def.Scope,
		&targetType,
		&def.Required,
		&def.Archived,
		&def.IndexExact,
		&def.IndexRange,
		&def.IndexFullText,
		&def.SortOrder,
		&def.OwnerUserID,
		&def.RegistrarClientID,
		&def.CreatedAt,
		&def.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := def.Kind.UnmarshalText([]byte(kind)); err != nil {
		return nil, err
	}
	if err := def.TargetType.UnmarshalText([]byte(targetType)); err != nil {
		return nil, err
	}
	return &def, nil
}

func scanCustomFieldStoredValue(row interface{ Scan(dest ...any) error }) (CustomFieldStoredValue, error) {
	var (
		stored       CustomFieldStoredValue
		id           string
		resourceType string
		decimal      *string
		userID       *string
		refID        *string
		updatedAt    *time.Time
		createdAt    time.Time
	)
	err := row.Scan(
		&id,
		&stored.DefinitionID,
		&stored.ResourceID,
		&resourceType,
		&stored.Value.Ordinal,
		&stored.Committed,
		&stored.Value.Text,
		&stored.Value.Integer,
		&decimal,
		&stored.Value.Boolean,
		&stored.Value.Date,
		&stored.Value.DateTime,
		&stored.Value.URL,
		&stored.Value.OptionKey,
		&userID,
		&refID,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return CustomFieldStoredValue{}, err
	}
	stored.Value.Decimal = decimal
	stored.Value.UserID, err = parseOptionalCompositeID(userID)
	if err != nil {
		return CustomFieldStoredValue{}, err
	}
	stored.Value.ResourceID, err = parseOptionalCompositeID(refID)
	if err != nil {
		return CustomFieldStoredValue{}, err
	}
	return stored, nil
}

func compositeIDArg(id *model.ID) any {
	if id == nil {
		return nil
	}
	return id.Composite()
}

func parseOptionalCompositeID(raw *string) (*model.ID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	id, err := model.ParseCompositeID(*raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func compileCustomFieldSearch(definitionID model.ID, pred CustomFieldPredicate, limit int) (string, []any, error) {
	column, value, indexClause, err := customFieldPredicateColumn(pred)
	if err != nil {
		return "", nil, err
	}

	op, err := customFieldSQLOp(pred.Op)
	if err != nil {
		return "", nil, err
	}

	var b strings.Builder
	b.WriteString(`SELECT DISTINCT resource_id FROM custom_field_values WHERE definition_id = $1 AND committed = TRUE AND `)
	b.WriteString(indexClause)
	b.WriteString(" AND ")
	if pred.Op == CustomFieldPredMatch {
		b.WriteString(`to_tsvector('simple', text_value) @@ plainto_tsquery('simple', $2)`)
	} else {
		b.WriteString(column)
		b.WriteString(" ")
		b.WriteString(op)
		b.WriteString(" $2")
	}
	fmt.Fprintf(&b, " ORDER BY resource_id LIMIT $%d", 3)
	return b.String(), []any{definitionID, value, limit}, nil
}

func customFieldSQLOp(op CustomFieldPredicateOp) (string, error) {
	switch op {
	case CustomFieldPredEq, CustomFieldPredMatch:
		return "=", nil
	case CustomFieldPredGt:
		return ">", nil
	case CustomFieldPredGte:
		return ">=", nil
	case CustomFieldPredLt:
		return "<", nil
	case CustomFieldPredLte:
		return "<=", nil
	default:
		return "", errors.New("unsupported custom field predicate")
	}
}

func customFieldPredicateColumn(pred CustomFieldPredicate) (column string, value any, indexClause string, err error) {
	switch {
	case pred.Text != nil:
		if pred.Op == CustomFieldPredMatch {
			return "text_value", *pred.Text, "index_fulltext", nil
		}
		return "text_value", *pred.Text, "index_exact", nil
	case pred.Integer != nil:
		clause := "index_exact"
		if pred.Op != CustomFieldPredEq {
			clause = "(index_exact OR index_range)"
		}
		return "integer_value", *pred.Integer, clause, nil
	case pred.Decimal != nil:
		clause := "index_exact"
		if pred.Op != CustomFieldPredEq {
			clause = "(index_exact OR index_range)"
		}
		return "decimal_value", *pred.Decimal, clause, nil
	case pred.Boolean != nil:
		return "boolean_value", *pred.Boolean, "index_exact", nil
	case pred.Date != nil:
		clause := "index_exact"
		if pred.Op != CustomFieldPredEq {
			clause = "(index_exact OR index_range)"
		}
		return "date_value", *pred.Date, clause, nil
	case pred.DateTime != nil:
		clause := "index_exact"
		if pred.Op != CustomFieldPredEq {
			clause = "(index_exact OR index_range)"
		}
		return "datetime_value", *pred.DateTime, clause, nil
	case pred.URL != nil:
		return "url_value", *pred.URL, "index_exact", nil
	case pred.OptionKey != nil:
		return "option_key", *pred.OptionKey, "index_exact", nil
	case pred.UserID != nil:
		return "user_id", *pred.UserID, "index_exact", nil
	case pred.ResourceID != nil:
		return "ref_resource_id", *pred.ResourceID, "index_exact", nil
	default:
		return "", nil, "", errors.New("predicate has no value")
	}
}
