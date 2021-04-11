#
#  Synopsis:
#	Fetch a zip file as blob.
#  Note:
#	The pdfbox-full-text-search-click needs to move to a separate
#	cgi-bin widget callable from web page!  Simply displaying a pdf
#	blob triggers a json blob for a web search!
#
#	WTF?  No validation is done on the blob.  In fact, the blob is
#	assumed to exist and be correct.  Terrible.
#
require 'dbi-pg.pl';

our %QUERY_ARG;

my $blob = $QUERY_ARG{blob};
my $q = dbi_pg_select(
		db =>	dbi_pg_connect(),
		tag =>	'select-mime-zip',
		argv =>	[
			$blob
		],
		sql =>	q(
SELECT
	dz.zip_path,
	fm.mime_type,
	bc.byte_count
  FROM
  	secedgar.daily_zip dz
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

ZIP not found: $blob
END
	return;
}

my $zip_path = $row->{'zip_path'};
$zip_path =~ s@.*/([^/]*)$@$1@;
my $content_length = $row->{'byte_count'};
my $mime_type = $row->{'mime_type'};

print <<END;
Content-Type: $mime_type
Content-Disposition: inline;  filename="$zip_path"
Content-Length: $content_length

END

my $SERVICE = $ENV{BLOBIO_SERVICE};
my $GET_SERVICE = $ENV{BLOBIO_GET_SERVICE} ? 
			$ENV{BLOBIO_GET_SERVICE} :
			$SERVICE
;

my $status = system("blobio get --service $GET_SERVICE --udig $blob");

print STDERR "mime.zip.d/blob: blobio get $blob failed: exit status=$status\n"
	unless $status == 0
;

1;
