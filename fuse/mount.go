package fuse

import (
	"context"
	"log"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"

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

func Mount(mountpoint string, pcClient *sdk.Client) (*Drive, error) {
	conn, err := fuse.Mount(
		mountpoint,
		fuse.FSName("pcloud"),
		fuse.Subtype("seborama"),
	)
	if err != nil {
		return nil, err
	}

	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	uid, err := strconv.ParseUint(user.Uid, 10, 32)
	if err != nil {
		return nil, err
	}
	gid, err := strconv.ParseUint(user.Gid, 10, 32)
	if err != nil {
		return nil, err
	}

	return &Drive{
		fs: &FS{
			pcClient:  pcClient,
			uid:       uint32(uid),
			gid:       uint32(gid),
			rdev:      531,
			dirPerms:  0o750,
			filePerms: 0o640,
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
	pcClient  *sdk.Client // TODO: define an interface
	uid       uint32
	gid       uint32
	rdev      uint32
	dirPerms  os.FileMode
	filePerms os.FileMode
}

// ensure interfaces conpliance
var (
	_ fs.FS = (*FS)(nil)
)

func (fs *FS) Root() (fs.Node, error) {
	log.Println("Root called")

	rootDir := &Dir{
		Type: fuse.DT_Dir,
		fs:   fs,
	}

	err := rootDir.materialiseFolder(context.Background())
	if err != nil {
		return nil, err
	}

	return rootDir, nil
}

// Dir implements both Node and Handle for the root directory.
type Dir struct {
	Type       fuse.DirentType
	Attributes fuse.Attr

	// TODO: we must be able to find something better than interface{}, either a proper interface or perhaps a generic type
	// TODO: we likely don't need this: we should always call `materialiseFolder()` because the source of truth is pCloud
	// TODO: contents is subject to changes at anytime, and we should allow the fuse driver to be the judge of whether to
	// TODO: ... refresh the folder or not via fuse.Attr.Validate
	Entries map[string]interface{}

	fs             *FS
	parentFolderID uint64
	folderID       uint64
}

// ensure interfaces conpliance
var (
	_ fs.Node               = (*Dir)(nil)
	_ fs.NodeStringLookuper = (*Dir)(nil)
	_ fs.HandleReadDirAller = (*Dir)(nil)
)

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Println("Dir.Attr called")
	*a = d.Attributes
	return nil
}

// TODO: add support for . and ..
func (d *Dir) materialiseFolder(ctx context.Context) error {
	fsList, err := d.fs.pcClient.ListFolder(ctx, sdk.T1FolderByID(d.folderID), false, false, false, false)
	if err != nil {
		return err
	}

	// TODO: is this necessary? perhaps only for the root folder?
	d.Attributes = fuse.Attr{
		Valid: time.Second,
		Inode: d.folderID,
		Atime: fsList.Metadata.Modified.Time,
		Mtime: fsList.Metadata.Modified.Time,
		Ctime: fsList.Metadata.Modified.Time,
		Mode:  os.ModeDir | d.fs.dirPerms,
		Nlink: 1, // TODO: is that right? How else can we find this value?
		Uid:   d.fs.uid,
		Gid:   d.fs.gid,
		Rdev:  d.fs.rdev,
	}

	d.parentFolderID = fsList.Metadata.ParentFolderID
	d.folderID = fsList.Metadata.FolderID

	entries := lo.SliceToMap(fsList.Metadata.Contents, func(item *sdk.Metadata) (string, interface{}) {
		if item.IsFolder {
			return item.Name, &Dir{
				Type: fuse.DT_Dir,
				Attributes: fuse.Attr{
					Valid: time.Second,
					Inode: item.FolderID,
					Atime: item.Modified.Time,
					Mtime: item.Modified.Time,
					Ctime: item.Modified.Time,
					Mode:  os.ModeDir | d.fs.dirPerms,
					Nlink: 1, // the official pCloud client can show other values that 1 - dunno how
					Uid:   d.fs.uid,
					Gid:   d.fs.gid,
					Rdev:  d.fs.rdev,
				},
				Entries:        nil, // will be populated by Dir.Lookup
				fs:             d.fs,
				parentFolderID: item.ParentFolderID,
				folderID:       item.FolderID,
			}
		}

		return item.Name, &File{
			Type: fuse.DT_File,
			// Content: content, // TODO
			Attributes: fuse.Attr{
				Valid: time.Second,
				Inode: item.FileID,
				Size:  item.Size,
				Atime: item.Modified.Time,
				Mtime: item.Modified.Time,
				Ctime: item.Modified.Time,
				Mode:  d.fs.filePerms,
				Nlink: 1, // TODO: is that right? How else can we find this value?
				Uid:   d.fs.uid,
				Gid:   d.fs.gid,
				Rdev:  d.fs.rdev,
			},
		}
	})

	d.Entries = entries

	return nil
}

// Lookup looks up a specific entry in the receiver,
// which must be a directory.  Lookup should return a Node
// corresponding to the entry.  If the name does not exist in
// the directory, Lookup should return ENOENT.
//
// Lookup need not to handle the names "." and "..".
func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("Dir.Lookup called on dir folderID:", d.folderID, "entries count:", len(d.Entries), "- with name:", name)
	// TODO: this test is likely incorrect: we should always list entries in case the folder has changed
	// TODO: ...at the very least, we should combine it with a TTL or simply rely on the fuse driver to manage that for us.
	if len(d.Entries) == 0 {
		// TODO: we can do better here: all this function wants is to get a single entry, not everything
		d.materialiseFolder(ctx)
	}

	node, ok := d.Entries[name]
	if ok {
		return node.(fs.Node), nil
	}

	return nil, syscall.ENOENT
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("Dir.ReadDirAll called - folderID:", d.folderID, "-", "parentFolderID:", d.parentFolderID)
	d.materialiseFolder(ctx) // TODO: this should not be required here

	dirEntries := lo.MapToSlice(d.Entries, func(key string, value interface{}) fuse.Dirent {
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
			log.Printf("unknown directory entry type '%T'", castEntry)
			return fuse.Dirent{
				Inode: 6_666_666_666_666_666_666,
				Type:  fuse.DT_Unknown,
				Name:  key,
			}
		}
	})

	return dirEntries, nil
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
