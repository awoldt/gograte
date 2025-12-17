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

type TableStructureQueryResponse struct {
	Tablename     string  `db:"table_name"`
	ColumnName    string  `db:"column_name"`
	DataType      string  `db:"data_type"`
	Nullable      string  `db:"is_nullable"`
	ColumnDefault *string `db:"column_default"`
}

type Column struct {
	ColumnName    string
	ColumnType    string
	Nullable      bool
	ColumnDefault *string // can be null
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

func GetTables(conn *pgx.Conn, ctx context.Context) ([]string, error) {

	tablesQuery, err := conn.Query(ctx, "SELECT tablename FROM pg_tables WHERE schemaname = 'public';")

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	tables, err := pgx.CollectRows(tablesQuery, pgx.RowToAddrOfStructByName[TablesQueryResponse])
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	var tableNames []string
	for _, table := range tables {
		tableNames = append(tableNames, table.Tablename)
	}

	return tableNames, nil
}

func ReplaceDatabase(sourceConn *pgx.Conn, ctx context.Context) error {
	databaseTablesStuctureQuery, err := sourceConn.Query(ctx, `
		SELECT
			table_name,
			column_name,
			data_type,
			is_nullable,
			column_default
		FROM information_schema.columns
		WHERE table_schema = 'public'
		ORDER BY table_name, ordinal_position;
	`)

	if err != nil {
		return fmt.Errorf("%s", "error while querying source databases tables\n"+err.Error())
	}

	databaseTables, err := pgx.CollectRows(databaseTablesStuctureQuery, pgx.RowToAddrOfStructByName[TableStructureQueryResponse])
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	tableStructures := map[string][]Column{}

	for _, dt := range databaseTables {
		_, exists := tableStructures[dt.Tablename]

		if !exists {
			tableStructures[dt.Tablename] = []Column{}
		}

		var isNullable bool
		if dt.Tablename == " YES" {
			isNullable = true
		}

		tableStructures[dt.Tablename] = append(tableStructures[dt.Tablename], Column{
			ColumnName:    dt.ColumnName,
			ColumnType:    dt.DataType,
			Nullable:      isNullable,
			ColumnDefault: dt.ColumnDefault,
		})
	}

	fmt.Println(tableStructures)

	return nil
}
