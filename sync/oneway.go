package sync

import (
	"context"
	"path/filepath"

	"seborama/pcloud/sdk"
	"seborama/pcloud/tracker/db"

	"github.com/pkg/errors"
)

type tracker interface {
	ListMutations(ctx context.Context) (db.FSMutations, error)
}

type sdkClient interface {
	FileOpen(ctx context.Context, flags uint64, file sdk.T4PathOrFileIDOrFolderIDName, opts ...sdk.ClientOption) (*sdk.File, error)
	FileClose(ctx context.Context, fd uint64, opts ...sdk.ClientOption) error
	FileRead(ctx context.Context, fd, count uint64, opts ...sdk.ClientOption) ([]byte, error)
}

type FSReader interface {
	StreamFileData(ctx context.Context, fsEntry db.FSEntry) (<-chan []byte, <-chan error)
}

type FSWriter interface {
	MkDir(ctx context.Context, path string) error
	MkFile(ctx context.Context, path string, dataCh <-chan []byte) error
	RmDir(ctx context.Context, path string) error
	RmFile(ctx context.Context, path string) error
	MvDir(ctx context.Context, fromPath, toPath string) error
	MvFile(ctx context.Context, fromPath, toPath string) error
}

type OneWay struct {
	from    FSReader
	to      FSWriter
	tracker tracker
}

func NewOneWay(from FSReader, to FSWriter, fsTracker tracker) *OneWay {
	return &OneWay{
		from:    from,
		to:      to,
		tracker: fsTracker,
	}
}

func (s *OneWay) Sync(ctx context.Context) error {
	// TODO: after the one-way sync has completed, delete extraneous entries that exist on the right
	//       ie files and folder that were created externally on the "to" side, not by the sync.
	mutations, err := s.tracker.ListMutations(ctx)
	if err != nil {
		return nil
	}

	for _, m := range mutations {
		switch m.Type {
		case db.MutationTypeCreated:
			s.create(ctx, m.Details)

		case db.MutationTypeDeleted:
			s.delete(ctx, m.Details)

		case db.MutationTypeModified:
			s.update(ctx, m.Details)

		case db.MutationTypeMoved:
			s.move(ctx, m.Details)

		default:
			return errors.Errorf("unknown mutation type '%s'", string(m.Type))
		}
	}

	return nil
}

func (s *OneWay) create(ctx context.Context, entryMutations db.EntryMutations) error {
	if len(entryMutations) != 1 {
		return errors.Errorf("expected 1 entry in mutation details but got '%d'", len(entryMutations))
	}

	fsEntry := entryMutations[0].FSEntry

	if fsEntry.IsFolder {
		return s.createFolder(ctx, fsEntry)
	}

	return s.createFile(ctx, fsEntry)
}

func (s *OneWay) createFolder(ctx context.Context, fsEntry db.FSEntry) error {
	return s.to.MkDir(ctx, filepath.Join(fsEntry.Path, fsEntry.Name))
}

func (s *OneWay) createFile(ctx context.Context, fsEntry db.FSEntry) error {
	var err error

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dataCh, errCh := s.from.StreamFileData(ctx, fsEntry)

	go func() {
		select {
		case <-ctx.Done():
		case err = <-errCh:
			if err != nil {
				err = errors.WithStack(err)
				cancel()
			}
		}
	}()

	err1 := s.to.MkFile(ctx, filepath.Join(fsEntry.Path, fsEntry.Name), dataCh)
	if err1 != nil && err == nil {
		return err1
	}

	return err
}

func (s *OneWay) delete(ctx context.Context, entryMutations db.EntryMutations) error {
	// TODO: some ordering of deletions should take place so that files in a folder are deleted before the folder is deleted
	if len(entryMutations) != 1 {
		return errors.Errorf("expected 1 entry in mutation details but got '%d'", len(entryMutations))
	}

	fsEntry := entryMutations[0].FSEntry

	if fsEntry.IsFolder {
		return s.deleteFolder(ctx, fsEntry)
	}

	return s.deleteFile(ctx, fsEntry)
}

func (s *OneWay) deleteFolder(ctx context.Context, fsEntry db.FSEntry) error {
	return s.to.RmDir(ctx, filepath.Join(fsEntry.Path, fsEntry.Name))
}

func (s *OneWay) deleteFile(ctx context.Context, fsEntry db.FSEntry) (err error) {
	return s.to.RmFile(ctx, filepath.Join(fsEntry.Path, fsEntry.Name))
}

func (s *OneWay) update(ctx context.Context, entryMutations db.EntryMutations) error {
	if len(entryMutations) != 2 {
		return errors.Errorf("expected 2 entries in mutation details but got '%d'", len(entryMutations))
	}

	fsEntry := entryMutations[1].FSEntry

	if fsEntry.IsFolder {
		return errors.Errorf("received update for a folder: folderID='%d' path='%s' name='%s'", fsEntry.EntryID, fsEntry.Path, fsEntry.Name)
	}

	// TODO: refactor and optimise for block-level (differential) copying
	return s.createFile(ctx, fsEntry)
}

func (s *OneWay) move(ctx context.Context, entryMutations db.EntryMutations) error {
	if len(entryMutations) != 2 {
		return errors.Errorf("expected 2 entries in mutation details but got '%d'", len(entryMutations))
	}

	fromFSEntry := entryMutations[0].FSEntry
	toFSEntry := entryMutations[1].FSEntry

	if fromFSEntry.IsFolder {
		return s.moveFolder(ctx, fromFSEntry, toFSEntry)
	}

	return s.moveFile(ctx, fromFSEntry, toFSEntry)
}

func (s *OneWay) moveFolder(ctx context.Context, fromFSEntry, toFSEntry db.FSEntry) error {
	return s.to.MvDir(ctx, filepath.Join(fromFSEntry.Path, fromFSEntry.Name), filepath.Join(toFSEntry.Path, toFSEntry.Name))
}

func (s *OneWay) moveFile(ctx context.Context, fromFSEntry, toFSEntry db.FSEntry) (err error) {
	return s.to.MvFile(ctx, filepath.Join(fromFSEntry.Path, fromFSEntry.Name), filepath.Join(toFSEntry.Path, toFSEntry.Name))
}
