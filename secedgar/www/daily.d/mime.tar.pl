#
#  Synopsis:
#	Fetch an SEC EDGAR nc tar file as blob.
#  Usage:
#	/sgi-bin/daily?out=mime.tar?blob=bc160:3f1fca04e46c32da3369df8a1 ...
#
require 'dbi-pg.pl';

our %QUERY_ARG;

my $blob = $QUERY_ARG{blob};
my $q = dbi_pg_select(
		db =>	dbi_pg_connect(),
		tag =>	'select-mime-nc-tar',
		argv =>	[
			$blob
		],
		sql =>	q(
SELECT
	dz.tar_path,
	fm.mime_type,
	bc.byte_count
  FROM
  	secedgar.edgar_put_daily dz
	  JOIN setcore.byte_count bc ON (
	  	bc.blob = dz.blob
	  )
  	  JOIN fffile.file_mime_type fm ON (
	  	fm.blob = dz.blob
	  )
  WHERE
  	dz.blob = $1
;
));

my $row = $q->fetchrow_hashref();
unless ($row) {
	print <<END;
Status: 404
Content-Type: text/html

TAR not found: $blob
END
	return;
}

my $tar__path = $row->{'tar__path'};
$tar__path =~ s@.*/([^/]*)$@$1@;
my $content_length = $row->{'byte_count'};
my $mime_type = $row->{'mime_type'};

print <<END;
Content-Type: $mime_type
Content-Disposition: inline;  filename="$tar__path"
Content-Length: $content_length

END

my $SERVICE = $ENV{BLOBIO_SERVICE};
my $GET_SERVICE = $ENV{BLOBIO_GET_SERVICE} ? 
			$ENV{BLOBIO_GET_SERVICE} :
			$SERVICE
;

my $status = system("blobio get --service $GET_SERVICE --udig $blob");

print STDERR "mime.tar: blobio get $blob failed: exit status=$status\n"
	unless $status == 0
;

1;
