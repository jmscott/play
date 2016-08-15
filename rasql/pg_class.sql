/*
 *  Synopsis:
 *	Fetch all classes in a pg_catalog.pg_class
 *
 *  Command Line Variables:
 *
 *  Usage:
 *	psql -f pg_class.sql
 */
SELECT
	n.nspname,
	c.relname,
	c.relnamespace,
	c.reltype,
	c.reloftype,
	c.relowner,
	c.relam,
	c.relfilenode,
	c.reltablespace,
	c.relpages,
	c.reltuples,
	c.relallvisible,
	c.reltoastrelid,
	c.relhasindex,
	c.relisshared,
	c.relpersistence,
	c.relkind,
	c.relnatts,
	c.relchecks,
	c.relhasoids,
	c.relhaspkey,
	c.relhasrules,
	c.relhastriggers,
	c.relhassubclass,
	c.relrowsecurity,
	c.relforcerowsecurity,
	c.relispopulated,
	c.relreplident,
	c.relfrozenxid,
	c.relminmxid,
	c.relacl,
	c.reloptions,
	c.oid
  FROM
  	pg_catalog.pg_class c
	  JOIN pg_catalog.pg_namespace n ON (n.oid = c.relnamespace)
  ORDER BY
  	n.nspname asc, c.relname asc
;
