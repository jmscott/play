/*
 *  Synopsis:
 *	All authorization identifiers (roles) in table pg_authid.
 */
\x auto on
SELECT
	oid,
	rolname,
	rolsuper,
	rolinherit,
	rolcreaterole,
	rolcreatedb,
	rolcanlogin,
	rolreplication,
	rolbypassrls,
	rolconnlimit,
	rolpassword,
	rolvaliduntil
  FROM
  	pg_catalog.pg_authid
  ORDER BY
  	rolname asc
;
