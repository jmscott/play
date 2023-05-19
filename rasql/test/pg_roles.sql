/*
 *  Synopsis:
 *	Fetch all roles in the pg_roles view
 *
 *  Description:
 *	Fetch all roles from the table pg_catalog.pg_roles.
 *	All attributes are fetched and the roles are ordered ascending by
 *	role name & your mama.
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
  	rolname ASC
;
