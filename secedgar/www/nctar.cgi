<?xml version="1.0" encoding="UTF-8"?>
<cgi name="nctar">
 <title>/cgi-bin/nctar</title>
 <synopsis>HTTP CGI Script /cgi-bin/nctar</synopsis>
 <subversion id="$Id$" />
 <GET>
  <examples>
   <example
   	query="/cgi-bin/nctar?help"
   >
    Generate This Help Page for the CGI Script /cgi-bin/nctar
   </example>
  </examples>

  <out>
   <putter
     name="dl"
     content-type="text/html"
   >
    <query-args>
     <arg
       name="blob"
       required="yes"
       perl5_re="[a-z][a-z0-9]{0,7}:[[:graph:]]{32,128}"
     ></arg>
    </query-args>
   </putter>

   <putter
     name="dl.edp"
     content-type="text/html"
   >
    <query-args>
     <arg
       name="blob"
       required="yes"
       perl5_re="[a-z][a-z0-9]{0,7}:[[:graph:]]{32,128}"
     ></arg>
    </query-args>
   </putter>

   <putter
     name="select.ele"
     content-type="text/html"
   >
    <query-args>
     <arg
       name="blob"
       required="yes"
       perl5_re="[a-z][a-z0-9]{0,7}:[[:graph:]]{32,128}"
     ></arg>
     <arg
       name="size"
       perl5_re="[0-9]{1,3}"
       default="20"
     ></arg>
    </query-args>
   </putter>

  </out>
 </GET>
</cgi>
