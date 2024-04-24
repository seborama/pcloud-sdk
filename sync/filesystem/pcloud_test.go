package filesystem_test

import (
	"context"
	"io"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/seborama/pcloud-sdk/sdk"
	"github.com/seborama/pcloud-sdk/sync/filesystem"
	"github.com/seborama/pcloud-sdk/tracker/db"
)

func TestPCloud_StreamFileData_MkFile(t *testing.T) {
	ctx := context.Background()

	data := []byte("Hello")

	fileIDMatcher := func(f func(q url.Values)) bool {
		q := url.Values{}
		f(q)
		return assert.Equal(t, "123", q.Get("fileid"))
	}

	pCloudSDK1 := &mockPCloudSDK{}
	defer func() { _ = pCloudSDK1.AssertExpectations(t) }()
	pCloudSDK1.
		On("FileOpen", ctx, uint64(0), mock.MatchedBy(fileIDMatcher), []sdk.ClientOption(nil)).
		Return(&sdk.File{FD: 124816}, nil).
		Once().
		On("FileRead", ctx, uint64(124816), uint64(1_048_576), []sdk.ClientOption(nil)).
		Return(data, io.EOF).
		Once().
		On("FileClose", ctx, uint64(124816), []sdk.ClientOption(nil)).
		Return(nil).
		Once()

	fileByPathMatcher := func(f func(q url.Values)) bool {
		q := url.Values{}
		f(q)
		return assert.Equal(t, "somewhere", q.Get("path"))
	}

	pCloudSDK2 := &mockPCloudSDK{}
	defer func() { _ = pCloudSDK2.AssertExpectations(t) }()
	pCloudSDK2.
		On("FileOpen", ctx, uint64(sdk.O_CREAT|sdk.O_TRUNC), mock.MatchedBy(fileByPathMatcher), []sdk.ClientOption(nil)).
		Return(&sdk.File{FD: 321684}, nil).
		Once().
		On("FileWrite", ctx, uint64(321684), data, []sdk.ClientOption(nil)).
		Return(&sdk.FileDataTransfer{Bytes: uint64(len(data))}, nil).
		Once().
		On("FileClose", ctx, uint64(321684), []sdk.ClientOption(nil)).
		Return(nil).
		Once()

	u1 := filesystem.NewPCloud(pCloudSDK1)
	u2 := filesystem.NewPCloud(pCloudSDK2)

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

	dataCh, errCh := u1.StreamFileData(ctx, fsEntry)

	err := u2.MkFile(ctx, "somewhere", dataCh)
	require.NoError(t, err)

	require.NoError(t, <-errCh)
}

type mockPCloudSDK struct {
	mock.Mock
}

func (m *mockPCloudSDK) FileOpen(ctx context.Context, flags uint64, file sdk.T4PathOrFileIDOrFolderIDName, opts ...sdk.ClientOption) (*sdk.File, error) {
	args := m.Called(ctx, flags, file, opts)
	return args.Get(0).(*sdk.File), args.Error(1)
}

func (m *mockPCloudSDK) FileClose(ctx context.Context, fd uint64, opts ...sdk.ClientOption) error {
	args := m.Called(ctx, fd, opts)
	return args.Error(0)
}

func (m *mockPCloudSDK) FileRead(ctx context.Context, fd, count uint64, opts ...sdk.ClientOption) ([]byte, error) {
	args := m.Called(ctx, fd, count, opts)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockPCloudSDK) FileWrite(ctx context.Context, fd uint64, data []byte, opts ...sdk.ClientOption) (*sdk.FileDataTransfer, error) {
	args := m.Called(ctx, fd, data, opts)
	return args.Get(0).(*sdk.FileDataTransfer), args.Error(1)
}
