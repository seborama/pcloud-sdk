package tracker

import (
	"bufio"
	"context"

	// nolint:gosec
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"seborama/pcloud/sdk"
	"seborama/pcloud/tracker/archos"
	"seborama/pcloud/tracker/db"

	"github.com/pkg/errors"
)

type sdkClient interface {
	ListFolder(ctx context.Context, folder sdk.T1PathOrFolderID, recursiveOpt, showDeletedOpt, noFilesOpt, noSharesOpt bool, opts ...sdk.ClientOption) (*sdk.FSList, error)
	Diff(ctx context.Context, diffID uint64, after time.Time, last uint64, block bool, limit uint64, opts ...sdk.ClientOption) (*sdk.DiffResult, error)
}

type storer interface {
	AddNewFileSystemEntries(ctx context.Context, opts ...db.Options) (chan<- db.FSEntry, <-chan error)
	GetLatestFileSystemEntries(ctx context.Context, fsType db.FSType) ([]db.FSEntry, error)
	GetPCloudMutations(ctx context.Context) (db.FSMutations, error)
	GetLocalMutations(ctx context.Context) (db.FSMutations, error)
	GetPCloudVsLocalMutations(ctx context.Context) (db.FSMutations, error)
	DeleteVersionNew(ctx context.Context, fsType db.FSType) error
	MarkNewFileSystemEntriesAsPrevious(ctx context.Context, fsType db.FSType) error
	MarkSyncRequired(ctx context.Context, fsType db.FSType) error
	MarkSyncInProgress(ctx context.Context, fsType db.FSType) error
	MarkSyncComplete(ctx context.Context, fsType db.FSType) error
	IsFileSystemEmpty(ctx context.Context, fsType db.FSType) (bool, error)
	GetSyncStatus(ctx context.Context, fsType db.FSType) (db.SyncStatus, error)
}

// Tracker contains the elements necessary to track file system mutations.
type Tracker struct {
	pCloudClient sdkClient
	store        storer
}

