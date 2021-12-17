# migrate

A tool to perform Postgres database schema migrations.

## Install

```sh
go env -w GOPRIVATE=github.com/aukilabs/*
go install github.com/aukilabs/go-tooling/bin/migrate@latest
```

## Help

```sh
migrate -h
```

## Run migration

Create 2 SQL files with the following pattern in the sources directory (usually `REPO/data/migrations` ):

- `VERSION_TITLE.up.sql`: contains sql modifications
- `VERSION_TITLE.down.sql`: reverts sql modifications

eg:

```
42_add_test_table.up.sql
42_add_test_table.down.sql
```

Then launch the migration:

```sh
migrate          # Performs migration up to the latest version
migrate -steps 1 # Performs migration X steps forward or backward
```
