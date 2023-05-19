/*
 *  Synopsis:
 *	Fetch all classes in a particular name space
 *
 *  Description:
 *	Fetch all the classes for a particular name space.
 *	The classses are ascendingly ordered by the relation name.
 *
 *  Command Line Variables:
 *	nspname	text
 *
 *  Usage:
 *	psql -f pg_class-by-nspname.sql --set nspname="'pg_catalog'"
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
  WHERE
  	n.nspname = :nspname
  ORDER BY
  	c.relname ASC
;
