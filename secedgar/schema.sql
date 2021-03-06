/*
 *  Synospsis:
 *	PostgreSQL schema for edgar secdata.
 *  Description:
 *	A schema for EDGAR data managed by the USA SEC/FTW organizations
 *	and NAICS/SIC categories managed by the census of the USA government.
 *
 *	The NAICS spreadsheets are pulled from these sources:
 *
 *		https://www.census.gov/programs-surveys/cbp/technical-documentation/reference/naics-descriptions.html
 *
 *	Load these spreadsheets into separate tables, one per spreadsheet.
 *
 *		naics.txt
 *		naics2002.txt
 *		naics2007.txt
 *		naics2012.txt
 *		naics2017.txt
 *		sic86_87.txt
 *		sic88_97.txt
 *  Note:
 *	Add source url in comments for SGML files!
 *
 *	Add SQL COMMENTS to all tables, dork!
 *
 *	More examples of \COPY are desperately needed on the PG web site:
 *
 *		https://www.postgresql.org/docs/current/sql-copy.html
 */

\set ON_ERROR_STOP 1
SET search_path to secedgar,public;

BEGIN;

DROP SCHEMA IF EXISTS secedgar CASCADE;
CREATE SCHEMA secedgar;

DROP DOMAIN IF EXISTS siccode CASCADE;
CREATE DOMAIN siccode AS text CHECK (
	value ~ '^[0-9][0-9/\\-]{0,6}$'
	OR
	value = '------'
	OR
	value = '----'
);

DROP DOMAIN IF EXISTS sicdesc CASCADE;
CREATE DOMAIN sicdesc AS text CHECK (
	value ~ '[[:graph:]]'
	AND
	value !~ '^[[:space:]]'		--  no leading space
	AND
	value !~ '[[:space:]]$'		--  no trailing space
	AND
	length(value) <= 128
	AND
	length(value) > 0
);

--  Create & Load csv naics2012.txt, stripping non-utf8

DROP TABLE IF EXISTS naics2017 CASCADE;
CREATE TABLE naics2017
(
	code		siccode PRIMARY KEY,

	"description"	sicdesc NOT NULL
);

--  Note: strip non-utf codes from certain entries.
\COPY naics2017 FROM PROGRAM 'iconv -f utf-8 -t utf-8 -c naics2017.txt; test $? -le 1' DELIMITER ',' CSV HEADER

--  Create and Load csv naics2012.txt, stripping non-utf8

DROP TABLE IF EXISTS naics2012 CASCADE;
CREATE TABLE naics2012
(
	code		siccode PRIMARY KEY,

	"description"	sicdesc NOT NULL
);

--  Note: strip non-utf codes from certain entries.
\COPY naics2012 FROM PROGRAM 'iconv -f utf-8 -t utf-8 -c naics2012.txt; test $? -le 1' DELIMITER ',' CSV HEADER

--  Create and Load csv naics2007.txt, stripping non-utf8

DROP TABLE IF EXISTS naics2007 CASCADE;
CREATE TABLE naics2007
(
	code		siccode PRIMARY KEY,

	"description"	sicdesc NOT NULL
);

--  Note: load fixed width codes
\COPY naics2007 FROM PROGRAM 'sed "s/  /	/" naics2007.txt | sed "s/ *$//"' DELIMITER E'\t' CSV HEADER

--  Create and Load csv naics2002.txt, stripping non-utf8

DROP TABLE IF EXISTS naics2002 CASCADE;
CREATE TABLE naics2002
(
	code		siccode PRIMARY KEY,

	"description"	sicdesc NOT NULL
);

--  Note: load fixed width codes

\COPY naics2002 FROM PROGRAM 'sed "s/  /	/" naics2002.txt | sed "s/ *$//"' DELIMITER E'\t' CSV HEADER
--  Create and Load csv naics2002.txt, stripping non-utf8

DROP TABLE IF EXISTS naics CASCADE;
CREATE TABLE naics
(
	code		siccode PRIMARY KEY,

	"description"	sicdesc NOT NULL
);

--  Note: load fixed width codes

\COPY naics FROM PROGRAM 'sed "s/  /	/" naics.txt | sed "s/ *$//"' DELIMITER E'\t' CSV HEADER

DROP TABLE IF EXISTS sic88_97 CASCADE;
CREATE TABLE sic88_97
(
	code		siccode CHECK (
				length(code) = 4
			) PRIMARY KEY,

	"description"	sicdesc NOT NULL
);

--  Note: load fixed width codes

\COPY sic88_97 FROM PROGRAM 'sed "s/  /	/" sic88_97.txt |tee x' DELIMITER E'\t' CSV HEADER

DROP TABLE IF EXISTS sic86_87 CASCADE;
CREATE TABLE sic86_87
(
	code		siccode CHECK (
				length(code) = 4
			),

	"description"	sicdesc NOT NULL,

	--  Note: duplicate codes exist (i.e, 5012).  not sure why

	PRIMARY KEY	(code, "description")
);

--  Note: load fixed width codes

\COPY sic86_87 FROM 'sic86_87.txt' DELIMITER E'\t' CSV HEADER

