.PHONY: init
init:
	PGDATA=postgres-data/ pg_ctl init
	PGDATA=postgres-data/ pg_ctl start
	createuser lakehouse
	psql -U sirodoht -d postgres -c "ALTER USER lakehouse CREATEDB;"
	psql -U lakehouse -d postgres -c "CREATE DATABASE lakehouse;"
	psql -U lakehouse -d lakehouse -f schema.sql

.PHONY: start
start:
	PGDATA=postgres-data/ pg_ctl start

.PHONY: stop
stop:
	PGDATA=postgres-data/ pg_ctl stop
