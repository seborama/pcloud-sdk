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
	Metadata *Metadata
}

// Metadata contains properties related to folder and file information.
type Metadata struct {
	Path string

	// Generic
	Name    string
	Created *APITime

	IsMine bool `json:"ismine"`
	// BEGIN: if IsMine == false
	CanRead   bool `json:"canread,omitempty"`
	CanModify bool `json:"canmodify,omitempty"`
	CanDelete bool `json:"candelete,omitempty"`
	CanCreate bool `json:"cancreate,omitempty"` // for folders only
	// END: if IsMine == false

	Thumb          bool
	Modified       *APITime
	Comments       uint64
	ID             string
	IsShared       bool `json:"isshared"`
	Icon           string
	IsFolder       bool   `json:"isfolder"`
	ParentFolderID uint64 `json:"parentfolderid"`
	IsDeleted      bool   `json:"isdeleted"`     // this may be set by DeleteFile, for instance
	DeletedFileID  uint64 `json:"deletedfileid"` // this may be set by RenameFile, for instance

	// Folder-specific
	FolderID uint64      `json:"folderid,omitempty"`
	Contents []*Metadata `json:"contents,omitempty"`

	// File-specific
	FileID uint64 `json:"fileid,omitempty"`
	Hash   uint64 `json:"hash,omitempty"`
	// Category is one of:
	// 0 - uncategorized
	// 1 - image
	// 2 - video
	// 3 - audio
	// 4 - document
	// 5 - archive
	Category    int32  `json:"category,omitempty"`
	Size        uint64 `json:"size,omitempty"`
	ContentType string `json:"contenttype,omitempty"`

	// optionally, image/video files may have:
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`

	// optionally, audio files may have:
	Artist  string `json:"artist,omitempty"`
	Album   string `json:"album,omitempty"`
	Title   string `json:"title,omitempty"`
	Genre   string `json:"genre,omitempty"`
	TrackNo string `json:"trackno,omitempty"`

	// optionally, video files may have (see also Width and Height in image files above):
	Duration        string `json:"duration,omitempty"`        // duration of the video in seconds (floating point number sent as string)
	FPS             string `json:"fps,omitempty"`             // frames per second rate of the video (floating point number sent as string)
	VideoCodec      string `json:"videocodec,omitempty"`      // codec used for enconding of the video
	AudioCodec      string `json:"audiocodec,omitempty"`      // codec used for enconding of the audio
	VideoBitrate    int    `json:"videobitrate,omitempty"`    // bitrate of the video in kilobits
	AudioBitrate    int    `json:"audiobitrate,omitempty"`    // bitrate of the audio in kilobits
	AudioSamplerate int    `json:"audiosamplerate,omitempty"` // sampling rate of the audio in Hz
	Rotate          int    `json:"rotate,omitempty"`          // indicates that video should be rotated (0, 90, 180 or 270) degrees when playing}
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

// T1PathOrFolderID is a type of parameters that some of the SDK functions take.
// Such functions have a dichotomic usage to reference a folder: either by path or by folderid.
type T1PathOrFolderID func(q url.Values)

// T1FolderByPath is a type of T1PathOrFolderID that references a folder by path alone.
func T1FolderByPath(path string) T1PathOrFolderID {
	return func(q url.Values) {
		q.Set("path", path)
	}
}

// T1FolderByID is a type of T1PathOrFolderID that references a folder by folderid alone.
func T1FolderByID(folderID uint64) T1PathOrFolderID {
	return func(q url.Values) {
		q.Set("folderid", fmt.Sprintf("%d", folderID))
	}
}

// T2PathOrFolderIDName is a type of parameters that some of the SDK functions take.
// Such functions have a dichotomic usage to reference a folder:
// either by path or by folderid+name.
type T2PathOrFolderIDName func(q url.Values)

// T2FolderByPath is a type of T2PathOrFolderIDName that references a folder by path alone.
func T2FolderByPath(path string) T2PathOrFolderIDName {
	return func(q url.Values) {
		q.Set("path", path)
	}
}

// T2FolderByIDName is a type of T2PathOrFolderIDName that references a folder by folderid+name.
func T2FolderByIDName(folderID uint64, name string) T2PathOrFolderIDName {
	return func(q url.Values) {
		q.Set("folderid", fmt.Sprintf("%d", folderID))
		q.Set("name", name)
	}
}

// ToT1PathOrFolderID is a type of parameters that some of the SDK functions take.
// It is similar to T1PathOrFolderID but applies when referencing a destination folder.
type ToT1PathOrFolderID func(q url.Values)

// ToT1FolderByPath is a type of ToT1PathOrFolderID that references a folder by path alone.
func ToT1FolderByPath(path string) ToT1PathOrFolderID {
	return func(q url.Values) {
		q.Set("topath", path)
	}
}

// ToT1FolderByID is a type of ToT1PathOrFolderID that references a folder by folderid alone.
func ToT1FolderByID(folderID uint64) ToT1PathOrFolderID {
	return func(q url.Values) {
		q.Set("tofolderid", fmt.Sprintf("%d", folderID))
	}
}

// ToT2PathOrFolderIDOrFolderIDName is a type of parameters that some of the SDK functions take.
// It applies when referencing a destination folder.
// Functions that use it have a trichotomic usage to reference a folder:
// by path alone, by folderid alone or by folderid+name.
type ToT2PathOrFolderIDOrFolderIDName func(q url.Values)

// ToT2FolderByPath is a type of ToT2PathOrFolderIDOrFolderIDName that references a folder by
// path alone.
func ToT2FolderByPath(path string) ToT2PathOrFolderIDOrFolderIDName {
	return func(q url.Values) {
		q.Set("topath", path)
	}
}

// ToT2FolderByID is a type of ToT2PathOrFolderIDOrFolderIDName that references a folder by
// folderid alone.
func ToT2FolderByID(folderID uint64) ToT2PathOrFolderIDOrFolderIDName {
	return func(q url.Values) {
		q.Set("tofolderid", fmt.Sprintf("%d", folderID))
	}
}

// ToT2FolderByIDName is a type of ToT2PathOrFolderIDOrFolderIDName that references a folder by
// folderid+name.
func ToT2FolderByIDName(folderID uint64, name string) ToT2PathOrFolderIDOrFolderIDName {
	return func(q url.Values) {
		q.Set("tofolderid", fmt.Sprintf("%d", folderID))
		q.Set("toname", name)
	}
}
