package main

import (
	"context"
	"fmt"
	"gograte/config"
	"gograte/postgres"
	"log"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/urfave/cli/v3"
)

func main() {
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
				Database:       cmd.String("database"),
				TargetDb:       cmd.String("target-db"),
				TargetUser:     cmd.String("target-user"),
				TargetPassword: cmd.String("target-password"),
				TargetPort:     cmd.String("target-port"),
				SourceDb:       cmd.String("source-db"),
				SourceUser:     cmd.String("source-user"),
				SourcePassword: cmd.String("source-password"),
				SourcePort:     cmd.String("source-port"),
			}

			database := dbConfig.Database
			driver := dbConfig.Driver

			targetDb := dbConfig.TargetDb
			targetUser := dbConfig.TargetUser
			targetPassword := dbConfig.TargetPassword
			targetPort := dbConfig.TargetPort

			sourcedb := dbConfig.SourceDb
			sourceUser := dbConfig.SourceUser
			sourcePassword := dbConfig.SourcePassword
			sourcePort := dbConfig.SourcePort

			// ensure all strings OTHER THAN PASSWORDS are not empty
			if database == "" || targetDb == "" || targetUser == "" || targetPort == "" || sourcedb == "" || sourceUser == "" || sourcePort == "" || driver == "" {
				return fmt.Errorf("must supply database, database driver, target-db, target-user, target-port, source-db, source-user, and source-port")
			}

			// make sure this is a valid database driver even before connecting to the databases
			validDriver := false
			for _, v := range config.SupportedDatabases {
				if v == strings.ToLower(driver) {
					validDriver = true
					break
				}
			}
			if !validDriver {
				return fmt.Errorf("'%v' is not a supported database driver", driver)
			}

			// connect to both the target and source databases
			sourceDbConn, err := postgres.ConnectToPostgres(sourcedb, database, sourceUser, sourcePassword, sourcePort)
			if err != nil {
				return err
			}
			defer sourceDbConn.Close(ctx)

			targetDbConn, err := postgres.ConnectToPostgres(targetDb, database, targetUser, targetPassword, targetPort)
			if err != nil {
				return err
			}
			defer targetDbConn.Close(ctx)

			switch method {
			case "replace":
				if dbConfig.Driver == "postgres" {
					if err := postgres.ReplaceMethod(targetDbConn, sourceDbConn, ctx, s); err != nil {
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
