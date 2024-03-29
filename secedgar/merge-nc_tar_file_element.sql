/*
 *  Synopsis:
 *	Extract/merge json blob into materialized table nc_tar_file_element
 */

\set ON_ERROR_STOP on
\timing
set search_path to secedgar,public;

/*
 *  Find the json blobs (not large json docs) of candiates for merging/
 *  materializing into the table nc_tar_file_element.
 */
WITH mergable AS (
  SELECT
  	jt.blob AS "json_blob"
    FROM
  	jsonorg.jsonb_255 jt
    WHERE
  	jsonb_path_exists(doc, '
                $."secedgar.play.jmscott.github.com"
                        ."command_line"
                        ."command"
                ? (@ == "nc-tar2file-element")
        ')
	AND NOT EXISTS (
	  SELECT
	  	true
	    FROM
	    	nc_tar_file_element mat
	    WHERE
	    	mat.blob = (jt.doc
				->'secedgar.play.jmscott.github.com'
				->'command_line'
				->>'tar_blob')::udig
	)
), elements AS (
  SELECT
  	(jj.doc->'command_line'->>'tar_blob')::udig AS tar_blob,
	ele->>'path' AS file_path,
	(ele->>'size')::bigint AS file_size
    FROM
    	mergable m,
	  LATERAL (
	  	SELECT
			j.doc->'secedgar.play.jmscott.github.com' AS doc
		  FROM
	  		jsonorg.jsonb_255 j
		  WHERE
		  	j.blob = m.json_blob
	  ) jj,
	  LATERAL jsonb_array_elements(
	  	jj.doc->'file_elements'
	  ) AS ele
) INSERT INTO nc_tar_file_element (
	blob,
	file_path,
	file_size
  ) SELECT
  	tar_blob,
	file_path,
	file_size
      FROM
      	elements
  ON CONFLICT
  	DO NOTHING
;

VACUUM ANALYZE nc_tar_file_element;

SELECT
	blob AS tar_blob,
	count(*) AS tar_file_count,
	pg_size_pretty(sum(file_size)) AS byte_count
  FROM
  	nc_tar_file_element
  GROUP BY
  	1
  ORDER BY
  	2 DESC
  LIMIT
  	10
;

SELECT
	blob AS tar_blob,
	pg_size_pretty(sum(file_size)) AS byte_count,
	count(*) AS tar_file_count
  FROM
  	nc_tar_file_element
  GROUP BY
  	1
  ORDER BY
  	sum(file_size) DESC
  LIMIT
  	10
;

SELECT
	count(DISTINCT blob) tar_count,
	count(*) AS total_file_count,
	pg_size_pretty(sum(file_size)) AS total_byte_count
  FROM
  	nc_tar_file_element
;
