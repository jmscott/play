REST API for SQL - RASQL
========================

# Synopsis
	Build a simple REST server by parsing PostgreSQL query files.

	Both http and ssl are supported, as well as anonymous and basic authentification.

	Logging is to standard error.
# Build and Test in Unix Environment
	#  Install lib/pq, the pure postgresql package

	GOPATH=/usr/local /usr/local/go/bin/go get -u github.com/lib/pq

	#  Build new rasqld program.

	make clean rasqld

	#  Verify PostgreSQL environment variables for access to database.
	#  lib/pq croaks on existing PGSYSCONFDIR, so undefine

	env | grep '^PG';  unset PGSYSCONFDIR

	#  Create a RASQL configuration for pg_catalog schema

	cp pg_catalog.rasql.example pg_catalog.rasql

	#  Start the server in the background, listening on localhost:8080

	rasqld pg_catalog.rasql >rasqld.log 2>&1 &

	#  Check for startup errors

	tail rasqld.log

	#  REST query to pull list of queries

	curl http://localhost:8080/pg_catalog

	#  View all PostgreSQL classses.  See file pg_class.sql.

	curl http://localhost:8080/pg_catalog/pg_class
