/*
 *  Synopsis:
 *	Top unused indexes.
 *  Note:
 *	Derived from a blog post at citusdb.
 */

\timing ON
\x

SELECT
	schemaname || '.' || relname AS table,
	indexrelname AS index,
	pg_size_pretty(pg_relation_size(i.indexrelid)) AS index_size,
	idx_scan AS index_scans
  FROM
  	pg_stat_user_indexes ui
	  JOIN pg_index i ON (ui.indexrelid = i.indexrelid)
  WHERE
  	NOT indisunique
	AND
	idx_scan < 50
	AND
	pg_relation_size(relid) > 5 * 8192
  ORDER BY
  	pg_relation_size(i.indexrelid) / nullif(idx_scan, 0) DESC NULLS FIRST,
	pg_relation_size(i.indexrelid) DESC
;
