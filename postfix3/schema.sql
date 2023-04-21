/*
 *  Synopsis:
 *	Schema for postfix3 log data
 *  Note:
 *	Not clear if column syslog2json.log_digest should be unique.
 *	Making syslog2json.log_digest unique simplifies the schema considerable,
 *	but does prevent multiple differnt invocations of "syslog2json"
 *	on the same log file.
 */

\set ON_ERROR_STOP 1
BEGIN;
\set search_path to postfix3,public

DROP SCHEMA IF EXISTS postfix3 CASCADE;
CREATE SCHEMA postfix3;
COMMENT ON SCHEMA postfix3 IS
  'postfix3 syslogs, roughly according to RFC3164 and RFC5424'
;

DROP TYPE IF EXISTS report_type CASCADE;
CREATE TYPE report_type AS ENUM ('full');
COMMENT ON TYPE report_type IS
  'report types for syslog2json command line program'
;

DROP DOMAIN IF EXISTS xx512x1 CASCADE;
CREATE DOMAIN xx512x1 AS bytea
  CHECK (
  	length(value) = 20
  ) NOT NULL
;
COMMENT ON DOMAIN xx512x1 IS
  '20 byte hash value sha1(sha512(sha512))'
;

DROP DOMAIN IF EXISTS in_time CASCADE;
CREATE DOMAIN in_time AS timestamptz
  CHECK (
  	value >= '2023-04-20 18:15:24.199107-05'
  ) NOT NULL
;
COMMENT ON DOMAIN in_time IS
  'Track insert time for immutable tuples'
;

DROP DOMAIN IF EXISTS unix_id CASCADE;
CREATE DOMAIN unix_id AS BIGINT
  CHECK (
  	value > 0
  ) NOT NULL
;
COMMENT ON DOMAIN unix_id IS
  'Typical unix process/user/group numeric ids'
;

DROP TABLE IF EXISTS syslog2json CASCADE;
CREATE TABLE syslog2json
(
	json_digest	xx512x1 PRIMARY KEY,
	log_digest	xx512x1 UNIQUE,

	insert_time	in_time,

	CHECK (
		json_digest != log_digest
	)
);
COMMENT ON TABLE syslog2json IS
  'json output of command "syslog2json"'
;
COMMENT ON COLUMN syslog2json.json_digest IS
  'The crypto digest of the json blob produced by syslog2json'
;
COMMENT ON COLUMN syslog2json.log_digest IS
  'The crypto digest of the syslog file pointed to a particular json blob'
;

DROP TABLE IF EXISTS source_host CASCADE;
CREATE TABLE source_host
(
	host		text CHECK (
				host ~ '^[[::graph::]]{1,64}'
			) PRIMARY KEY,
	insert_time	timestamptz DEFAULT now() NOT NULL
);
COMMENT ON TABLE source_host IS
  'Source hosts (second field) for all syslog files in database'
;

DROP TABLE IF EXISTS syslog2json_os_core CASCADE;
CREATE TABLE syslog2json_os_core 
(
	json_digest	xx512x1 REFERENCES syslog2json PRIMARY KEY,

	args		text[] CHECK (
				array_length(args, 1) > 0
				AND
				array_length(args, 1) < 256
			) NOT NULL,
	executable	text CHECK (
				executable ~ '^[[:graph:]]{1,255}$'
			) NOT NULL,
	pid		unix_id,
	ppid		unix_id,
	uid		unix_id,
	euid		unix_id,
	gid		unix_id,
	egid		unix_id
);
COMMENT ON TABLE syslog2json_os_core IS
  'attributes of particular process invocation from golang core package os.*'
;
COMMENT ON COLUMN syslog2json_os_core.args IS
  'Command line arguments of process invocation: golang os.Args'
;
COMMENT ON COLUMN syslog2json_os_core.executable IS
  'Full file system path to executable syslog2json'
;
COMMENT ON COLUMN syslog2json_os_core.pid IS
  'Process id of invocation of program syslog2json'
;
COMMENT ON COLUMN syslog2json_os_core.ppid IS
  'Parent process id of invocation of program syslog2json'
;
COMMENT ON COLUMN syslog2json_os_core.uid IS
  'User id of invocation of program syslog2json'
;
COMMENT ON COLUMN syslog2json_os_core.euid IS
  'Effective user id of invocation of program syslog2json'
;
COMMENT ON COLUMN syslog2json_os_core.gid IS
  'Group id of invocation of program syslog2json'
;
COMMENT ON COLUMN syslog2json_os_core.gid IS
  'Effective group id of invocation of program syslog2json'
;

DROP TABLE IF EXISTS syslog2json_os_environ CASCADE;
CREATE TABLE syslog2json_os_environ
(
	json_digest	xx512x1 REFERENCES syslog2json PRIMARY KEY,
	env_kv	text[] CHECK (
				array_length(env_kv, 1) > 0
				AND
				array_length(env_kv, 1) < 65536
			) NOT NULL
);
COMMENT ON TABLE syslog2json_os_environ IS
  'Process enviroment of invocation of program syslog2json'
;
COMMENT ON COLUMN syslog2json_os_environ.env_kv IS
  'Array of "key=value" pairs from golang os.Environ()'
;

REVOKE UPDATE ON ALL TABLES IN SCHEMA postfix3 FROM PUBLIC;

COMMIT;
