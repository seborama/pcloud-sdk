package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "pcloud-username",
				EnvVars:  []string{"PCLOUD_USERNAME"},
				Usage:    "pCloud account username",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "pcloud-password",
				EnvVars:  []string{"PCLOUD_PASSWORD"},
				Usage:    "pCloud account password",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "pcloud-otp-code",
				EnvVars: []string{"PCLOUD_OTP_CODE"},
				Usage:   "pCloud account login One-Time-Password (for two-factor authentication)",
			},
		},

		Commands: []*cli.Command{
			{
				Name:    "analyse",
				Aliases: []string{"a"},
				Usage:   "analyse filesystem",
				Action:  analyse,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "db-path",
						EnvVars:  []string{"DB_PATH"},
						Usage:    "Location of the database (it will be created if inexistent)",
						Required: true,
					},
				},
			},
			{
				Name:    "cli",
				Aliases: []string{"a"},
				Usage:   "pCloud CLI",
				Action:  pCLI,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "from",
						Usage:    "Location of source (use prefix 'r:' for pCloud remote)",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "to",
						Usage:    "Location of destination (use prefix 'r:' for pCloud remote)",
						Required: true,
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
