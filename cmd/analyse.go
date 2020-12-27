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
	err = track.ListLatestPCloudContents(ctx)
	if err != nil {
		return err
	}

	fmt.Println(`
	TODO: shifting state from New to Previous (and hence drop the 'old' Previous) should ONLY
	ONLY happen when the analysis has completed and sync completed successfully.
	The "Previous" state MUST NOT be dropped until it has been fully sync'ed with a newer state.
	Failing to follow this rule would lead an incomplete sync and loss of data"`)

	fmt.Println("FindPCloudMutations...")
	fsm, err := track.FindPCloudMutations(ctx)
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(fsm, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("PCloud mutations: count=%d\nFirst few:\n%s\n", len(fsm), string(j[:512]))

	return nil
}
