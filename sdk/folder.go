package sdk

import (
	"context"
	"fmt"
	"net/url"
)

// RootFolderID is the folderID of the root folder (i.e. '/').
const RootFolderID = uint64(0)

// FSList contains a file system list, i.e. the properties returned by several folder or file
// operating APIs such as:
// CreeateFolder, DeleteFolder, ListFolder, etc.
type FSList struct {
	result
	Metadata Metadata
}

// Metadata contains properties related to folder and file information.
type Metadata struct {
	Path string

	// Generic
	Name           string
	Created        *APITime
	IsMine         bool // TODO: when true, there are more fields available. See: https://github.com/pcloudcom/pclouddoc/blob/master/api.txt
	Thumb          bool
	Modified       *APITime
	Comments       uint64
	ID             string
	IsShared       bool
	Icon           string
	IsFolder       bool
	ParentFolderID uint64
	IsDeleted      bool   // this may be set by DeleteFile, for instance
	DeletedFileID  uint64 // this may be set by RenameFile, for instance

	// Folder-specific
	FolderID uint64     `json:"folderid,omitempty"`
	Contents []Metadata `json:"contents,omitempty"`

	// File-specific
	FileID      uint64 `json:"fileid,omitempty"`
	Hash        uint64 `json:"hash,omitempty"`
	Category    int32  `json:"category,omitempty"`
	Size        uint64 `json:"size,omitempty"`
	ContentType string `json:"contenttype,omitempty"`
}

// DeleteResult contains the properties returned by DeleteFolderRecursive.
type DeleteResult struct {
	result
	DeletedFiles   uint64
	DeletedFolders uint64
}

// ListFolder receives data for a folder.
// Expects folderid or path parameter, returns folder's metadata.
// The metadata will have contents field that is array of metadatas of folder's contents.
// Recursively listing the root folder is not an expensive operation.
// https://docs.pcloud.com/methods/folder/listfolder.html
func (c *Client) ListFolder(ctx context.Context, folder T1PathOrFolderID, recursiveOpt, showDeletedOpt, noFilesOpt, noSharesOpt bool, opts ...ClientOption) (*FSList, error) {
	q := toQuery(opts...)
	folder(q)

	if recursiveOpt {
		q.Add("recursive", "1")
	}

	if showDeletedOpt {
		q.Add("showdeleted", "1")
	}

	if noFilesOpt {
		q.Add("nofiles", "1")
	}

	if noSharesOpt {
		q.Add("noshares", "1")
	}

	lf := &FSList{}

	err := parseAPIOutput(lf)(c.get(ctx, "listfolder", q))
	if err != nil {
		return nil, err
	}

	return lf, nil
}

// CreateFolder creates a folder.
// Expects either path string parameter (discouraged) or int folderid and string name parameters.
// https://docs.pcloud.com/methods/folder/createfolder.html
func (c *Client) CreateFolder(ctx context.Context, folder T2PathOrFolderIDName, opts ...ClientOption) (*FSList, error) {
	q := toQuery(opts...)
	folder(q)

	lf := &FSList{}

	err := parseAPIOutput(lf)(c.get(ctx, "createfolder", q))
	if err != nil {
		return nil, err
	}

	return lf, nil
}

// CreateFolderIfNotExists creates a folder if the folder doesn't exist or returns the existing
// folder's metadata.
// Expects either path string parameter (discouraged) or int folderid and string name parameters.
// https://docs.pcloud.com/methods/folder/createfolderifnotexists.html
func (c *Client) CreateFolderIfNotExists(ctx context.Context, folder T2PathOrFolderIDName, opts ...ClientOption) (*FSList, error) {
	q := toQuery(opts...)
	folder(q)

	lf := &FSList{}

	err := parseAPIOutput(lf)(c.get(ctx, "createfolderifnotexists", q))
	if err != nil {
		return nil, err
	}

	return lf, nil
}

// DeleteFolderRecursive deletes a folder recursively.
// Expects either path string parameter (discouraged) or int folderid parameter.
// Note: This function deletes files, directories, and removes sharing. Use with extreme care.
// https://docs.pcloud.com/methods/folder/deletefolderrecursive.html
func (c *Client) DeleteFolderRecursive(ctx context.Context, folder T1PathOrFolderID, opts ...ClientOption) (*DeleteResult, error) {
	q := toQuery(opts...)
	folder(q)

	dr := &DeleteResult{}

	err := parseAPIOutput(dr)(c.get(ctx, "deletefolderrecursive", q))
	if err != nil {
		return nil, err
	}

	return dr, nil
}

