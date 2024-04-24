package filesystem_test

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/seborama/pcloud-sdk/sync/filesystem"
	"github.com/seborama/pcloud-sdk/tracker/db"
)

func TestUnix_StreamFileData_MkFile(t *testing.T) {
	rc := &mockReadCloser{}
	defer func() { _ = rc.AssertExpectations(t) }()
	rc.
		On("Read", mock.Anything).
		Return(5, nil).
		Once().
		On("Read", mock.Anything).
		Return(0, io.EOF).
		Once().
		On("Close").
		Return(nil).
		Once()

	wc := &mockWriteCloser{}
	defer func() { _ = wc.AssertExpectations(t) }()
	wc.
		On("Write", mock.MatchedBy(func(p []byte) bool {
			return assert.Equal(t, []byte("Hello"), p)
		})).
		Return(5, nil).
		Once().
		On("Close").
		Return(nil).
		Once()

	fso := &mockFSOperations{}
	defer func() { _ = fso.AssertExpectations(t) }()
	fso.
		On("OpenFile", "somewhere", os.O_CREATE|os.O_TRUNC, mock.MatchedBy(func(perm os.FileMode) bool {
			return assert.EqualValues(t, 0640, perm)
		})).
		Return(wc, nil).
		Once().
		On("Open", "unix_test.go").
		Return(rc, nil).
		Once()

	u := filesystem.NewUnix(fso)

	ctx := context.Background()

	fsEntry := db.FSEntry{
		FSName:         db.LocalFileSystem,
		DeviceID:       "dev-id-1",
		EntryID:        123,
		IsFolder:       false,
		Path:           ".",
		Name:           "unix_test.go",
		ParentFolderID: 456,
		Created:        time.Time{},
		Modified:       time.Time{},
		Size:           123,
		Hash:           "hash",
	}

	dataCh, errCh := u.StreamFileData(ctx, fsEntry)

	err := u.MkFile(ctx, "somewhere", dataCh)
	require.NoError(t, err)

	require.NoError(t, <-errCh)
}

type mockFSOperations struct {
	mock.Mock
}

func (m *mockFSOperations) Open(name string) (io.ReadCloser, error) {
	args := m.Called(name)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *mockFSOperations) MkdirAll(path string, perm os.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

func (m *mockFSOperations) OpenFile(name string, flag int, perm os.FileMode) (io.WriteCloser, error) {
	args := m.Called(name, flag, perm)
	return args.Get(0).(io.WriteCloser), args.Error(1)
}

type mockReadCloser struct {
	mock.Mock
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	copy(p, "Hello")
	return args.Int(0), args.Error(1)
}

func (m *mockReadCloser) Close() error {
	args := m.Called()
	return args.Error(0)
}

type mockWriteCloser struct {
	mock.Mock
}

func (m *mockWriteCloser) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockWriteCloser) Close() error {
	args := m.Called()
	return args.Error(0)
}
