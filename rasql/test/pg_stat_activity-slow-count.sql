/*
 *  Synopsis:
 *	Select count of queries slower than a certain duration.
 *
 *  Description:
 *	Select the count of all query processes which have been running longer
 *	than a particular duration in time.  The query start time is compared
 *	to the attribute pg_stat_activity.query_start.	The duration must be
 *	expressed as a negative time interval, such as -1min.
 *
 *  Command Line Variables:
 *	duration text
 *
 *  Usage:
 *	psql --file pg_stat_activity-slow-count.sql --set duration="'-1min'"
 */

SELECT
	count(*) as slow_count
  FROM
  	pg_catalog.pg_stat_activity
  WHERE
  	query_start <= (now() + :duration)
;
