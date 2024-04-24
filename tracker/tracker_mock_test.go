package tracker

import (
	"context"

	"github.com/seborama/pcloud-sdk/tracker/db"
	"github.com/stretchr/testify/mock"
)

type StorerMock struct {
	mock.Mock
}

func (m *StorerMock) AddNewFileSystemEntries(ctx context.Context, opts ...db.Options) (chan<- db.FSEntry, <-chan error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(chan<- db.FSEntry), args.Get(1).(<-chan error)
}

func (m *StorerMock) GetFileSystemMutations(ctx context.Context, fsName db.FSName) (db.FSMutations, error) {
	args := m.Called(ctx, fsName)
	return args.Get(0).(db.FSMutations), args.Error(1)
}

func (m *StorerMock) DeleteVersionNew(ctx context.Context, fsName db.FSName) error {
	args := m.Called(ctx, fsName)
	return args.Error(0)
}

func (m *StorerMock) RotateFileSystemVersions(ctx context.Context, fsName db.FSName) error {
	args := m.Called(ctx, fsName)
	return args.Error(0)
}

func (m *StorerMock) MarkFileSystemAsChanged(ctx context.Context, fsName db.FSName) error {
	args := m.Called(ctx, fsName)
	return args.Error(0)
}

func (m *StorerMock) GetFileSystemInfo(ctx context.Context, fsName db.FSName) (*db.FSInfo, error) {
	args := m.Called(ctx, fsName)
	return args.Get(0).(*db.FSInfo), args.Error(1)
}

func (m *StorerMock) GetSyncDetails(ctx context.Context, fsName db.FSName) (db.FSDriver, string, error) {
	args := m.Called(ctx, fsName)
	return args.Get(0).(db.FSDriver), args.String(1), args.Error(2)
}
