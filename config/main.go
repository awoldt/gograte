package config

import "github.com/urfave/cli/v3"

type StringFlagType struct {
	name  string
	usage string
}

var SupportedDatabases []string = []string{"postgres"}

var flags []StringFlagType = []StringFlagType{
	{name: "driver", usage: "Database driver type"},
	{name: "host", usage: "Database host value"},
	{name: "database", usage: "The database within the specified driver to connect to"},
	{name: "user", usage: "Database user value"},
	{name: "password", usage: "Database password value"},
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
