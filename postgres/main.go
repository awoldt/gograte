package postgres

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5"
)

func ConnetToPostgres(host, database, user, password string) (*pgx.Conn, error) {
	if host == "" || database == "" || user == "" {
		return nil, fmt.Errorf("Must supply a host, database, and user")
	}
	var connectionString string

	if password == "" {
		connectionString = fmt.Sprintf("postgres://%v@%v/%v", user, host, database)

	} else {
		encodedPwd := url.QueryEscape(password)

		connectionString = fmt.Sprintf("postgres://%v:%v@%v/%v", user, encodedPwd, host, database)

	}

	connectionConfig, err := pgx.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	ctx := context.Background()
	conn, err := pgx.ConnectConfig(ctx, connectionConfig)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return conn, nil
}
