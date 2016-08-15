/*
 *  Synopsis:
 *	Fetch all roles in the pg_rols view
 *
 *  Command Line Variables:
 *
 *  Usage:
 *	psql -f pg_roles.sql
 */
select
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
  from
  	pg_catalog.pg_roles
  order by
  	rolname asc
;
