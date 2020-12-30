package tracker

import (
	"context"
	"database/sql"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"seborama/pcloud/tracker/db"
)

func TestHashFileData(t *testing.T) {
	const dbPath = "/tmp/data_test_hashFileData"

	err := os.MkdirAll(dbPath, 0700)
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(dbPath) }()

	fName := uuid.New().String()
	err = ioutil.WriteFile(filepath.Join(dbPath, fName), []byte("This is File000"), 0600)
	require.NoError(t, err)

	h, err := hashFileData(filepath.Join(dbPath, fName))
	require.NoError(t, err)
	assert.Equal(t, "01ce643e7c1ca98f6fb21e61b5d03f547813edae", h)
}

func TestMarkSyncAsRequired(t *testing.T) {
	const dbPath = "/tmp/data_test_markSyncAsRequired"
	ctx := context.Background()

	err := os.MkdirAll(dbPath, 0700)
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(dbPath) }()

	sqlDB, err := db.NewSQLite3(ctx, dbPath)
	require.NoError(t, err)

	tr := &Tracker{
		store: sqlDB,
	}

	err = tr.markSyncAsRequired(ctx, db.LocalFileSystem)
	require.EqualError(t, err, sql.ErrNoRows.Error())

	err = tr.initSyncStatus(ctx)
	require.NoError(t, err)

	err = tr.markSyncAsRequired(ctx, db.LocalFileSystem)
	require.NoError(t, err)

	status, err := sqlDB.GetSyncStatus(ctx, db.LocalFileSystem)
	require.NoError(t, err)
	assert.Equal(t, db.SyncStatusRequired, status)

	err = tr.markSyncAsRequired(ctx, db.LocalFileSystem)
	require.EqualError(t, err, "cannot transition sync status from 'Required' to 'Required'")

	err = sqlDB.MarkSyncInProgress(ctx, db.LocalFileSystem)
	require.NoError(t, err)

	err = tr.markSyncAsRequired(ctx, db.LocalFileSystem)
	require.EqualError(t, err, "cannot transition sync status from 'In progress' to 'Required'")
}

func TestMarkNewFileSystemEntriesAsPrevious(t *testing.T) {
	const dbPath = "/tmp/data_test_markNewFileSystemEntriesAsPrevious"
	ctx := context.Background()

	err := os.MkdirAll(dbPath, 0700)
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(dbPath) }()

	sqlDB, err := db.NewSQLite3(ctx, dbPath)
	require.NoError(t, err)

	tr := &Tracker{
		store: sqlDB,
	}

	err = tr.markNewFileSystemEntriesAsPrevious(ctx, db.LocalFileSystem)
	require.EqualError(t, err, "database error or sync has not been initialised: "+sql.ErrNoRows.Error())

	err = tr.initSyncStatus(ctx)
	require.NoError(t, err)

	err = tr.markNewFileSystemEntriesAsPrevious(ctx, db.LocalFileSystem)
	require.NoError(t, err)

	err = tr.markSyncAsRequired(ctx, db.LocalFileSystem)
	require.NoError(t, err)

	err = tr.markNewFileSystemEntriesAsPrevious(ctx, db.LocalFileSystem)
	require.EqualError(t, err, "markNewFileSystemEntriesAsPrevious requires sync status 'Complete' but status is currently 'Required'")
}
