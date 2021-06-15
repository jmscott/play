/*
 *  Synopsis:
 *	Merge into materialized table nc_submission from json jobs blobs.
 */
\set ON_ERROR_STOP on

SET search_path TO secedgar,public;

INSERT INTO nc_submission(
	nc_tar_blob,
	nc_file_path,
	line_number,
	element,
	value
) SELECT
	(jj.doc
		->'secedgar.play.jmscott.github.com'
		->'command_line'
		->>'nc_tar_blob')::udig
		AS nc_tar_blob,
	replace(jj.doc
		->'secedgar.play.jmscott.github.com'
		->'command_line'
		->>'nc_file_path',
		E'\n',
		''
	) AS nc_file_path,
	(h->>'line_number')::bigint AS "submission_line_number",
	h->>'element' AS "submission_element",
	h->>'value' AS "submission_value"
  FROM
  	jsonorg.jsonb_255 jj,
	  LATERAL jsonb_array_elements(
	  	jj.doc->'secedgar.play.jmscott.github.com'->'submission_header'
	  ) AS h
  WHERE
  	jsonb_path_exists(jj.doc, '
                $."secedgar.play.jmscott.github.com"
                        ."command_line"
                        ."command"
                ? (@ == "nc2submission")
        ')
  ON CONFLICT
  	DO NOTHING
;