// DeleteFolder deletes a folder.
// Expects either path string parameter (discouraged) or int folderid parameter.
// Note: Folders must be empty before calling deletefolder.
// https://docs.pcloud.com/methods/folder/deletefolder.html
func (c *Client) DeleteFolder(ctx context.Context, folder T1PathOrFolderID, opts ...ClientOption) (*FSList, error) {
	q := toQuery(opts...)
	folder(q)

	lf := &FSList{}

	err := parseAPIOutput(lf)(c.get(ctx, "deletefolder", q))
	if err != nil {
		return nil, err
	}

	return lf, nil
}

// RenameFolder renames (and/or moves) a folder identified by folderid or path to either
// topath (if topath is an existing folder, to place the source folder without new name for the
// folder it MUST end with slash - /newpath/) or tofolderid/toname (one or both can be provided).
// https://docs.pcloud.com/methods/folder/renamefolder.html
func (c *Client) RenameFolder(ctx context.Context, folder T1PathOrFolderID, toFolder ToT2PathOrFolderIDOrFolderIDName, opts ...ClientOption) (*FSList, error) {
	q := toQuery(opts...)
	folder(q)
	toFolder(q)

	lf := &FSList{}

	err := parseAPIOutput(lf)(c.get(ctx, "renamefolder", q))
	if err != nil {
		return nil, err
	}

	return lf, nil
}

// CopyFolder copies a folder identified by folderid or path to either topath or tofolderid.
// https://docs.pcloud.com/methods/folder/copyfolder.html
func (c *Client) CopyFolder(ctx context.Context, folder T1PathOrFolderID, toFolder ToT1PathOrFolderID, noOverOpt, skipExisting, copyContentOnly bool, opts ...ClientOption) (*FSList, error) {
	q := toQuery(opts...)
	folder(q)
	toFolder(q)

	lf := &FSList{}

	err := parseAPIOutput(lf)(c.get(ctx, "copyfolder", q))
	if err != nil {
		return nil, err
	}

	return lf, nil
}

type T1PathOrFolderID func(q url.Values)

func T1FolderByPath(path string) T1PathOrFolderID {
	return func(q url.Values) {
		q.Set("path", path)
	}
}

func T1FolderByID(folderID uint64) T1PathOrFolderID {
	return func(q url.Values) {
		q.Set("folderid", fmt.Sprintf("%d", folderID))
	}
}

type T2PathOrFolderIDName func(q url.Values)

func T2FolderByPath(path string) T2PathOrFolderIDName {
	return func(q url.Values) {
		q.Set("path", path)
	}
}

func T2FolderByIDName(folderID uint64, name string) T2PathOrFolderIDName {
	return func(q url.Values) {
		q.Set("folderid", fmt.Sprintf("%d", folderID))
		q.Set("name", name)
	}
}

type ToT1PathOrFolderID func(q url.Values)

func ToT1FolderByPath(path string) ToT1PathOrFolderID {
	return func(q url.Values) {
		q.Set("topath", path)
	}
}

func ToT1FolderByID(folderID uint64) ToT1PathOrFolderID {
	return func(q url.Values) {
		q.Set("tofolderid", fmt.Sprintf("%d", folderID))
	}
}

type ToT2PathOrFolderIDOrFolderIDName func(q url.Values)

func ToT2FolderByPath(path string) ToT2PathOrFolderIDOrFolderIDName {
	return func(q url.Values) {
		q.Set("topath", path)
	}
}

func ToT2FolderByID(folderID uint64) ToT2PathOrFolderIDOrFolderIDName {
	return func(q url.Values) {
		q.Set("tofolderid", fmt.Sprintf("%d", folderID))
	}
}

func ToT2FolderByIDName(folderID uint64, name string) ToT2PathOrFolderIDOrFolderIDName {
	return func(q url.Values) {
		q.Set("tofolderid", fmt.Sprintf("%d", folderID))
		q.Set("toname", name)
	}
}
