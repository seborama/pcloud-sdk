package sdk

import (
	"context"
	"fmt"
	"net/url"
)

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
// https://docs.pcloud.com/methods/streaming/getfilelink.html.
func (c *Client) GetFileLink(ctx context.Context, path string, fileID uint64, forceDownloadOpt bool, contentTypeOpt string, maxSpeedOpt uint64, skipFilenameOpt bool, opts ...ClientOption) (*FileLink, error) {
	q := url.Values{}
	if path != "" {
		q.Add("path", path)
	} else {
		q.Add("fileid", fmt.Sprintf("%d", fileID))
	}
	q.Add("forcedownload", fromBool(forceDownloadOpt))
	q.Add("contenttype", contentTypeOpt)
	if maxSpeedOpt > 0 {
		q.Add("maxspeed", fmt.Sprintf("%d", maxSpeedOpt))
	}
	q.Add("skipfilename", fromBool(skipFilenameOpt))

	fl := &FileLink{}

	err := parseAPIOutput(fl)(c.request(ctx, "getfilelink", q))
	if err != nil {
		return nil, err
	}

	for i, host := range fl.Hosts {
		fl.Hosts[i] = "https://" + host
	}

	return fl, nil
}