// NewTracker creates a new initiliased Tracker.
func NewTracker(ctx context.Context, pCloudClient sdkClient, store storer) (*Tracker, error) {
	t := &Tracker{
		pCloudClient: pCloudClient,
		store:        store,
	}

	err := t.initSyncStatus(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// ListLatestPCloudContents moves all entries marked as VersionNew to VersionPrevious
// (includes removing all entries marked as VersionPrevious) and then queries the contents
// of path from PCloud recursively and stores the results as VersionNew.
func (t *Tracker) ListLatestPCloudContents(ctx context.Context, path string, opts ...Options) error {
	cfg := config{
		entriesChSize: 100,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	lf, err := t.pCloudClient.ListFolder(ctx, sdk.T1FolderByPath(path), false, false, true, true)
	if err != nil {
		return err
	}
	folderID := lf.Metadata.FolderID

	err = t.markNewFileSystemEntriesAsPrevious(ctx, db.PCloudFileSystem)
	if err != nil {
		return err
	}

	lf, err = t.pCloudClient.ListFolder(ctx, sdk.T1FolderByID(folderID), true, false, false, false)
	if err != nil {
		return err
	}

	if lf.Metadata.Name == "" {
		return errors.New("cannot list pCloud drive contents: no data")
	}

	if lf.Metadata.IsDeleted {
		return errors.New("cannot list pCloud drive contents: root folder is reportedly deleted")
	}

	err = func() error {
		fsEntriesCh, errCh := t.store.AddNewFileSystemEntries(ctx, db.WithEntriesChannelSize(cfg.entriesChSize))
		var entries stack
		entries.add(lf.Metadata)

		for entries.hasNext() {
			entry := entries.pop()

			hash := ""
			entryID := entry.FileID
			if entry.IsFolder {
				for _, e := range entry.Contents {
					if e.IsDeleted {
						continue
					}

					e.Path = filepath.Join(entry.Path, entry.Name)
					entries.add(e)
				}

				entryID = entry.FolderID
			} else {
				hash = fmt.Sprintf("%d", entry.Hash)
			}

			fsEntry := db.FSEntry{
				FSType:         db.PCloudFileSystem,
				EntryID:        entryID,
				IsFolder:       entry.IsFolder,
				Path:           entry.Path,
				Name:           entry.Name,
				ParentFolderID: entry.ParentFolderID,
				Created:        entry.Created.Time,
				Modified:       entry.Modified.Time,
				Size:           entry.Size,
				Hash:           hash,
			}

			select {
			case err = <-errCh:
				close(fsEntriesCh)
				return errors.WithStack(err)
			case fsEntriesCh <- fsEntry:
			}
		}
		close(fsEntriesCh)

		return errors.WithStack(<-errCh)
	}()
	if err != nil {
		return err
	}

	err = t.markSyncAsRequired(ctx, db.PCloudFileSystem)
	if err != nil {
		return err
	}

	return nil
}

// ListLatestLocalContents moves all entries marked as VersionNew to VersionPrevious
// (includes removing all entries marked as VersionPrevious) and then queries the contents
// of '/' from PCloud recursively and stores the results as VersionNew.
func (t *Tracker) ListLatestLocalContents(ctx context.Context, path string, opts ...Options) error {
	cfg := config{
		entriesChSize: 100,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	err := t.markNewFileSystemEntriesAsPrevious(ctx, db.LocalFileSystem)
	if err != nil {
		return err
	}

	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return errors.Errorf("path is not pointing at a directory: %s", path)
	}

	deviceID := archos.Device(fi)

	folderIDs := map[string]uint64{}

	err = func() error {
		fsEntriesCh, errCh := t.store.AddNewFileSystemEntries(ctx, db.WithEntriesChannelSize(cfg.entriesChSize))
		isFSEntriesChOpened := true

		err = filepath.Walk(path,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return errors.WithStack(err)
				}

				if archos.Device(info) != deviceID {
					return filepath.SkipDir
				}

				hash := ""
				dir := filepath.Dir(path) // NOTE: this also calls filepath.Clean
				if info.IsDir() {
					dir = filepath.Clean(path)
					folderIDs[dir] = archos.Inode(info)
				} else {
					hash, err = hashFileData(path)
					if err != nil {
						return err
					}
				}

				createdTime := archos.CreatedTime(info)

				parentFolderID, ok := folderIDs[dir]
				if !ok {
					return errors.Errorf("unable to determine parent folder ID for '%s' using key='%s'", path, dir)
				}

				// tips for Windows support:
				// - go/src/os/types_windows.go
				// - https://stackoverflow.com/questions/7162164/does-windows-have-inode-numbers-like-linux
				fsEntry := db.FSEntry{
					FSType:         db.LocalFileSystem,
					DeviceID:       fmt.Sprintf("%d", deviceID),
					EntryID:        archos.Inode(info),
					IsFolder:       info.IsDir(),
					Path:           filepath.Dir(path),
					Name:           info.Name(),
					ParentFolderID: parentFolderID,
					Created:        createdTime,
					Modified:       info.ModTime(),
					Size:           uint64(info.Size()),
					Hash:           hash,
				}

				select {
				case err := <-errCh:
					close(fsEntriesCh)
					isFSEntriesChOpened = false
					return errors.WithStack(err)
				case fsEntriesCh <- fsEntry:
					return nil
				}
			})
		if isFSEntriesChOpened {
			close(fsEntriesCh)
			isFSEntriesChOpened = false
			return errors.WithStack(<-errCh)
		}
		return err
	}()
	if err != nil {
		return err
	}

	err = t.markSyncAsRequired(ctx, db.LocalFileSystem)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tracker) initSyncStatus(ctx context.Context) error {
	for _, fsType := range []db.FSType{db.PCloudFileSystem, db.LocalFileSystem} {
		fsEmpty, err := t.store.IsFileSystemEmpty(ctx, fsType)
		if err != nil {
			return err
		}

		if fsEmpty {
			err = t.store.MarkSyncComplete(ctx, fsType)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *Tracker) markNewFileSystemEntriesAsPrevious(ctx context.Context, fsType db.FSType) error {
	status, err := t.store.GetSyncStatus(ctx, fsType)
	if err != nil {
		return errors.WithMessage(err, "database error or sync has not been initialised")
	}

	switch status {
	case db.SyncStatusRequired:
		// TODO: add a test for this scenario
		return t.store.DeleteVersionNew(ctx, fsType)

	case db.SyncStatusComplete:
		return t.store.MarkNewFileSystemEntriesAsPrevious(ctx, fsType)

	default:
		return errors.Errorf("markNewFileSystemEntriesAsPrevious requires sync status '%s' but status is currently '%s'", db.SyncStatusComplete, status)
	}
}

func (t *Tracker) markSyncAsRequired(ctx context.Context, fsType db.FSType) error {
	status, err := t.store.GetSyncStatus(ctx, fsType)
	if err != nil {
		return err
	}

	if status != db.SyncStatusComplete {
		return errors.Errorf("cannot transition sync status from '%s' to '%s'", status, db.SyncStatusRequired)
	}

	return t.store.MarkSyncRequired(ctx, fsType)
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

func hashFileData(path string) (string, error) {
	// nolint: gosec
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	// nolint: gosec
	cs := sha1.New()

	r := bufio.NewReader(f)

	data := make([]byte, 2_097_152)

	for {
		n, err := r.Read(data)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}

		_, err = cs.Write(data[:n])
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", cs.Sum(nil)), nil
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

type stack struct {
	entries []*sdk.Metadata
}

func (s *stack) add(entry *sdk.Metadata) {
	s.entries = append(s.entries, entry)
}

func (s *stack) hasNext() bool {
	return len(s.entries) > 0
}

func (s *stack) pop() *sdk.Metadata {
	if len(s.entries) == 0 {
		return nil
	}

	entry := s.entries[0]
	if len(s.entries) > 1 {
		s.entries = s.entries[1:] // drop the fist element
	} else {
		s.entries = s.entries[:0] // empty stack
	}

	return entry
}
