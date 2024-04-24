package sync_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/seborama/pcloud-sdk/sync"
	"github.com/seborama/pcloud-sdk/tracker/db"
)

func TestOneWay_Sync_Created(t *testing.T) {
	ctx := context.Background()

	expectedFSMutations := db.FSMutations{
		{
			Type: db.MutationTypeCreated,
			Details: db.EntryMutations{
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "left",
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
			},
		},
		{
			Type: db.MutationTypeCreated,
			Details: db.EntryMutations{
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "left",
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
			},
		},
		{
			Type: db.MutationTypeCreated,
			Details: db.EntryMutations{
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "left",
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
			},
		},
	}

	dataCh := make(chan []byte)
	errCh := make(chan error)

	fsEntryMatcher := mock.MatchedBy(
		func(fsEntry db.FSEntry) bool {
			return assert.EqualValues(t, 100201, fsEntry.EntryID)
		},
	)

	pCloudFS := MockPCloudFileSystem{}
	defer func() { _ = pCloudFS.AssertExpectations(t) }()
	pCloudFS.
		On("StreamFileData", mock.Anything, fsEntryMatcher).
		Return(
			func() <-chan []byte {
				go func() {
					defer close(dataCh)
					dataCh <- []byte("Hello")
				}()
				return dataCh
			}(),
			func() <-chan error {
				go func() {
					defer close(errCh)
					errCh <- nil
				}()
				return errCh
			}(),
		).
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
				return assert.ElementsMatch(t, []byte("Hello"), data)
			})
	}()

	localClient := MockLocalFileSystem{}
	defer func() { _ = localClient.AssertExpectations(t) }()
	localClient.
		On("MkDir", ctx, "/Folder1").
		Return(nil).
		Once().
		On("MkDir", ctx, "/Folder2").
		Return(nil).
		Once().
		On("MkFile", mock.Anything, "/Folder2/File2-1", channelMatcher).
		Return(nil).
		Once()

	fsTracker := MockFSTracker{}
	defer func() { _ = fsTracker.AssertExpectations(t) }()
	fsTracker.
		On("ListMutations", ctx).
		Return(expectedFSMutations, nil).
		Once()

	s := sync.NewOneWay(&pCloudFS, &localClient, &fsTracker)
	err := s.Sync(ctx)
	require.NoError(t, err)
}

func TestOneWay_Sync_Deleted(t *testing.T) {
	ctx := context.Background()

	expectedFSMutations := db.FSMutations{
		{
			Type: db.MutationTypeDeleted,
			Details: db.EntryMutations{
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "left",
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
			},
		},
		{
			Type: db.MutationTypeDeleted,
			Details: db.EntryMutations{
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "left",
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
			},
		},
		{
			Type: db.MutationTypeDeleted,
			Details: db.EntryMutations{
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "left",
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
			},
		},
	}

	pCloudFS := MockPCloudFileSystem{}
	defer func() { _ = pCloudFS.AssertExpectations(t) }()

	localClient := MockLocalFileSystem{}
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
		On("ListMutations", ctx).
		Return(expectedFSMutations, nil).
		Once()

	s := sync.NewOneWay(&pCloudFS, &localClient, &fsTracker)
	err := s.Sync(ctx)
	require.NoError(t, err)
}

func TestOneWay_Sync_Modified(t *testing.T) {
	ctx := context.Background()

	expectedFSMutations := db.FSMutations{
		{
			Type: db.MutationTypeModified,
			Details: db.EntryMutations{
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "left",
						DeviceID:       "dev-id",
						EntryID:        100201,
						IsFolder:       false,
						Path:           "/Folder2",
						Name:           "File2-1",
						ParentFolderID: 1002,
						Created:        time.Time{},
						Modified:       time.Time{}.Add(time.Hour),
						Size:           100123,
						Hash:           "file2-1-hash-2",
					},
				},
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "local",
						DeviceID:       "local-dev-id",
						EntryID:        100100201,
						IsFolder:       false,
						Path:           "/Folder2",
						Name:           "File2-1",
						ParentFolderID: 1001002,
						Created:        time.Time{},
						Modified:       time.Time{},
						Size:           100,
						Hash:           "file2-1-hash",
					},
				},
			},
		},
	}

	dataCh := make(chan []byte)
	errCh := make(chan error)

	fsEntryMatcher := mock.MatchedBy(
		func(fsEntry db.FSEntry) bool {
			return assert.EqualValues(t, 100201, fsEntry.EntryID)
		},
	)

	pCloudFS := MockPCloudFileSystem{}
	defer func() { _ = pCloudFS.AssertExpectations(t) }()
	pCloudFS.
		On("StreamFileData", mock.Anything, fsEntryMatcher).
		Return(
			func() <-chan []byte {
				go func() {
					defer close(dataCh)
					dataCh <- []byte("Hello")
				}()
				return dataCh
			}(),
			func() <-chan error {
				go func() {
					defer close(errCh)
					errCh <- nil
				}()
				return errCh
			}(),
		).
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

	localClient := MockLocalFileSystem{}
	defer func() { _ = localClient.AssertExpectations(t) }()
	localClient.
		On("MkFile", mock.Anything, "/Folder2/File2-1", channelMatcher).
		Return(nil).
		Once()

	fsTracker := MockFSTracker{}
	defer func() { _ = fsTracker.AssertExpectations(t) }()
	fsTracker.
		On("ListMutations", ctx).
		Return(expectedFSMutations, nil).
		Once()

	s := sync.NewOneWay(&pCloudFS, &localClient, &fsTracker)
	err := s.Sync(ctx)
	require.NoError(t, err)
}

