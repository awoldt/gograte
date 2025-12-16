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
			database := cmd.String("database")

			targetDb := cmd.String("target-db")
			targetUser := cmd.String("target-user")
			targetPassword := cmd.String("target-password")
			targetPort := cmd.String("target-port")

			sourcedb := cmd.String("source-db")
			sourceUser := cmd.String("source-user")
			sourcePassword := cmd.String("source-password")
			sourcePort := cmd.String("source-port")

			fmt.Printf(`
		Config used:
		- database driver -> "%v"
		- database -> "%v"
		- target-db -> "%v"
		- target-user -> "%v"
		- target-password -> "%v"
		- target-port -> "%v"
		- source-db -> "%v"
		- source-user -> "%v"
		- source-password -> "%v"
		- source-port -> "%v"
		`, dbDriver, database, targetDb, targetUser, targetPassword, targetPort, sourcedb, sourceUser, sourcePassword, sourcePort)

			switch dbDriver {
			case "postgres":
				{
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

					tables, err := postgres.GetTables(sourceDbConn, ctx)
					if err != nil {
						return fmt.Errorf(err.Error())
					}

					fmt.Printf("there are %v tables in the source database", len(tables))

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
