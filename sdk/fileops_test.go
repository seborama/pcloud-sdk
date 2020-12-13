package sdk_test

import (
	"context"
	"net/http"
	"net/url"
	"seborama/pcloud/sdk"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_Exploratory(t *testing.T) {
	c := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:          1,
			MaxIdleConnsPerHost:   1,
			DisableKeepAlives:     false,
			MaxConnsPerHost:       1,
			IdleConnTimeout:       60 * time.Second,
			ResponseHeaderTimeout: 15 * time.Second,
		},
		Timeout: 10 * time.Second,
	}

	u, err := url.Parse("https://eapi.pcloud.com")
	require.NoError(t, err)

	q, err := url.ParseQuery("")
	require.NoError(t, err)

	u.RawQuery = q.Encode()

	r, err := c.Get(u.String())
	require.NoError(t, err)
	_ = r
}

func Test_FileOps_ByPath(t *testing.T) {
	folderPath := "/go_pCloud_" + uuid.New().String()
	fileName := "go_pCloud_" + uuid.New().String() + ".txt"

	_, err := pcc.CreateFolder(context.Background(), folderPath, 0, "")
	require.NoError(t, err)

	f, err := pcc.FileOpen(context.Background(), sdk.O_CREAT|sdk.O_EXCL, folderPath+"/"+fileName, 0, 0, "")
	require.NoError(t, err)

	fdt, err := pcc.FileWrite(context.Background(), f.FD, []byte(Lipsum))
	require.NoError(t, err)
	require.EqualValues(t, len(Lipsum), fdt.Bytes)

	err = pcc.FileClose(context.Background(), f.FD)
	require.NoError(t, err)

	// TODO: add FileDelete (when available)

	_, err = pcc.DeleteFolderRecursive(context.Background(), folderPath, 0)
	require.NoError(t, err)
}

func Test_FileOpen_FileClose_ByFileID(t *testing.T) {
	t.Skip() // not yet written
}
