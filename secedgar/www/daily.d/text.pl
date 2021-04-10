#
#  Synopsis:
#
require 'jmscott/dbi-pg.pl';

our %QUERY_ARG;

my $db = dbi_pg_connect();
my $q = dbi_pg_select(
	tag =>	'select-daily-zip',
	db =>	$db,
	argv => [],
	sql =>  q(
SELECT
	count(*) AS zip_count
  FROM
  	secedgar.daily_zip dz
	  JOIN setcore.byte_count bc ON (
	  	bc.blob = dz.blob
	  )
;));

print $q->fetchrow_hashref()->{zip_count}, " Zips Downloaded";
