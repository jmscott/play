/*
 *  Synopsis:
 *	Fetch all classes in a pg_catalog.pg_class
 *
 *  Command Line Arguments: {
 *	"nspname": {
 *		"type":	"text"
 *	}
 *  }
 *
 *  Usage:
 *	psql -f pg_classes.sql --set nspname="'pg_catalog'"
 */
select
	c.oid,
	n.nspname,
	c.*
  from
  	pg_class c
	  join pg_namespace n on (n.oid = c.relnamespace)
  where
  	n.nspname = :nspname
  order by
  	c.relname
;
