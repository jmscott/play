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

DROP DOMAIN IF EXISTS uint63 CASCADE;
CREATE DOMAIN uint63 AS BIGINT
  CHECK (
  	value >= 0
  )
;
COMMENT ON DOMAIN uint63 IS
  'unsigned int in [0, 2^63]'
;

DROP DOMAIN IF EXISTS log_count CASCADE;
CREATE DOMAIN log_count AS uint63 NOT NULL;
COMMENT ON DOMAIN log_count IS
  'counts always existing and >= 0 in syslog messages'
;

DROP DOMAIN IF EXISTS log_line_number CASCADE;
CREATE DOMAIN log_line_number AS uint63
  CHECK (
  	value > 0
  ) NOT NULL
;
COMMENT ON DOMAIN log_line_number IS
  'line number in a syslog file'
;

DROP DOMAIN IF EXISTS log_seek_offset CASCADE;
CREATE DOMAIN log_seek_offset AS uint63
  CHECK (
  	value >= 0
  ) NOT NULL
;
COMMENT ON DOMAIN log_seek_offset IS
  'seek offset in a syslog file'
;

DROP DOMAIN IF EXISTS uint15 CASCADE;
CREATE DOMAIN uint15 AS BIGINT
  CHECK (
  	value >= 0
  )
;
COMMENT ON DOMAIN uint15 IS
  'unsigned int in [0, 2^15]'
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

DROP DOMAIN IF EXISTS log_time CASCADE;
CREATE DOMAIN log_time AS timestamptz
  CHECK (
  	value >= '1970-01-01'
  ) NOT NULL
;
COMMENT ON DOMAIN log_time IS
  'Reasonable timestamp for syslog messages'
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
	json_digest	xx512x1
				REFERENCES syslog2json
				ON DELETE CASCADE
				PRIMARY KEY,

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
	json_digest	xx512x1
				REFERENCES syslog2json
				ON DELETE CASCADE
				PRIMARY KEY,
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

DROP TABLE IF EXISTS syslog2json_custom_regexp CASCADE;
CREATE TABLE syslog2json_custom_regexp
(
	json_digest	xx512x1
				REFERENCES syslog2json
				ON DELETE CASCADE,

	tag		text CHECK (
				tag ~ '^[a-z_-]{1,16}$'
			),
	regexp		text CHECK (
				length(regexp) > 0
				AND
				length(regexp)  < 256
			),
	PRIMARY KEY	(json_digest, tag)
);

DROP TABLE IF EXISTS syslog2json_scan;
CREATE TABLE syslog2json_scan
(
	json_digest	xx512x1
				REFERENCES syslog2json
				ON DELETE CASCADE
				PRIMARY KEY,
	report_type	report_type NOT NULL,
	line_count	log_count,
	byte_count	log_count,
	input_digest	xx512x1 NOT NULL,
	time_location_name	text CHECK (
				length(time_location_name) > 0
				AND
				length(time_location_name) < 128
			) NOT NULL,
	year		uint15 NOT NULL
);

DROP TABLE IF EXISTS syslog2json_source_host CASCADE;
CREATE TABLE syslog2json_source_host
(
	json_digest     xx512x1
			REFERENCES syslog2json_scan
			ON DELETE CASCADE,

	host_name	text CHECK (
				host_name ~ '^[[::graph::]]{1,64}$'
			),
	min_log_time	log_time,
	max_log_time	log_time,
	
	min_line_number	log_line_number,
	min_line_seek_offset	log_seek_offset,
	
	max_line_number	log_line_number,
	max_line_seek_offset	log_seek_offset,

	PRIMARY KEY	(json_digest, host_name),

	CONSTRAINT log_time_range CHECK (
		min_log_time <= max_log_time
	),

	CONSTRAINT line_number_range CHECK (
		min_line_number <= max_line_number
	),

	CONSTRAINT line_seek_offsetrange CHECK (
		min_line_seek_offset <= max_line_seek_offset
	)
);
COMMENT ON TABLE syslog2json_source_host IS
  'Hosts seen in a scan of a syslog file'
