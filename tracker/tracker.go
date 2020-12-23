package tracker

import (
	"context"
	"seborama/pcloud/sdk"
	"seborama/pcloud/tracker/db"
	"time"

	"github.com/pkg/errors"
)

type sdkClient interface {
	ListFolder(ctx context.Context, folder sdk.T1PathOrFolderID, recursiveOpt, showDeletedOpt, noFilesOpt, noSharesOpt bool, opts ...sdk.ClientOption) (*sdk.FSList, error)
	Diff(ctx context.Context, diffID uint64, after time.Time, last uint64, block bool, limit uint64, opts ...sdk.ClientOption) (*sdk.DiffResult, error)
}

type storer interface {
	GetLatestTrackingInfo(ctx context.Context) (*db.TrackingInfo, error)
	AddNewFileSystemEntry(ctx context.Context, entry db.FSEntry) error
	GetLatestFileSystemEntries(ctx context.Context) ([]db.FSEntry, error)
	GetPCloudMutations(ctx context.Context) ([]db.FSMutation, error)
	MarkNewFileSystemEntriesAsPrevious(ctx context.Context) error
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
	// ti, err := t.store.GetLatestTrackingInfo(ctx)
	// if err != nil && !errors.Is(err, sql.ErrNoRows) {
	// 	return err
	// }

	err := t.ListLatestPCloudContents(ctx)
	if err != nil {
		return err
	}

	_, err = t.FindPCloudMutations(ctx)
	if err != nil {
		return err
	}

	// dr, err := t.pCloudClient.Diff(ctx, 0, time.Now().Add(-10*time.Minute), 0, false, 0)
	// if err != nil {
	// 	return err
	// }
	// _ = dr

	return nil
}

// ListLatestPCloudContents moves all entries marked as VersionNew to VersionPrevious
// (includes removing all entries marked as VersionPrevious) and then queries list all PCloud
// contents from '/' recursively and stores the results as VersionNew.
func (t *Tracker) ListLatestPCloudContents(ctx context.Context) error {
	err := t.store.MarkNewFileSystemEntriesAsPrevious(ctx)
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

	var entries stack
	entries.add(lf.Metadata)

	for entries.hasNext() {
		entry := entries.pop()

		entryID := entry.FileID
		if entry.IsFolder {
			for _, e := range entry.Contents {
				entries.add(e)
			}
			entryID = entry.FolderID
		}

		fsEntry := db.FSEntry{
			EntryID:        entryID,
			IsFolder:       entry.IsFolder,
			IsDeleted:      entry.IsDeleted,
			DeletedFileID:  entry.DeletedFileID,
			Name:           entry.Name,
			ParentFolderID: entry.ParentFolderID,
			Created:        entry.Created.Time,
			Modified:       entry.Modified.Time,
			Size:           entry.Size,
			Hash:           entry.Hash,
		}

		err = t.store.AddNewFileSystemEntry(ctx, fsEntry)
		if err != nil {
			return errors.WithMessagef(err, "entryID: %d", entry.FileID)
		}
	}

	return nil
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
