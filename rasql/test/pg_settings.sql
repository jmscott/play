/*
 *  Synopsis:
 *	Show all run time system settings in the view pg_catalog.pg_settings.
 *  Description:
 *	Show all attributes for all run time system settings in the view
 *	pg_catalog.pg_settings.  The settings are ordered by the name
 *	in ascending order.
 *  Usage:
 *	psql --file pg_settings.sql
 */
SELECT
	name,
	setting,
	unit,
	category,
	short_desc,
	extra_desc,
	context,
	vartype,
	source,
	min_val,
	max_val,
	enumvals,
	boot_val
	reset_val,
	sourcefile,
	sourceline,
	pending_restart
  FROM
  	pg_catalog.pg_settings
  ORDER BY
  	name ASC
;
