package main

import (
	"fmt"
	"gograte/config"
	"slices"
)

// this will determine if the user's command is valid or not
// kill prgrogram if not
// HAVE to have all the flags
func ValidCommand(dbDriver, host, database, user string) error {

	if dbDriver == "" || !slices.Contains(config.SupportedDatabases, dbDriver) {
		return fmt.Errorf("invalid database driver")
	}

	if host == "" || database == "" || user == "" {
		return fmt.Errorf("missing flags")
	}

	return nil
}
