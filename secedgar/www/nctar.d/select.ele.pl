#
#  Synopsis:
#	Write an html <dl> of daily nc tar files
#
require 'jmscott/dbi-pg.pl';

our %QUERY_ARG;
my $blob = $QUERY_ARG{blob};
my $size = $QUERY_ARG{size};

print <<END;
<select
  size="$size"
  $QUERY_ARG{id_att}
  $QUERY_ARG{class_att}
> 
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
	file_path,
	file_size,
	pg_size_pretty(file_size) AS file_size_english
  FROM
  	secedgar.nc_tar_file_element
  WHERE
  	blob = $1
  ORDER BY
  	file_path ASC
;
));

while (my $r = $q->fetchrow_hashref()) {
	my $file_path = encode_html_entities($r->{file_path});
	my $file_size_english = encode_html_entities($r->{file_size_english});

	print <<END;
 <option value="$file_path">$file_path: $file_size_english</option>
END
}

print <<END;
</select>
END
