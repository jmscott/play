#
#  Synopsis:
#
require 'jmscott/dbi-pg.pl';

our %QUERY_ARG;

print <<END;
<dl$QUERY_ARG{id_att}$QUERY_ARG{class_att}>
END

my $db = dbi_pg_connect();
my $q = dbi_pg_select(
	tag =>	'select-daily-zip',
	db =>	$db,
	argv => [],
	sql =>  q(
SELECT
	dz.zip_path,
	dz.blob,
	to_char(dz.job_time, 'FMDay, FMDD  HH12:MI:SS') AS job_time,
	pg_size_pretty(bc.byte_count) AS byte_count
  FROM
  	secedgar.daily_zip dz
	  JOIN setcore.byte_count bc ON (
	  	bc.blob = dz.blob
	  )
  ORDER BY
  	dz.job_time DESC
;));

while (my $row = $q->fetchrow_hashref()) {
	print <<END;
 <dt>$row->{zip_path}</dt>
 <dd>$row->{byte_count} @ $row->{job_time}</dd>
END
}

print <<END;
</dl>
END