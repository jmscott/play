#
#  Synopsis:
#	Write an html <dl> of daily nc zip files
#
require 'jmscott/dbi-pg.pl';

our %QUERY_ARG;
my $lim = $QUERY_ARG{lim};
my $off = $QUERY_ARG{off};

print <<END;
<dl$QUERY_ARG{id_att}$QUERY_ARG{class_att}>
END

my $db = dbi_pg_connect();
my $q = dbi_pg_select(
	tag =>	'select-daily-nc-zip',
	db =>	$db,
	argv => [
			$lim,
			$off
		],
	sql =>  q(
SELECT
	dz.zip_path,
	dz.blob,
	to_char(dz.job_time, 'Dy, Mon dd, yyyy') AS job_time,
	pg_size_pretty(bc.byte_count) AS byte_count
  FROM
  	secedgar.daily_nc_zip dz
	  JOIN setcore.byte_count bc ON (
	  	bc.blob = dz.blob
	  )
  ORDER BY
  	dz.job_time DESC
  LIMIT
  	$1
  OFFSET
  	$2
;));

while (my $row = $q->fetchrow_hashref()) {
	print <<END;
 <dt>
  <a href="/cgi-bin/daily?out=mime.zip&blob=$row->{blob}">$row->{zip_path}</a>
 </dt>
 <dd>$row->{byte_count} @ $row->{job_time}</dd>
END
}

print <<END;
</dl>
END
