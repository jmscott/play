#
#  Synopsis:
#	Write html <span> of for navigating daily tar search results
#
use utf8;

#  stop apache log message 'Wide character in print at ...' from arrow chars
binmode(STDOUT, ":utf8");

require 'jmscott/dbi-pg.pl';

our %QUERY_ARG;

my $off = $QUERY_ARG{off};
my $lim = $QUERY_ARG{lim};

my $tar_count = dbi_pg_select(
		db =>	dbi_pg_connect(),
		tag =>	'select-daily-tar-count',
		argv =>	[],
		sql => q(
SELECT
	count(*) AS tar_count
  FROM
  	secedgar.edgar_put_daily
)
)->fetchrow_hashref()->{tar_count};

print <<END;
<span
  $QUERY_ARG{id_att}
  $QUERY_ARG{class_att}
>
END

if ($tar_count == 0) {
	print <<END;
No daily nc tar files.</span>
END
	return 1;
}

if ($tar_count <= $lim) {
	my $plural = 's';
	$plural = '' if $tar_count == 1;

	print <<END;
$tar_count tar$plural fetched
END
	return 1;
}

my $arrow_off;
if ($off >= $lim) {
	$arrow_off = $off - $lim;
	print <<END;
<a href=
"/daily-tar.shtml?off=$arrow_off&amp;lim=$lim">◀</a>
END
}

my $tar_lower = $off + 1;
my $tar_up = $tar_lower + $lim - 1;
$tar_up = $tar_count if $tar_up > $tar_count;

#  Note:
#	Rewrite with commas, english style.
#	Need a library here.
#
1 while $tar_lower =~ s/^(\d+)(\d{3})/$1,$2/;
1 while $tar_up =~ s/^(\d+)(\d{3})/$1,$2/;

print <<END;
$tar_lower to $tar_up 
END

$arrow_off = $off + $lim;
print <<END if $arrow_off < $tar_count;
<a href="/daily-tar.shtml?off=$arrow_off&amp;lim=$lim">▶</a>
END

print <<END;
of $tar_count tars fetched
</span>
END

1;
