/*
 *  Synopsis:
 *	Fetch all classes in a pg_catalog.pg_class
 *
 *  Command Line Arguments: {
 *  }
 *
 *  Usage:
 *	psql -f pg_classes.sql
 */
select
	c.oid,
	n.nspname,
	c.*
  from
  	pg_class c
	  join pg_namespace n on (n.oid = c.relnamespace)
  order by
  	n.nspname, c.relname
;
