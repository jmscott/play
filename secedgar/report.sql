/*
 *  Synopsis:
 *	Summarize json for daily edgar pull of nc tar from sec database.
 *  Note:
 *	Query slowed dramatically when adding min/max tar sizes.  why?
 *
 *	Do a daily summary of all jobs.
 *
 *	Investigate creating parital index on
 *	"secedgar.play.jmscott.github.com".  In PostgreSQL docs, read scrion
 *	8.14.4. jsonb Indexing.
 */
\set ON_ERROR_STOP 1
\timing 1
\x on

SET search_path TO secedgar,public;
WITH put_daily AS (
  SELECT
  	doc
	  ->'secedgar.play.jmscott.github.com'
	  ->'command_line'
	AS doc
    FROM
    	jsonorg.jsonb_255
    WHERE
  	jsonb_path_exists(doc, '
		$."secedgar.play.jmscott.github.com"
			."command_line"
			."command"
		? (@ == "edgar-put-daily")
	')
	AND
	length(doc->
		'secedgar.play.jmscott.github.com'->
		'command_line'->>
		'now'
	) > 0
), recent_tar AS (
  SELECT
  	(doc->>'tar_blob')::udig AS blob,
  	substring(doc->>'tar_path' FROM '[^/]+$') AS tar_name
    FROM
    	put_daily
    ORDER BY
    	2 DESC
    LIMIT
    	1
), tar_sizes AS (
  SELECT
  	min(bc.byte_count) AS "min_size",
  	max(bc.byte_count) AS "max_size"
    FROM
    	edgar_put_daily d
	  JOIN setcore.byte_count bc ON (
	  	bc.blob = bc.blob
	  )
	  
) SELECT
	count(d.doc) || ' Invocations' AS "edgar-put-daily",
	min((d.doc->'now')::text::timestamp) AS "Earliest Job Time",
	max((d.doc->'now')::text::timestamp) AS "Recent Job Time",
	r.tar_name AS "Recent TAR",
	r.blob AS "Recent TAR Blob",
	pg_size_pretty(bc.byte_count) AS "Recent TAR Size",
	pg_size_pretty(sz.min_size) AS "Min TAR Size",
	pg_size_pretty(sz.max_size) AS "Max TAR Size"
  FROM
 	put_daily d,
	recent_tar r
	  JOIN setcore.byte_count bc ON (
	  	bc.blob = r.blob
	  ),
	tar_sizes sz
  GROUP BY
  	r.blob,
	r.tar_name,
	bc.byte_count,
	sz.min_size,
	sz.max_size
  ORDER BY
  	random()
;

\x off
\i lib/select-json-job-count.sql
