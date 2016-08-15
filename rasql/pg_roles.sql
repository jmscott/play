/*
 *  Synopsis:
 *	Fetch all roles in the pg_rols view
 *
 *  Command Line Variables:
 *
 *  Usage:
 *	psql -f pg_roles.sql
 */
SELECT
	rolname,
	rolsuper,
	rolinherit,
	rolcreaterole,
	rolcreatedb,
	rolcanlogin,
	rolreplication,
	rolconnlimit,
	rolpassword,
	rolvaliduntil,
	rolbypassrls,
	rolconfig,
	oid
  FROM
  	pg_catalog.pg_roles
  ORDER BY
  	rolname asc
;
