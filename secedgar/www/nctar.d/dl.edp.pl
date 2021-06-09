#
#  Synopsis:
#	Write an html <dl> ... <pre> of json job blobs for a tar blob.
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
	sql =>  q(
SELECT
	j.blob,
	jsonb_pretty(j.doc) AS doc,
	to_char(nc.job_time, 'YYYY/mm/dd HH24:mi:ss') AS job_time
  FROM
  	secedgar.daily_nc_tar nc,
	jsonorg.jsonb_255 j
  WHERE
  	nc.blob = $1
	AND
	j.blob = nc.doc_blob
  ORDER BY
  	nc.job_time DESC
;));

while (my $r = $q->fetchrow_hashref()) {
	my $doc = encode_html_entities($r->{doc});
	print <<END;
 <dt>$r->{job_time} - <span class="udig">$r->{blob}</span></dt>
 <dd>
  <pre class="json">$doc</pre>
 </dd>
END
}
print "</dl>\n";
