package sdk

import (
	"context"
	"fmt"
	"net/url"
)

type ListFolder struct {
	result
	Metadata Metadata
}

type Metadata struct {
	Path     string
	Contents []Contents
}

type Contents struct {
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
	Contents []Contents `json:"contents,omitempty"`

	// File-specific
	FileID      uint64 `json:"fileid,omitempty"`
	Hash        uint64 `json:"hash,omitempty"`
	Category    int32  `json:"category,omitempty"`
	Size        uint64 `json:"size,omitempty"`
	ContentType string `json:"contenttype,omitempty"`
}

func fromBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// ListFolder receives data for a folder.
// Expects folderid or path parameter, returns folder's metadata.
// The metadata will have contents field that is array of metadatas of folder's contents.
// Recursively listing the root folder is not an expensive operation.
// https://docs.pcloud.com/methods/folder/listfolder.html.
func (c *Client) ListFolder(ctx context.Context, path string, folderID uint64, recursiveOpt, showDeletedOpt, noFilesOpt, noSharesOpt bool) (*ListFolder, error) {
	q := url.Values{}
	q.Add("auth", c.auth)
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
