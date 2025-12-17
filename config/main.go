package config

import "github.com/urfave/cli/v3"

type StringFlagType struct {
	name  string
	usage string
}

var SupportedDatabases []string = []string{"postgres"}

var flags []StringFlagType = []StringFlagType{
	{name: "driver", usage: "Database driver type"},
	{name: "database", usage: "The database within the specified driver to connect to"},

	{name: "target-db", usage: "Database that will updated"},
	{name: "target-user", usage: "Database user value"},
	{name: "target-password", usage: "Database password value"},
	{name: "target-port", usage: "Database port"},

	{name: "source-db", usage: "The database that will have the schemas read from (what you want to clone)"},
	{name: "source-user", usage: "Database user value"},
	{name: "source-password", usage: "Database password value"},
	{name: "source-port", usage: "Database port"},
}

type DatabaseConfig struct {
	Driver         string
	Database       string
	TargetDb       string
	TargetUser     string
	TargetPassword string
	TargetPort     string
	SourceDb       string
	SourceUser     string
	SourcePassword string
	SourcePort     string
}

func InitiateFlags() []cli.Flag {
	var data []cli.Flag = []cli.Flag{}

	for _, flagName := range flags {
		data = append(data, &cli.StringFlag{
			Name:  flagName.name,
			Value: "",
			Usage: flagName.usage,
		})
	}

	return data
}
