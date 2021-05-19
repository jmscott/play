#
#  Synopsis:
#	Write an html <dl> of daily nc tar files
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
	tag =>	'select-daily-nc-tar',
	db =>	$db,
	argv => [
			$lim,
			$off
		],
	sql =>  q(
SELECT DISTINCT
	regexp_replace(dz.tar_path, '^.+[/\\\\]', '') AS tar_name,
	dz.blob,
	pg_size_pretty(bc.byte_count) AS byte_count
  FROM
  	secedgar.daily_nc_tar dz
	  JOIN setcore.byte_count bc ON (
	  	bc.blob = dz.blob
	  )
  ORDER BY
  	tar_name DESC
  LIMIT
  	$1
  OFFSET
  	$2
;));

while (my $row = $q->fetchrow_hashref()) {
	print <<END;
 <dt>
  <a href="/cgi-bin/daily?out=mime.tar&blob=$row->{blob}">$row->{tar_name}</a>
 </dt>
 <dd>$row->{byte_count}</dd>
END
}

print <<END;
</dl>
END
