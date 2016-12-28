/*
 *  Synopsis:
 *	Which tuples qualify for auto vacuum.
 *   Note:
 *	What about per table values for settings?
 */
\x auto
WITH rel_stat as (
  SELECT
  	ut.schemaname,
	ut.relname,
	ut.last_vacuum,
	ut.last_autovacuum,
	pg_class.reltuples,
	ut.n_dead_tup,
	current_setting('autovacuum_vacuum_threshold')::bigint +
		(current_setting('autovacuum_vacuum_scale_factor')::numeric *
		 pg_class.reltuples) AS av_threshold
  FROM
  	pg_catalog.pg_stat_user_tables ut
	  JOIN pg_catalog.pg_class on ut.relid = pg_class.oid
)
SELECT
	st.schemaname AS "Schema",
	st.relname AS "Relation",
	to_char(st.last_vacuum, 'YYYY-MM-DD HH24:MI') AS
		"Most Recent Vacuum",
	to_char(st.last_autovacuum, 'YYYY-MM-DD HH24:MI') AS 
		"Most Recent AutoVacuum",
	to_char(st.reltuples, '9G999G999G999') AS "Tuple Count",
	to_char(st.n_dead_tup, '9G999G999G999') AS "Dead Tuple Count",
	to_char(st.av_threshold, '9G999G999G999') AS
		"Autovacuum Threshold",
	CASE
         WHEN
	 	st.av_threshold < st.n_dead_tup
         THEN
	 	'Yes'
         ELSE
	 	'No'
	END AS "Expect Autovacuum"
 FROM
 	rel_stat st
 ORDER BY
 	"Expect Autovacuum" DESC,
	st.schemaname,
	st.relname;

