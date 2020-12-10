package sdk

import (
	"context"
	"fmt"
	"net/url"
)

type File struct {
	result
	FD     uint64
	FileID uint64
}

const (
	// O_WRITE you do not need to specify O_WRITE even if you intend to write to the file.
	//However that will preform write access control and quota checking and you will
	//get possible errors during open, not at the first write.
	O_WRITE = 0x0002

	// O_CREAT if is set, file_open will create the file. In this case full "path"
	// or "folderid" and "name" MUST be provided for the new file. If the file already
	// exists the old file will be open unless O_EXCL is set, in which case open will
	// fail.
	// If O_CREAT is not set, than full "path" or "fileid" MUST be provided. The
	// function will fail if the file does not exist.
	O_CREAT = 0x0040

	// O_EXCL when used with O_CREAT, file must not exist
	O_EXCL = 0x0080

	// O_TRUNC will truncate files when opening existing files.
	O_TRUNC = 0x0200

	// O_APPEND if set, will always write to the end of file (unless you
	// use pwrite). That is the only reliable method without race conditions for
	// writing in the end of file when there are multiple writers.
	O_APPEND = 0x0400
)

// FileOpen opens a file descriptor.
// https://docs.pcloud.com/methods/fileops/file_open.html.
func (c *Client) FileOpen(ctx context.Context, flags uint64, pathOpt string, fileIDOpt uint64, folderIDOpt uint64, nameOpt string, opts ...ClientOptions) (*File, error) {
	q := url.Values{}
	q.Add("auth", c.auth)
	q.Add("flags", fmt.Sprintf("%d", flags))
	if pathOpt != "" {
		q.Add("path", pathOpt)
	}
	if fileIDOpt != 0 {
		q.Add("fileid", fmt.Sprintf("%d", fileIDOpt))
	}
	if folderIDOpt != 0 {
		q.Add("folderid", fmt.Sprintf("%d", folderIDOpt))
	}
	if nameOpt != "" {
		q.Add("name", nameOpt)
	}

	f := &File{}

	err := parseAPIOutput(f)(c.request(ctx, "file_open", q))
	if err != nil {
		return nil, err
	}

	return f, nil
}

// FileClose closes a file descriptor.
// https://docs.pcloud.com/methods/fileops/file_close.html.
func (c *Client) FileClose(ctx context.Context, fd uint64, opts ...ClientOptions) error {
	q := url.Values{}
	q.Add("auth", c.auth)
	q.Add("fd", fmt.Sprintf("%d", fd))

	f := &result{}

	err := parseAPIOutput(f)(c.request(ctx, "file_close", q))
	if err != nil {
		return err
	}

	return nil
}
