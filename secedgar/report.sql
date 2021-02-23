/*
 *  Synopsis:
 *	Summarize json for daily edgar pull from sec database.
 *  Note:
 *	Do a daily summary of all jobs.
 *
 *	Investigate creating parital index on
 *	"secedgar.play.jmscott.github.com".  In PostgreSQL docs, read scrion
 *	8.14.4. jsonb Indexing.
 */
\set ON_ERROR_STOP 1
\timing 1
\x on
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
), recent_zip AS(
  SELECT
  	(doc->>'zip-blob')::udig AS zip_blob,
  	substring(doc->>'zip-path' FROM '[^/]+$') AS zip_name
    FROM
    	put_daily
    ORDER BY
    	2 DESC
    LIMIT
    	1
)
SELECT
	count(d.doc) || ' Invocations' AS "edgar-put-daily",
	min((d.doc->'now')::text::timestamp) AS "Earliest Time",
	max((d.doc->'now')::text::timestamp) AS "Recent Time",
	r.zip_name AS "Recent Zip",
	r.zip_blob AS "Recent Blob",
	pg_size_pretty(bc.byte_count) AS "Zip Size"
  FROM
 	put_daily d,
	recent_zip r
	  JOIN setcore.byte_count bc ON (
	  	bc.blob = r.zip_blob
	  )
  GROUP BY
  	r.zip_blob,
	r.zip_name,
	bc.byte_count
  ORDER BY
  	random()
;
