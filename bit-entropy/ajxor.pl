#!/usr/bin/env perl
#
#  Synopsis:
#	For N in bits, XOr adjacent bits to produce vector <N-1 bit, hamming>
#  Usage:
#	IN='01010101  11110000  00100110  11110000	15'
#	echo "$IN" | ajxor.pl | ajxor.pl | ajxor.pl | ajxor.pl
#  See:
#	https://github.com/jmscott/play/tree/master/bit-entropy
#

STDOUT->autoflush(1);

my ($in, $bit_count);
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
