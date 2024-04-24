package main

import (
	"context"
	"log"
	"net/http"
	"time"

	ucli "github.com/urfave/cli/v2"

	"github.com/seborama/pcloud/fuse"
	"github.com/seborama/pcloud/sdk"
)

func drive(c *ucli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sdkHTTPClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   2,
			MaxConnsPerHost:       10,
			ResponseHeaderTimeout: 20 * time.Second,
			Proxy:                 http.ProxyFromEnvironment,
		},
		Timeout: 0,
	}

	pCloudClient := sdk.NewClient(sdkHTTPClient)

	log.Println("Logging into pCloud")
	err := pCloudClient.Login(
		ctx,
		c.String("pcloud-otp-code"),
		sdk.WithGlobalOptionUsername(c.String("pcloud-username")),
		sdk.WithGlobalOptionPassword(c.String("pcloud-password")),
	)
	if err != nil {
		return err
	}

	log.Println("Creating drive")
	drive, err := fuse.NewDrive(
		c.String("mount-point"),
		pCloudClient,
	)
	if err != nil {
		panic(err)
	}
	defer func() { _ = drive.Unmount() }()

	log.Println("Mouting FS at", c.String("mount-point"))
	err = drive.Mount()
	if err != nil {
		panic(err)
	}

	return nil
}
