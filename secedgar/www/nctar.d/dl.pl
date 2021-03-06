#
#  Synopsis:
#	Write an html <dl> of daily nc tar files
#
require 'jmscott/dbi-pg.pl';

our %QUERY_ARG;
my $blob = $QUERY_ARG{blob};

print <<END;
<dl$QUERY_ARG{id_att}$QUERY_ARG{class_att}>
END

my $db = dbi_pg_connect();
my $q = dbi_pg_select(
	tag =>	'select-daily-nc-tar',
	db =>	$db,
	argv => [
			$blob,
		],
#
#  Note:
#	query not correct
#
	sql =>  q(
SELECT DISTINCT
	regexp_replace(nc.tar_path, '^.+[/\\\\]', '') AS tar_name,
	bc.byte_count,
	pg_size_pretty(bc.byte_count) AS byte_count_english
  FROM
  	secedgar.edgar_put_daily nc
  	  JOIN setcore.byte_count bc ON (
	  	bc.blob = nc.blob
	  )
  WHERE
  	nc.blob = $1
;));

my $r = $q->fetchrow_hashref();
unless ($r) {
	print <<END;
 <dt class="err">Blob Not Found</dt>
 <dd class="err">$blob</dd>
</dl>
END
	return 0;
}

print <<END;
 <dt>Tar File Name</dt>
 <dd>$r->{tar_name}</dd>

 <dt>Byte Count English</dt>
 <dd>$r->{byte_count_english}</dd>

 <dt>Byte Count</dt>
 <dd>$r->{byte_count}</dd>

 <dt>NC Tar Blob UDig</dt>
 <dd>
  <a
    title="$r->{tar_name}"
    href="/cgi-bin/daily?out=mime.tar&blob=$blob"
  >$blob</a></dd>
</dl>
END
