package fuse

import (
	"context"
	"log"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
	"github.com/samber/lo"
	"github.com/seborama/pcloud/sdk"
)

type Drive struct {
	fs   fs.FS
	conn *fuse.Conn // TODO: define an interface
}

func Mount(mountpoint string, pcClient *sdk.Client) (Drive, error) {
	conn, err := fuse.Mount(
		mountpoint,
		fuse.FSName("pcloud"),
		fuse.Subtype("seborama"),
	)
	if err != nil {
		log.Fatal(err)
	}

	return Drive{
		fs: &FS{
			pcClient: pcClient,
		},
		conn: conn,
	}, nil
}

func (d *Drive) Unmount() error {
	return d.conn.Close()
}

func (d *Drive) Serve() error {
	return fs.Serve(d.conn, d.fs)
}

// FS implements the pCloud file system.
type FS struct {
	pcClient *sdk.Client // TODO: define an interface
}

// ensure interfaces conpliance
var (
	_ fs.FS = (*FS)(nil)
)

func (fs *FS) Root() (fs.Node, error) {
	log.Println("Root called")
	fsList, err := fs.pcClient.ListFolder(context.Background(), sdk.T1FolderByID(sdk.RootFolderID), false, false, false, false)
	if err != nil {
		return nil, err
	}

	entries := lo.SliceToMap(fsList.Metadata.Contents, func(item *sdk.Metadata) (string, interface{}) {
		if item.IsFolder {
			return item.Name, &Dir{
				Type: fuse.DT_Dir,
				Attributes: fuse.Attr{
					Inode: item.FolderID,
					Atime: item.Modified.Time,
					Mtime: item.Modified.Time,
					Ctime: item.Modified.Time,
					Mode:  os.ModeDir | 0o777,
				},
				Entries: map[string]interface{}{}, // TODO
			}
		}

		return item.Name, &File{
			Type: fuse.DT_File,
			// Content: content, // TODO
			Attributes: fuse.Attr{
				Inode: item.FileID,
				Size:  item.Size,
				Atime: item.Modified.Time,
				Mtime: item.Modified.Time,
				Ctime: item.Modified.Time,
				Mode:  0o777,
			},
		}
	})

	rootDir := &Dir{
		Type: fuse.DT_Dir,
		Attributes: fuse.Attr{
			Inode: sdk.RootFolderID,
			Atime: fsList.Metadata.Modified.Time,
			Mtime: fsList.Metadata.Modified.Time,
			Ctime: fsList.Metadata.Modified.Time,
			Mode:  os.ModeDir | 0o777,
		},
		Entries: entries,
	}

	return rootDir, nil
}

// Dir implements both Node and Handle for the root directory.
type Dir struct {
	Type       fuse.DirentType
	Attributes fuse.Attr
	Entries    map[string]interface{}
}

// ensure interfaces conpliance
var (
	_ fs.Node               = (*Dir)(nil)
	_ fs.NodeStringLookuper = (*Dir)(nil)
)

func (d Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Println("Dir.Attr called")
	log.Println("File.Attr called")
	*a = d.Attributes
	return nil
}

func (d Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("Dir.Lookup called")
	node, ok := d.Entries[name]
	if ok {
		return node.(fs.Node), nil
	}
	return nil, syscall.ENOENT
}

func (d Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("Dir.ReadDirAll called")
	entries := lo.MapToSlice(d.Entries, func(key string, value interface{}) fuse.Dirent {
		switch castEntry := value.(type) {
		case *File:
			return fuse.Dirent{
				Inode: castEntry.Attributes.Inode,
				Type:  castEntry.Type,
				Name:  key,
			}
		case *Dir:
			return fuse.Dirent{
				Inode: castEntry.Attributes.Inode,
				Type:  castEntry.Type,
				Name:  key,
			}
		default:
			log.Printf("unknown directory entry '%T'", castEntry)
			return fuse.Dirent{
				Inode: 6_666_666_666_666_666_666,
				Type:  fuse.DT_Unknown,
				Name:  key,
			}
		}
	})
	return entries, nil
}

// File implements both Node and Handle for the hello file.
type File struct {
	Type       fuse.DirentType
	Content    []byte
	Attributes fuse.Attr
}

// ensure interfaces conpliance
var (
	_ = (fs.Node)((*File)(nil))
	// _ = (fs.HandleWriter)((*File)(nil))
	_ = (fs.HandleReadAller)((*File)(nil))
	// _ = (fs.NodeSetattrer)((*File)(nil))
	// _ = (EntryGetter)((*File)(nil))
)

func (f File) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Println("File.Attr called")
	*a = f.Attributes
	return nil
}

func (File) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte(nil), nil // TODO
}
