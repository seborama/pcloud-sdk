package tracker

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/seborama/pcloud/tracker/db"
	"github.com/seborama/pcloud/tracker/filesystem"
)

type storer interface {
	AddNewFileSystemEntries(ctx context.Context, opts ...db.Options) (chan<- db.FSEntry, <-chan error)
	GetLatestFileSystemEntries(ctx context.Context, fsName db.FSName) ([]db.FSEntry, error)
	GetPCloudMutations(ctx context.Context) (db.FSMutations, error)
	GetLocalMutations(ctx context.Context) (db.FSMutations, error)
	GetPCloudVsLocalMutations(ctx context.Context) (db.FSMutations, error)
	DeleteVersionNew(ctx context.Context, fsName db.FSName) error
	RotateFileSystemVersions(ctx context.Context, fsName db.FSName) error
	MarkFileSystemAsChanged(ctx context.Context, fsName db.FSName) error
	MarkSyncInProgress(ctx context.Context, fsName db.FSName) error
	MarkSyncComplete(ctx context.Context, fsName db.FSName) error
	IsFileSystemEmpty(ctx context.Context, fsName db.FSName) (bool, error)
	GetFileSystemInfo(ctx context.Context, fsName db.FSName) (*db.FSInfo, error)
	GetSyncDetails(ctx context.Context, fsName db.FSName) (db.FSDriver, string, error)
}

// Tracker contains the elements necessary to track file system mutations.
type Tracker struct {
	store  storer
	logger *zap.Logger
}

// NewTracker creates a new initiliased Tracker.
func NewTracker(ctx context.Context, store storer) (*Tracker, error) {
	t := &Tracker{
		store: store,
		// TODO: add 'logger'!
	}

	return t, nil
}

// FSDriver represents the operations that may be performed on a file system.
type FSDriver interface {
	// Walk traverses the file system entries and writes each entry to fsEntriesCh.
	// It must check for an error in errCh (which indicates the receiver of fsEntriesCh encountered
	// a problem and terminate if one is present.
	// Walk is the PRODUCER on fsEntriesCh and IS RESPONSIBLE FOR CLOSING IT!!
	Walk(ctx context.Context, fsName db.FSName, path string, fsEntriesCh chan<- db.FSEntry, errCh <-chan error) error
}

// TODO: the fact that this method returns an interface indicates a problem.
//       the implementation of this method likely belongs to the sync package, not the tracker.
func (t *Tracker) GetSyncDetails(ctx context.Context, fsName db.FSName) (FSDriver, string, error) {
	fsDriver, rootPath, err := t.store.GetSyncDetails(ctx, fsName)
	if err != nil {
		return nil, "", err
	}

	switch fsDriver {
	case db.FSDriverPCloud:
		return filesystem.NewPCloud(), rootPath, nil

	case db.FSDriverLocal:
		return filesystem.NewLocal(), rootPath, nil

	default:
		return nil, "", errors.Errorf("unknown file system type '%s' for file system with name '%s'", fsDriver, fsName)
	}
}

// RefreshFSContents walks the specified file system and saves the new contents as VersionNew.
// In order to proceed, RefreshFSContents first drops all VersionPrevious entries and moves the
// current VersionNew entries as VersionPrevious.
func (t *Tracker) RefreshFSContents(ctx context.Context, fsName db.FSName, opts ...Options) error {
	cfg := config{
		entriesChSize: 100,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	panic("this whole method should be inside a transaction for data consistency")
	err := t.rotateFileSystemVersions(ctx, fsName)
	if err != nil {
		return err
	}

	fsEntriesCh, errCh := t.store.AddNewFileSystemEntries(ctx, db.WithEntriesChannelSize(cfg.entriesChSize))

	fsDriver, rootPath, err := t.GetSyncDetails(ctx, fsName)
	if err != nil {
		return err
	}

	err = fsDriver.Walk(ctx, fsName, rootPath, fsEntriesCh, errCh)
	if err != nil {
		return err
	}

	err = t.markFileSystemAsChanged(ctx, fsName)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tracker) rotateFileSystemVersions(ctx context.Context, fsName db.FSName) error {
	fsInfo, err := t.store.GetFileSystemInfo(ctx, fsName)
	if err != nil {
		return errors.WithMessage(err, "database error or sync has not been initialised")
	}

	if fsInfo.FSChanged {
		// The file system is marked as changed. This indicates a sync is pending.
		// So VersionPrevious should not be modified and we only need to refresh VersionNew
		// to allow a more up-to-date sync when it takes place.
		t.logger.Info("clearing down latest version of file system while preserving previous version intact", zap.String("fs_name", string(fsName)))
		return t.store.DeleteVersionNew(ctx, fsName)
	}

	// The file system is not marked as changed. We can replace VersionPrevious with the
	// current VersionNew. After this operation, VersionNew is empty.
	t.logger.Info("rotating versions of file system", zap.String("fs_name", string(fsName)))
	return t.store.RotateFileSystemVersions(ctx, fsName)
}

func (t *Tracker) markFileSystemAsChanged(ctx context.Context, fsName db.FSName) error {
	fsInfo, err := t.store.GetFileSystemInfo(ctx, fsName)
	if err != nil {
		return err
	}

	if fsInfo.FSChanged {
		t.logger.Warn("state of file system is already marked as 'changed'", zap.String("fs_name", string(fsName)))
		return nil
	}

	return t.store.MarkFileSystemAsChanged(ctx, fsName)
}

type config struct {
	entriesChSize int
}

type Options func(*config)

// WithEntriesChannelSize is a functional parameter that allows to choose the size of the entries
// channel used by AddNewFileSystemEntries.
func WithEntriesChannelSize(n int) Options {
	return func(obj *config) {
		obj.entriesChSize = n
	}
}

// FindPCloudVsLocalMutations determines all mutations that have taken place between PCloud
// vs Local.
func (t *Tracker) FindPCloudVsLocalMutations(ctx context.Context) (db.FSMutations, error) {
	fsMutations, err := t.store.GetPCloudVsLocalMutations(ctx)
	if err != nil {
		return nil, err
	}
	return fsMutations, nil
}

// FindPCloudMutations determines all mutations that have taken place in PCloud between
// VersionPrevious and VersionNew.
func (t *Tracker) FindPCloudMutations(ctx context.Context) (db.FSMutations, error) {
	fsMutations, err := t.store.GetPCloudMutations(ctx)
	if err != nil {
		return nil, err
	}
	return fsMutations, nil
}

// FindLocalMutations determines all mutations that have taken place in the Local file system
// between VersionPrevious and VersionNew.
func (t *Tracker) FindLocalMutations(ctx context.Context) (db.FSMutations, error) {
	fsMutations, err := t.store.GetLocalMutations(ctx)
	if err != nil {
		return nil, err
	}
	return fsMutations, nil
}
