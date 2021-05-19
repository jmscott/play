#
#  Synopsis:
#
require 'jmscott/dbi-pg.pl';

our %QUERY_ARG;

my $db = dbi_pg_connect();
my $q = dbi_pg_select(
	tag =>	'select-text-daily-nc-tar',
	db =>	$db,
	argv => [],
	sql =>  q(
SELECT
	count(*) AS tar_count
  FROM
  	secedgar.daily_nc_tar dz
	  JOIN setcore.byte_count bc ON (
	  	bc.blob = dz.blob
	  )
;));

print $q->fetchrow_hashref()->{tar_count}, " NC TARs Downloaded";
