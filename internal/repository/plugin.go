package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

var (
	ErrPluginCreate   = errors.New("failed to create plugin installation")
	ErrPluginRead     = errors.New("failed to read plugin installation")
	ErrPluginUpdate   = errors.New("failed to update plugin installation")
	ErrPluginDelete   = errors.New("failed to delete plugin installation")
	ErrPluginConflict = errors.New("plugin is already installed")
)

//go:generate go tool mockgen -source=plugin.go -destination=mock/mock_plugin_gen.go -package=mockrepo
type PluginRepository interface {
	UpsertInstallation(ctx context.Context, inst *model.PluginInstallation) (*model.PluginInstallation, error)
	GetInstallation(ctx context.Context, pluginID string) (*model.PluginInstallation, error)
	ListInstallations(ctx context.Context) ([]*model.PluginInstallation, error)
	DeleteInstallation(ctx context.Context, pluginID string) error

	UpsertActivation(ctx context.Context, act *model.PluginActivation) (*model.PluginActivation, error)
	GetActivation(ctx context.Context, pluginID string, scope model.ID) (*model.PluginActivation, error)
	ListActivations(ctx context.Context, pluginID string) ([]*model.PluginActivation, error)
	ListActivationsByScope(ctx context.Context, scopes []model.ID) ([]*model.PluginActivation, error)
	DeleteActivations(ctx context.Context, pluginID string) error

	GetStorage(ctx context.Context, pluginID string, scope model.ID, key string) (*model.PluginStorageEntry, error)
	SetStorage(ctx context.Context, entry *model.PluginStorageEntry) (*model.PluginStorageEntry, error)
	ListStorage(ctx context.Context, pluginID string, scope model.ID) ([]*model.PluginStorageEntry, error)
	DeleteStorage(ctx context.Context, pluginID string, scope model.ID, key string) error
	DeleteStorageForPlugin(ctx context.Context, pluginID string) error
}

type PGPluginRepository struct {
	*pgBaseRepository
}

func NewPluginRepository(opts ...PGRepositoryOption) (*PGPluginRepository, error) {
	base, err := newPGRepository(opts...)
	if err != nil {
		return nil, err
	}
	return &PGPluginRepository{pgBaseRepository: base}, nil
}

func (r *PGPluginRepository) UpsertInstallation(
	ctx context.Context,
	inst *model.PluginInstallation,
) (*model.PluginInstallation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/UpsertInstallation")
	defer span.End()

	if inst == nil {
		return nil, ErrPluginCreate
	}
	if inst.ID == "" {
		inst.ID = model.NewRawID()
	}
	now := time.Now().UTC().Round(time.Microsecond)
	if inst.CreatedAt == nil {
		inst.CreatedAt = convert.ToPointer(now)
	}
	inst.UpdatedAt = convert.ToPointer(now)
	if inst.Status == 0 {
		inst.Status = model.PluginStatusInstalled
	}

	manifest, err := json.Marshal(inst.Manifest)
	if err != nil {
		return nil, errors.Join(ErrPluginCreate, err)
	}

	_, err = r.db.pool.Exec(ctx, `
		INSERT INTO plugin_installations (id, plugin_id, version, status, manifest, error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (plugin_id) DO UPDATE SET
			id = EXCLUDED.id,
			version = EXCLUDED.version,
			status = EXCLUDED.status,
			manifest = EXCLUDED.manifest,
			error_message = EXCLUDED.error_message,
			updated_at = EXCLUDED.updated_at
	`, inst.ID, inst.PluginID, inst.Version, inst.Status.String(), manifest, inst.Error, *inst.CreatedAt, inst.UpdatedAt)
	if err != nil {
		return nil, errors.Join(ErrPluginCreate, err)
	}
	return inst, nil
}

func (r *PGPluginRepository) GetInstallation(ctx context.Context, pluginID string) (*model.PluginInstallation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/GetInstallation")
	defer span.End()

	row := r.db.pool.QueryRow(ctx, `
		SELECT id, plugin_id, version, status, manifest, error_message, created_at, updated_at
		FROM plugin_installations WHERE plugin_id = $1
	`, pluginID)
	inst, err := scanPluginInstallation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(ErrPluginRead, err)
	}
	return inst, nil
}

