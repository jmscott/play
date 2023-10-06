/*
 *  Synopsis:
 *	Hit rate of cache.
 *  Usage:
 *	psql -f cache-hit-rate.sql
 */
SELECT
	sum(heap_blks_read) AS "Heap Blocks Read",
	sum(heap_blks_hit)  AS "Heap Blocks Hit",
	to_char(
	  sum(heap_blks_hit) / (sum(heap_blks_hit) +
	  sum(heap_blks_read)) * 100,
	  'FM99.0') || '%'
		AS "Cache Hit Ratio"
  FROM
  	pg_statio_user_tables
;
