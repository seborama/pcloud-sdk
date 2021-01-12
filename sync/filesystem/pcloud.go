package filesystem

import (
	"context"
	"io"
	"seborama/pcloud/sdk"
	"seborama/pcloud/tracker/db"

	"github.com/pkg/errors"
)

type PCloud struct {
	sdk *sdk.Client
}

func (fs *PCloud) StreamFileData(ctx context.Context, fsEntry db.FSEntry) (<-chan []byte, <-chan error) {
	dataCh := make(chan []byte, 100)
	errCh := make(chan error)

	go func() {
		defer close(errCh)
		defer close(dataCh)

		f, err := fs.sdk.FileOpen(ctx, 0, sdk.T4FileByID(fsEntry.EntryID))
		if err != nil {
			errCh <- err
			return
		}
		defer func() {
			e := fs.sdk.FileClose(ctx, f.FD)
			if e != nil {
				if err != nil {
					err = errors.Wrapf(err, "additional error on unwinding call stack with existing error: %s", e.Error())
					return
				}
				err = e
			}
		}()

		eof := false
		for !eof {
			data, err := fs.sdk.FileRead(ctx, f.FD, 1_048_576)
			if err != nil {
				if err != io.EOF {
					errCh <- err
					return
				}
				eof = true
			}

			dataCh <- data
		}
	}()

	return dataCh, errCh
}

func (fs *PCloud) MkDir(ctx context.Context, path string) error {
	panic("not implemented") // TODO: Implement
}

func (fs *PCloud) MkFile(ctx context.Context, path string, dataCh <-chan []byte) error {
	panic("not implemented") // TODO: Implement
}

func (fs *PCloud) RmDir(ctx context.Context, path string) error {
	panic("not implemented") // TODO: Implement
}

func (fs *PCloud) RmFile(ctx context.Context, path string) error {
	panic("not implemented") // TODO: Implement
}

func (fs *PCloud) MvDir(ctx context.Context, fromPath string, toPath string) error {
	panic("not implemented") // TODO: Implement
}

func (fs *PCloud) MvFile(ctx context.Context, fromPath string, toPath string) error {
	panic("not implemented") // TODO: Implement
}
