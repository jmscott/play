#
#  Synopsis:
#	Write an html <dl> of daily nc tar files
#  Note:
#	Is each tar file element == comapany?
#
require 'jmscott/dbi-pg.pl';

our %QUERY_ARG;
my $blob = $QUERY_ARG{blob};

my $db = dbi_pg_connect();
my $q = dbi_pg_select(
	tag =>	'select-daily-nc-tar',
	db =>	$db,
	argv => [
			$blob,
		],
	sql =>  q(
SELECT
	count(*) AS element_count
  FROM
  	secedgar.nc_tar_file_element
  WHERE
  	blob = $1
;
));

if (my $r = $q->fetchrow_hashref()) {
	my $plural = 's';

	$plural = '' if $r->{element_count} == 1;
	print encode_html_entities($r->{element_count}), " Submission$plural";
} else {
	print 'blob not found';
}
