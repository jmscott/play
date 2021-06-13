\timing on

/*
 *  Synopsis:
 *	Summarize json job json blobs in {secedgar.play.jmscott.github.com}.
 *  Job Scripts:
 *	nc-tar2file-element
 *	edgar-put-daily
 *	nc2submission
 */
WITH secedgar AS (
  SELECT
	blob
    FROM
  	jsonorg.jsonb_255
    WHERE
    	doc ? 'secedgar.play.jmscott.github.com'
), canonical_sec AS (
  SELECT
	j.doc->'secedgar.play.jmscott.github.com' AS doc,
	j.doc->'secedgar.play.jmscott.github.com'->'command_line'
		AS command_line
    FROM
    	secedgar rj natural join jsonorg.jsonb_255 j
) SELECT
	doc->>'hostname' AS "Host",
	command_line->>'command' AS "Job Command",
	count(*) AS "Run Count"
    FROM
    	canonical_sec
    GROUP BY
    	"Host",
	"Job Command"
    ORDER BY
    	"Host" ASC,
    	"Run Count" DESC,
	"Job Command" ASC
;
