package tracker

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/seborama/pcloud/tracker/db"
)

func TestTracker_rotateFileSystemVersions(t *testing.T) {
	t.Skip("mock the 'store' and write a separate db test, if one does not already exist")
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
}
