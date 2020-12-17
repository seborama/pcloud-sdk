package sdk

import (
	"context"
	"fmt"
	"net/url"
)

// FileLink contains the details of a file link, as provided by GetFileLink.
type FileLink struct {
	result
	Path    string
	Expires APITime
	Hosts   []string
}

// GetFileLink gets a download link for file Takes fileid (or path) as parameter and provides
// links from which the file can be downloaded.
// If the optional parameter forcedownload is set, the file will be served by the content server
// with content type application/octet-stream, which typically forces user agents to save the
// file.
// Alternatively you can provide parameter contenttype with the Content-Type you wish the
// content server to send. If these parameters are not set, the content type will depend on the
// extension of the file.
// Parameter maxspeed may be used if you wish to limit the download speed (in bytes per second)
// for this download.
// Finally you can set skipfilename so the link generated will not include the name of the file.
// https://docs.pcloud.com/methods/streaming/getfilelink.html
func (c *Client) GetFileLink(ctx context.Context, file T3PathOrFileID, forceDownloadOpt bool, contentTypeOpt string, maxSpeedOpt uint64, skipFilenameOpt bool, opts ...ClientOption) (*FileLink, error) {
	q := toQuery(opts...)
	file(q)

	if forceDownloadOpt {
		q.Add("forcedownload", "1")
	}

	if contentTypeOpt != "" {
		q.Add("contenttype", contentTypeOpt)
	}

	if maxSpeedOpt > 0 {
		q.Add("maxspeed", fmt.Sprintf("%d", maxSpeedOpt))
	}

	if skipFilenameOpt {
		q.Add("skipfilename", "1")
	}

	fl := &FileLink{}

	err := parseAPIOutput(fl)(c.get(ctx, "getfilelink", q))
	if err != nil {
		return nil, err
	}

	for i, host := range fl.Hosts {
		fl.Hosts[i] = "https://" + host
	}

	return fl, nil
}

// T3PathOrFileID is a type of parameters that some of the SDK functions take.
// Such functions have a diadic aspect to reference a file: either by path or by fileid.
type T3PathOrFileID func(q url.Values)

// T3FileByPath is a type of T3PathOrFileID that references a file by path alone.
func T3FileByPath(path string) T3PathOrFileID {
	return func(q url.Values) {
		q.Set("path", path)
	}
}

// T3FileByID is a type of T3PathOrFileID that references a file by path alone.
func T3FileByID(fileID uint64) T3PathOrFileID {
	return func(q url.Values) {
		q.Set("fileid", fmt.Sprintf("%d", fileID))
	}
}
