package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func wipe(ctx context.Context, d *deps, queriesDir string) error {
	d.logger.Info(ctx, "wiping neo4j")
	if err := repository.Neo4jExecuteWriteAndConsume(
		ctx,
		d.graphDB,
		"MATCH (n) DETACH DELETE n",
		nil,
	); err != nil {
		return fmt.Errorf("wipe neo4j: %w", err)
	}

	cypherPath := filepath.Join(queriesDir, "bootstrap.cypher")
	d.logger.Info(ctx, "applying neo4j bootstrap")
	if err := applyCypherFile(ctx, d.graphDB, cypherPath); err != nil {
		return err
	}

	d.logger.Info(ctx, "wiping meilisearch")
	if err := d.searchRepo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("wipe meilisearch: %w", err)
	}
	if err := d.searchRepo.EnsureIndex(ctx); err != nil {
		return fmt.Errorf("ensure search index: %w", err)
	}

	d.logger.Info(ctx, "flushing redis")
	if err := d.cacheDB.Client().FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("flush redis: %w", err)
	}

	sqlPath := filepath.Join(queriesDir, "bootstrap.sql")
	d.logger.Info(ctx, "applying postgres bootstrap")
	if err := applySQLFile(ctx, d.relDB, sqlPath); err != nil {
		return err
	}
	if _, err := d.relDB.Pool().Exec(
		ctx,
		"TRUNCATE TABLE user_tokens, notifications RESTART IDENTITY CASCADE",
	); err != nil {
		return fmt.Errorf("truncate postgres tokens: %w", err)
	}
	if _, err := d.relDB.Pool().Exec(
		ctx,
		"TRUNCATE TABLE plugin_storage, plugin_activations, plugin_installations",
	); err != nil {
		return fmt.Errorf("truncate plugin installations: %w", err)
	}
	d.logger.Info(ctx, "wiping plugin packages")
	if err := wipePluginPackages(d.cfg.Plugin.Directory); err != nil {
		return err
	}

	return nil
}

func wipePluginPackages(dir string) error {
	if dir == "" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read plugin directory: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		root := filepath.Join(dir, entry.Name())
		_, err := os.Stat(filepath.Join(root, model.PluginManifestFileName))
		if err == nil {
			continue
		}
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat plugin manifest: %w", err)
		}
		if err := os.RemoveAll(root); err != nil {
			return fmt.Errorf("remove plugin install %s: %w", root, err)
		}
	}
	return nil
}

func applyCypherFile(ctx context.Context, db *repository.Neo4jDatabase, path string) error {
	script, err := os.ReadFile(path) // #nosec G304 -- operator-supplied queries directory
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	for _, statement := range splitStatements(string(script)) {
		if err := repository.Neo4jExecuteWriteAndConsume(ctx, db, statement, nil); err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "EquivalentSchemaRuleAlreadyExists") ||
				strings.Contains(errStr, "ConstraintCreationFailed") ||
				strings.Contains(errStr, "already exists") {
				continue
			}
			return fmt.Errorf("cypher %s: %w", statement, err)
		}
	}
	return nil
}

func applySQLFile(ctx context.Context, db *repository.PGDatabase, path string) error {
	script, err := os.ReadFile(path) // #nosec G304 -- operator-supplied queries directory
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	for _, statement := range splitStatements(string(script)) {
		if _, err := db.Pool().Exec(ctx, statement); err != nil {
			return fmt.Errorf("sql %s: %w", statement, err)
		}
	}
	return nil
}

func splitStatements(script string) []string {
	raw := strings.Split(script, ";")
	out := make([]string, 0, len(raw))
	for _, statement := range raw {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		out = append(out, statement)
	}
	return out
}
