package sync_test

import (
	"context"
	"testing"
	"time"

	"seborama/pcloud/sync"
	"seborama/pcloud/tracker/db"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRightSync_Created(t *testing.T) {
	ctx := context.Background()

	leftMutations := db.FSMutations{
		{
			Type:    "123",
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         "type",
				DeviceID:       "dev-id",
				EntryID:        1001,
				IsFolder:       true,
				Path:           "/",
				Name:           "Folder1",
				ParentFolderID: 1000,
				Created:        time.Time{},
				Modified:       time.Time{},
				Size:           0,
				Hash:           "",
			},
		},
		{
			Type:    "123",
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         "type",
				DeviceID:       "dev-id",
				EntryID:        1002,
				IsFolder:       true,
				Path:           "/",
				Name:           "Folder2",
				ParentFolderID: 1000,
				Created:        time.Time{},
				Modified:       time.Time{},
				Size:           0,
				Hash:           "",
			},
		},
		{
			Type:    "123",
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         "type",
				DeviceID:       "dev-id",
				EntryID:        100201,
				IsFolder:       true,
				Path:           "/Folder2",
				Name:           "File2-1",
				ParentFolderID: 1002,
				Created:        time.Time{},
				Modified:       time.Time{},
				Size:           100,
				Hash:           "file2-1-hash",
			},
		},
	}
	rightMutations := db.FSMutations{}

	pCloudClient := MockSDKClient{}
	pCloudClient.
		On("").
		Return().
		Once()

	s := sync.NewSync(&pCloudClient)
	err := s.Right(ctx, leftMutations, rightMutations)
	require.NoError(t, err)
}

func TestRightSync_Deleted(t *testing.T) {}

func TestRightSync_Modified(t *testing.T) {}

func TestRightSync_Moved(t *testing.T) {}

func TestRightSync_MutatedAndMoved(t *testing.T) {}

type MockSDKClient struct {
	mock.Mock
}
