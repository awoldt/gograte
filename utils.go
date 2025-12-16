package main

import (
	"fmt"
	"gograte/config"
	"slices"
)

// this will determine if the user's command is valid or not
// kill prgrogram if not
// HAVE to have all the flags
func ValidCommand(dbDriver, targetHost, database, user, sourceHost string) error {

	if dbDriver == "" || !slices.Contains(config.SupportedDatabases, dbDriver) {
		return fmt.Errorf("invalid database driver")
	}

	if targetHost == "" || database == "" || user == "" || sourceHost == "" {
		return fmt.Errorf("missing flags")
	}

	return nil
}
