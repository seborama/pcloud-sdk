package main

import (
	"context"
	"net/http"
	"time"

	ucli "github.com/urfave/cli/v2"

	pcli "github.com/seborama/pcloud-sdk/cli"
	"github.com/seborama/pcloud-sdk/sdk"
)

func pCLI(c *ucli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sdkHTTPClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   1,
			MaxConnsPerHost:       1,
			ResponseHeaderTimeout: 20 * time.Second,
			Proxy:                 http.ProxyFromEnvironment,
		},
		Timeout: 0,
	}

	pCloudClient := sdk.NewClient(sdkHTTPClient)

	err := pCloudClient.Login(
		ctx,
		c.String("pcloud-otp-code"),
		sdk.WithGlobalOptionUsername(c.String("pcloud-username")),
		sdk.WithGlobalOptionPassword(c.String("pcloud-password")),
	)
	if err != nil {
		return err
	}

	cliHTTPClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   1,
			MaxConnsPerHost:       1,
			ResponseHeaderTimeout: 20 * time.Second,
			Proxy:                 http.ProxyFromEnvironment,
		},
		Timeout: 0,
	}

	pCli := pcli.NewCLI(pCloudClient, cliHTTPClient)

	err = pCli.Copy(ctx, c.String("from"), c.String("to"))
	if err != nil {
		return err
	}

	return nil
}
