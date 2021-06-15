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
	count(distinct blob) AS tar_count
  FROM
  	secedgar.edgar_put_daily dz
;
));

print $q->fetchrow_hashref()->{tar_count};
