package sync_test

import (
	"context"
	"io"
	"net/url"
	"testing"
	"time"

	"seborama/pcloud/sdk"
	"seborama/pcloud/sync"
	"seborama/pcloud/tracker/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRightSync_Created(t *testing.T) {
	ctx := context.Background()

	expectedFSMutations := db.FSMutations{
		{
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         db.PCloudFileSystem,
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
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         db.PCloudFileSystem,
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
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         db.PCloudFileSystem,
				DeviceID:       "dev-id",
				EntryID:        100201,
				IsFolder:       false,
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

	fileMatcher := mock.MatchedBy(
		func(p func(q url.Values)) bool {
			q := url.Values{}
			p(q)
			return assert.Equal(t, "100201", q.Get("fileid"))
		})

	pCloudClient := MockSDKClient{}
	defer func() { _ = pCloudClient.AssertExpectations(t) }()
	pCloudClient.
		On("FileOpen", ctx, uint64(0), fileMatcher, []sdk.ClientOption(nil)).
		Return(&sdk.File{
			FD:     123,
			FileID: 100201,
		}, nil).
		Once().
		On("FileRead", ctx, uint64(123), []sdk.ClientOption(nil)).
		Return([]byte("Hello"), io.EOF).
		Once().
		On("FileClose", ctx, uint64(123), []sdk.ClientOption(nil)).
		Return(nil).
		Once()

	channelMatcher := func() interface{} {
		count := 0
		return mock.MatchedBy(
			func(dataCh <-chan []byte) bool {
				count++
				if count >= 1 {
					return true
				}
				data := <-dataCh
				return assert.Equal(t, []byte("Hello"), data)
			})
	}()

	localClient := MockLocalClient{}
	defer func() { _ = localClient.AssertExpectations(t) }()
	localClient.
		On("MkDir", ctx, "/Folder1").
		Return(nil).
		Once().
		On("MkDir", ctx, "/Folder2").
		Return(nil).
		Once().
		On("MkFile", ctx, "/Folder2/File2-1", channelMatcher).
		Return(nil).
		Once()

	fsTracker := MockFSTracker{}
	defer func() { _ = fsTracker.AssertExpectations(t) }()
	fsTracker.
		On("FindPCloudVsLocalMutations", ctx).
		Return(expectedFSMutations, nil).
		Once()

	s := sync.NewSync(&fsTracker, &pCloudClient, &localClient)
	err := s.Right(ctx)
	require.NoError(t, err)
}

func TestRightSync_Deleted(t *testing.T) {
	ctx := context.Background()

	expectedFSMutations := db.FSMutations{
		{
			Type:    db.MutationTypeDeleted,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         db.PCloudFileSystem,
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
			Type:    db.MutationTypeDeleted,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         db.PCloudFileSystem,
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
			Type:    db.MutationTypeDeleted,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         db.PCloudFileSystem,
				DeviceID:       "dev-id",
				EntryID:        100201,
				IsFolder:       false,
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

	pCloudClient := MockSDKClient{}
	defer func() { _ = pCloudClient.AssertExpectations(t) }()

	localClient := MockLocalClient{}
	defer func() { _ = localClient.AssertExpectations(t) }()
	localClient.
		On("RmFile", ctx, "/Folder2/File2-1").
		Return(nil).
		Once().
		On("RmDir", ctx, "/Folder1").
		Return(nil).
		Once().
		On("RmDir", ctx, "/Folder2").
		Return(nil).
		Once()

	fsTracker := MockFSTracker{}
	defer func() { _ = fsTracker.AssertExpectations(t) }()
	fsTracker.
		On("FindPCloudVsLocalMutations", ctx).
		Return(expectedFSMutations, nil).
		Once()

	s := sync.NewSync(&fsTracker, &pCloudClient, &localClient)
	err := s.Right(ctx)
	require.NoError(t, err)
}

func TestRightSync_Modified(t *testing.T) {
	ctx := context.Background()

	expectedFSMutations := db.FSMutations{
		{
			Type:    db.MutationTypeCreated,
			Version: db.VersionNew,
			FSEntry: db.FSEntry{
				FSType:         db.PCloudFileSystem,
				DeviceID:       "dev-id",
				EntryID:        100201,
				IsFolder:       false,
				Path:           "/Folder2",
				Name:           "File2-1",
				ParentFolderID: 1002,
				Created:        time.Time{},
				Modified:       time.Time{},
				Size:           100,
				Hash:           "f7ff9e8b7bb2e09b70935a5d785e0cc5d9d0abf0", // SHA1 of "Hello"
			},
		},
	}

	fileMatcher := mock.MatchedBy(
		func(p func(q url.Values)) bool {
			q := url.Values{}
			p(q)
			return assert.Equal(t, "100201", q.Get("fileid"))
		})

	pCloudClient := MockSDKClient{}
	defer func() { _ = pCloudClient.AssertExpectations(t) }()
	pCloudClient.
		On("FileOpen", ctx, uint64(0), fileMatcher, []sdk.ClientOption(nil)).
		Return(&sdk.File{
			FD:     123,
			FileID: 100201,
		}, nil).
		Once().
		On("FileRead", ctx, uint64(123), []sdk.ClientOption(nil)).
		Return([]byte("Hello"), io.EOF).
		Once().
		On("FileClose", ctx, uint64(123), []sdk.ClientOption(nil)).
		Return(nil).
		Once()

	channelMatcher := func() interface{} {
		count := 0
		return mock.MatchedBy(
			func(dataCh <-chan []byte) bool {
				count++
				if count >= 1 {
					return true
				}
				data := <-dataCh
				return assert.Equal(t, []byte("Hello"), data)
			})
	}()

	localClient := MockLocalClient{}
	defer func() { _ = localClient.AssertExpectations(t) }()
	localClient.
		On("MkFile", ctx, "/Folder2/File2-1", channelMatcher).
		Return(nil).
		Once()

	fsTracker := MockFSTracker{}
	defer func() { _ = fsTracker.AssertExpectations(t) }()
	fsTracker.
		On("FindPCloudVsLocalMutations", ctx).
		Return(expectedFSMutations, nil).
		Once()

	s := sync.NewSync(&fsTracker, &pCloudClient, &localClient)
	err := s.Right(ctx)
	require.NoError(t, err)
}

func TestRightSync_Moved(t *testing.T) {}

func TestRightSync_MutatedAndMoved(t *testing.T) {}

type MockSDKClient struct {
	mock.Mock
}

func (m *MockSDKClient) FileOpen(ctx context.Context, flags uint64, file sdk.T4PathOrFileIDOrFolderIDName, opts ...sdk.ClientOption) (*sdk.File, error) {
	args := m.Called(ctx, flags, file, opts)
	return args.Get(0).(*sdk.File), args.Error(1)
}

func (m *MockSDKClient) FileClose(ctx context.Context, fd uint64, opts ...sdk.ClientOption) error {
	args := m.Called(ctx, fd, opts)
	return args.Error(0)
}

func (m *MockSDKClient) FileRead(ctx context.Context, fd, count uint64, opts ...sdk.ClientOption) ([]byte, error) {
	args := m.Called(ctx, fd, opts)
	return args.Get(0).([]byte), args.Error(1)
}

type MockLocalClient struct {
	mock.Mock
}

func (m *MockLocalClient) MkDir(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockLocalClient) MkFile(ctx context.Context, path string, dataCh <-chan []byte) error {
	args := m.Called(ctx, path, dataCh)
	return args.Error(0)
}

func (m *MockLocalClient) RmDir(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockLocalClient) RmFile(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

type MockFSTracker struct {
	mock.Mock
}

func (m *MockFSTracker) FindPCloudVsLocalMutations(ctx context.Context) (db.FSMutations, error) {
	args := m.Called(ctx)
	return args.Get(0).(db.FSMutations), args.Error(1)
}
