/*
 *  Synopsis:
 *	Fetch all classes in a particular name space
 *  Command Line Variables:
 *	nspname	text
 *  Usage:
 *	psql -f pg_class-by-nsname.sql --set nspname="'pg_catalog'"
 */
select
	c.oid,
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
	c.reloptions
  from
  	pg_catalog.pg_class c
	  join pg_catalog.pg_namespace n on (n.oid = c.relnamespace)
  where
  	n.nspname = :nspname
  order by
  	c.relname
;
