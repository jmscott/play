package main

import (
	"encoding/json"
	"fmt"
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
		for _, q := range cf.SQLQuerySet {
			for _, qa := range q.SQLQueryArgSet {
				if qa.path != a {
					continue
				}
				log("  %s -> %s", qa.path, ha.name)
				found = true
				qa.http_arg = ha
			}
		}
		if !found {
			die(`http arg "%s": no sql variable for alias "%s"`,
				ha.name, a)
		}
	}

	log("%s: loaded", cf.source_path)
}

func (cf *Config) new_handler_query_json(sqlq *SQLQuery) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		sqlq.handle_query_json(w, r, cf)
	}
}

func (cf *Config) new_handler_query_tsv(sqlq *SQLQuery) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		sqlq.handle_query_tsv(w, r, cf)
	}
}

func (cf *Config) handle_query_index_json(
	w http.ResponseWriter,
	r *http.Request,
) {
	putf := func(format string, args ...interface{}) {
		fmt.Fprintf(w, format, args...)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	//  write bytes string to client

	putb := func(b []byte) {
		_, err := w.Write(b)
		if err != nil {
			panic(err)
		}
	}

	puts := func(s string) {
		putb([]byte(s))
	}

	// put json string to client

	putjs := func(s string) {
		b, err := json.Marshal(s)
		if err != nil {
			panic(err)
		}
		putb(b)
	}

	//  write a json array to the client

	puta := func(a []string) {
		b, err := json.Marshal(a)
		if err != nil {
			panic(err)
		}
		putb(b)
	}

	puts(`[
    "duration,colums,rows",
    0.0,
    `,
	)

	//  write the columns

	var columns = [...]string{
		"path",
		"synopsis",
		"description",
	}

	puta(columns[:])
	puts(",\n\n    [\n")

	count := uint64(0)
	for n, q := range cf.SQLQuerySet {

		if count > 0 {
			putf(",\n")
		}
		count++

		puts("[")
		putjs(n)
		puts(",")
		putjs(q.synopsis)
		puts(",")
		putjs(q.description)

		putf("]")
	}
	putf("\n    ]\n]\n")
}
