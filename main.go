package main

import (
	"context"
	"fmt"
	"gograte/config"
	"log"
	"os"
	"slices"

	"github.com/urfave/cli/v3"
)

func main() {

	cmd := &cli.Command{
		Flags: config.InitiateFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// make sure its a legit db driver
			dbDriver := cmd.String("driver")
			if dbDriver == "" || !slices.Contains(config.SupportedDatabases, dbDriver) {
				return fmt.Errorf("invalid driver")
			}

			return nil

		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
