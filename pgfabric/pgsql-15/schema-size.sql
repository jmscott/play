/*
 *  Synopsis:
 *	Enumerate total size of schemas in a database
 *  Note:
 *	Higher, inaccurate results are obtained when not qualifying on
 *	pg_class.relkind = 'r' .  Why?  Apparently the same object
 *	exists in different relkind.
 */
WITH schema_size(schema_name, table_total) AS (
  SELECT
    	n.nspname,
	sum(pg_total_relation_size(c.oid))
    FROM
      	pg_catalog.pg_class c,
	pg_catalog.pg_namespace n
    WHERE
    	c.relnamespace = n.oid
	AND
	c.relkind = 'r'
    GROUP BY
    	n.nspname
)
SELECT
	schema_name AS "Schema",
	pg_size_pretty(table_total) AS "Total Size",
	(table_total/ pg_database_size(current_database()) * 100) ::int
		as "Percentage"
  FROM
  	schema_size
  ORDER BY
	"Percentage" DESC,
	"Total Size" DESC
;
