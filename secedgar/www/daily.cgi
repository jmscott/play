<?xml version="1.0" encoding="UTF-8"?>
<cgi name="daily">
 <title>/cgi-bin/daily</title>
 <synopsis>HTTP CGI Script /cgi-bin/daily</synopsis>
 <subversion id="$Id$" />
 <blame>jmscott</blame>
 <GET>
  <examples>
   <example
   	query="/cgi-bin/daily?help"
   >
    Generate This Help Page for the CGI Script /cgi-bin/daily
   </example>
  </examples>

  <out>
   <putter
     name="dl"
     content-type="text/html"
   >
    <query-args>
     <arg
       name="lim"
       required="no"
       default="10"
       perl5_re="10|100"
     ></arg>
     <arg
       name="off"
       default="10"
       perl5_re="[0-9]{1,4}"
     ></arg>
    </query-args>
   </putter>

   <putter
     name="span.nav"
     content-type="text/html"
   >
    <query-args>
     <arg
       name="lim"
       required="no"
       default="10"
       perl5_re="10|100"
     ></arg>
     <arg
       name="off"
       default="0"
       perl5_re="[0-9]{1,4}"
     ></arg>
    </query-args>
   </putter>

   <putter
     name="text"
     content-type="text/html"
   ></putter>

   <putter name="mime.zip">
    <query-args>
     <arg
       name="blob"
       required="yes"
       perl5_re="[a-z][a-z0-9]{0,7}:[[:graph:]]{32,128}"
     ></arg>
    </query-args>
   </putter>

  </out>
 </GET>
</cgi>
