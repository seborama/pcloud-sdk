package filesystem

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/seborama/pcloud/tracker/db"
)

// fsOperations wraps basic OS implemented operations on a file system.
type fsOperations interface {
	Open(name string) (io.ReadCloser, error)
	MkdirAll(path string, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (io.WriteCloser, error)
}

// Unix is a file system abstraction for Unix-type file systems.
type Unix struct {
	fsOps fsOperations
}

// NewUnix creates a new initialised Unix structure.
func NewUnix(fsOps fsOperations) *Unix {
	return &Unix{
		fsOps: fsOps,
	}
}

// StreamFileData reads the contents of the file pointed to by fsEntry and streams it to the
// channel the method returns.
// TODO: wrap the dataCh into a io.ReadWriter so to keep the code simple and offer a familiar Go feel.
func (fs *Unix) StreamFileData(ctx context.Context, fsEntry db.FSEntry) (<-chan []byte, <-chan error) {
	dataCh := make(chan []byte, 100)
	errCh := make(chan error)

	go func() {
		defer close(errCh)
		defer close(dataCh)

		f, err := fs.fsOps.Open(filepath.Join(fsEntry.Path, fsEntry.Name))
		if err != nil {
			errCh <- errors.WithStack(err)
			return
		}
		defer func() {
			e := f.Close()
			if e != nil && err == nil {
				errCh <- errors.WithStack(e)
				return
			}
		}()

		br := bufio.NewReader(f)
		data := make([]byte, 1_048_576)

		for {
			n, err := br.Read(data)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					errCh <- errors.WithStack(err)
					return
				}
				break
			}

			dataCh <- data[:n]
		}
	}()

	return dataCh, errCh
}

// MkDir creates a directory.
func (fs *Unix) MkDir(ctx context.Context, path string) error {
	return fs.fsOps.MkdirAll(filepath.Join(path), 0750)
}

// MkFile creates a file with the contents streamed through dataCh.
// TODO: wrap the dataCh into a io.ReadWriter so to keep the code simple and offer a familiar Go feel.
func (fs *Unix) MkFile(ctx context.Context, path string, dataCh <-chan []byte) (err error) {
	f, err := fs.fsOps.OpenFile(path, os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		e := f.Close()
		if e != nil && err == nil {
			err = errors.WithStack(e)
			return
		}
	}()

	bw := bufio.NewWriter(f)

	for data := range dataCh {
		_, err = bw.Write(data)
		if err != nil {
			return
		}
	}

	return errors.WithStack(bw.Flush())
}

// RmDir removes a directory.
func (fs *Unix) RmDir(ctx context.Context, path string) error {
	return os.Remove(path)
}

// RmFile removes a file.
func (fs *Unix) RmFile(ctx context.Context, path string) error {
	return os.Remove(path)
}

// MvDir moves a directory.
func (fs *Unix) MvDir(ctx context.Context, fromPath string, toPath string) error {
	return os.Rename(fromPath, toPath)
}

// MvFile moves a file.
func (fs *Unix) MvFile(ctx context.Context, fromPath string, toPath string) error {
	return os.Rename(fromPath, toPath)
}

// GoFSOperations provides abstractions stubs for some of Go's OS operations.
type GoFSOperations struct{}

// Open is glue code for os.Open.
func (fso *GoFSOperations) Open(name string) (*os.File, error) {
	// nolint: gosec
	return os.Open(name)
}

// MkdirAll is glue code for os.MkdirAll.
func (fso *GoFSOperations) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// OpenFile is glue code for os.OpenFile.
func (fso *GoFSOperations) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	// nolint: gosec
	return os.OpenFile(name, flag, perm)
}
