/*
 *  Synopsis:
 *	Which tuples qualify for auto analyze.
 *   Note:
 *	What about per table values for settings?
 */
\x auto
WITH rel_stat as (
  SELECT
  	ut.schemaname,
	ut.relname,
	ut.last_analyze,
	ut.last_autoanalyze,
	pg_class.reltuples,
	ut.n_mod_since_analyze,
	current_setting('autovacuum_analyze_threshold')::bigint +
		(current_setting('autovacuum_analyze_scale_factor')::numeric *
		 pg_class.reltuples) AS aa_threshold
  FROM
  	pg_catalog.pg_stat_user_tables ut
	  JOIN pg_catalog.pg_class on ut.relid = pg_class.oid
) SELECT
	st.schemaname AS "Schema",
	st.relname AS "Relation",
	to_char(st.last_analyze, 'YYYY-MM-DD HH24:MI') AS
		"Most Recent Analyze",
	to_char(st.last_autoanalyze, 'YYYY-MM-DD HH24:MI') AS 
		"Most Recent AutoAnalyze",
	to_char(st.reltuples, '9G999G999G999') AS "Tuple Count",
	to_char(st.n_mod_since_analyze, '9G999G999G999') AS
		"Modified Tuple Count",
	to_char(st.aa_threshold, '9G999G999G999') AS
		"AutoAnalyze Threshold",
	CASE
         WHEN
	 	st.aa_threshold < st.n_mod_since_analyze
         THEN
	 	'Yes'
         ELSE
	 	'No'
	END AS "Expect AutoAnalyze"
 FROM
 	rel_stat st
 ORDER BY
 	"Expect AutoAnalyze" DESC,
	st.schemaname,
	st.relname;

