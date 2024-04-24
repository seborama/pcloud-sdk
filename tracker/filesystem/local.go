package filesystem

import (
	"bufio"
	"context"

	// nolint:gosec
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/seborama/pcloud-sdk/tracker/archos"
	"github.com/seborama/pcloud-sdk/tracker/db"
)

// Local is a file system abstraction for a local file system.
type Local struct{}

// NewLocal creates a new initialised Local structure.
func NewLocal() *Local {
	return &Local{}
}

// Walk traverses the file system entries and writes each entry to fsEntriesCh.
// It must check for an error in errCh (which indicates the receiver of fsEntriesCh encountered
// a problem and terminate if one is present.
// Walk is the PRODUCER on fsEntriesCh and IS RESPONSIBLE FOR CLOSING IT!!
// nolint: gocognit, gocyclo
func (fs *Local) Walk(ctx context.Context, fsName db.FSName, path string, fsEntriesCh chan<- db.FSEntry, errCh <-chan error) error {
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
					FSName:         fsName,
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

	return err
}

func hashFileData(path string) (string, error) {
	// nolint: gosec
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	return hashData(bufio.NewReader(f))
}

func hashData(r io.Reader) (string, error) {
	// nolint: gosec
	cs := sha1.New()

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
