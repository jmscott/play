/*
 *  Synopsis:
 *	Fetch all classes in a particular name space
 *
 *  Command Line Arguments: {
 *	"nspname": {
 *		"type":	"text"
 *	}
 *  }
 *
 *  Usage:
 *	psql -f pg_class-by-nsname.sql --set nspname="'pg_catalog'"
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
