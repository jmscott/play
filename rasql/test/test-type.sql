/*
 *  Synopsis:
 *	Select various postgres types for testing.
 *
 *  Command Line:
 *
 *  Usage:
 *	psql --file test-type
 */
SELECT
	false AS bool_false_val,
	true AS bool_true_val,
	NULL::bool AS bool_null,

	'i am a text string' AS text_string_val,
	NULL::text AS text_string_null,

	'A'::char AS char_val,
	NULL::char AS char_null,

	2345::smallint AS smallint_val,
	NULL::smallint AS smallint_null,

	123456::integer AS integer_val,
	NULL::integer AS integer_null,

	1152921504606846976::bigint AS bigint_val,
	NULL::bigint AS bigint_null,

	3.141592654::real AS real_val,
	NULL::real AS real_null,

	1::oid AS oid_val,
	NULL::oid as oid_null,

	now()::timestamptz AS timestamp_val,
	NULL::timestamptz AS timestamp_null
;
