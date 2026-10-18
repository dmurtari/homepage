package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/driver"
)

//go:embed migrations/*
var migrations embed.FS

func Open(path string, namespace string, logger *slog.Logger) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign_keys: %w", err)
	}

	migrator := NewMigrator(db, migrations, logger)
	if err := migrator.Migrate(namespace); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

type Migrator struct {
	db     *sql.DB
	logger *slog.Logger
	fs     embed.FS
}

func NewMigrator(db *sql.DB, fs embed.FS, logger *slog.Logger) *Migrator {
	return &Migrator{db: db, fs: fs, logger: logger}
}

func (m *Migrator) Migrate(namespace string) error {
	dir := "migrations/" + namespace

	entries, err := m.fs.ReadDir(dir)
	if err != nil {
		m.logger.Info("no migrations found", "namespace", namespace)
		return nil
	}

	var version int
	if err := m.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}

	m.logger.Info("migrating", "namespace", namespace, "from_version", version, "to_version", len(entries))

	for i := version; i < len(entries); i++ {
		name := entries[i].Name()

		q, err := m.fs.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}

		tx, err := m.db.Begin()
		if err != nil {
			return err
		}

		if _, err := tx.Exec(string(q)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s.%s: %w", namespace, name, err)
		}

		if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", i+1)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update user_version: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s.%s: %w", namespace, name, err)
		}
	}

	m.logger.Info("migration completed", "namespace", namespace, "to_version", len(entries))

	return nil
}
