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
	godotenv.Load() // load these values

	cmd := &cli.Command{
		Flags: config.InitiateFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			method := strings.ToLower(cmd.Args().Get(0))
			if method == "" {
				return fmt.Errorf("you must add an argument to the gograte command")
			}

			s := spinner.New(spinner.CharSets[2], 100*time.Millisecond)
			defer s.Stop()

			dbConfig := config.DatabaseConfig{
				Driver:         cmd.String("driver"),
				TargetDatabase: cmd.String("target-database"),
				TargetHost:     cmd.String("target-host"),
				TargetUser:     cmd.String("target-user"),
				TargetPassword: cmd.String("target-password"),
				TargetPort:     cmd.String("target-port"),
				TargetSchema:   cmd.String("target-schema"),
				SourceHost:     cmd.String("source-host"),
				SourceUser:     cmd.String("source-user"),
				SourcePassword: cmd.String("source-password"),
				SourcePort:     cmd.String("source-port"),
				SourceSchema:   cmd.String("source-schema"),
				SourceDatabase: cmd.String("source-database"),
			}

			driver := dbConfig.Driver

			targetDb := dbConfig.TargetDatabase
			targetUser := dbConfig.TargetUser
			targetPassword := dbConfig.TargetPassword
			targetPort := dbConfig.TargetPort
			targetSchema := dbConfig.TargetSchema
			targetHost := dbConfig.TargetHost

			sourcedb := dbConfig.SourceDatabase
			sourceUser := dbConfig.SourceUser
			sourcePassword := dbConfig.SourcePassword
			sourcePort := dbConfig.SourcePort
			sourceSchema := dbConfig.SourceSchema
			sourceHost := dbConfig.SourceHost

			// ensure all required flags are here
			if driver == "" || targetHost == "" || targetPort == "" || targetDb == "" || targetUser == "" || sourceHost == "" || sourcePort == "" || sourcedb == "" || sourceUser == "" {
				return fmt.Errorf("missing required flags")
			}

			if valid := slices.Contains(config.SupportedDatabases, strings.ToLower(driver)); !valid {
				return fmt.Errorf("'%v' is not a supported database driver", driver)
			}

			// connect to both the target and source databases
			sourceDbConn, err := postgres.ConnectToPostgres(sourceHost, sourcedb, sourceUser, sourcePassword, sourcePort, sourceSchema)
			if err != nil {
				return err
			}
			defer sourceDbConn.Close(ctx)

			targetDbConn, err := postgres.ConnectToPostgres(targetHost, targetDb, targetUser, targetPassword, targetPort, targetSchema)
			if err != nil {
				return err
			}
			defer targetDbConn.Close(ctx)

			switch method {
			case "replace":
				if dbConfig.Driver == "postgres" {
					if err := postgres.ReplaceMethod(targetDbConn, sourceDbConn, ctx, s, sourceSchema, targetSchema); err != nil {
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
