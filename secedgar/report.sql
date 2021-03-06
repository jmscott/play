/*
 *  Synopsis:
 *	Summarize json for daily edgar pull from sec database.
 *  Note:
 *	Query slowed dramatically when adding min/max zip sizes.  why?
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
	  ->'command-line'
	AS doc
    FROM
    	jsonorg.jsonb_255
    WHERE
  	jsonb_path_exists(doc, '
		$."secedgar.play.jmscott.github.com"
			."command-line"
			."command"
		? (@ == "edgar-put-daily")
	')
), recent_zip AS (
  SELECT
  	(doc->>'zip-blob')::udig AS blob,
  	substring(doc->>'zip-path' FROM '[^/]+$') AS zip_name
    FROM
    	put_daily
    ORDER BY
    	2 DESC
    LIMIT
    	1
), zip_sizes AS (
  SELECT
  	min(bc.byte_count) AS "min_size",
  	max(bc.byte_count) AS "max_size"
    FROM
    	daily_zip d
	  JOIN setcore.byte_count bc ON (
	  	bc.blob = bc.blob
	  )
	  
) SELECT
	count(d.doc) || ' Invocations' AS "edgar-put-daily",
	min((d.doc->'now')::text::timestamp) AS "Earliest Job Time",
	max((d.doc->'now')::text::timestamp) AS "Recent Job Time",
	r.zip_name AS "Recent Zip",
	r.blob AS "Recent Zip Blob",
	pg_size_pretty(bc.byte_count) AS "Recent Zip Size",
	pg_size_pretty(sz.min_size) AS "Min Zip Size",
	pg_size_pretty(sz.max_size) AS "Max Zip Size"
  FROM
 	put_daily d,
	recent_zip r
	  JOIN setcore.byte_count bc ON (
	  	bc.blob = r.blob
	  ),
	zip_sizes sz
  GROUP BY
  	r.blob,
	r.zip_name,
	bc.byte_count,
	sz.min_size,
	sz.max_size
  ORDER BY
  	random()
;
