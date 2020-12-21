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
 *	More examples of \COPY are desparatly needed on the PG web site:
 *
 *		https://www.postgresql.org/docs/current/sql-copy.html
 *
 *	Is "secedgar" a reasonable name for the schema?
 *	Perhaps "usasec" would be more accurate.
 */

\set ON_ERROR_STOP 1
SET search_path to secedgar,public;

BEGIN;

DROP SCHEMA IF EXISTS secedgar;

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

COMMIT;
