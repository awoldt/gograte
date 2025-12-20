package config

import "github.com/urfave/cli/v3"

type StringFlagType struct {
	name  string
	usage string
}

var SupportedDatabases []string = []string{"postgres"}

var flags []StringFlagType = []StringFlagType{
	{name: "driver", usage: "Database driver type (postgres, mysql, etc)"},

	// target connection
	{name: "target-host", usage: "Target database host"},
	{name: "target-port", usage: "Target database port"},
	{name: "target-database", usage: "Target database name"},
	{name: "target-schema", usage: "Target schema within the database"},
	{name: "target-user", usage: "Target database user"},
	{name: "target-password", usage: "Target database password"},

	// source connection
	{name: "source-host", usage: "Source database host"},
	{name: "source-port", usage: "Source database port"},
	{name: "source-database", usage: "Source database name"},
	{name: "source-schema", usage: "Source schema within the database"},
	{name: "source-user", usage: "Source database user"},
	{name: "source-password", usage: "Source database password"},
}

type DatabaseConfig struct {
	Driver         string
	TargetHost     string
	TargetPort     string
	TargetDatabase string
	TargetSchema   string
	TargetUser     string
	TargetPassword string
	SourceHost     string
	SourcePort     string
	SourceDatabase string
	SourceSchema   string
	SourceUser     string
	SourcePassword string
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
