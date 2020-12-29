package cli

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"seborama/pcloud/sdk"
	"strings"

	"github.com/pkg/errors"
)

const PCloudPrefix = "r:"

type sdkClient interface {
	GetFileLink(ctx context.Context, file sdk.T3PathOrFileID, forceDownloadOpt bool, contentTypeOpt string, maxSpeedOpt uint64, skipFilenameOpt bool, opts ...sdk.ClientOption) (*sdk.FileLink, error)
}

type CLI struct {
	pCloudClient sdkClient
	httpClient   *http.Client
}

func NewCLI(pCloudClient sdkClient, httpClient *http.Client) *CLI {
	return &CLI{
		pCloudClient: pCloudClient,
		httpClient:   httpClient,
	}
}

func (cli *CLI) Copy(ctx context.Context, from, to string) error {
	if strings.HasPrefix(from, PCloudPrefix) && !strings.HasPrefix(to, PCloudPrefix) {
		return cli.copyFromPCloudToLocal(ctx, from, to)
	}

	return errors.New("this type of Copy is not yet implemented")
}

// TODO: does not support folders - use pCloud's Archiving methods or folder.ListFolder?
func (cli *CLI) copyFromPCloudToLocal(ctx context.Context, from, to string) error {
	fl, err := cli.pCloudClient.GetFileLink(ctx, sdk.T3FileByPath(from[2:]), true, "", 0, false)
	if err != nil {
		return err
	}
	if len(fl.Hosts) == 0 {
		return errors.New("no hosts available to download the file from pCloud")
	}

	u := fl.Hosts[0] + fl.Path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return errors.Wrap(err, "unable to prepare request to download from pCloud")
	}

	resp, err := cli.httpClient.Do(req)
	if resp != nil {
		defer func() {
			_, err = io.Copy(ioutil.Discard, resp.Body)
			if err != nil {
				fmt.Println("discarding remainder of response body:", err.Error())
			}
			err = resp.Body.Close()
			if err != nil {
				fmt.Println("closing the response body:", err.Error())
			}
		}()
	}
	if err != nil {
		return errors.Wrap(err, "executing HTTP request to download from pCloud")
	}

	// TODO: optimise this to read in chunks
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "reading body of the response to download data from pCloud")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}

	err = ioutil.WriteFile(to, body, 0600)

	return errors.WithStack(err)
}
