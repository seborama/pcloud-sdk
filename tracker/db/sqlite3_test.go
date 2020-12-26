package db_test

import (
	"context"
	"os"
	"seborama/pcloud/tracker/db"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestSQLite3_MigrationsSuccess(t *testing.T) {
	const dbPath = "/tmp/data_schema_migrations_test"
	ctx := context.Background()

	err := os.RemoveAll(dbPath)
	require.NoError(t, err)

	err = os.MkdirAll(dbPath, 0700)
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(dbPath) }()

	_, err = db.NewSQLite3(ctx, dbPath)
	require.NoError(t, err)
}
