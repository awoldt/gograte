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
	ForeignColumnName string
	ForeignTableName  string
	SourceColumn      string
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

func DiffMethod(targetDbConn, sourceDbConn *pgx.Conn, ctx context.Context, spinner *spinner.Spinner, targetSchema, sourceSchema string) error {
	/*
		showcases between the source and target table:
		- new tables
		- removed tables
		- new columns
		- removed columns
	*/

	spinner.Start()
	spinner.Suffix = " getting diff"

	var newTables []string
	var removedTables []string

	sourceTablesQuery, err := sourceDbConn.Query(ctx, `SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1`, sourceSchema)

	if err != nil {
		return fmt.Errorf("%s", "error while querying source databases tables\n"+err.Error())
	}
	sourcetables, err := pgx.CollectRows(sourceTablesQuery, pgx.RowToAddrOfStructByName[struct {
		Tablename string `db:"table_name"`
	}])

	targetTableQuery, err := targetDbConn.Query(ctx, `SELECT table_name
	FROM information_schema.tables
	WHERE table_schema = $1`, targetSchema)

	if err != nil {
		return fmt.Errorf("%s", "error while querying target databases tables\n"+err.Error())
	}
	targetTables, err := pgx.CollectRows(targetTableQuery, pgx.RowToAddrOfStructByName[struct {
		Tablename string `db:"table_name"`
	}])

	// new tables
	// tables that exist in the source but not the target
	for _, sourceTable := range sourcetables {
		new := true

		for _, targetTable := range targetTables {
			if sourceTable.Tablename == targetTable.Tablename {
				new = false
				break
			}

		}

		if new {
			newTables = append(newTables, sourceTable.Tablename)
		}

	}

	// removed tables
	for _, targetTable := range targetTables {
		removed := true
		for _, sourceTable := range sourcetables {
			if targetTable.Tablename == sourceTable.Tablename {
				removed = false
				break
			}
		}

		if removed {
			removedTables = append(removedTables, targetTable.Tablename)
		}
	}

	spinner.Stop()

	if len(newTables) > 0 {
		fmt.Printf("found %d new tables to be created:\n", len(newTables))
		for _, table := range newTables {
			fmt.Printf("\t + %s\n", table)
		}
	} else {
		fmt.Print("no new tables found.")
	}

	if len(removedTables) > 0 {
		fmt.Printf("found %d tables to be removed:\n", len(removedTables))
		for _, table := range removedTables {
			fmt.Printf("\t - %s\n", table)
		}
	} else {
		fmt.Print(" no tables to be removed.")
	}

	return nil
}

