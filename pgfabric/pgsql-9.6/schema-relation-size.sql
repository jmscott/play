/*
 *  Synopsis:
 *	Enumerate total size of each relation in a schema in a database
 *  Command Line Arguments:
 *	1	nspname::text	Name Space
 */
WITH schema_relation_size(relation_name, relation_size) AS (
  SELECT
	c.relname,
	pg_total_relation_size(c.oid)::numeric
    FROM
      	pg_catalog.pg_class c,
	pg_catalog.pg_namespace n
    WHERE
    	n.nspname = :nspname
	and
    	c.relnamespace = n.oid
	AND
	c.relkind = 'r'
)
SELECT
	relation_name as "Relation Name",
	pg_size_pretty(relation_size) AS "Total Relation Size",
	(relation_size / pg_database_size(current_database()) * 100)::int
			as "Percentage of DB Size"
  FROM
  	schema_relation_size
  ORDER BY
  	relation_size desc,
	relation_name asc
;
