package main

import (
	"context"
	"fmt"
	"gograte/config"
	"gograte/postgres"
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/urfave/cli/v3"
)

func main() {

	cmd := &cli.Command{
		Flags: config.InitiateFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			dbDriver := cmd.String("driver")
			database := cmd.String("database")

			targetDb := cmd.String("target-db")
			targetUser := cmd.String("target-user")
			targetPassword := cmd.String("target-password")
			targetPort := cmd.String("target-port")

			sourcedb := cmd.String("source-db")
			sourceUser := cmd.String("source-user")
			sourcePassword := cmd.String("source-password")
			sourcePort := cmd.String("source-port")

			replace := cmd.String("replace") // optional, will delete this database locally and replace with the sources structure (DEFAULT)

			s := spinner.New(spinner.CharSets[2], 100*time.Millisecond)
			s.Start()
			defer s.Stop()

			switch dbDriver {
			case "postgres":
				{
					s.Suffix = " loading postgres driver"
					sourceDbConn, err := postgres.ConnetToPostgres(sourcedb, database, sourceUser, sourcePassword, sourcePort)
					if err != nil {
						return fmt.Errorf(err.Error())
					}
					defer sourceDbConn.Close(ctx) // CLOSE THIS FUKCIN THING

					targetDbConn, err := postgres.ConnetToPostgres(targetDb, database, targetUser, targetPassword, targetPort)
					if err != nil {
						return fmt.Errorf(err.Error())
					}
					defer targetDbConn.Close(ctx) // CLOSE THIS FUKCIN THING

					startTime := time.Now()

					s.Suffix = " getting table details"

					// get the number of tables for both the source and target database
					_, err = postgres.GetTables(sourceDbConn, ctx)
					if err != nil {
						return fmt.Errorf(err.Error())
					}
					_, err = postgres.GetTables(targetDbConn, ctx)
					if err != nil {
						return fmt.Errorf(err.Error())
					}

					if replace == "true" {
						err := postgres.ReplaceDatabase(sourceDbConn, targetDbConn, database, ctx, s)
						if err != nil {
							return fmt.Errorf(err.Error())
						}
					} else {
						return fmt.Errorf("must provide replace")
					}

					fmt.Printf("\nFinished in %v seconds", time.Since(startTime))
					break
				}
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
