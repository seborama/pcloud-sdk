package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"time"

	"github.com/pkg/errors"
)

// FileResult contains properties about an operation on a file such as:
// DeleteFile, RenameFile, etc.
type FileResult struct {
	result
	ID       string
	Metadata Metadata
}

// DeleteFile deletes a file identified by fileid or path.
// https://docs.pcloud.com/methods/file/deletefile.html
func (c *Client) DeleteFile(ctx context.Context, file T3PathOrFileID, opts ...ClientOption) (*FileResult, error) {
	q := toQuery(opts...)
	file(q)

	r := &FileResult{}

	err := parseAPIOutput(r)(c.get(ctx, "deletefile", q))
	if err != nil {
		return nil, err
	}

	return r, nil
}

// RenameFile renames a file identified by fileid or path.
// Renames (and/or moves) a file identified by fileid or path to either topath (if topath is a
// foldername without new filename it MUST end with slash - /newpath/) or tofolderid/toname
// (one or both can be provided).
// If the destination file already exists it will be replaced atomically with the source file,
// in this case the metadata will include deletedfileid with the fileid of the old file at the
// destination, and the source and destination files revisions will be merged together.
// https://docs.pcloud.com/methods/file/renamefile.html
func (c *Client) RenameFile(ctx context.Context, file T3PathOrFileID, destination ToT3PathOrFolderIDName, opts ...ClientOption) (*FileResult, error) {
	q := toQuery(opts...)
	file(q)
	destination(q)

	r := &FileResult{}

	err := parseAPIOutput(r)(c.get(ctx, "renamefile", q))
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Stat returns information about the file pointed to by fileid or path.
// It's is recomended to use fileid.
// https://docs.pcloud.com/methods/file/stat.html
func (c *Client) Stat(ctx context.Context, file T3PathOrFileID, opts ...ClientOption) (*FileResult, error) {
	q := toQuery(opts...)
	file(q)

	r := &FileResult{}

	err := parseAPIOutput(r)(c.get(ctx, "stat", q))
	if err != nil {
		return nil, err
	}

	return r, nil
}

// CopyFile takes one file and copies it as another file in the user's filesystem.
// Expects fileid or path to identify the source file and tofolderid+toname or topath to
// identify destination filename.
// If toname is omitted, original filename is used.
// The same is true if the last character of topath is '/' (slash), thus identifying only the
// target folder. The target file will be separate, newly created (with current creation time
// unless old file is overwritten) independent file.
// Any future operations on either the source or destination file will not modify the other one.
// This call is useful when you want to create a public link from somebody else's file (shared
// with you).
// If ctime is set, file created time is set. It's required to provide mtime to set ctime.
// https://docs.pcloud.com/methods/file/copyfile.html
func (c *Client) CopyFile(ctx context.Context, file T3PathOrFileID, destination ToT3PathOrFolderIDName, noOverOpt bool, mTime, cTime time.Time, opts ...ClientOption) (*FileResult, error) {
	q := toQuery(opts...)
	file(q)
	destination(q)

	if noOverOpt {
		q.Add("noover", "1")
	}

	if !mTime.IsZero() {
		q.Add("mtime", fmt.Sprintf("%d", mTime.UTC().Unix()))
	}

	if !cTime.IsZero() {
		q.Add("ctime", fmt.Sprintf("%d", cTime.UTC().Unix()))
	}

	r := &FileResult{}

	err := parseAPIOutput(r)(c.get(ctx, "copyfile", q))
	if err != nil {
		return nil, err
	}

	return r, nil
}

// FileChecksum is returned by the SDK FileChecksum() method.
type FileChecksum struct {
	result
	SHA1     string
	MD5      string
	SHA256   string
	Metadata Metadata
}

// ChecksumFile calculates checksums of a given file.
// Note that fileid or path could be used at once (at the same time??).
// Upon success returns metadata.
// sha1 checksum is returned from both US and Europe API servers.
// md5 is returned only from US API servers, not added in Europe as it's quite old and has
// collions.
// sha256 is returned in Europe only.
// https://docs.pcloud.com/methods/file/checksumfile.html
func (c *Client) ChecksumFile(ctx context.Context, file T3PathOrFileID, opts ...ClientOption) (*FileChecksum, error) {
	q := toQuery(opts...)
	file(q)

	fc := &FileChecksum{}

	err := parseAPIOutput(fc)(c.get(ctx, "checksumfile", q))
	if err != nil {
		return nil, err
	}

	return fc, nil
}

// FileUpload is returned by the SDK UploadFile() method.
type FileUpload struct {
	result
	FileIDs   []uint64
	Checksums []*ChecksumSet
	Metadata  []*Metadata
}

// ChecksumSet contains various checksum hashes.
type ChecksumSet struct {
	SHA1   string
	MD5    string
	SHA256 string
}

// UploadFile Upload a file.
// String path or int folderid specify the target directory. If both are omitted the root folder
// is selected.
// Parameter string progresshash can be passed. Same should be passed to uploadprogress method.
// If nopartial is set, partially uploaded files will not be saved (that is when the connection
// breaks before file is read in full). If renameifexists is set, on name conflict, files will
// not be overwritten but renamed to name like filename (2).ext.
// Multiple files can be uploaded, using POST with multipart/form-data encoding. If passed by
// POST, the parameters must come before files. All files are accepted, the name of the form
// field is ignored. Multiple files can come one or more HTML file controls.
// Filenames must be passed as filename property of each file, that is - the way browsers send
// the file names.
// If a file with the same name already exists in the directory, it is overwritten and old one
// is saved as revision. Overwriting a file with the same data does nothing except updating the
// modification time of the file.
//
// files is a map whose keys are filenames (no path as it is specified by `folder`) and values
// are file descriptors to the corresponding files.
// IMPORTANT: the file descriptors should be rewinded to the beginning of the file or only the
// data (if any) from the current position will be uplaoded!
//
// https://docs.pcloud.com/methods/file/uploadfile.html
func (c *Client) UploadFile(ctx context.Context, folder T1PathOrFolderID, files map[string]*os.File, noPartialOpt bool, progressHashOpt string, renameIfExistsOpt bool, mTimeOpt, cTimeOpt time.Time, opts ...ClientOption) (*FileUpload, error) {
	q := toQuery(opts...)
	folder(q)

	if noPartialOpt {
		q.Add("nopartial", "1")
	}

	if progressHashOpt != "" {
		q.Add("progresshash", progressHashOpt)
	}

	if renameIfExistsOpt {
		q.Add("renameifexists", "1")
	}

	if !mTimeOpt.IsZero() {
		q.Add("mtime", fmt.Sprintf("%d", mTimeOpt.UTC().Unix()))
	}

	if !cTimeOpt.IsZero() {
		q.Add("ctime", fmt.Sprintf("%d", cTimeOpt.UTC().Unix()))
	}

	fu := &FileUpload{}

	contentType, data, err := prepareForm(files)
	if err != nil {
		return nil, err
	}

	err = parseAPIOutput(fu)(c.post(ctx, "uploadfile", q, contentType, data))
	if err != nil {
		return nil, err
	}

	return fu, nil
}

// ToT3PathOrFolderIDName is a type of parameters that some of the SDK functions take.
// It applies when referencing a destination folder.
// Functions that use it have a dichotomic usage to reference a folder:
// by path alone or by folderid+name.
type ToT3PathOrFolderIDName func(q url.Values)

// ToT3ByPath is a type of ToT3PathOrFolderIDName that references a folder (must end with '/')
// or afile, by path alone.
func ToT3ByPath(path string) ToT3PathOrFolderIDName {
	return func(q url.Values) {
		q.Set("topath", path)
	}
}

// ToT3ByIDName is a type of ToT3PathOrFolderIDName that references a file by
// folderid+name or, if name is empty, a folder.
func ToT3ByIDName(folderID uint64, name string) ToT3PathOrFolderIDName {
	return func(q url.Values) {
		q.Set("tofolderid", fmt.Sprintf("%d", folderID))
		q.Set("toname", name)
	}
}

func prepareForm(files map[string]*os.File) (string, []byte, error) {
	var b bytes.Buffer

	w := multipart.NewWriter(&b)
	defer func() { _ = w.Close() }()

	for destName, f := range files {
		fw, err := w.CreateFormFile(destName, destName)
		if err != nil {
			return "", nil, errors.WithStack(err)
		}

		_, err = io.Copy(fw, f)
		if err != nil {
			return "", nil, errors.WithStack(err)
		}
	}

	// close the multipart writer to ensure the terminating boundary is written.
	err := w.Close()
	if err != nil {
		return "", nil, errors.WithStack(err)
	}

	return w.FormDataContentType(), b.Bytes(), nil
}