func (r *PGPluginRepository) ListInstallations(ctx context.Context) ([]*model.PluginInstallation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/ListInstallations")
	defer span.End()

	rows, err := r.db.pool.Query(ctx, `
		SELECT id, plugin_id, version, status, manifest, error_message, created_at, updated_at
		FROM plugin_installations ORDER BY plugin_id
	`)
	if err != nil {
		return nil, errors.Join(ErrPluginRead, err)
	}
	defer rows.Close()

	var out []*model.PluginInstallation
	for rows.Next() {
		inst, err := scanPluginInstallation(rows)
		if err != nil {
			return nil, errors.Join(ErrPluginRead, err)
		}
		out = append(out, inst)
	}
	return out, rows.Err()
}

func (r *PGPluginRepository) DeleteInstallation(ctx context.Context, pluginID string) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/DeleteInstallation")
	defer span.End()

	tag, err := r.db.pool.Exec(ctx, `DELETE FROM plugin_installations WHERE plugin_id = $1`, pluginID)
	if err != nil {
		return errors.Join(ErrPluginDelete, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PGPluginRepository) UpsertActivation(
	ctx context.Context,
	act *model.PluginActivation,
) (*model.PluginActivation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/UpsertActivation")
	defer span.End()

	now := time.Now().UTC().Round(time.Microsecond)
	if act.CreatedAt == nil {
		act.CreatedAt = convert.ToPointer(now)
	}
	act.UpdatedAt = convert.ToPointer(now)

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO plugin_activations (plugin_id, scope_id, enabled, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (plugin_id, scope_id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			config = EXCLUDED.config,
			updated_at = EXCLUDED.updated_at
	`, act.PluginID, act.ScopeID, act.Enabled, configBytes(act.Config), *act.CreatedAt, act.UpdatedAt)
	if err != nil {
		return nil, errors.Join(ErrPluginUpdate, err)
	}
	return act, nil
}

func (r *PGPluginRepository) GetActivation(
	ctx context.Context,
	pluginID string,
	scope model.ID,
) (*model.PluginActivation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/GetActivation")
	defer span.End()

	row := r.db.pool.QueryRow(ctx, `
		SELECT plugin_id, scope_id, enabled, config, created_at, updated_at
		FROM plugin_activations WHERE plugin_id = $1 AND scope_id = $2
	`, pluginID, scope)
	act, err := scanPluginActivation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(ErrPluginRead, err)
	}
	return act, nil
}

func (r *PGPluginRepository) ListActivations(ctx context.Context, pluginID string) ([]*model.PluginActivation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/ListActivations")
	defer span.End()

	rows, err := r.db.pool.Query(ctx, `
		SELECT plugin_id, scope_id, enabled, config, created_at, updated_at
		FROM plugin_activations WHERE plugin_id = $1
	`, pluginID)
	if err != nil {
		return nil, errors.Join(ErrPluginRead, err)
	}
	defer rows.Close()
	return scanPluginActivations(rows)
}

func (r *PGPluginRepository) ListActivationsByScope(
	ctx context.Context,
	scopes []model.ID,
) ([]*model.PluginActivation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/ListActivationsByScope")
	defer span.End()

	keys := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		keys = append(keys, scope.Composite())
	}
	rows, err := r.db.pool.Query(ctx, `
		SELECT plugin_id, scope_id, enabled, config, created_at, updated_at
		FROM plugin_activations WHERE scope_id = ANY($1) AND enabled = TRUE
	`, keys)
	if err != nil {
		return nil, errors.Join(ErrPluginRead, err)
	}
	defer rows.Close()
	return scanPluginActivations(rows)
}

func (r *PGPluginRepository) DeleteActivations(ctx context.Context, pluginID string) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/DeleteActivations")
	defer span.End()

	_, err := r.db.pool.Exec(ctx, `DELETE FROM plugin_activations WHERE plugin_id = $1`, pluginID)
	if err != nil {
		return errors.Join(ErrPluginDelete, err)
	}
	return nil
}

func (r *PGPluginRepository) GetStorage(
	ctx context.Context,
	pluginID string,
	scope model.ID,
	key string,
) (*model.PluginStorageEntry, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/GetStorage")
	defer span.End()

	row := r.db.pool.QueryRow(ctx, `
		SELECT plugin_id, scope_id, storage_key, value, created_at, updated_at
		FROM plugin_storage WHERE plugin_id = $1 AND scope_id = $2 AND storage_key = $3
	`, pluginID, scope, key)
	entry, err := scanPluginStorage(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(ErrPluginRead, err)
	}
	return entry, nil
}

func (r *PGPluginRepository) SetStorage(
	ctx context.Context,
	entry *model.PluginStorageEntry,
) (*model.PluginStorageEntry, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/SetStorage")
	defer span.End()

	now := time.Now().UTC().Round(time.Microsecond)
	if entry.CreatedAt == nil {
		entry.CreatedAt = convert.ToPointer(now)
	}
	entry.UpdatedAt = convert.ToPointer(now)
	if entry.Value == nil {
		entry.Value = []byte("null")
	}

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO plugin_storage (plugin_id, scope_id, storage_key, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (plugin_id, scope_id, storage_key) DO UPDATE SET
			value = EXCLUDED.value,
			updated_at = EXCLUDED.updated_at
	`, entry.PluginID, entry.ScopeID, entry.Key, entry.Value, *entry.CreatedAt, entry.UpdatedAt)
	if err != nil {
		return nil, errors.Join(ErrPluginUpdate, err)
	}
	return entry, nil
}

func (r *PGPluginRepository) ListStorage(
	ctx context.Context,
	pluginID string,
	scope model.ID,
) ([]*model.PluginStorageEntry, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/ListStorage")
	defer span.End()

	rows, err := r.db.pool.Query(ctx, `
		SELECT plugin_id, scope_id, storage_key, value, created_at, updated_at
		FROM plugin_storage WHERE plugin_id = $1 AND scope_id = $2 ORDER BY storage_key
	`, pluginID, scope)
	if err != nil {
		return nil, errors.Join(ErrPluginRead, err)
	}
	defer rows.Close()

	var out []*model.PluginStorageEntry
	for rows.Next() {
		entry, err := scanPluginStorage(rows)
		if err != nil {
			return nil, errors.Join(ErrPluginRead, err)
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

func (r *PGPluginRepository) DeleteStorage(ctx context.Context, pluginID string, scope model.ID, key string) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/DeleteStorage")
	defer span.End()

	tag, err := r.db.pool.Exec(ctx, `
		DELETE FROM plugin_storage WHERE plugin_id = $1 AND scope_id = $2 AND storage_key = $3
	`, pluginID, scope, key)
	if err != nil {
		return errors.Join(ErrPluginDelete, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PGPluginRepository) DeleteStorageForPlugin(ctx context.Context, pluginID string) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.PluginRepository/DeleteStorageForPlugin")
	defer span.End()

	_, err := r.db.pool.Exec(ctx, `DELETE FROM plugin_storage WHERE plugin_id = $1`, pluginID)
	if err != nil {
		return errors.Join(ErrPluginDelete, err)
	}
	return nil
}

type pluginRow interface {
	Scan(dest ...any) error
}

func scanPluginInstallation(row pluginRow) (*model.PluginInstallation, error) {
	var inst model.PluginInstallation
	var status string
	var manifest []byte
	if err := row.Scan(
		&inst.ID, &inst.PluginID, &inst.Version, &status, &manifest, &inst.Error, &inst.CreatedAt, &inst.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if err := inst.Status.UnmarshalText([]byte(status)); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(manifest, &inst.Manifest); err != nil {
		return nil, err
	}
	return &inst, nil
}

func scanPluginActivation(row pluginRow) (*model.PluginActivation, error) {
	var act model.PluginActivation
	if err := row.Scan(&act.PluginID, &act.ScopeID, &act.Enabled, &act.Config, &act.CreatedAt, &act.UpdatedAt); err != nil {
		return nil, err
	}
	return &act, nil
}

func configBytes(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte("{}")
	}
	return raw
}

func scanPluginActivations(rows pgx.Rows) ([]*model.PluginActivation, error) {
	var out []*model.PluginActivation
	for rows.Next() {
		act, err := scanPluginActivation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, act)
	}
	return out, rows.Err()
}

func scanPluginStorage(row pluginRow) (*model.PluginStorageEntry, error) {
	var entry model.PluginStorageEntry
	if err := row.Scan(&entry.PluginID, &entry.ScopeID, &entry.Key, &entry.Value, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
		return nil, err
	}
	return &entry, nil
}
