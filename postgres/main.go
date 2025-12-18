package postgres

import (
	"context"
	"fmt"
	"gograte/config"
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

func ReplaceMethod(dbConfig config.DatabaseConfig, ctx context.Context, spinner *spinner.Spinner) error {
	// this is one of the main methods to run
	// will delete the target db and rebuild based on targets schema
	// ALL DATA WILL BE LOST

	database := dbConfig.Database

	targetDb := dbConfig.TargetDb
	targetUser := dbConfig.TargetUser
	targetPassword := dbConfig.TargetPassword
	targetPort := dbConfig.TargetPort

	sourcedb := dbConfig.SourceDb
	sourceUser := dbConfig.SourceUser
	sourcePassword := dbConfig.SourcePassword
	sourcePort := dbConfig.SourcePort

	// ensure all strings OTHER THAN PASSWORDS are not empty
	if database == "" || targetDb == "" || targetUser == "" || targetPort == "" || sourcedb == "" || sourceUser == "" || sourcePort == "" {
		return fmt.Errorf("must supply database, target-db, target-user, target-port, source-db, source-user, and source-port")
	}

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

	spinner.Start()
	startTime := time.Now()

	sourceDbConn, err := connectToPostgres(sourcedb, database, sourceUser, sourcePassword, sourcePort)
	if err != nil {
		return err
	}
	defer sourceDbConn.Close(ctx)

	targetDbConn, err := connectToPostgres(targetDb, database, targetUser, targetPassword, targetPort)
	if err != nil {
		return err
	}
	defer targetDbConn.Close(ctx)

	spinner.Suffix = " getting table details"

	// WRAP EVERYTHING IN A TRANSACTION TO PREVENT THE WORST!
	tx, err := targetDbConn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // rollback if we dont commit!!!!!!

	tableStructures, err := getDatabaseTablesSchema(sourceDbConn, ctx, spinner)
	if err != nil {
		return err
	}

	// generate a create table query for every table in the source db
	for key, value := range tableStructures {
		spinner.Suffix = fmt.Sprintf(" creating table %v", key)
		// drop this table from teh database
		_, err := tx.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %v CASCADE;", key))
		if err != nil {
			return err
		}

		// attempt to create this table in target db
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

func getDatabaseTablesSchema(sourceDbConn *pgx.Conn, ctx context.Context, spinner *spinner.Spinner) (map[string][]Column, error) {

	databaseTablesStuctureQuery, err := sourceDbConn.Query(ctx, `
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

	databaseTables, err := pgx.CollectRows(databaseTablesStuctureQuery, pgx.RowToAddrOfStructByName[TableStructureQueryResponse])
	if err != nil {
		return nil, err
	}

	spinner.Suffix = " loading tables and columns"

	tableStructures := map[string][]Column{}
	for _, dt := range databaseTables {
		_, exists := tableStructures[dt.Tablename]

		if !exists {
			tableStructures[dt.Tablename] = []Column{}
		}

		var isNullable bool
		if dt.Nullable == "YES" {
			isNullable = true
		}

		tableStructures[dt.Tablename] = append(tableStructures[dt.Tablename], Column{
			ColumnName:    dt.ColumnName,
			ColumnType:    dt.DataType,
			Nullable:      isNullable,
			ColumnDefault: dt.ColumnDefault,
		})
	}

	return tableStructures, nil
}

func connectToPostgres(host, database, user, password, port string) (*pgx.Conn, error) {
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
	stringBuilder.WriteString(fmt.Sprintf("CREATE TABLE %s(\n", table))

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
