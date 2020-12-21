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
 *	Is "secedgar" a reasonable name for the schema?
 *	Perhaps "usasec" would be more accurate.
 */

\set ON_ERROR_STOP 1
SET search_path to secedgar,public;

BEGIN;

DROP SCHEMA IF EXISTS secedgar;

DROP DOMAIN IF EXISTS siccode CASCADE;
CREATE DOMAIN siccode AS text CHECK (
	value ~ '^[0-9][0-9/-]{0,6}$'
	OR
	value = '------'
);

DROP DOMAIN IF EXISTS sicdesc CASCADE;
CREATE DOMAIN sicdesc AS text CHECK (
	value ~ '[[:graph:]]'
	AND
	value !~ '^[[:space:]]'
	AND
	value !~ '[[:space:]]$'
	AND
	length(value) <= 255
	AND
	length(value) > 0
);

DROP TABLE IF EXISTS naics2017 CASCADE;
CREATE TABLE naics2017
(
	code		siccode PRIMARY KEY,

	"description"	sicdesc NOT NULL
);

--  Note: strip non-utf codes from certain entries.
\COPY naics2017 FROM PROGRAM 'iconv -f utf-8 -t utf-8 -c naics2017.txt; test $? -le 1' DELIMITER ',' CSV HEADER

COMMIT;
