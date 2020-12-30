package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"seborama/pcloud/sdk"
	"seborama/pcloud/tracker"
	"seborama/pcloud/tracker/db"
	"time"

	"github.com/urfave/cli/v2"
)

func analyse(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   1,
			MaxConnsPerHost:       1,
			ResponseHeaderTimeout: 20 * time.Second,
			Proxy:                 http.ProxyFromEnvironment,
		},
		Timeout: 0,
	}

	pCloudClient := sdk.NewClient(httpClient)

	err := pCloudClient.Login(
		ctx,
		c.String("pcloud-otp-code"),
		sdk.WithGlobalOptionUsername(c.String("pcloud-username")),
		sdk.WithGlobalOptionPassword(c.String("pcloud-password")),
	)
	if err != nil {
		return err
	}

	store, err := db.NewSQLite3(ctx, c.String("db-path"))
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	track, err := tracker.NewTracker(ctx, pCloudClient, store)
	if err != nil {
		return err
	}

	fmt.Println("ListLatestPCloudContents...")
	err = track.ListLatestPCloudContents(ctx, "/pcloud")
	if err != nil {
		return err
	}

	fmt.Println("FindPCloudMutations...")
	fsm, err := track.FindPCloudMutations(ctx)
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(fsm, "", "  ")
	if err != nil {
		return err
	}
	s := 1024
	if len(j) < 1024 {
		s = len(j)
	}
	fmt.Printf("PCloud mutations: count=%d\nFirst few:\n%s\n", len(fsm), string(j[:s]))

	fmt.Println("ListLatestLocalContents...")
	err = track.ListLatestLocalContents(ctx, "/tmp/pcloudLocalFS")
	if err != nil {
		return err
	}

	fmt.Println("FindLocalMutations...")
	fsm, err = track.FindLocalMutations(ctx)
	if err != nil {
		return err
	}

	j, err = json.MarshalIndent(fsm, "", "  ")
	if err != nil {
		return err
	}
	s = 1024
	if len(j) < 1024 {
		s = len(j)
	}
	fmt.Printf("PCloud vs Local mutations: count=%d\nFirst few:\n%s\n", len(fsm), string(j[:s]))

	return nil
}
