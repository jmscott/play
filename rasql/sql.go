package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"io"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SQLQueryArg struct {
	name     string
	PGType   string `json:"type"`
	pgtype_re *regexp.Regexp
	gokind   reflect.Kind
	position uint8
	http_arg *HTTPQueryArg
}

type SQLQueryArgSet map[string]*SQLQueryArg

type SQLQuery struct {
	name           string
	SourcePath     string `json:"source-path"`
	SQLQueryArgSet `json:"query-arg-set"`
	sql_text       string
	stmt           *sql.Stmt
	qargv          []*SQLQueryArg
}

var (
	db                      *sql.DB
	pgsql_command_prefix_re = regexp.MustCompile(`^[ \t]*\\`)
	pgsql_colon_var         = regexp.MustCompile(`(?:[^:]|\A):[\w]+`)

	pgtype2re = map[string]*regexp.Regexp{
		//  Note: what about null in the string?
		//        1000 is a limit imposed by package regexp
		"text": regexp.MustCompile(`^.{0,1000}$`),

		//  0 - 65535
		"uint16": regexp.MustCompile(
			`^(6553[0-5]|655[0-2][0-9]|65[0-4][0-9]{2}|6[0-4][0-9]{3}|[1-5][0-9]{4}|[1-9][0-9]{0,3}|0)$`),

		//  0 - 4294967295
		"uint32": regexp.MustCompile(
			`^(429496729[0-5]|42949672[0-8][0-9]|4294967[01][0-9]{2}|429496[0-6][0-9]{3}|42949[0-5][0-9]{4}|4294[0-8][0-9]{5}|429[0-3][0-9]{6}|42[0-8][0-9]{7}|4[01][0-9]{8}|[1-3][0-9]{9}|[1-9][0-9]{0,8}|0)$`),

		//  0 - 9223372036854775807
		"ubigint": regexp.MustCompile(
			`^(922337203685477580[0-7]|9223372036854775[0-7][0-9]{2}|922337203685477[0-4][0-9]{3}|92233720368547[0-6][0-9]{4}|9223372036854[0-6][0-9]{5}|922337203685[0-3][0-9]{6}|92233720368[0-4][0-9]{7}|9223372036[0-7][0-9]{8}|922337203[0-5][0-9]{9}|92233720[0-2][0-9]{10}|922337[01][0-9]{12}|92233[0-6][0-9]{13}|9223[0-2][0-9]{14}|922[0-2][0-9]{15}|92[01][0-9]{16}|9[01][0-9]{17}|[1-8][0-9]{18}|[1-9][0-9]{0,17}|0)$`),
	}
)

type SQLQuerySet map[string]*SQLQuery

func (queries SQLQuerySet) load() {

	log("%d sql query files in config {", len(queries))
	for n, q := range queries {
		q.name = n
		log("  %s: {", q.name)
		log("    source-path: %s", q.SourcePath)
		log("  }")
	}
	log("}")

	//  load sql queries from external files

	log("loading sql queries from %d files", len(queries))
	for n := range queries {
		q := queries[n]
		q.load()
	}
}

func (qset SQLQuerySet) open() {

	var err error

	db, err = sql.Open(
		"postgres",
		"sslmode=disable",
	)
	if err != nil {
		panic(err)
	}

	log("preparing %d queries in sql database", len(qset))
	for n, q := range qset {
		log("	%s", n)
		q.stmt, err = db.Prepare(q.sql_text)
		if err != nil {
			ERROR("sql prepare failed:\n%s", q.sql_text)
			die("%s", err)
		}
	}
	log("all queries prepared")
}

func (q *SQLQuery) die(format string, args ...interface{}) {

	die("sql query: %s: %s", q.SourcePath, fmt.Sprintf(format, args...))
}

func (q *SQLQuery) WARN(format string, args ...interface{}) {

	log("WARN: sql query: %s: %s", q.SourcePath,
		fmt.Sprintf(format, args...))
}

