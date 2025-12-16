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
			host := cmd.String("host")
			database := cmd.String("database")
			user := cmd.String("user")

			validCommand := ValidCommand(dbDriver, host, database, user)
			if validCommand != nil {
				return fmt.Errorf(validCommand.Error())
			}

			password := cmd.String("password") // dont always needs password so define here

			switch dbDriver {
			case "postgres":
				{
					conn, err := postgres.ConnetToPostgres(host, database, user, password)
					if err != nil {
						return fmt.Errorf(err.Error())
					}

					// CLOSE THIS FUKCIN THING
					defer conn.Close(ctx)

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
