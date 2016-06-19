/*
 *  Synopsis:
 *	Select all tuples from pg_stat_activity
 *  Command Line Arguments:
 *	{}
 *  Ordered By:
 *	datname, state, application_name, query_start
 */

select
	datid,
	datname,
	pid,
	usesysid,
	usename,
	application_name,
	client_addr,
	client_hostname,
	client_port,
	backend_start,
	xact_start,
	query_start,
	state_change,
	wait_event_type,
	wait_event,
	state,
	backend_xid,
	backend_xmin,
	query
  from
  	pg_catalog.pg_stat_activity
  order by
  	datname, state, application_name, query_start
;
