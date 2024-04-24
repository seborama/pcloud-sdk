package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/seborama/pcloud-sdk/sdk"
	"github.com/seborama/pcloud-sdk/tracker"
	"github.com/seborama/pcloud-sdk/tracker/db"
	"github.com/seborama/pcloud-sdk/tracker/filesystem"
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

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	pCloudFS := filesystem.NewPCloud(pCloudClient)

	track, err := tracker.NewTracker(ctx, logger, store, pCloudFS, "pcloud")
	if err != nil {
		return err
	}

	fmt.Println("RefreshFSContents...")
	err = track.RefreshFSContents(ctx)
	if err != nil {
		return err
	}

	fmt.Println("ListMutations...")
	fsm, err := track.ListMutations(ctx)
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

	fmt.Println("RefreshFSContents...")
	err = track.RefreshFSContents(ctx)
	if err != nil {
		return err
	}

	fmt.Println("ListMutations...")
	fsm, err = track.ListMutations(ctx)
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

	fmt.Printf("Local mutations: count=%d\nFirst few:\n%s\n", len(fsm), string(j[:s]))

	return nil
}
