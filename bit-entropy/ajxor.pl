#!/usr/bin/env perl
#
#  Synopsis:
#	Derive a population vector of adjacent xors of a bit sequence.
#  Usage:
#	echo '01010101  11110000  00100110  11110000' | ajxor.pl
#  See:
#	https://github.com/jmscott/play/tree/master/bit-entropy
#

my $in;
while (<STDIN>) {
	s/^([^\t]*)\t.*$/$1/;	#  only first field examined
	s/[^01]*//g;		#  remove non 0 or 1 chars
	$in .= $_;
}

sub XOR
{
	my $p = shift;
	my $q = shift;

	return '0' if $p eq $q;
	$bit_count++;
	return '1';
}

my $in_length = length($in) - 1;
for (my $i = 0;  $i < $in_length;  $i++) {
	print ' ' if ($i > 0 && ($i % 8 == 0));
	print XOR(substr($in, $i, 1), substr($in, $i + 1, 1));
}
print "	$bit_count\n";
