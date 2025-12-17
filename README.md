# gograte - PostgreSQL Database Migration Tool

A command-line tool for migrating PostgreSQL database schemas from a source database to a target database.

### Command Structure

```bash
go run main.go \
  --driver <DATABASE_DRIVER> \
  --database <DATABASE_NAME> \
  --source-db <SOURCE_HOST> \
  --source-user <SOURCE_USERNAME> \
  --source-password <SOURCE_PASSWORD> \
  --source-port <SOURCE_PORT> \
  --target-db <TARGET_HOST> \
  --target-user <TARGET_USERNAME> \
  --target-password <TARGET_PASSWORD> \
  --target-port <TARGET_PORT>
```

## Required Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--driver` | Database driver type (currently supports: `postgres`) | `postgres` |
| `--database` | The database name within the specified driver to connect to | `myapp_db` |
| `--source-db` | Source database host (what you want to clone from) | `localhost` |
| `--source-user` | Source database username | `postgres` |
| `--source-password` | Source database password | `password123` |
| `--source-port` | Source database port | `5432` |
| `--target-db` | Target database host (what will be updated) | `localhost` |
| `--target-user` | Target database username | `postgres` |
| `--target-password` | Target database password | `password456` |
| `--target-port` | Target database port | `5433` |
