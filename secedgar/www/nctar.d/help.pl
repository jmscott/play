#
#  Synopsis:
#	Write <div> help page for script nctar.
#  Source Path:
#	nctar.cgi
#  Source SHA1 Digest:
#	No SHA1 Calculated
#  Note:
#	nctar.d/help.pl was generated automatically by cgi2perl5.
#
#	Do not make changes directly to this script.
#

our (%QUERY_ARG);

print <<END;
<div$QUERY_ARG{id_att}$QUERY_ARG{class_att}>
END
print <<'END';
 <h1>Help Page for <code>/cgi-bin/nctar</code></h1>
 <div class="overview">
  <h2>Overview</h2>
  <dl>
<dt>Title</dt>
<dd>/cgi-bin/nctar</dd>
<dt>Synopsis</dt>
<dd>HTTP CGI Script /cgi-bin/nctar</dd>
  </dl>
 </div>
 <div class="GET">
  <h2><code>GET</code> Request.</h2>
   <div class="out">
    <div class="handlers">
    <h3>Output Scripts in <code>$SERVER_ROOT/lib/nctar.d</code></h3>
    <dl>
     <dt>dl</dt>
     <dd>
<div class="query-args">
 <h4>Query Args</h4>
 <dl>
  <dt>blob</dt>
  <dd>
   <ul>
    <li><code>perl5_re:</code> [a-z][a-z0-9]{0,7}:[[:graph:]]{32,128}</li>
    <li><code>required:</code> yes</li>
   </ul>
  </dd>
  </dl>
</div>
     </dd>
     <dt>dl.edp</dt>
     <dd>
<div class="query-args">
 <h4>Query Args</h4>
 <dl>
  <dt>blob</dt>
  <dd>
   <ul>
    <li><code>perl5_re:</code> [a-z][a-z0-9]{0,7}:[[:graph:]]{32,128}</li>
    <li><code>required:</code> yes</li>
   </ul>
  </dd>
  </dl>
</div>
     </dd>
  </dl>
 </div>
</div>
<div class="examples">
 <h3>Examples</h3>
 <dl>
   <dt><a href="/cgi-bin/nctar?/cgi-bin/nctar?help">/cgi-bin/nctar?/cgi-bin/nctar?help</a></dt>
   <dd>Generate This Help Page for the CGI Script /cgi-bin/nctar</dd>
 </dl>
</div>
 </div>
</div>
END

1;