func ReplaceMethod(targetDbConn, sourceDbConn *pgx.Conn, ctx context.Context, spinner *spinner.Spinner, sourceSchema, targetSchema string) error {
	// will delete the target db and rebuild based on targets schema
	// ALL DATA WILL BE LOST

	var yesno string
	for {
		fmt.Println("source: " + sourceDbConn.Config().Host)
		fmt.Println("target: " + targetDbConn.Config().Host)
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

	sourceTableStructures, err := getSchemaDetails(sourceDbConn, ctx, spinner, sourceSchema, true)
	if err != nil {
		fmt.Println("error while getting source table schema")
		return err
	}

	targetTableStructures, err := getSchemaDetails(targetDbConn, ctx, spinner, targetSchema, false)
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
	// add all columns with tables as well
	numOfTablesCreated := 0
	numOfColumnsCreated := 0
	for key, value := range sourceTableStructures {
		spinner.Suffix = fmt.Sprintf(" creating table %v", key)

		_, err = tx.Exec(ctx, generateCreateTableQuery(key, value.Columns))
		if err != nil {
			fmt.Println("error while creating table")
			return err
		}
		numOfTablesCreated++
		numOfColumnsCreated += len(value.Columns)
	}

	// INSERT ALL PKS BEFORE FKS BELOW!!!!!!!!!!!!!!
	for table, tableDetails := range sourceTableStructures {
		spinner.Suffix = " adding constraints to table " + table

		// insert pk
		if tableDetails.PrimaryKey != "" {
			_, err = tx.Exec(ctx, fmt.Sprintf("ALTER TABLE %s ADD PRIMARY KEY (%s);", table, tableDetails.PrimaryKey))
			if err != nil {
				fmt.Println("error while adding primary key to table " + table + " with value " + tableDetails.PrimaryKey)
				return err
			}
		}

	}

	// insert all foreign keys
	for table, tableDetails := range sourceTableStructures {
		spinner.Suffix = " adding constraints to table " + table
		for _, fk := range tableDetails.ForeignKeys {
			// insert fk
			_, err = tx.Exec(ctx, fmt.Sprintf(`
			ALTER TABLE %s
			ADD FOREIGN KEY (%s)
			REFERENCES %s(%s);
		`, table, fk.SourceColumn, fk.ForeignTableName, fk.ForeignColumnName))
			if err != nil {
				fmt.Printf("error while adding fk key to table %s: column %s referencing %s(%s)\n", table, fk.SourceColumn, fk.ForeignTableName, fk.ForeignColumnName)
				return err
			}
		}
	}

	tx.Commit(ctx)
	spinner.Stop()

	fmt.Printf("\nReplaced %v tables and %v columns in %v seconds\n", numOfTablesCreated, numOfColumnsCreated, time.Since(startTime))
	return nil
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

	connectionConfig.ConnectTimeout = 10 * time.Second // 10 second timeout

	ctx := context.Background()
	conn, err := pgx.ConnectConfig(ctx, connectionConfig)
	if err != nil {
		fmt.Println("error while connecting to database")
		return nil, err
	}

	return conn, nil
}

func getSchemaDetails(dbConn *pgx.Conn, ctx context.Context, spinner *spinner.Spinner, schema string, getConstraints bool) (map[string]Table, error) {
	/*
		this function is fukin insane...
		basically get all the tables inside this schema along with their constraints
	*/

	tables := make(map[string]Table) // table name is key

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
		WHERE table_schema = $1
		ORDER BY table_name, ordinal_position;
	`, schema)

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

	// now get the constraints
	// we only need to get constraints from source db AND if the user wants to put in their target tables
	if getConstraints {
		schemaConstraintsQuery, err := dbConn.Query(ctx, `
					SELECT
					tc.table_name,
					tc.constraint_name,
					tc.constraint_type,
					kcu.column_name,
					ccu.table_name  AS foreign_table_name,
					ccu.column_name AS foreign_column_name
					FROM information_schema.table_constraints tc
					LEFT JOIN information_schema.key_column_usage kcu
					ON tc.constraint_name = kcu.constraint_name
					AND tc.table_schema   = kcu.table_schema
					LEFT JOIN information_schema.constraint_column_usage ccu
					ON tc.constraint_name = ccu.constraint_name
					AND tc.table_schema   = ccu.table_schema
					WHERE tc.table_schema = $1
					AND tc.constraint_type IN ('PRIMARY KEY', 'FOREIGN KEY')
					ORDER BY tc.table_name, tc.constraint_type, kcu.ordinal_position;
`, schema)

		if err != nil {
			fmt.Println("error while querying constraint quries")
			return nil, err
		}

		schemaConstraints, err := pgx.CollectRows(schemaConstraintsQuery, pgx.RowToAddrOfStructByName[struct {
			Tablename         string `db:"table_name"`
			ConstraintName    string `db:"constraint_name"`
			ConstraintType    string `db:"constraint_type"`
			SourceColumnName  string `db:"column_name"`
			ForeignColumnName string `db:"foreign_column_name"`
			ForeignTableName  string `db:"foreign_table_name"`
		}])

		for _, v := range schemaConstraints {

			tableDetails := tables[v.Tablename]

			// pk
			if v.ConstraintType == "PRIMARY KEY" {
				tableDetails.PrimaryKey = v.SourceColumnName
				tables[v.Tablename] = tableDetails
			}

			// foreign key
			if v.ConstraintType == "FOREIGN KEY" {
				tableDetails.ForeignKeys = append(tableDetails.ForeignKeys, ForeignKey{
					ForeignTableName:  v.ForeignTableName,
					ForeignColumnName: v.ForeignColumnName,
					SourceColumn:      v.SourceColumnName,
				})
				tables[v.Tablename] = tableDetails
			}
		}
	}

	return tables, nil
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
