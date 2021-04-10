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
   ></putter>
   <putter
     name="text"
     content-type="text/html"
   ></putter>
  </out>
 </GET>
</cgi>
