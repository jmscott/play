/*
 *  Synopsis:
 *	Select all query processes in the server, regardless of state.
 *
 *  Command Line Variables:
 *
 *  Usage:
 *	psql -f pg_stat_activity-count.sql
 */

SELECT
	count(*) AS "activity_count"
  FROM
  	pg_catalog.pg_stat_activity
;
