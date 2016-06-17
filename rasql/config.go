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

	//  add default http query args for sql args
	for _, sq := range cf.SQLQuerySet {
		for n, sqa := range sq.SQLQueryArgSet {
			if cf.HTTPQueryArgSet == nil {
				cf.HTTPQueryArgSet = make(HTTPQueryArgSet)
			}
			ha := cf.HTTPQueryArgSet[n]
			if ha != nil {
				continue
			}
			re := pgtype2re[sqa.PGType]
			cf.HTTPQueryArgSet[n] = &HTTPQueryArg{
				name:       sqa.name,
				Matches:    re.String(),
				matches_re: re,
			}
			log("added http query arg: %s(%s)", n, re.String())
		}
	}

	log("%s: loaded", cf.source_path)
}

func (cf *Config) new_sql_handler(query_name string) http.HandlerFunc {

	sqlq := cf.SQLQuerySet[query_name]
	if sqlq == nil {
		panic("no sql query: " + query_name)
	}

	return func(w http.ResponseWriter, r *http.Request) {

		sqlq.handle(w, r, cf)
	}
}
