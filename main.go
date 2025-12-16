package main

import (
	"context"
	"fmt"
	"gograte/config"
	"gograte/postgres"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {

	cmd := &cli.Command{
		Flags: config.InitiateFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			dbDriver := cmd.String("driver")
			targetDb := cmd.String("target-host")
			database := cmd.String("database")
			user := cmd.String("user")
			sourcedb := cmd.String("source-host")

			validCommand := ValidCommand(dbDriver, targetDb, database, user, sourcedb)
			if validCommand != nil {
				return fmt.Errorf(validCommand.Error())
			}

			password := cmd.String("password") // dont always needs password so define here

			switch dbDriver {
			case "postgres":
				{
					sourceDbConn, err := postgres.ConnetToPostgres(sourcedb, database, user, password)
					if err != nil {
						return fmt.Errorf(err.Error())
					}
					defer sourceDbConn.Close(ctx) // CLOSE THIS FUKCIN THING

					targetDbConn, err := postgres.ConnetToPostgres(targetDb, database, user, password)
					if err != nil {
						return fmt.Errorf(err.Error())
					}
					defer targetDbConn.Close(ctx) // CLOSE THIS FUKCIN THING

					_, err = postgres.ListAllTables(sourceDbConn, ctx)
					if err != nil {
						return fmt.Errorf(err.Error())
					}

					break
				}
			}

			fmt.Println("DONE!")

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
