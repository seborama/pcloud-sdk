package tracker

import (
	"context"
	"testing"

	"github.com/seborama/pcloud/tracker/db"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTracker_rotateFileSystemVersions_FSNotChanged(t *testing.T) {
	t.Log("mock the 'store' and write a separate db test, if one does not already exist")
	ctx := context.Background()

	const fsName db.FSName = "some_fs"

	fsInfo := &db.FSInfo{
		FSName:    fsName,
		FSDriver:  db.FSDriverLocal,
		FSRoot:    "/tmp",
		FSChanged: false,
	}

	sqlDB := &StorerMock{}
	defer sqlDB.AssertExpectations(t)

	sqlDB.On("GetFileSystemInfo", ctx, fsName).
		Return(fsInfo, nil).
		Once().
		On("RotateFileSystemVersions", ctx, fsName).
		Return(nil).
		Once()

	tr := &Tracker{
		logger:   zap.NewNop(),
		store:    sqlDB,
		fsDriver: nil,
		fsName:   fsName,
	}

	err := tr.rotateFileSystemVersions(ctx)
	require.NoError(t, err)
}

func TestTracker_rotateFileSystemVersions_FSChanged(t *testing.T) {
	t.Log("mock the 'store' and write a separate db test, if one does not already exist")
	ctx := context.Background()

	const fsName db.FSName = "some_fs"

	fsInfo := &db.FSInfo{
		FSName:    fsName,
		FSDriver:  db.FSDriverLocal,
		FSRoot:    "/tmp",
		FSChanged: true,
	}

	sqlDB := &StorerMock{}
	defer sqlDB.AssertExpectations(t)

	sqlDB.On("GetFileSystemInfo", ctx, fsName).
		Return(fsInfo, nil).
		Once().
		On("DeleteVersionNew", ctx, fsName).
		Return(nil).
		Once()

	tr := &Tracker{
		logger:   zap.NewNop(),
		store:    sqlDB,
		fsDriver: nil,
		fsName:   fsName,
	}

	err := tr.rotateFileSystemVersions(ctx)
	require.NoError(t, err)
}
