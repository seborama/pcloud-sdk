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
			dirPerms:  0o750,
			filePerms: 0o640,
			dirValid:  2 * time.Second,
			fileValid: time.Second,
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
	dirPerms  os.FileMode
	filePerms os.FileMode
	dirValid  time.Duration
	fileValid time.Duration
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
		Valid: d.fs.dirValid,
		Inode: d.folderID,
		Atime: fsList.Metadata.Modified.Time,
		Mtime: fsList.Metadata.Modified.Time,
		Ctime: fsList.Metadata.Modified.Time,
		Mode:  os.ModeDir | d.fs.dirPerms,
		Nlink: 1, // TODO: is that right? How else can we find this value?
		Uid:   d.fs.uid,
		Gid:   d.fs.gid,
	}
	d.parentFolderID = fsList.Metadata.ParentFolderID
	d.folderID = fsList.Metadata.FolderID

	entries := lo.SliceToMap(fsList.Metadata.Contents, func(item *sdk.Metadata) (string, interface{}) {
		if item.IsFolder {
			return item.Name, &Dir{
				Type: fuse.DT_Dir,
				Attributes: fuse.Attr{
					Valid: d.fs.dirValid,
					Inode: item.FolderID,
					Atime: item.Modified.Time,
					Mtime: item.Modified.Time,
					Ctime: item.Modified.Time,
					Mode:  os.ModeDir | d.fs.dirPerms,
					Nlink: 1, // the official pCloud client can show other values that 1 - dunno how
					Uid:   d.fs.uid,
					Gid:   d.fs.gid,
				},
				Entries:        nil, // will be populated upon access by Dir.Lookup or Dir.ReadDirAll
				fs:             d.fs,
				parentFolderID: item.ParentFolderID,
				folderID:       item.FolderID,
			}
		}

		return item.Name, &File{
			Type: fuse.DT_File,
			// Content: content, // TODO
			Attributes: fuse.Attr{
				Valid: d.fs.fileValid,
				Inode: item.FileID,
				Size:  item.Size,
				Atime: item.Modified.Time,
				Mtime: item.Modified.Time,
				Ctime: item.Modified.Time,
				Mode:  d.fs.filePerms,
				Nlink: 1, // TODO: is that right? How else can we find this value?
				Uid:   d.fs.uid,
				Gid:   d.fs.gid,
			},
			fs:       d.fs,
			folderID: item.FolderID,
			fileID:   item.FileID,
			file:     nil,
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
	fs         *FS
	folderID   uint64 // TODO: not needed??
	fileID     uint64
	file       *sdk.File
}

// ensure interfaces conpliance
var (
	_ = (fs.Node)((*File)(nil))
	// _ = (fs.HandleWriter)((*File)(nil))
	// _ = (fs.HandleReadAller)((*File)(nil)) // NOTE: it's best avoiding to implement this method to avoid costly memory operations with large files.
	_ = (fs.HandleReader)((*File)(nil))
	_ = (fs.NodeOpener)((*File)(nil))
	_ = (fs.HandleFlusher)((*File)(nil))
	_ = (fs.HandleReleaser)((*File)(nil))
	// _ = (fs.NodeSetattrer)((*File)(nil))
	// _ = (EntryGetter)((*File)(nil))
)

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Println("File.Attr called")
	*a = f.Attributes
	return nil
}

type fileHandle struct {
	file *sdk.File
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Println("File.Open - 1 - req.ID", req.ID, " - resp.Handle:", resp.Handle)

	// TODO: this is temporary - for now, let's focus on a read-only implementation
	if !req.Flags.IsReadOnly() {
		return nil, fuse.Errno(syscall.EACCES)
	}

	file, err := f.fs.pcClient.FileOpen(ctx, 0, sdk.T4FileByID(f.fileID))
	if err != nil {
		return nil, err
	}

	f.file = file
	log.Println("File.Open - 2 - req.ID", req.ID, " - file.FD:", file.FD)
	resp.Flags |= fuse.OpenKeepCache

	return &File{
		Type:       f.Type,
		Attributes: f.Attributes,
		fs:         f.fs,
		folderID:   f.folderID,
		fileID:     f.fileID,
		file:       file,
	}, nil
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	log.Println("File.Read called - req.ID", req.ID, " - req.Handle:", req.Handle)
	if f.file == nil {
		return syscall.ENOENT
	}

	data, err := f.fs.pcClient.FilePRead(ctx, f.file.FD, uint64(req.Size), uint64(req.Offset))
	if err != nil {
		return err
	}
	resp.Data = data

	return nil
}

// Flush is called each time the file or directory is closed.
// Because there can be multiple file descriptors referring to a
// single opened file, Flush can be called multiple times.
func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	log.Println("File.Flush called - req.ID", req.ID, " - req.Handle:", req.Handle)
	if f.file != nil {
		err := f.fs.pcClient.FileClose(ctx, f.file.FD)
		f.file = nil
		return err
	}

	return nil
}

// A ReleaseRequest asks to release (close) an open file handle.
func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	if f.file != nil {
		err := f.fs.pcClient.FileClose(ctx, f.file.FD)
		f.file = nil
		return err
	}

	return nil
}
