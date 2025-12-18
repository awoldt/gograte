package postgres

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/jackc/pgx/v5"
)

type TablesQueryResponse struct {
	Tablename string `db:"tablename"`
}

type DatabaseTablesColumnsQuery struct {
	Tablename     string  `db:"table_name"`
	ColumnName    string  `db:"column_name"`
	DataType      string  `db:"data_type"`
	Nullable      string  `db:"is_nullable"`
	ColumnDefault *string `db:"column_default"`
}

type DatabaseTableQuery struct {
	Tablename string `db:"table_name"`
}

type Column struct {
	ColumnName    string
	ColumnType    string
	Nullable      bool
	ColumnDefault *string // can be null
}

func ReplaceMethod(targetDbConn, sourceDbConn *pgx.Conn, ctx context.Context, spinner *spinner.Spinner) error {
	// will delete the target db and rebuild based on targets schema
	// ALL DATA WILL BE LOST

	var yesno string
	for {
		fmt.Print("replacing a database is permanent and will remove all data. are you sure? (y/n): ")
		fmt.Scan(&yesno)

		answer := strings.TrimSpace(strings.ToLower(yesno))

		if answer != "y" && answer != "n" {
			continue
		}

		if answer == "n" {
			os.Exit(0)
		} else if answer == "y" {
			break
		} else {
			continue
		}
	}

	startTime := time.Now()
	spinner.Start()

	spinner.Suffix = " getting table details"

	// WRAP EVERYTHING IN A TRANSACTION TO PREVENT THE WORST!
	tx, err := targetDbConn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // rollback if we dont commit!!!!!!

	sourceTableStructures, err := getDatabaseTablesSchema(sourceDbConn, ctx, spinner)
	if err != nil {
		return err
	}

	targetTableStructures, err := getDatabaseTablesSchema(targetDbConn, ctx, spinner)
	if err != nil {
		return err
	}

	// delete all tables in the target db before creating tables
	for key := range targetTableStructures {
		spinner.Suffix = fmt.Sprintf(" deleting table %v", key)

		_, err := tx.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %v CASCADE;", key))
		if err != nil {
			return err
		}
	}

	// generate a create table query for every table detected in source db
	for key, value := range sourceTableStructures {
		spinner.Suffix = fmt.Sprintf(" creating table %v", key)

		_, err = tx.Exec(ctx, generateCreateTableQuery(key, value))
		if err != nil {
			return err
		}
	}

	tx.Commit(ctx)
	spinner.Stop()

	fmt.Printf("\nFinished in %v seconds\n", time.Since(startTime))
	return nil
}

func getDatabaseTablesSchema(dbConn *pgx.Conn, ctx context.Context, spinner *spinner.Spinner) (map[string][]Column, error) {

	spinner.Suffix = " loading tables"

	// first, get only the tables
	// this will return all tables regardless or not if it has any columns
	databaseTablesQuery, err := dbConn.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		ORDER BY table_name;
	`)

	if err != nil {
		return nil, fmt.Errorf("%s", "error while querying source databases tables\n"+err.Error())
	}

	databaseTables, err := pgx.CollectRows(databaseTablesQuery, pgx.RowToAddrOfStructByName[DatabaseTableQuery])
	if err != nil {
		return nil, err
	}

	tables := make(map[string][]Column)
	for _, t := range databaseTables {
		_, exists := tables[t.Tablename]

		if !exists {
			tables[t.Tablename] = []Column{}
		}
	}

	spinner.Suffix = " loading columns"

	databaseTablesColumnsQuery, err := dbConn.Query(ctx, `
		SELECT
			table_name,
			column_name,
			CASE
				WHEN data_type = 'ARRAY' THEN REPLACE(udt_name, '_', '') || '[]'
				ELSE data_type
			END as data_type,
			is_nullable,
			column_default
		FROM information_schema.columns
		WHERE table_schema = 'public'
		ORDER BY table_name, ordinal_position;
	`)

	if err != nil {
		return nil, fmt.Errorf("%s", "error while querying source databases tables\n"+err.Error())
	}

	databaseTablesColumns, err := pgx.CollectRows(databaseTablesColumnsQuery, pgx.RowToAddrOfStructByName[DatabaseTablesColumnsQuery])
	if err != nil {
		return nil, err
	}

	for _, dt := range databaseTablesColumns {
		var isNullable bool
		if dt.Nullable == "YES" {
			isNullable = true
		}

		tables[dt.Tablename] = append(tables[dt.Tablename], Column{
			ColumnName:    dt.ColumnName,
			ColumnType:    dt.DataType,
			Nullable:      isNullable,
			ColumnDefault: dt.ColumnDefault,
		})
	}

	return tables, nil
}

func ConnectToPostgres(host, database, user, password, port string) (*pgx.Conn, error) {
	if host == "" || database == "" || user == "" {
		return nil, fmt.Errorf("must supply a host, database, and user")
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
		return nil, err
	}

	ctx := context.Background()
	conn, err := pgx.ConnectConfig(ctx, connectionConfig)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func generateCreateTableQuery(table string, columns []Column) string {
	var stringBuilder strings.Builder
	stringBuilder.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(\n", table))

	numOfCols := len(columns)

	// all the crazy stuff here
	for i, col := range columns {
		nullConstraintString := ""
		if !col.Nullable {
			nullConstraintString = "NOT NULL"
		}
		stringBuilder.WriteString(fmt.Sprintf("%v %v %v", col.ColumnName, col.ColumnType, nullConstraintString))

		// add comma if not the last column
		if i < numOfCols-1 {
			stringBuilder.WriteString(", ")
		}
	}

	stringBuilder.WriteString("\n);")

	return stringBuilder.String()
}
