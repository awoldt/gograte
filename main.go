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

			switch method {
			case "replace":
				{
					switch dbConfig.Driver {
					case "postgres":
						{
							if err := postgres.ReplaceMethod(dbConfig, ctx, s); err != nil {
								return err
							}
							break
						}
					default:
						{
							if dbConfig.Driver == "" {
								return fmt.Errorf("must supply a database driver")
							}
							return fmt.Errorf("'%v' is not a supported database", dbConfig.Driver)
						}
					}
					break
				}
			default:
				{
					return fmt.Errorf("'%v' is not a valid command", method)
				}
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
