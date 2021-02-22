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
\x
WITH put_daily AS (
  SELECT
  	doc
		->'secedgar.play.jmscott.github.com'
		->'command-line'
	AS 
		doc
    FROM
    	jsonorg.jsonb_255
    WHERE
  	jsonb_path_exists(doc, '
		$."secedgar.play.jmscott.github.com"
			."command-line"
			."command"
		? (@ == "edgar-put-daily")
	')
) SELECT
	count(doc) || ' Invocations' AS "edgar-put-daily",
	min((doc->'now')::text::timestamp) AS "Earliest Time",
	max((doc->'now')::text::timestamp) AS "Recent Time"
  FROM
 	put_daily
;
