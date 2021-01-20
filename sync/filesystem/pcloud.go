package filesystem

import (
	"context"
	"io"
	"seborama/pcloud/sdk"
	"seborama/pcloud/tracker/db"

	"github.com/pkg/errors"
)

// pCloudSDK defines the SDK methods used to perform operations on the PCloud file system.
type pCloudSDK interface {
	FileOpen(ctx context.Context, flags uint64, file sdk.T4PathOrFileIDOrFolderIDName, opts ...sdk.ClientOption) (*sdk.File, error)
	FileClose(ctx context.Context, fd uint64, opts ...sdk.ClientOption) error
	FileRead(ctx context.Context, fd, count uint64, opts ...sdk.ClientOption) ([]byte, error)
	FileWrite(ctx context.Context, fd uint64, data []byte, opts ...sdk.ClientOption) (*sdk.FileDataTransfer, error)
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

// StreamFileData reads the contents of the file pointed to by fsEntry and streams it to the
// channel the method returns.
// TODO: wrap the dataCh into a io.ReadWriter so to keep the code simple and offer a familiar
//       Go feel.
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
			if e != nil && err == nil {
				errCh <- e
				return
			}
		}()

		eof := false
		for !eof {
			data, err := fs.sdk.FileRead(ctx, f.FD, 1_048_576)
			if err != nil {
				if !errors.Is(err, io.EOF) {
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

// MkDir creates a directory.
func (fs *PCloud) MkDir(ctx context.Context, path string) error {
	panic("not implemented")
}

// MkFile creates a file with the contents streamed through dataCh.
// TODO: wrap the dataCh into a io.ReadWriter so to keep the code simple and offer a familiar
//       Go feel.
func (fs *PCloud) MkFile(ctx context.Context, path string, dataCh <-chan []byte) (err error) {
	f, err := fs.sdk.FileOpen(ctx, sdk.O_CREAT|sdk.O_TRUNC, sdk.T4FileByPath(path))
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		e := fs.sdk.FileClose(ctx, f.FD)
		if e != nil && err == nil {
			err = e
			return
		}
	}()

	for data := range dataCh {
		_, err = fs.sdk.FileWrite(ctx, f.FD, data)
		if err != nil {
			return
		}
	}

	return nil
}

// RmDir removes a directory.
func (fs *PCloud) RmDir(ctx context.Context, path string) error {
	panic("not implemented")
}

// RmFile removes a file.
func (fs *PCloud) RmFile(ctx context.Context, path string) error {
	panic("not implemented")
}

// MvDir moves a directory.
func (fs *PCloud) MvDir(ctx context.Context, fromPath string, toPath string) error {
	panic("not implemented")
}

// MvFile moves a file.
func (fs *PCloud) MvFile(ctx context.Context, fromPath string, toPath string) error {
	panic("not implemented")
}
