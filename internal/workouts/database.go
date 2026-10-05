package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	_ "github.com/ncruces/go-sqlite3/driver"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Open(path string, logger *slog.Logger) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	err = migrate(db, migrations)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB, migrations embed.FS) error {
	mig, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}

	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}

	for i := version; i < len(mig); i++ {
		q, err := migrations.ReadFile("migrations/" + mig[i].Name())
		if err != nil {
			return err
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		if _, err := tx.Exec(string(q)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", mig[i].Name(), err)
		}

		if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", i+1)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update user_version: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", mig[i].Name(), err)
		}
	}

	return nil
}
