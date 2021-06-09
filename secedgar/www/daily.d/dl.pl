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
/*
 *  Find the most recent "not-cooked" tar files per day.
 * 
 *  Note:
 *	Incorrectly assume that the same job can not run at exactly
 *	the same time!  Man, i love hacking json+pg.
 */
WITH recent_tar AS (		--  limited set of most recent jobs
  SELECT
  	regexp_replace(tar_path, '^.+[/\\\\]', '') AS tar_name,
	blob AS blob,
	max(job_time) as max_job_time
  FROM
	secedgar.edgar_put_daily
  GROUP BY
  	tar_name,
	blob
  ORDER BY
  	tar_name DESC
  LIMIT
  	$1
  OFFSET
  	$2
)
  SELECT DISTINCT
	rj.tar_name,
	pg_size_pretty(bc.byte_count) AS byte_count,
	rj.blob
  FROM
  	recent_tar rj
	  JOIN setcore.byte_count bc ON (
	  	bc.blob = rj.blob
	  )
  ORDER BY
  	tar_name DESC
;));

while (my $r = $q->fetchrow_hashref()) {
	print <<END;
 <dt>
  <a href="/cgi-bin/daily?out=mime.tar&blob=$r->{blob}">$r->{tar_name}</a>
 </dt>
 <dd>
   $r->{byte_count},
   <a
     href="/nctar.shtml?blob=$r->{blob}"
     class="detail"
   >Detail</a>
 </dd>
END
}

print <<END;
</dl>
END
