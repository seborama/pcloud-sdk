package db

import (
	"context"
	"database/sql"
	"fmt"
	"seborama/pcloud/tracker/db/migrations"

	// sqllite3 sql driver.
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{
		db: db,
	}
}

func (m *Migrator) MigrateUp(ctx context.Context) (err error) {
	stmt := `
		BEGIN;
		CREATE TABLE IF NOT EXISTS "schema_migrations" ( "version" INTEGER PRIMARY KEY, "status" VARCHAR );
		COMMIT;`

	_, err = m.db.ExecContext(ctx, stmt)
	if err != nil {
		return errors.WithStack(err)
	}

	version := 0
	var status migrationStatus

	err = m.db.QueryRowContext(ctx, `SELECT version, status FROM "schema_migrations"`).Scan(&version, &status)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.WithStack(err)
	}

	if status == migrationStatusInProgress {
		return errors.Errorf("unable to apply database migration: migration version %d is recorded 'in progress', probably as the result of a previous failure", version)
	}

	if version == 0 {
		version = -1
	}
	fmt.Printf("current migrations version: %d\n", version)

	if version == len(migrations.SQLite3) {
		fmt.Printf("migrations up-to-date - version: %d\n", version)
		return nil
	}

	if version > len(migrations.SQLite3) {
		return errors.Errorf("migrations corruption - last executed version: %d - highest available migration version: %d\n", version, len(migrations.SQLite3))
	}

	for version, stmt := range migrations.SQLite3[version+1:] {
		fmt.Printf("applying migrations: %d\n", version)

		err = m.recordMigrationVersion(ctx, version, migrationStatusInProgress)
		if err != nil {
			return err
		}

		_, err = m.db.ExecContext(ctx, stmt)
		if err != nil {
			return errors.WithStack(err)
		}

		err = m.recordMigrationVersion(ctx, version, migrationStatusApplied)
		if err != nil {
			return err
		}
	}

	return nil
}

type migrationStatus string

const (
	migrationStatusInProgress migrationStatus = "in progress"
	migrationStatusApplied    migrationStatus = "applied"
)

func (m *Migrator) recordMigrationVersion(ctx context.Context, version int, status migrationStatus) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM "schema_migrations"`)
	if err != nil {
		return doRollback(tx, err)
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO "schema_migrations" VALUES (?, ?)`, version, status)
	if err != nil {
		return doRollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		return doRollback(tx, err)
	}

	return nil
}