DROP VIEW IF EXISTS edgar_put_daily;
CREATE VIEW edgar_put_daily AS
  WITH tars AS (
    SELECT
  	doc
	  ->'secedgar.play.jmscott.github.com'
	  ->'command_line'
	AS doc,
	blob
      FROM
    	jsonorg.jsonb_255
      WHERE
  	jsonb_path_exists(doc, '
		$."secedgar.play.jmscott.github.com"
			."command_line"
			."command"
		? (@ == "edgar-put-daily")
	')
) SELECT
	(doc->>'tar_blob')::udig AS blob,
	doc->>'tar_path' AS tar_path,
	(doc->>'now')::timestamptz AS job_time,
	blob AS doc_blob
    FROM
    	tars
    WHERE
	length(doc->>'now') > 0
;
COMMENT ON VIEW edgar_put_daily IS
  'Fetched Edgar SEC Daily Tar Files'
;

/*
 *  Note:
 *	No cross-indexes in PostgreSQL.
 */
DROP INDEX IF EXISTS jsonorg.idx_edgar_put_daily_path;
CREATE INDEX idx_edgar_put_daily_path
  ON jsonorg.jsonb_255 (
  	jsonb_path_exists(doc, '
		$."secedgar.play.jmscott.github.com"
			."command_line"
			."command"
		? (@ == "edgar-put-daily")
	')
);
/*
 *  Note:
 *	No cross-indexes in PostgreSQL.
 */
DROP INDEX IF EXISTS jsonorg.idx_secedgar_nc_submission;
CREATE INDEX idx_secedgar_nc_submission
  ON jsonorg.jsonb_255 (
  	jsonb_path_exists(doc, '
		$."secedgar.play.jmscott.github.com"
			."command_line"
			."command"
		? (@ == "nc2submission")
	')
);

DROP DOMAIN IF EXISTS tsv_text CASCADE;
CREATE DOMAIN tsv_text AS text
  CHECK (
	value ~ '[[:graph:]]'
	AND
	length(value) < 512
	AND
	value !~ E'\t'
  ) NOT NULL
;
COMMENT ON DOMAIN tsv_text IS
  'Text field extracted from SGML files of SEC'
;

DROP TABLE IF EXISTS tsv_SGML_SUBMISSION;
CREATE TABLE tsv_SGML_SUBMISSION (
	DATA_ELEMENT	tsv_text,
	TAG		tsv_text,
	DESCRIPTION	tsv_text,
	LENGTH		tsv_text,
	END_TAG		tsv_text,
	CHARACTERISTIC	tsv_text,
	LIMITS		tsv_text,
	FORMAT		tsv_text
);
COMMENT ON TABLE tsv_SGML_SUBMISSION IS
  'Contents of scrubbed SEC file SGML_SUBMISSION.tsv'
;

\COPY tsv_SGML_SUBMISSION FROM 'SGML-SUBMISSION.tsv' DELIMITER E'\t' CSV HEADER

DROP TABLE IF EXISTS tsv_SGML_DOCUMENT;
CREATE TABLE tsv_SGML_DOCUMENT (
	DATA_ELEMENT	tsv_text,
	TAG		tsv_text,
	DESCRIPTION	tsv_text,
	LENGTH		tsv_text,
	END_TAG		tsv_text,
	CHARACTERISTIC	tsv_text,
	LIMITS		tsv_text,
	FORMAT		tsv_text
);
COMMENT ON TABLE tsv_SGML_DOCUMENT IS
  'Contents of scrubbed SEC file SGML_DOCUMENT.tsv'
;

\COPY tsv_SGML_DOCUMENT FROM 'SGML-DOCUMENT.tsv' DELIMITER E'\t' CSV HEADER

/*
 *  Note:
 *	Uggh.  Missing the order of the file elements.  The gnutar schema
 *	must not do this!
 *
 *	The eventual gnutar.tar_file_element schema/table needs to extract
 *	the timestamptz.
 *
 *	A view will replace this table, changing to reference table
 *	gnutar.tar_file_element.
 */
DROP TABLE IF EXISTS nc_tar_file_element;
CREATE TABLE nc_tar_file_element
(
	blob	udig	REFERENCES setcore.service,
	file_path	text	CHECK (
				file_path ~ '[[:graph:]]'
				AND
				length(file_path) > 0
				AND
				length(file_path) < 256
			)
			NOT NULL,
	file_size	bigint CHECK (
				file_size >= 0
			) NOT NULL,
	PRIMARY KEY	(blob, file_path)
);
COMMENT ON TABLE nc_tar_file_element IS
  'Files Elements in SEC EDGAR tar file YYYYMMDD.nc.tar.gz'
;

DROP TABLE IF EXISTS nc_submission CASCADE;
CREATE TABLE nc_submission(
	nc_tar_blob	udig,
	nc_file_path	text CHECK (
				nc_file_path ~ '[[:graph:]]'
				AND
				length(nc_file_path) < 256
			),
	line_number	bigint CHECK (
				line_number > 0
			),
	element		text CHECK (
				element ~ '^[A-Z].*[A-Z0-9_-].*$'
			),
	value		text CHECK (
				length(value) < 256
			),
	PRIMARY KEY	(nc_tar_blob, nc_file_path, line_number)
);

COMMENT ON TABLE nc_submission IS
  'The submission section of nc|corr* file in daily nc tar, no <TEXT>'
;

COMMIT;