func (q *SQLQuery) load() {

	log("  %s", q.SourcePath)

	sqlf, err := os.Open(q.SourcePath)
	if err != nil {
		q.die("%s", err)
	}
	defer sqlf.Close()

	in := bufio.NewReader(sqlf)

	//  first line of sql file must be "/*"

	line, err := in.ReadString('\n')
	if err != nil {
		q.die(err.Error())
	}
	if line != "/*\n" {
		q.die("first line is not \"/*\"")
	}

	//  load the preamble in the sql file

	var pre CcommentPreamble = make(CcommentPreamble)
	var line_count int

	line_count, err = pre.parse(in)
	line_count++
	if err != nil {
		q.die("preamble: %s near line %d", err, line_count)
	}
	if len(pre) == 0 {
		q.WARN("preamble is empty")
		return
	}

	//  decode the json description of the command line arguments

	cla := pre["Command Line Arguments"]
	if cla == "" {
		q.WARN("no \"Command Line Arguments\" section")
		q.WARN("add empty {} section to eliminate this warning")
	}
	dec := json.NewDecoder(strings.NewReader(cla))
	err = dec.Decode(&q.SQLQueryArgSet)
	if err != nil && err != io.EOF {
		q.die("failed to decode json in command line arguments")
	}

	if len(q.SQLQueryArgSet) > 255 {
		q.die("> 255 sql query arguments")
	}

	//  verify pg sql types

	log("    %d arguments: {", len(q.SQLQueryArgSet))
	for n, qa := range q.SQLQueryArgSet {
		qa.name = n
		log("      %s:{pgtype:%s}", qa.name, qa.PGType)

		// verify PostgreSQL types

		switch qa.PGType {
		case "text":
			qa.gokind = reflect.String
		case "uint16":
			qa.gokind = reflect.Uint16
		case "uint32":
			qa.gokind = reflect.Uint32
		case "ubigint":
			qa.gokind = reflect.Uint64
		default:
			q.die("unknown pgtype: %s", qa.PGType)
		}
		qa.pgtype_re = pgtype2re[qa.PGType]
	}
	log("    }")
	q.qargv = make([]*SQLQueryArg, len(q.SQLQueryArgSet))
	q.parse_pgsql(in)

	//  build query argument argument vector

	for _, qa := range q.SQLQueryArgSet {
		qa.position--
		q.qargv[qa.position] = qa
	}
}

//  Reply to an sql query request from a url