func TestOneWay_Sync_Moved(t *testing.T) {
	ctx := context.Background()

	expectedFSMutations := db.FSMutations{
		{
			Type: db.MutationTypeMoved,
			Details: db.EntryMutations{
				{
					Version: db.VersionPrevious,
					FSEntry: db.FSEntry{
						FSName:         "left",
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
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "left",
						DeviceID:       "dev-id",
						EntryID:        1001,
						IsFolder:       true,
						Path:           "/",
						Name:           "MovedFolder1",
						ParentFolderID: 1000,
						Created:        time.Time{},
						Modified:       time.Time{},
						Size:           0,
						Hash:           "",
					},
				},
			},
		},
		{
			Type: db.MutationTypeMoved,
			Details: db.EntryMutations{
				{
					Version: db.VersionPrevious,
					FSEntry: db.FSEntry{
						FSName:         "left",
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
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "left",
						DeviceID:       "dev-id",
						EntryID:        1002,
						IsFolder:       true,
						Path:           "/Moved",
						Name:           "MovedFolder2",
						ParentFolderID: 1000,
						Created:        time.Time{},
						Modified:       time.Time{},
						Size:           0,
						Hash:           "",
					},
				},
			},
		},
		{
			Type: db.MutationTypeMoved,
			Details: db.EntryMutations{
				{
					Version: db.VersionPrevious,
					FSEntry: db.FSEntry{
						FSName:         "left",
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
				{
					Version: db.VersionNew,
					FSEntry: db.FSEntry{
						FSName:         "left",
						DeviceID:       "dev-id",
						EntryID:        100201,
						IsFolder:       false,
						Path:           "/Moved/MovedFolder2",
						Name:           "MovedFile2-1",
						ParentFolderID: 1002,
						Created:        time.Time{},
						Modified:       time.Time{},
						Size:           100,
						Hash:           "file2-1-hash",
					},
				},
			},
		},
	}

	pCloudFS := MockPCloudFileSystem{}
	defer func() { _ = pCloudFS.AssertExpectations(t) }()

	localClient := MockLocalFileSystem{}
	defer func() { _ = localClient.AssertExpectations(t) }()
	localClient.
		On("MvFile", ctx, "/Folder2/File2-1", "/Moved/MovedFolder2/MovedFile2-1").
		Return(nil).
		Once().
		On("MvDir", ctx, "/Folder1", "/MovedFolder1").
		Return(nil).
		Once().
		On("MvDir", ctx, "/Folder2", "/Moved/MovedFolder2").
		Return(nil).
		Once()

	fsTracker := MockFSTracker{}
	defer func() { _ = fsTracker.AssertExpectations(t) }()
	fsTracker.
		On("ListMutations", ctx).
		Return(expectedFSMutations, nil).
		Once()

	s := sync.NewOneWay(&pCloudFS, &localClient, &fsTracker)
	err := s.Sync(ctx)
	require.NoError(t, err)
}

func TestOneWay_Sync_MutatedAndMoved(t *testing.T) {
	t.Skip("NOT YET WRITTEN")
}

type MockPCloudFileSystem struct {
	mock.Mock
}

func (m *MockPCloudFileSystem) StreamFileData(ctx context.Context, fsEntry db.FSEntry) (<-chan []byte, <-chan error) {
	args := m.Called(ctx, fsEntry)
	return args.Get(0).(<-chan []byte), args.Get(1).(<-chan error)
}

type MockLocalFileSystem struct {
	mock.Mock
}

func (m *MockLocalFileSystem) MkDir(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockLocalFileSystem) MkFile(ctx context.Context, path string, dataCh <-chan []byte) error {
	args := m.Called(ctx, path, dataCh)
	return args.Error(0)
}

func (m *MockLocalFileSystem) RmDir(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockLocalFileSystem) RmFile(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockLocalFileSystem) MvDir(ctx context.Context, fromPath, toPath string) error {
	args := m.Called(ctx, fromPath, toPath)
	return args.Error(0)
}

func (m *MockLocalFileSystem) MvFile(ctx context.Context, fromPath, toPath string) error {
	args := m.Called(ctx, fromPath, toPath)
	return args.Error(0)
}

type MockFSTracker struct {
	mock.Mock
}

func (m *MockFSTracker) ListMutations(ctx context.Context) (db.FSMutations, error) {
	args := m.Called(ctx)
	return args.Get(0).(db.FSMutations), args.Error(1)
}
