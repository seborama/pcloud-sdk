package tracker

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/seborama/pcloud/tracker/db"
)

type storer interface {
	AddNewFileSystemEntries(ctx context.Context, opts ...db.Options) (chan<- db.FSEntry, <-chan error)
	GetFileSystemMutations(ctx context.Context, fsName db.FSName) (db.FSMutations, error)
	DeleteVersionNew(ctx context.Context, fsName db.FSName) error
	RotateFileSystemVersions(ctx context.Context, fsName db.FSName) error
	MarkFileSystemAsChanged(ctx context.Context, fsName db.FSName) error
	GetFileSystemInfo(ctx context.Context, fsName db.FSName) (*db.FSInfo, error)
	GetSyncDetails(ctx context.Context, fsName db.FSName) (db.FSDriver, string, error)
}

// Tracker contains the elements necessary to track file system mutations.
type Tracker struct {
	logger   *zap.Logger
	store    storer
	fsDriver FSDriver
	fsName   db.FSName // TODO: not ideal to have this coupling with "db"
}

// NewTracker creates a new initiliased Tracker.
func NewTracker(ctx context.Context, logger *zap.Logger, store storer, fsDriver FSDriver, fsName db.FSName) (*Tracker, error) {
	t := &Tracker{
		logger:   logger,
		store:    store,
		fsDriver: fsDriver,
		fsName:   fsName,
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
func (t *Tracker) GetRootPath(ctx context.Context) (string, error) {
	_, rootPath, err := t.store.GetSyncDetails(ctx, t.fsName)
	if err != nil {
		return "", err
	}

	return rootPath, nil
}

// RefreshFSContents walks the specified file system and saves the new contents as VersionNew.
// In order to proceed, RefreshFSContents first drops all VersionPrevious entries and moves the
// current VersionNew entries as VersionPrevious.
func (t *Tracker) RefreshFSContents(ctx context.Context, opts ...RefreshOption) error {
	cfg := config{
		entriesChSize: 100,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	fmt.Println("THIS WHOLE METHOD SHOULD BE INSIDE A TRANSACTION FOR DATA CONSISTENCY")
	err := t.rotateFileSystemVersions(ctx)
	if err != nil {
		return err
	}

	fsEntriesCh, errCh := t.store.AddNewFileSystemEntries(ctx, db.WithEntriesChannelSize(cfg.entriesChSize))

	rootPath, err := t.GetRootPath(ctx)
	if err != nil {
		return err
	}

	err = t.fsDriver.Walk(ctx, t.fsName, rootPath, fsEntriesCh, errCh)
	if err != nil {
		return err
	}

	err = t.markFileSystemAsChanged(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tracker) rotateFileSystemVersions(ctx context.Context) error {
	fsInfo, err := t.store.GetFileSystemInfo(ctx, t.fsName)
	if err != nil {
		return errors.WithMessage(err, "database error or sync has not been initialised")
	}

	if fsInfo.FSChanged {
		// The file system is marked as changed. This indicates a sync is pending.
		// So VersionPrevious should not be modified and we only need to refresh VersionNew
		// to allow a more up-to-date sync when it takes place.
		t.logger.Debug("clearing down latest version of file system while preserving previous version intact", zap.String("fs_name", string(t.fsName)))
		return t.store.DeleteVersionNew(ctx, t.fsName)
	}

	// The file system is not marked as changed. We can replace VersionPrevious with the
	// current VersionNew. After this operation, VersionNew is empty.
	t.logger.Debug("rotating versions of file system", zap.String("fs_name", string(t.fsName)))
	return t.store.RotateFileSystemVersions(ctx, t.fsName)
}

func (t *Tracker) markFileSystemAsChanged(ctx context.Context) error {
	fsInfo, err := t.store.GetFileSystemInfo(ctx, t.fsName)
	if err != nil {
		return err
	}

	if fsInfo.FSChanged {
		t.logger.Warn("state of file system is already marked as 'changed'", zap.String("fs_name", string(t.fsName)))
		return nil
	}

	return t.store.MarkFileSystemAsChanged(ctx, t.fsName)
}

type config struct {
	entriesChSize int
}

type RefreshOption func(*config)

// WithEntriesChannelSize is a functional parameter that allows to choose the size of the entries
// channel used by AddNewFileSystemEntries.
func WithEntriesChannelSize(n int) RefreshOption {
	return func(obj *config) {
		obj.entriesChSize = n
	}
}

// ListMutations finds all mutations that have taken place in the file system between
// VersionPrevious and VersionNew.
func (t *Tracker) ListMutations(ctx context.Context) (db.FSMutations, error) {
	fsMutations, err := t.store.GetFileSystemMutations(ctx, t.fsName)
	if err != nil {
		return nil, err
	}
	return fsMutations, nil
}
