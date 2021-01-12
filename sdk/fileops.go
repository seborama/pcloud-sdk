package sdk

import (
	"context"
	"fmt"
	"net/url"
)

// File contains properties about an opened file, notably the file descriptor FD.
type File struct {
	result
	FD     uint64
	FileID uint64
}

// nolint: golint, stylecheck
const (
	// O_WRITE you do not need to specify O_WRITE even if you intend to write to the file.
	// However that will preform write access control and quota checking and you will
	// get possible errors during open, not at the first write.
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
// https://docs.pcloud.com/methods/fileops/file_open.html
func (c *Client) FileOpen(ctx context.Context, flags uint64, file T4PathOrFileIDOrFolderIDName, opts ...ClientOption) (*File, error) {
	q := toQuery(opts...)
	file(q)

	q.Add("flags", fmt.Sprintf("%d", flags))

	f := &File{}

	err := parseAPIOutput(f)(c.get(ctx, "file_open", q))
	if err != nil {
		return nil, err
	}

	return f, nil
}

// FileDataTransfer is returned by FileWrite and contains the result and number of
// bytes transferred.
type FileDataTransfer struct {
	result
	Bytes uint64
}

// FileWrite writes as much data as you send to the file descriptor fd to the current file
// offset and adjusts the offset.
// You can see how to send data here: https://docs.pcloud.com/methods/fileops/index.html
// https://docs.pcloud.com/methods/fileops/file_write.html
func (c *Client) FileWrite(ctx context.Context, fd uint64, data []byte, opts ...ClientOption) (*FileDataTransfer, error) {
	q := toQuery(opts...)

	q.Add("fd", fmt.Sprintf("%d", fd))

	fdt := &FileDataTransfer{}

	err := parseAPIOutput(fdt)(c.put(ctx, "file_write", q, data))
	if err != nil {
		return nil, err
	}

	return fdt, nil
}

// FileRead tries to read at most count bytes at the current offset of the file.
// If currentofset+count<=filesize this method will satisfy the request and read count bytes,
// otherwise it will return just the bytes available (this is the only way to discover the EOF
// condition).
// You can see how to send data here: https://docs.pcloud.com/methods/fileops/index.html
// https://docs.pcloud.com/methods/fileops/file_read.html
func (c *Client) FileRead(ctx context.Context, fd, count uint64, opts ...ClientOption) ([]byte, error) {
	// TODO - BUG? - For better Golang compliance, EOF should be returned when all data was read
	//        AND EOF has been reached. In the current implementation, EOF is returned when some
	//        data was read but EOF was reached. This means both data and EOF are returned at the
	//        same time. This is not as per os.File.Read()'s specification which returns "0, EOF".
	q := toQuery(opts...)

	q.Add("fd", fmt.Sprintf("%d", fd))
	q.Add("count", fmt.Sprintf("%d", count))

	data, err := c.binget(ctx, "file_read", q)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// FilePRead tries to read at most count bytes at the given offset of the file.
// You can see how to send data here: https://docs.pcloud.com/methods/fileops/index.html
// offset starts at 0.
// https://docs.pcloud.com/methods/fileops/file_pread.html
func (c *Client) FilePRead(ctx context.Context, fd, count, offset uint64, opts ...ClientOption) ([]byte, error) {
	q := toQuery(opts...)

	q.Add("fd", fmt.Sprintf("%d", fd))
	q.Add("count", fmt.Sprintf("%d", count))
	q.Add("offset", fmt.Sprintf("%d", offset))

	data, err := c.binget(ctx, "file_pread", q)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// T5SHA1OrMD5 is a type of parameters that is used by some of the SDK functions that take
// a checksum (SHA1 or MD5).
// Functions that use it have a dichotomic usage to provide a checksum: by SHA1 or MD5.
type T5SHA1OrMD5 func(q url.Values)

// T5SHA1 is a type of T5SHA1OrMD5 that provides a SHA1 checksum.
func T5SHA1(sha1 string) T5SHA1OrMD5 {
	return func(q url.Values) {
		q.Set("sha1", sha1)
	}
}

// T5MD5 is a type of T5SHA1OrMD5 that provides a MD5 checksum.
func T5MD5(md5 string) T5SHA1OrMD5 {
	return func(q url.Values) {
		q.Set("md5", md5)
	}
}

// FilePReadIfMod same as file_pread, but additionally expects sha1 or md5 parameter (hex).
// If the checksum of the data to be read matches the sha1 or md5 checksum, it returns error
// code 6000 Not modified.
// This call is useful if the application has the data cached and wants to verify if it still
// current.
// offset starts at 0.
// You can see how to send data here: https://docs.pcloud.com/methods/fileops/index.html
// https://docs.pcloud.com/methods/fileops/file_pread_ifmod.html
func (c *Client) FilePReadIfMod(ctx context.Context, fd, count, offset uint64, checksum T5SHA1OrMD5, opts ...ClientOption) ([]byte, error) {
	q := toQuery(opts...)
	checksum(q)

	q.Add("fd", fmt.Sprintf("%d", fd))
	q.Add("count", fmt.Sprintf("%d", count))
	q.Add("offset", fmt.Sprintf("%d", offset))

	data, err := c.binget(ctx, "file_pread_ifmod", q)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// PFileChecksum is returned by the SDK FileChecksum method.
// P indicates this is the checksum of a file part, not the entire file.
// The exact amount of data the checksum applies to is indicated by Size.
type PFileChecksum struct {
	result
	SHA1 string
	MD5  string
	Size uint64
}

// FileChecksum calculates checksums of count bytes at offset from the file descripor fd.
//
// DO NOT use this function to calculate checksums of an ENTIRE, UNMODIFIED file,
// USE checksumfile INSTEAD.
//
// Returns sha1, md5 and size.
// size will be equal to count unless bytes past current filesize are requested to be
// checksummed.
//
// You can see how to send data here: https://docs.pcloud.com/methods/fileops/index.html
// https://docs.pcloud.com/methods/fileops/file_checksum.html
func (c *Client) FileChecksum(ctx context.Context, fd, count, offset uint64, opts ...ClientOption) (*PFileChecksum, error) {
	q := toQuery(opts...)

	q.Add("fd", fmt.Sprintf("%d", fd))
	q.Add("count", fmt.Sprintf("%d", count))
	q.Add("offset", fmt.Sprintf("%d", offset))

	pfc := &PFileChecksum{}

	err := parseAPIOutput(pfc)(c.get(ctx, "file_checksum", q))
	if err != nil {
		return nil, err
	}

	return pfc, nil
}

// Whence defines from where an offset applies when seeking a file position.
type Whence int8

const (
	// WhenceFromBeginning moves offset from beginning of the file.
	WhenceFromBeginning Whence = iota

	// WhenceFromCurrent moves offset from current position in file.
	WhenceFromCurrent

	// WhenceFromEnd moves offset from end of file.
	WhenceFromEnd
)

// FileSeek is returned by the SDK FileSeek() method.
type FileSeek struct {
	result
	Offset uint64
}

// FileSeek sets the current offset of the file descriptor to offset bytes.
// This method works in the following modes, depending on the whence parameter:
// whence	Description
// 0        moves after beginning of the file
// 1        after current position
// 2        after end of the file.
// https://docs.pcloud.com/methods/fileops/file_seek.html
func (c *Client) FileSeek(ctx context.Context, fd, offset uint64, whenceOpt Whence, opts ...ClientOption) (*FileSeek, error) {
	q := toQuery(opts...)

	q.Add("fd", fmt.Sprintf("%d", fd))
	q.Add("offset", fmt.Sprintf("%d", offset))
	q.Add("whence", fmt.Sprintf("%d", whenceOpt))

	fs := &FileSeek{}

	err := parseAPIOutput(fs)(c.get(ctx, "file_seek", q))
	if err != nil {
		return nil, err
	}

	return fs, nil
}

// FileClose closes a file descriptor.
// https://docs.pcloud.com/methods/fileops/file_close.html
func (c *Client) FileClose(ctx context.Context, fd uint64, opts ...ClientOption) error {
	q := toQuery(opts...)

	q.Add("fd", fmt.Sprintf("%d", fd))

	f := &result{}

	err := parseAPIOutput(f)(c.get(ctx, "file_close", q))
	if err != nil {
		return err
	}

	return nil
}
