package config

import "github.com/urfave/cli/v3"

type StringFlagType struct {
	name   string
	usage  string
	envVar string
}

var SupportedDatabases []string = []string{"postgres"}

var flags []StringFlagType = []StringFlagType{
	{name: "driver", usage: "Database driver type (postgres, mysql, etc)", envVar: "DRIVER"},

	// target connection
	{name: "target-host", usage: "Target database host", envVar: "TARGET_HOST"},
	{name: "target-port", usage: "Target database port", envVar: "TARGET_PORT"},
	{name: "target-database", usage: "Target database name", envVar: "TARGET_DATABASE"},
	{name: "target-schema", usage: "Target schema within the database", envVar: "TARGET_SCHEMA"},
	{name: "target-user", usage: "Target database user", envVar: "TARGET_USER"},
	{name: "target-password", usage: "Target database password", envVar: "TARGET_PASSWORD"},

	// source connection
	{name: "source-host", usage: "Source database host", envVar: "SOURCE_HOST"},
	{name: "source-port", usage: "Source database port", envVar: "SOURCE_PORT"},
	{name: "source-database", usage: "Source database name", envVar: "SOURCE_DATABASE"},
	{name: "source-schema", usage: "Source schema within the database", envVar: "SOURCE_SCHEMA"},
	{name: "source-user", usage: "Source database user", envVar: "SOURCE_USER"},
	{name: "source-password", usage: "Source database password", envVar: "SOURCE_PASSWORD"},
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
			Name:    flagName.name,
			Value:   "",
			Usage:   flagName.usage,
			Sources: cli.EnvVars(flagName.envVar),
		})
	}

	return data
}
