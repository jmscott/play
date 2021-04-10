#
#  Synopsis:
#	Write <div> help page for script daily.
#  Source Path:
#	daily.cgi
#  Source SHA1 Digest:
#	No SHA1 Calculated
#  Note:
#	daily.d/help.pl was generated automatically by cgi2perl5.
#
#	Do not make changes directly to this script.
#

our (%QUERY_ARG);

print <<END;
<div$QUERY_ARG{id_att}$QUERY_ARG{class_att}>
END
print <<'END';
 <h1>Help Page for <code>/cgi-bin/daily</code></h1>
 <div class="overview">
  <h2>Overview</h2>
  <dl>
<dt>Title</dt>
<dd>/cgi-bin/daily</dd>
<dt>Synopsis</dt>
<dd>HTTP CGI Script /cgi-bin/daily</dd>
<dt>Blame</dt>
<dd>jmscott</dd>
  </dl>
 </div>
 <div class="GET">
  <h2><code>GET</code> Request.</h2>
   <div class="out">
    <div class="handlers">
    <h3>Output Scripts in <code>$SERVER_ROOT/lib/daily.d</code></h3>
    <dl>
     <dt>dl</dt>
     <dd>
     </dd>
     <dt>text</dt>
     <dd>
     </dd>
  </dl>
 </div>
</div>
<div class="examples">
 <h3>Examples</h3>
 <dl>
   <dt><a href="/cgi-bin/daily?/cgi-bin/daily?help">/cgi-bin/daily?/cgi-bin/daily?help</a></dt>
   <dd>Generate This Help Page for the CGI Script /cgi-bin/daily</dd>
 </dl>
</div>
 </div>
</div>
END

1;
