package fuse_test

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	pfuse "github.com/seborama/pcloud/fuse"
	"github.com/seborama/pcloud/sdk"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	pcClient := newPCloudClient(t)

	mountpoint := "/tmp/pcloud_mnt"

	drive, err := pfuse.Mount(
		mountpoint,
		pcClient,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = drive.Unmount() }()

	log.Println("Mouting FS")
	err = drive.Serve()
	if err != nil {
		log.Fatal(err)
	}
}

func newPCloudClient(t *testing.T) *sdk.Client {
	t.Helper()

	username := os.Getenv("GO_PCLOUD_USERNAME")
	require.NotEmpty(t, username)

	password := os.Getenv("GO_PCLOUD_PASSWORD")
	require.NotEmpty(t, password)

	otpCode := os.Getenv("GO_PCLOUD_TFA_CODE")

	c := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   1,
			MaxConnsPerHost:       1,
			ResponseHeaderTimeout: 20 * time.Second,
			// Proxy:           http.ProxyFromEnvironment,
			// TLSClientConfig: &tls.Config{
			// InsecureSkipVerify: true, // only use this for debugging environments
			// },
		},
		Timeout: 0,
	}

	pcc := sdk.NewClient(c)

	err := pcc.Login(
		context.Background(),
		otpCode,
		sdk.WithGlobalOptionUsername(username),
		sdk.WithGlobalOptionPassword(password),
		sdk.WithGlobalOptionAuthInactiveExpire(5*time.Minute),
	)
	require.NoError(t, err)

	return pcc

}
