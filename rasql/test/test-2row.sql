/*
 *  Synopsis:
 *	Select various postgres types for testing.
 *  Command Line:
 *
 *  Usage:
 *	psql --file test-type
 */
SELECT
	1 AS val_1,
	2 AS val_2
UNION
SELECT
	3,
	4
;
