#
#  Synopsis:
#	Write html <span> of for navigating daily zip search results
#
use utf8;

#  stop apache log message 'Wide character in print at ...' from arrow chars
binmode(STDOUT, ":utf8");

require 'jmscott/dbi-pg.pl';

our %QUERY_ARG;

my $off = $QUERY_ARG{off};
my $lim = $QUERY_ARG{lim};

my $zip_count = dbi_pg_select(
		db =>	dbi_pg_connect(),
		tag =>	'select-daily-zip-count',
		argv =>	[],
		sql => q(
SELECT
	count(*) AS zip_count
  FROM
  	secedgar.daily_zip
)
)->fetchrow_hashref()->{zip_count};

print <<END;
<span
  $QUERY_ARG{id_att}
  $QUERY_ARG{class_att}
>
END

if ($zip_count == 0) {
	print <<END;
No daily zip files.</span>
END
	return 1;
}

if ($zip_count <= $lim) {
	my $plural = 's';
	$plural = '' if $zip_count == 1;

	print <<END;
$zip_count zip$plural fetched
END
	return 1;
}

my $arrow_off;
if ($off >= $lim) {
	$arrow_off = $off - $lim;
	print <<END;
<a href=
"/daily-zip.shtml?off=$arrow_off&amp;lim=$lim">◀</a>
END
}

my $zip_lower = $off + 1;
my $zip_up = $zip_lower + $lim - 1;
$zip_up = $zip_count if $zip_up > $zip_count;

#  Note:
#	Rewrite with commas, english style.
#	Need a library here.
#
1 while $zip_lower =~ s/^(\d+)(\d{3})/$1,$2/;
1 while $zip_up =~ s/^(\d+)(\d{3})/$1,$2/;

print <<END;
$zip_lower to $zip_up 
END

$arrow_off = $off + $lim;
print <<END if $arrow_off < $zip_count;
<a href="/daily-zip.shtml?off=$arrow_off&amp;lim=$lim">▶</a>
END

print <<END;
of $zip_count zips fetched
</span>
END

1;
