package config

import "github.com/urfave/cli/v3"

type StringFlagType struct {
	name   string
	usage  string
	EnvVar string
}

var SupportedDatabases []string = []string{"postgres"}

var Flags []StringFlagType = []StringFlagType{
	{name: "driver", usage: "Database driver type (postgres, mysql, etc)", EnvVar: "DRIVER"},

	// target connection
	{name: "target-host", usage: "Target database host", EnvVar: "TARGET_HOST"},
	{name: "target-port", usage: "Target database port", EnvVar: "TARGET_PORT"},
	{name: "target-database", usage: "Target database name", EnvVar: "TARGET_DATABASE"},
	{name: "target-schema", usage: "Target schema within the database", EnvVar: "TARGET_SCHEMA"},
	{name: "target-user", usage: "Target database user", EnvVar: "TARGET_USER"},
	{name: "target-password", usage: "Target database password", EnvVar: "TARGET_PASSWORD"},

	// source connection
	{name: "source-host", usage: "Source database host", EnvVar: "SOURCE_HOST"},
	{name: "source-port", usage: "Source database port", EnvVar: "SOURCE_PORT"},
	{name: "source-database", usage: "Source database name", EnvVar: "SOURCE_DATABASE"},
	{name: "source-schema", usage: "Source schema within the database", EnvVar: "SOURCE_SCHEMA"},
	{name: "source-user", usage: "Source database user", EnvVar: "SOURCE_USER"},
	{name: "source-password", usage: "Source database password", EnvVar: "SOURCE_PASSWORD"},
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

	for _, flagName := range Flags {
		data = append(data, &cli.StringFlag{
			Name:    flagName.name,
			Value:   "",
			Usage:   flagName.usage,
			Sources: cli.EnvVars(flagName.EnvVar),
		})
	}

	return data
}
