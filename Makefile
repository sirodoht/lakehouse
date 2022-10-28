.PHONY: all
all: lint serve

.PHONY: pginit
pginit:
	PGDATA=postgres-data/ pg_ctl init
	PGDATA=postgres-data/ pg_ctl start
	createuser lakehouse
	psql -U sirodoht -d postgres -c "ALTER USER lakehouse CREATEDB;"
	psql -U lakehouse -d postgres -c "CREATE DATABASE lakehouse;"
	psql -U lakehouse -d lakehouse -f schema.sql

.PHONY: pgstart
pgstart:
	PGDATA=postgres-data/ pg_ctl start

.PHONY: pgstop
pgstop:
	PGDATA=postgres-data/ pg_ctl stop

.PHONY: lint
lint:
	$(info Running Go linters)
	@GOGC=off golangci-lint run

.PHONY: format
format:
	go fmt ./...

.PHONY: serve
serve:
	modd
