package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Config struct {
	source_path string

	Synopsis        string `json:"synopsis"`
	HTTPListen      string `json:"http-listen"`
	RESTPathPrefix  string `json:"rest-path-prefix"`
	SQLQuerySet     `json:"sql-query-set"`
	HTTPQueryArgSet `json:"http-query-arg-set"`

	//  Note:  also want to log slow http requests!
	//         consider moving into WARN section.

	WarnSlowSQLQueryDuration float64 `json:"warn-slow-sql-query-duration"`
}

func (cf *Config) load(path string) {

	log("loading config file: %s", path)

	cf.source_path = path

	//  slurp config file into string

	b, err := ioutil.ReadFile(cf.source_path)
	if err != nil {
		die("config read failed: %s", err)
	}

	//  decode json in config file

	dec := json.NewDecoder(strings.NewReader(string(b)))
	err = dec.Decode(&cf)
	if err != nil && err != io.EOF {
		die("config json decoding failed: %s", err)
	}

	if cf.HTTPListen == "" {
		die("config: http-listen not defined or empty")
	}
	log("http listen: %s", cf.HTTPListen)

	if cf.RESTPathPrefix == "" {
		cf.RESTPathPrefix = "/"
	}
	log("rest path prefix: %s", cf.RESTPathPrefix)
	log("warn slow sql query duration: %0.9fs", cf.WarnSlowSQLQueryDuration)

	cf.SQLQuerySet.load()
	cf.HTTPQueryArgSet.load()

	//  wire up sql aliases for the http query arguments

	log("map http/sql query args ...")
	for _, ha := range cf.HTTPQueryArgSet {
		a := ha.SQLAlias
		if a == "" {
			if ha.Matches == "" {
				die("query arg: missing \"matches\" regexp: %s",
					ha.name)
			}
			continue
		}

		//  point sql arguments to current http query argument

		found := false
		for _, sq := range cf.SQLQuerySet {
			for _, sqa := range sq.SQLQueryArgSet {
				if sqa.name != a {
					continue
				}
				log("  %s -> %s", sqa.name, ha.name)
				found = true
				sqa.http_arg = ha
			}
		}
		if !found {
			die("sql alias '%s' has no query arg in http arg '%s'",
					a, ha.name)
		}
	}

	log("%s: loaded", cf.source_path)
}

func (cf *Config) new_sql_handler(sqlq *SQLQuery) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		sqlq.handle(w, r, cf)
	}
}
