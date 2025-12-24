# gograte - PostgreSQL Schema Migration Tool

`gograte` is a lightweight Go-based CLI tool designed to synchronize PostgreSQL database schemas. It allows you to clone the table structures from a **source** database and apply them to a **target** database.

## Features

- **Schema Mirroring**: Automatically detects tables and columns (types and nullability) from a source database.
- **Constraints Handling**: Identifies and applies primary keys and foreign keys to maintain data integrity.
- **Transactional Safety**: Uses database transactions to ensure that changes are only committed if the entire process succeeds.

---

## Usage

The primary command is `replace`, which drops existing tables in the target database and recreates them based on the source schema.

```bash
go run main.go replace \
  --driver postgres \
  --source-host <SRC_HOST> \
  --source-port 5432 \
  --source-database <SRC_DB_NAME> \
  --source-schema <SRC_SCHEMA> \
  --source-user <SRC_USER> \
  --source-password <SRC_PWD> \
  --target-host <TGT_HOST> \
  --target-port 5433 \
  --target-database <TGT_DB_NAME> \
  --target-schema <TGT_SCHEMA> \
  --target-user <TGT_USER> \
  --target-password <TGT_PWD>
```

### Required Flags

| Flag | Description |
|------|-------------|
| `--driver` | Database type (currently only `postgres` is supported). |
| `--source-host` | The hostname or IP address of the source database server. |
| `--source-port` | The port number of the source database server. |
| `--source-database` | The database name on the source server. |
| `--source-user` | The username for the source database. |
| `--target-host` | The hostname or IP address of the target database server. |
| `--target-port` | The port number of the target database server. |
| `--target-database` | The database name on the target server. |
| `--target-user` | The username for the target database. |

### Optional Flags

| Flag | Description |
|------|-------------|
| `--source-password` | The password for the source database (omit if not required). |
| `--target-password` | The password for the target database (omit if not required). |
| `--source-schema` | The schema within the source database (defaults to `public`). |
| `--target-schema` | The schema within the target database (defaults to `public`). |

### Environment Variables

You can also provide these configurations via a `.env` file in the root of the project. If a flag is provided in the command line, it will override the value in the `.env` file.

Each key must be in uppercase:

| Environment Variable | Description |
|----------------------|-------------|
| `DRIVER` | Database type (e.g., `postgres`). |
| `SOURCE_HOST` | Hostname/IP of the source database. |
| `SOURCE_PORT` | Port of the source database. |
| `SOURCE_DATABASE` | Name of the source database. |
| `SOURCE_USER` | Username for the source database. |
| `SOURCE_PASSWORD` | Password for the source database. |
| `SOURCE_SCHEMA` | Schema for the source database. |
| `TARGET_HOST` | Hostname/IP of the target database. |
| `TARGET_PORT` | Port of the target database. |
| `TARGET_DATABASE` | Name of the target database. |
| `TARGET_USER` | Username for the target database. |
| `TARGET_PASSWORD` | Password for the target database. |
| `TARGET_SCHEMA` | Schema for the target database. |


## ⚠️ Warning

The `replace` command is **destructive**. It will permanently remove all existing data and tables in the target database before recreating the schema. Always ensure you have backups before running this tool against production environments.
