package filesystem

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/seborama/pcloud-sdk/sdk"
	"github.com/seborama/pcloud-sdk/tracker/db"
)

// pCloudSDK defines the SDK methods used to perform operations on the PCloud file system.
type pCloudSDK interface {
	ListFolder(ctx context.Context, folder sdk.T1PathOrFolderID, recursiveOpt, showDeletedOpt, noFilesOpt, noSharesOpt bool, opts ...sdk.ClientOption) (*sdk.FSList, error)
}

// PCloud is a file system abstraction for the PCloud file system.
type PCloud struct {
	sdk pCloudSDK
}

// NewPCloud creates a new initialised PCloud structure.
func NewPCloud(sdk pCloudSDK) *PCloud {
	return &PCloud{
		sdk: sdk,
	}
}

// Walk traverses the file system entries and writes each entry to fsEntriesCh.
// It must check for an error in errCh (which indicates the receiver of fsEntriesCh encountered
// a problem and terminate if one is present.
// Walk is the PRODUCER on fsEntriesCh and IS RESPONSIBLE FOR CLOSING IT!!
// nolint: gocognit
func (fs *PCloud) Walk(ctx context.Context, fsName db.FSName, path string, fsEntriesCh chan<- db.FSEntry, errCh <-chan error) error {
	lf, err := fs.sdk.ListFolder(ctx, sdk.T1FolderByPath(path), true, false, false, false)
	if err != nil {
		return err
	}

	err = func() error {
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
				FSName:         fsName,
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

	return err
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
