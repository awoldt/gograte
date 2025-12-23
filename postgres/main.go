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

type ForeignKey struct {
	TargetTable  string
	TargetColumn string
	SourceColumn string
}

type Table struct {
	PrimaryKey  string
	ForeignKeys []ForeignKey
	Columns     []Column
}

type Column struct {
	ColumnName    string
	ColumnType    string
	Nullable      bool
	ColumnDefault *string // can be null
}

func ReplaceMethod(targetDbConn, sourceDbConn *pgx.Conn, ctx context.Context, spinner *spinner.Spinner, sourceSchema, targetSchema string) error {
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
		fmt.Println("error while beginning transaction")
		return err
	}
	defer tx.Rollback(ctx) // rollback if we dont commit!!!!!!

	sourceTableStructures, err := getDatabaseTablesSchema(sourceDbConn, ctx, spinner, sourceSchema)
	if err != nil {
		fmt.Println("error while getting source table schema")
		return err
	}

	targetTableStructures, err := getDatabaseTablesSchema(targetDbConn, ctx, spinner, targetSchema)
	if err != nil {
		fmt.Println("error while getting target table schema")
		return err
	}

	// delete all tables in the target db before creating tables
	for key := range targetTableStructures {
		spinner.Suffix = fmt.Sprintf(" deleting table %v", key)

		_, err := tx.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %v CASCADE;", key))
		if err != nil {
			fmt.Println("error while deleting target table")
			return err
		}
	}

	// generate a create table query for every table detected in source db
	for key, value := range sourceTableStructures {
		spinner.Suffix = fmt.Sprintf(" creating table %v", key)

		_, err = tx.Exec(ctx, generateCreateTableQuery(key, value.Columns))
		if err != nil {
			fmt.Println("error while creating table")
			return err
		}
	}

	tx.Commit(ctx)
	spinner.Stop()

	fmt.Printf("\nFinished in %v seconds\n", time.Since(startTime))
	return nil
}

func getDatabaseTablesSchema(dbConn *pgx.Conn, ctx context.Context, spinner *spinner.Spinner, schema string) (map[string]Table, error) {

	spinner.Suffix = " loading tables"

	// if no schema is provided, default to public
	if schema == "" {
		schema = "public"
	}

	// first, get only the tables
	// this will return all tables regardless or not if it has any columns
	databaseTablesQuery, err := dbConn.Query(ctx, fmt.Sprintf(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = '%v'
		ORDER BY table_name;
	`, schema))

	if err != nil {
		fmt.Println("error while querying source databases tables")
		return nil, fmt.Errorf("%s", "error while querying source databases tables\n"+err.Error())
	}

	databaseTables, err := pgx.CollectRows(databaseTablesQuery, pgx.RowToAddrOfStructByName[struct {
		Tablename string `db:"table_name"`
	}])
	if err != nil {
		fmt.Println("error while collecting table rows")
		return nil, err
	}

	tables := make(map[string]Table) // table name is key
	for _, t := range databaseTables {
		_, exists := tables[t.Tablename]

		if !exists {
			tables[t.Tablename] = Table{}
		}

	}

	spinner.Suffix = " loading columns"

	// now get all the columns
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
		fmt.Println("error while querying source databases tables")
		return nil, fmt.Errorf("%s", "error while querying source databases tables\n"+err.Error())
	}

	databaseTablesColumns, err := pgx.CollectRows(databaseTablesColumnsQuery, pgx.RowToAddrOfStructByName[struct {
		Tablename     string  `db:"table_name"`
		ColumnName    string  `db:"column_name"`
		DataType      string  `db:"data_type"`
		Nullable      string  `db:"is_nullable"`
		ColumnDefault *string `db:"column_default"`
	}])
	if err != nil {
		fmt.Println("error while collecting column rows")
		return nil, err
	}

	for _, dt := range databaseTablesColumns {
		var isNullable bool
		if dt.Nullable == "YES" {
			isNullable = true
		}

		t := tables[dt.Tablename].Columns
		t = append(t, Column{
			ColumnName:    dt.ColumnName,
			ColumnType:    dt.DataType,
			Nullable:      isNullable,
			ColumnDefault: dt.ColumnDefault,
		})

		table := tables[dt.Tablename]
		table.Columns = t
		tables[dt.Tablename] = table
	}

	return tables, nil
}

func ConnectToPostgres(host, database, user, password, port, schema string) (*pgx.Conn, error) {
	if host == "" || database == "" || user == "" {
		fmt.Println("must supply a host, database, and user")
		return nil, fmt.Errorf("must supply a host, database, and user")
	}
	var connectionString string

	if password == "" {
		connectionString = fmt.Sprintf("postgres://%v@%v:%v/%v?search_path=%v", user, host, port, database, schema)
	} else {
		encodedPwd := url.QueryEscape(password)
		connectionString = fmt.Sprintf("postgres://%v:%v@%v:%v/%v?search_path=%v", user, encodedPwd, host, port, database, schema)
	}

	connectionConfig, err := pgx.ParseConfig(connectionString)
	if err != nil {
		fmt.Println("error while parsing connection string")
		return nil, err
	}

	ctx := context.Background()
	conn, err := pgx.ConnectConfig(ctx, connectionConfig)
	if err != nil {
		fmt.Println("error while connecting to database")
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
