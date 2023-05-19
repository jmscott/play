/*
 *  Synopsis:
 *	Select attributes for all names spaces from the table pg_name
 *  Description:
 *	Select all the attributes for all the names spaces stored in the
 *	table pg_catalog.pg_namespace.  The tuples are ordered by
 *	the namespace name in ascending order.
 *  Usage:
 *	psql --file pg_namespace.sql
 */
SELECT
	oid,
	nspname,
	nspowner,
	nspacl
  FROM
  	pg_catalog.pg_namespace
  ORDER BY
  	nspname ASC
;
