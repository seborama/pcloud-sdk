package tracker

import (
	"bufio"
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"seborama/pcloud/sdk"
	"seborama/pcloud/tracker/db"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

type sdkClient interface {
	ListFolder(ctx context.Context, folder sdk.T1PathOrFolderID, recursiveOpt, showDeletedOpt, noFilesOpt, noSharesOpt bool, opts ...sdk.ClientOption) (*sdk.FSList, error)
	Diff(ctx context.Context, diffID uint64, after time.Time, last uint64, block bool, limit uint64, opts ...sdk.ClientOption) (*sdk.DiffResult, error)
}

type storer interface {
	AddNewFileSystemEntry(ctx context.Context, fsType db.FSType) (chan<- db.FSEntry, <-chan error)
	GetLatestFileSystemEntries(ctx context.Context, fsType db.FSType) ([]db.FSEntry, error)
	GetPCloudMutations(ctx context.Context) ([]db.FSMutation, error)
	MarkNewFileSystemEntriesAsPrevious(ctx context.Context, fsType db.FSType) error
}

// Tracker contains the elements necessary to track file system mutations.
type Tracker struct {
	pCloudClient sdkClient
	store        storer
}

// NewTracker creates a new initiliased Tracker.
func NewTracker(ctx context.Context, pCloudClient sdkClient, store storer) (*Tracker, error) {
	return &Tracker{
		pCloudClient: pCloudClient,
		store:        store,
	}, nil
}

func (t *Tracker) Track(ctx context.Context) error {
	err := t.ListLatestPCloudContents(ctx)
	if err != nil {
		return err
	}

	_, err = t.FindPCloudMutations(ctx)
	if err != nil {
		return err
	}

	return nil
}

type folderChildren struct {
	Name     string
	ChildIDs []uint64
}

// ListLatestPCloudContents moves all entries marked as VersionNew to VersionPrevious
// (includes removing all entries marked as VersionPrevious) and then queries the contents
// of '/' from PCloud recursively and stores the results as VersionNew.
func (t *Tracker) ListLatestPCloudContents(ctx context.Context) error {
	err := t.store.MarkNewFileSystemEntriesAsPrevious(ctx, db.PCloudFileSystem)
	if err != nil {
		return err
	}

	lf, err := t.pCloudClient.ListFolder(ctx, sdk.T1FolderByID(sdk.RootFolderID), true, true, false, false)
	if err != nil {
		return err
	}

	if lf.Metadata.Name == "" {
		return errors.New("cannot list pCloud drive contents: no data")
	}

	fsEntriesCh, errCh := t.store.AddNewFileSystemEntry(ctx, db.PCloudFileSystem)
	var entries stack
	entries.add(lf.Metadata)

	func() {
		defer close(fsEntriesCh)

		for entries.hasNext() {
			entry := entries.pop()

			hash := ""
			entryID := entry.FileID
			if entry.IsFolder {
				for _, e := range entry.Contents {
					e.Path = filepath.Join(entry.Path, entry.Name)
					entries.add(e)
				}

				entryID = entry.FolderID
			} else {
				hash = fmt.Sprintf("%d", entry.Hash)
			}

			fsEntry := db.FSEntry{
				EntryID:        entryID,
				IsFolder:       entry.IsFolder,
				IsDeleted:      entry.IsDeleted,
				DeletedFileID:  entry.DeletedFileID,
				Path:           entry.Path,
				Name:           entry.Name,
				ParentFolderID: entry.ParentFolderID,
				Created:        entry.Created.Time,
				Modified:       entry.Modified.Time,
				Size:           entry.Size,
				Hash:           hash,
			}

			fsEntriesCh <- fsEntry
		}
	}()

	return <-errCh
}

// ListLatestLocalContents moves all entries marked as VersionNew to VersionPrevious
// (includes removing all entries marked as VersionPrevious) and then queries the contents
// of '/' from PCloud recursively and stores the results as VersionNew.
func (t *Tracker) ListLatestLocalContents(ctx context.Context, path string) error {
	err := t.store.MarkNewFileSystemEntriesAsPrevious(ctx, db.LocalFileSystem)
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

	deviceID := fi.Sys().(*syscall.Stat_t).Dev // Unix only

	folderIDs := map[string]uint64{}

	fsEntriesCh, errCh := t.store.AddNewFileSystemEntry(ctx, db.LocalFileSystem)

	func() {
		defer close(fsEntriesCh)

		err = filepath.Walk(path,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return errors.WithStack(err)
				}

				if info.Sys().(*syscall.Stat_t).Dev != deviceID {
					return filepath.SkipDir
				}

				hash := ""
				dir := filepath.Dir(path) // NOTE: this also calls filepath.Clean
				if info.IsDir() {
					dir = filepath.Clean(path)
					folderIDs[dir] = info.Sys().(*syscall.Stat_t).Ino // Unix only
				} else {
					hash, err = hashFileData(path)
					if err != nil {
						return err
					}
				}

				// TODO: OSX-specific!!
				createdTime := time.Unix(int64(info.Sys().(*syscall.Stat_t).Ctimespec.Sec), int64(info.Sys().(*syscall.Stat_t).Ctimespec.Nsec))

				parentFolderID, ok := folderIDs[dir]
				if !ok {
					return errors.Errorf("unable to determine parent folder ID for '%s' using key='%s'", path, dir)
				}

				// TODO: this is currently unix-specific, make more generic to at least include Windows
				//       see: go/src/os/types_windows.go
				//       see: https://stackoverflow.com/questions/7162164/does-windows-have-inode-numbers-like-linux
				fsEntry := db.FSEntry{
					DeviceID:       fmt.Sprintf("%d", deviceID),
					EntryID:        info.Sys().(*syscall.Stat_t).Ino, // Unix only
					IsFolder:       info.IsDir(),
					Path:           filepath.Dir(path),
					Name:           info.Name(),
					ParentFolderID: parentFolderID,
					Created:        createdTime,
					Modified:       info.ModTime(),
					Size:           uint64(info.Size()),
					Hash:           hash, // TODO: needs calculating but only if new / modified file
				}

				fsEntriesCh <- fsEntry

				return nil
			})
	}()

	return <-errCh
}

func hashFileData(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { f.Close() }()

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

// FindPCloudMutations determines all mutations that have taken place in PCloud between
// VersionPrevious and VersionNew.
func (t *Tracker) FindPCloudMutations(ctx context.Context) ([]db.FSMutation, error) {
	fsMutations, err := t.store.GetPCloudMutations(ctx)
	if err != nil {
		return nil, err
	}
	return fsMutations, nil
}

type stack struct {
	entries []*sdk.Metadata
}

func (s *stack) add(entry sdk.Metadata) {
	s.entries = append(s.entries, &entry)
}

func (s *stack) hasNext() bool {
	return len(s.entries) > 0
}

func (s *stack) len() int {
	return len(s.entries)
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
