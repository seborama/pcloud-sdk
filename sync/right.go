package sync

import (
	"context"
	"io"
	"path/filepath"

	"seborama/pcloud/sdk"
	"seborama/pcloud/tracker/db"

	"github.com/pkg/errors"
)

type tracker interface {
	FindPCloudVsLocalMutations(ctx context.Context) (db.FSMutations, error)
}

type sdkClient interface {
	FileOpen(ctx context.Context, flags uint64, file sdk.T4PathOrFileIDOrFolderIDName, opts ...sdk.ClientOption) (*sdk.File, error)
	FileClose(ctx context.Context, fd uint64, opts ...sdk.ClientOption) error
	FileRead(ctx context.Context, fd, count uint64, opts ...sdk.ClientOption) ([]byte, error)
}

type localClient interface {
	MkDir(ctx context.Context, path string) error
	MkFile(ctx context.Context, path string, dataCh <-chan []byte) error
	RmDir(ctx context.Context, path string) error
	RmFile(ctx context.Context, path string) error
}

type Sync struct {
	tracker      tracker
	pCloudClient sdkClient
	localClient  localClient
}

func NewSync(fsTracker tracker, pCloudClient sdkClient, localClient localClient) *Sync {
	return &Sync{
		tracker:      fsTracker,
		pCloudClient: pCloudClient,
		localClient:  localClient,
	}
}

func (s *Sync) Right(ctx context.Context) error {
	mutations, err := s.tracker.FindPCloudVsLocalMutations(ctx)
	if err != nil {
		return nil
	}

	for _, m := range mutations {
		switch m.Type {
		case db.MutationTypeCreated:
			s.create(ctx, m)

		case db.MutationTypeDeleted:
			s.delete(ctx, m)

		case db.MutationTypeModified:
			s.update(ctx, m)

		case db.MutationTypeMoved:
			panic("not yet implemented")

		default:
			return errors.Errorf("unknown mutation type '%s'", string(m.Type))
		}
	}

	return nil
}

func (s *Sync) create(ctx context.Context, fsMutation db.FSMutation) error {
	if fsMutation.IsFolder {
		return s.createFolder(ctx, fsMutation)
	}

	return s.createFile(ctx, fsMutation)
}

func (s *Sync) createFolder(ctx context.Context, fsMutation db.FSMutation) error {
	switch fsMutation.FSType {
	case db.PCloudFileSystem:
		return s.localClient.MkDir(ctx, filepath.Join(fsMutation.Path, fsMutation.Name))

	case db.LocalFileSystem:
		panic("not yet implemented")

	default:
		return errors.Errorf("unknown file system type '%s'", string(fsMutation.FSType))
	}
}

func (s *Sync) createFile(ctx context.Context, fsMutation db.FSMutation) (err error) {
	switch fsMutation.FSType {
	case db.PCloudFileSystem:
		f, err := s.pCloudClient.FileOpen(ctx, 0, sdk.T4FileByID(fsMutation.EntryID))
		if err != nil {
			return err
		}
		defer func() {
			e := s.pCloudClient.FileClose(ctx, f.FD)
			if e != nil {
				if err != nil {
					err = errors.Wrapf(err, "additional error on unwinding call stack with existing error: %s", e.Error())
					return
				}
				err = e
			}
		}()

		dataCh := make(chan []byte, 100)
		defer close(dataCh)

		eof := false
		for !eof {
			data, err := s.pCloudClient.FileRead(ctx, f.FD, 1_048_576)
			if err != nil {
				if err != io.EOF {
					return err
				}
				eof = true
			}

			dataCh <- data

			err = s.localClient.MkFile(ctx, filepath.Join(fsMutation.Path, fsMutation.Name), dataCh)
			if err != nil {
				return err
			}
		}

		return nil

	case db.LocalFileSystem:
		panic("not yet implemented")

	default:
		return errors.Errorf("unknown file system type '%s'", string(fsMutation.FSType))
	}
}

func (s *Sync) delete(ctx context.Context, fsMutation db.FSMutation) error {
	if fsMutation.IsFolder {
		return s.deleteFolder(ctx, fsMutation)
	}

	return s.deleteFile(ctx, fsMutation)
}

func (s *Sync) deleteFolder(ctx context.Context, fsMutation db.FSMutation) error {
	switch fsMutation.FSType {
	case db.PCloudFileSystem:
		return s.localClient.RmDir(ctx, filepath.Join(fsMutation.Path, fsMutation.Name))

	case db.LocalFileSystem:
		panic("not yet implemented")

	default:
		return errors.Errorf("unknown file system type '%s'", string(fsMutation.FSType))
	}
}

func (s *Sync) deleteFile(ctx context.Context, fsMutation db.FSMutation) (err error) {
	switch fsMutation.FSType {
	case db.PCloudFileSystem:
		return s.localClient.RmFile(ctx, filepath.Join(fsMutation.Path, fsMutation.Name))

	case db.LocalFileSystem:
		panic("not yet implemented")

	default:
		return errors.Errorf("unknown file system type '%s'", string(fsMutation.FSType))
	}
}

func (s *Sync) update(ctx context.Context, fsMutation db.FSMutation) error {
	if fsMutation.IsFolder {
		return errors.Errorf("unexpected: received update for a folder: folderID='%d' path='%s' name='%s'", fsMutation.EntryID, fsMutation.Path, fsMutation.Name)
	}

	// TODO: refactor and optimise for block-level (differential) copying
	return s.createFile(ctx, fsMutation)
}
