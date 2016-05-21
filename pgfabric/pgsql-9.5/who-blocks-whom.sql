/*
 *  Synopsis:
 *	Who blocks whom, query wise
 *  See:
 *	http://www.postgresql.org/docs/9.5/static/explicit-locking.html
 *	http://big-elephants.com/2013-09/exploring-query-locks-in-postgres/
 *  Note:
 *	Need to add locktype and duration.
 */
SELECT
	blockinga.datname,
	blockeda.pid AS blocked_pid,
	blockeda.query AS blocked_query,
  	blockinga.pid AS blocking_pid,
	blockinga.query AS blocking_query
  FROM
  	pg_catalog.pg_locks blockedl
  	JOIN pg_stat_activity blockeda ON blockedl.pid = blockeda.pid
  	JOIN pg_catalog.pg_locks blockingl ON(
		blockingl.transactionid = blockedl.transactionid
  		AND
		blockedl.pid != blockingl.pid
	)
	JOIN pg_stat_activity blockinga ON blockingl.pid = blockinga.pid
  WHERE
  	NOT blockedl.granted
;
