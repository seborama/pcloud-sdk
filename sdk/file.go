package sdk

import (
	"context"
	"fmt"
	"net/url"
	"time"
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

// CopyFile takes one file and copies it as another file in the user's filesystem.
// Expects fileid or path to identify the source file and tofolderid+toname or topath to
// identify destination filename.
// If toname is ommited, original filename is used.
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
