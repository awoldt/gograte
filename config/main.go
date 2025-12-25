package config

import "github.com/urfave/cli/v3"

type StringFlagType struct {
	name     string
	usage    string
	required bool
	EnvVar   string
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

var SupportedDatabases []string = []string{"postgres"}

var Flags []StringFlagType = []StringFlagType{
	{name: "driver", usage: "Database driver type (postgres, mysql, etc)", EnvVar: "DRIVER", required: true},

	// target connection
	{name: "target-host", usage: "Target database host", EnvVar: "TARGET_HOST", required: true},
	{name: "target-port", usage: "Target database port", EnvVar: "TARGET_PORT", required: true},
	{name: "target-database", usage: "Target database name", EnvVar: "TARGET_DATABASE", required: true},
	{name: "target-schema", usage: "Target schema within the database", EnvVar: "TARGET_SCHEMA", required: false},
	{name: "target-user", usage: "Target database user", EnvVar: "TARGET_USER", required: true},
	{name: "target-password", usage: "Target database password", EnvVar: "TARGET_PASSWORD", required: false},

	// source connection
	{name: "source-host", usage: "Source database host", EnvVar: "SOURCE_HOST", required: true},
	{name: "source-port", usage: "Source database port", EnvVar: "SOURCE_PORT", required: true},
	{name: "source-database", usage: "Source database name", EnvVar: "SOURCE_DATABASE", required: true},
	{name: "source-schema", usage: "Source schema within the database", EnvVar: "SOURCE_SCHEMA", required: false},
	{name: "source-user", usage: "Source database user", EnvVar: "SOURCE_USER", required: true},
	{name: "source-password", usage: "Source database password", EnvVar: "SOURCE_PASSWORD", required: false},
}

func GetConfig(cmd *cli.Command) DatabaseConfig {
	dbConfig := DatabaseConfig{
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

	return dbConfig
}

func InitiateFlags() []cli.Flag {
	var data []cli.Flag = []cli.Flag{}

	for _, flagName := range Flags {
		data = append(data, &cli.StringFlag{
			Name:     flagName.name,
			Value:    "",
			Usage:    flagName.usage,
			Sources:  cli.EnvVars(flagName.EnvVar),
			Required: flagName.required,
		})
	}

	return data
}
