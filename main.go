package main

import (
	"context"
	"fmt"
	"gograte/config"
	"gograte/postgres"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"
)

func main() {
	// flags passed from terminal will override env set flags
	// you can mix and match the two
	godotenv.Load() // ignore err

	cmd := &cli.Command{
		Flags: config.InitiateFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			method := strings.ToLower(cmd.Args().Get(0))
			if method == "" {
				return fmt.Errorf("no command provided")
			}

			if method == "init" {
				// init will create a .env file in the root of the project with all the
				// needed flag names placed for user to change quickly
				var data strings.Builder
				for _, v := range config.Flags {
					data.WriteString(fmt.Sprintf("%s=\n", v.EnvVar))
				}

				os.WriteFile(".env", []byte(data.String()), 0644)
				fmt.Println(".env file created in project root")
				os.Exit(0)
			}

			s := spinner.New(spinner.CharSets[2], 100*time.Millisecond)
			defer s.Stop()

			dbConfig := config.GetConfig(cmd)

			if valid := slices.Contains(config.SupportedDatabases, strings.ToLower(dbConfig.Driver)); !valid {
				return fmt.Errorf("'%v' is not a supported database driver", dbConfig.Driver)
			}

			// connect to both the target and source databases
			sourceDbConn, err := postgres.ConnectToPostgres(dbConfig.SourceHost, dbConfig.SourceDatabase, dbConfig.SourceUser, dbConfig.SourcePassword, dbConfig.SourcePort, dbConfig.SourceSchema)
			if err != nil {
				return err
			}
			defer sourceDbConn.Close(ctx)

			targetDbConn, err := postgres.ConnectToPostgres(dbConfig.TargetHost, dbConfig.TargetDatabase, dbConfig.TargetUser, dbConfig.TargetPassword, dbConfig.TargetPort, dbConfig.TargetSchema)
			if err != nil {
				return err
			}
			defer targetDbConn.Close(ctx)

			switch method {
			case "replace":
				if dbConfig.Driver == "postgres" {
					if err := postgres.ReplaceMethod(targetDbConn, sourceDbConn, ctx, s, dbConfig.SourceSchema, dbConfig.TargetSchema); err != nil {
						return err
					}
				}

			default:
				return fmt.Errorf("'%v' is not a valid command", method)
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
