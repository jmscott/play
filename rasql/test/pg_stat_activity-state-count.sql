/*
 *  Synopsis:
 *	Select count of queries in a particular state.
 *
 *  Command Line Variables:
 *	state	text
 *
 *  Usage:
 *	psql -f pg_stat_activity-state-count.sql --set state="'active'"
 */

SELECT
	count(*) as "activity_count"
  FROM
  	pg_catalog.pg_stat_activity
  WHERE
  	state = :state
;