;

DROP TABLE IF EXISTS syslog2json_source_host_count_stat CASCADE;
CREATE TABLE syslog2json_source_host_count_stat
(
	json_digest     xx512x1,

	host_name	text,
	PRIMARY KEY	(json_digest, host_name),
	FOREIGN KEY	(json_digest, host_name)
				REFERENCES syslog2json_source_host
				ON DELETE CASCADE,

	unknown_line_count	log_count,
	warning_count		log_count,
	statistics_count	log_count,
	fatal_count		log_count,
	daemon_started_count	log_count,
	refresh_postfix_count	log_count,
	reload_count		log_count,
	connect_from_count	log_count,
	lost_connect_count	log_count,
	disconnect_from_count	log_count,
	connect_to_count	log_count,
	backwards_compat_count	log_count,
	message_repeated_count	log_count,
	start_postfix_count	log_count,
	status_sent_count	log_count,
	status_bounced_count	log_count,
	status_deferred_count	log_count,
	status_expired_count	log_count
);
COMMENT ON TABLE syslog2json_source_host_count_stat IS
  'Count stats associated with a particular source host'
;

DROP TABLE IF EXISTS syslog2json_source_host_process_count CASCADE;
CREATE TABLE syslog2json_source_host_process_count
(
	json_digest     xx512x1,

	host_name	text,
	process_name	text CHECK (
				process_name ~ '^[[::graph::]]{1,63}$'
			),
	line_count	log_count,

	PRIMARY KEY (json_digest, host_name, process_name),

	FOREIGN KEY (json_digest, host_name)
			REFERENCES syslog2json_source_host_count_stat
			ON DELETE CASCADE
);
COMMENT ON TABLE syslog2json_source_host_process_count IS
  'Counts of the (third) process field in syslog messages'
;

DROP TABLE IF EXISTS syslog2json_source_host_queue_id CASCADE;
CREATE TABLE syslog2json_source_host_queue_id
(
	json_digest     xx512x1,
	host_name       text,
	queue_id	text CHECK(
				queue_id ~ '^([A-Z0-9]{8,12})$'
			),
	PRIMARY KEY	(json_digest, host_name, queue_id),

	FOREIGN KEY	(json_digest, host_name)
				REFERENCES syslog2json_source_host
				ON DELETE CASCADE,

	line_count	log_count,

	min_log_time	log_time,
	max_log_time	log_time,

	min_line_number		log_line_number,
	min_line_seek_offset	log_seek_offset,

	max_line_number		log_line_number,
	max_line_seek_offset	log_seek_offset,

	status_sent_count	log_count,
	status_bounced_count	log_count,
	status_deferred_count	log_count,
	status_expired_count	log_count,

	CONSTRAINT log_time_range CHECK (
		min_log_time <= max_log_time
	),

	CONSTRAINT line_number_range CHECK (
		min_line_number <= max_line_number
	),

	CONSTRAINT line_seek_offset_range CHECK (
		min_line_seek_offset <= max_line_seek_offset
	),

	empty_message_id_count	log_count
);
COMMENT ON TABLE syslog2json_source_host_queue_id IS
  'All log lines associated with a particular queue id'
;

DROP TABLE IF EXISTS syslog2json_source_host_queue_id_message_id CASCADE;
CREATE TABLE syslog2json_source_host_queue_id_message_id
(
	json_digest     xx512x1,
	host_name       text,
	queue_id	text,
	message_id	text CHECK (
				message_id ~ '^[[::graph::]]{1,255}$'
			),

	PRIMARY KEY	(json_digest, host_name, queue_id, message_id),

	FOREIGN KEY	(json_digest, host_name, queue_id)
				REFERENCES syslog2json_source_host_queue_id
				ON DELETE CASCADE
);
COMMENT ON TABLE syslog2json_source_host_queue_id_message_id IS
  'Mail message ID associated with a queue id for a particular source host'
;

REVOKE UPDATE ON ALL TABLES IN SCHEMA postfix3 FROM PUBLIC;

COMMIT;
