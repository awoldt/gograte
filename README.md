# gograte - PostgreSQL Schema Migration Tool

`gograte` is a lightweight Go-based CLI tool designed to synchronize PostgreSQL database schemas. It allows you to clone the table structures from a **source** database and apply them to a **target** database.

## Features

- **Schema Mirroring**: Automatically detects tables and columns (types and nullability) from a source database.
- **Transactional Safety**: Uses database transactions to ensure that changes are only committed if the entire process succeeds.

---

## Usage

The primary command is `replace`, which drops existing tables in the target database and recreates them based on the source schema.

```bash
go run main.go replace \
  --driver postgres \
  --database <DB_NAME> \
  --source-db <SRC_HOST> \
  --source-user <SRC_USER> \
  --source-password <SRC_PWD> \
  --source-port 5432 \
  --source-schema <SRC_SCHEMA> \
  --target-db <TGT_HOST> \
  --target-user <TGT_USER> \
  --target-password <TGT_PWD> \
  --target-port 5433 \
  --target-schema <TGT_SCHEMA>
```

### Required Flags

| Flag | Description |
|------|-------------|
| `--driver` | Database type (currently only `postgres` is supported). |
| `--database` | The specific database name to connect to on both hosts. |
| `--source-*` | Connection details for the **source** (the "template" database). |
| `--target-*` | Connection details for the **target** (the database to be updated). |

> **Note:** Passwords are optional and can be omitted if the database doesn't require them. The `--source-schema` and `--target-schema` flags are also optional and allow you to specify which schema within each database to perform actions on. The 'public' schema is selected by default if not provided.


## ⚠️ Warning

The `replace` command is **destructive**. It will permanently remove all existing data and tables in the target database before recreating the schema. Always ensure you have backups before running this tool against production environments.