func (q *SQLQuery) handle(w http.ResponseWriter, r *http.Request, cf *Config) {

	if r.Method != http.MethodGet {
		herror(
			w,
			http.StatusMethodNotAllowed,
			"unknown method: %s",
			r.Method,
		)
		return
	}
	url := r.URL

	//  build the argv []interface{} for the sql query to execute

	argv := make([]interface{}, len(q.SQLQueryArgSet))
	req_qa := url.Query()
	for _, qa := range q.qargv {

		bada := func(format string, args ...interface{}) {
			msg := fmt.Sprintf(
				"query arg: %s: %s",
				qa.name,
				fmt.Sprintf(format, args...),
			)
			http.Error(w, msg, http.StatusBadRequest)
			ERROR("%s", msg)
		}

		var an string

		//  does the sql query arg have an http alias?

		ha := qa.http_arg
		if ha == nil {
			ha = cf.HTTPQueryArgSet[qa.name]
			an = qa.name
		} else {
			an = ha.name
		}

		//  verify http query arg exists and matches regular expression

		rqa := req_qa[an]

		//  no query arg on url so try default for http args

		if rqa == nil {
			if ha == nil || ha.Default == "" {
				bada("missing url arg: %s", an)
				return
			}
			rqa = make([]string, 1)
			rqa[0] = ha.Default
		}

		//  sql query arguments can only be given once

		if len(rqa) != 1 {
			bada("given more than once")
			return
		}
		ra := rqa[0]

		//  verify that the url query argment value matches proper
		//  regular expression

		var re *regexp.Regexp
		re_what := ""
		if ha == nil || ha.matches_re == nil {
			re = qa.pgtype_re
			re_what = "sql"
		} else {
			re = ha.matches_re
			re_what = "http"
		}
		if !re.MatchString(ra) {
			bada("value does not match %s regexp: %s: %s",
							re_what, ra, re)
			return
		}

		//  parse http query arg into sql arg for prepared query

		switch qa.gokind {
		case reflect.String:
			argv[qa.position] = ra
		case reflect.Uint16:
			i64, err := strconv.ParseInt(ra, 10, 16)
			if err != nil {
				bada("can not parse int16: %s", ra)
				return
			}
			argv[qa.position] = int16(i64)
		case reflect.Uint32:
			i64, err := strconv.ParseInt(ra, 10, 32)
			if err != nil {
				bada("can not parse int32: %s", ra)
				return
			}
			argv[qa.position] = int32(i64)
		case reflect.Uint64:
			i64, err := strconv.ParseInt(ra, 10, 32)
			if err != nil {
				bada("can not parse int32: %s", ra)
				return
			}
			argv[qa.position] = i64
		default:
			panic("impossible gokind")
		}
	}

	//  run the sql query
	start_time := time.Now()
	rows, err := q.stmt.Query(argv...)
	if err != nil {
		panic(err)
	}
	duration := time.Since(start_time).Seconds()
	defer rows.Close()

	//  grumble about slow queries.

	if duration > cf.WarnSlowSQLQueryDuration {
		WARN("slow query: %s: %.9fs: %s", q.name, duration, url)
	}

	cols, err := rows.Columns()
	if err != nil {
		panic(err)
	}

	put := func(format string, args ...interface{}) {
		fmt.Fprintf(w, format, args...)
	}

	//  make the row string vector

	rowv := make([]interface{}, len(cols))
	for i := range rowv {
		rowv[i] = new(sql.NullString)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	//  write the reply with the query duration

	put(`{
  "sql-query-reply": {
    "duration": %.9f,

    "columns": [`,
		duration,
	)

	//  write the columns

	for i, c := range cols {
		put(`%q`, c)
		if i+1 < len(cols) {
			put(", ")
		}
	}

	put(`],

    "rows": [
`,
	)

	//  write the rows

	count := uint64(0)
	for rows.Next() {

		if count > 0 {
			put(",")
		}
		count++

		err = rows.Scan(rowv...)
		if err != nil {
			panic(err)
		}
		put("      [")
		for i, si := range rowv {
			if i > 0 {
				put(", ")
			}
			s := si.(*sql.NullString)
			if s.Valid {
				put("%q", s.String)
			} else {
				put("null")
			}
		}
		put("]\n")
	}
	put("    ]\n  }\n}\n")
}

//  parse a typical postgres sql file into a string suitable for Prepare()
//  in particular, :name variables are extracted and \<directives> are stripped.

func (q *SQLQuery) parse_pgsql(in *bufio.Reader) {

	var sql_text bytes.Buffer

	position := uint8(0)

	replace_re := make(map[string]*regexp.Regexp)

	for {
		line, err := in.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			die("parse_pgsql: io error: %s", err)
		}

		//  skip a psql command prefix "\d ..."

		if pgsql_command_prefix_re.MatchString(line) {
			continue
		}

		//  find all command line :var references

		vars := pgsql_colon_var.FindAllString(line, -1)
		if len(vars) == 0 {
			sql_text.WriteString(line)
			continue
		}

		//  swap the parameter name with $<pos>

		for _, v := range vars {
			if v[0:1] != ":" {
				v = v[1:]
			}
			qa := q.SQLQueryArgSet[v[1:]]
			if qa == nil {
				q.WARN("pgsql variable not in preamble: %s", v)
				continue
			}
			if qa.position == 0 {
				if position == 255 {
					q.die("pgsql variables: count > 254")
				}
				position++
				qa.position = position

				//  make a replacement re specific to
				//  this argument
				//
				//  Note: must test that default values
				//        matches the pattern?
				//	  also, matches MUST exist

				replace_re[v] = regexp.MustCompile(
					fmt.Sprintf(`([^:]|\A)(%s)`, v))
			}
			//  Note: why no ReplaceString()!!!

			line = replace_re[v].ReplaceAllString(
				line,
				fmt.Sprintf(`$1$$%d`, qa.position),
			)
		}
		sql_text.WriteString(line)
	}
	q.sql_text = sql_text.String()
}
