/*
 *  Synopsis:
 *	PostgreSQL schema "jmscott" for bit-entropy functions.
 */

\set ON_ERROR_STOP 1

BEGIN;

set search_path to jmscott,setcore,public;

CREATE SCHEMA IF NOT EXISTS jmscott;	-- do not drop schema

DROP TABLE IF EXISTS bit_pop_count;
CREATE TABLE bit_pop_count
(
	blob		udig
				PRIMARY KEY,
	one_count	bigint CHECK (
				one_count >= 0
			)
);
COMMENT ON TABLE bit_pop_count IS
  'Count of ones bits in blob'
;

DROP TABLE IF EXISTS shift_xor_64 CASCADE;
CREATE TABLE shift_xor_64
(
	blob		udig
				PRIMARY KEY,
	one_counts	bigint[64]
				NOT NULL
);

COMMENT ON TABLE shift_xor_64 IS
  'Count of ones bits of first 64 shift-xor values of blob'
;

COMMENT ON COLUMN shift_xor_64.blob IS
  'The blob shift-xorded up to 64 times'
;

COMMENT ON COLUMN shift_xor_64.one_counts IS
  'The array of first 64 shift-xor bit population counts'
;

COMMIT;
