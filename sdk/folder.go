package sdk

import (
	"context"
	"fmt"
)

// ListFolder contains the properties returned by several folder-operating APIs such as:
// CreeateFolder, DeleteFolder, ListFolder, etc.
type ListFolder struct {
	result
	Metadata Metadata
}

// Metadata contains properties related to folder and file information.
type Metadata struct {
	Path string

	// Generic
	Name           string
	Created        *APITime
	IsMine         bool // When true, there are more fields available. See: https://github.com/pcloudcom/pclouddoc/blob/master/api.txt
	Thumb          bool
	Modified       *APITime
	Comments       uint64
	ID             string
	IsShared       bool
	Icon           string
	IsFolder       bool
	ParentFolderID uint64

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

// fromBool convert a boolean to a value (it may be any) when true or an empty string otherwise.
func fromBool(b bool) string {
	if b {
		return "1"
	}
	return ""
}

// ListFolder receives data for a folder.
// Expects folderid or path parameter, returns folder's metadata.
// The metadata will have contents field that is array of metadatas of folder's contents.
// Recursively listing the root folder is not an expensive operation.
// https://docs.pcloud.com/methods/folder/listfolder.html
func (c *Client) ListFolder(ctx context.Context, path string, folderID uint64, recursiveOpt, showDeletedOpt, noFilesOpt, noSharesOpt bool, opts ...ClientOption) (*ListFolder, error) {
	q := toQuery(opts...)

	if path != "" {
		q.Add("path", path)
	} else {
		q.Add("folderid", fmt.Sprintf("%d", folderID))
	}
	q.Add("recursive", fromBool(recursiveOpt))
	q.Add("showdeleted", fromBool(showDeletedOpt))
	q.Add("nofiles", fromBool(noFilesOpt))
	q.Add("noshares", fromBool(noSharesOpt))

	lf := &ListFolder{}

	err := parseAPIOutput(lf)(c.request(ctx, "listfolder", q))
	if err != nil {
		return nil, err
	}

	return lf, nil
}

// CreateFolder creates a folder.
// Expects either path string parameter (discouraged) or int folderid and string name parameters.
// https://docs.pcloud.com/methods/folder/createfolder.html
func (c *Client) CreateFolder(ctx context.Context, path string, folderID uint64, name string, opts ...ClientOption) (*ListFolder, error) {
	q := toQuery(opts...)

	if path != "" {
		q.Add("path", path)
	} else {
		q.Add("folderid", fmt.Sprintf("%d", folderID))
		q.Add("name", name)
	}

	lf := &ListFolder{}

	err := parseAPIOutput(lf)(c.request(ctx, "createfolder", q))
	if err != nil {
		return nil, err
	}

	return lf, nil
}

// CreateFolderIfNotExists creates a folder if the folder doesn't exist or returns the existing
// folder's metadata.
// Expects either path string parameter (discouraged) or int folderid and string name parameters.
// https://docs.pcloud.com/methods/folder/createfolderifnotexists.html
func (c *Client) CreateFolderIfNotExists(ctx context.Context, path string, folderID uint64, name string, opts ...ClientOption) (*ListFolder, error) {
	q := toQuery(opts...)

	if path != "" {
		q.Add("path", path)
	} else {
		q.Add("folderid", fmt.Sprintf("%d", folderID))
		q.Add("name", name)
	}

	lf := &ListFolder{}

	err := parseAPIOutput(lf)(c.request(ctx, "createfolderifnotexists", q))
	if err != nil {
		return nil, err
	}

	return lf, nil
}

// DeleteFolderRecursive deletes a folder recursively.
// Expects either path string parameter (discouraged) or int folderid parameter.
// Note: This function deletes files, directories, and removes sharing. Use with extreme care.
// https://docs.pcloud.com/methods/folder/deletefolderrecursive.html
func (c *Client) DeleteFolderRecursive(ctx context.Context, path string, folderID uint64, opts ...ClientOption) (*DeleteResult, error) {
	q := toQuery(opts...)

	if path != "" {
		q.Add("path", path)
	} else {
		q.Add("folderid", fmt.Sprintf("%d", folderID))
	}

	dr := &DeleteResult{}

	err := parseAPIOutput(dr)(c.request(ctx, "deletefolderrecursive", q))
	if err != nil {
		return nil, err
	}

	return dr, nil
}

// DeleteFolder deletes a folder.
// Expects either path string parameter (discouraged) or int folderid parameter.
// Note: Folders must be empty before calling deletefolder.
// https://docs.pcloud.com/methods/folder/deletefolder.html
func (c *Client) DeleteFolder(ctx context.Context, path string, folderID uint64, opts ...ClientOption) (*ListFolder, error) {
	q := toQuery(opts...)

	if path != "" {
		q.Add("path", path)
	} else {
		q.Add("folderid", fmt.Sprintf("%d", folderID))
	}

	lf := &ListFolder{}

	err := parseAPIOutput(lf)(c.request(ctx, "deletefolder", q))
	if err != nil {
		return nil, err
	}

	return lf, nil
}

// RenameFolder renames (and/or moves) a folder identified by folderid or path to either
// topath (if topath is an existing folder, to place the source folder without new name for the
// folder it MUST end with slash - /newpath/) or tofolderid/toname (one or both can be provided).
// https://docs.pcloud.com/methods/folder/renamefolder.html
func (c *Client) RenameFolder(ctx context.Context, path string, folderID uint64, toPath string, toFolderID uint64, toName string, opts ...ClientOption) (*ListFolder, error) {
	q := toQuery(opts...)

	if path != "" {
		q.Add("path", path)
	} else {
		q.Add("folderid", fmt.Sprintf("%d", folderID))
	}

	if toPath != "" {
		q.Add("topath", toPath)
	} else {
		q.Add("tofolderid", fmt.Sprintf("%d", toFolderID))
		q.Add("toname", toName)
	}

	lf := &ListFolder{}

	err := parseAPIOutput(lf)(c.request(ctx, "renamefolder", q))
	if err != nil {
		return nil, err
	}

	return lf, nil
}

// CopyFolder copies a folder identified by folderid or path to either topath or tofolderid.
// https://docs.pcloud.com/methods/folder/copyfolder.html
func (c *Client) CopyFolder(ctx context.Context, path string, folderID uint64, toPath string, toFolderID uint64, noOverOpt, skipExisting, copyContentOnly bool, opts ...ClientOption) (*ListFolder, error) {
	q := toQuery(opts...)

	if path != "" {
		q.Add("path", path)
	} else {
		q.Add("folderid", fmt.Sprintf("%d", folderID))
	}

	if toPath != "" {
		q.Add("topath", toPath)
	} else {
		q.Add("tofolderid", fmt.Sprintf("%d", toFolderID))
	}

	lf := &ListFolder{}

	err := parseAPIOutput(lf)(c.request(ctx, "copyfolder", q))
	if err != nil {
		return nil, err
	}

	return lf, nil
}
