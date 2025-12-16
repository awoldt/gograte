package postgres

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5"
)

type TablesQueryResponse struct {
	Tablename string `db:"tablename"`
}

func ConnetToPostgres(host, database, user, password, port string) (*pgx.Conn, error) {
	if host == "" || database == "" || user == "" {
		return nil, fmt.Errorf("Must supply a host, database, and user")
	}
	var connectionString string

	if password == "" {
		connectionString = fmt.Sprintf("postgres://%v@%v:%v/%v", user, host, port, database)
	} else {
		encodedPwd := url.QueryEscape(password)
		connectionString = fmt.Sprintf("postgres://%v:%v@%v:%v/%v", user, encodedPwd, host, port, database)
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

func GetTables(conn *pgx.Conn, ctx context.Context) ([]*TablesQueryResponse, error) {

	tablesQuery, err := conn.Query(ctx, "SELECT tablename FROM pg_tables WHERE schemaname = 'public';")

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	tables, err := pgx.CollectRows(tablesQuery, pgx.RowToAddrOfStructByName[TablesQueryResponse])
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	for _, table := range tables {
		fmt.Println(table.Tablename)
	}

	return tables, nil
}
