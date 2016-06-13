package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type SQLQueryArg struct {
	name     string
	PGType   string `json:"type"`
	position uint8
}

type SQLQueryArgSet map[string]*SQLQueryArg

type SQLQuery struct {
	name           string
	SourcePath     string `json:"source-path"`
	SQLQueryArgSet `json:"query-arg-set"`
	sql_text       string
	stmt		*sql.Stmt
	qargv		[]*SQLQueryArg
}

var (
	db *sql.DB
	pgsql_command_prefix_re = regexp.MustCompile(`^[ \t]*\\`)
	pgsql_colon_var = regexp.MustCompile(`(?:[^:]|\A):[\w]+`)
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

	saw_PG := false
	log("dumping PG* environment variables")
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "PG") {
			log("	%s", env)
			saw_PG = true
		}
	}
	if !saw_PG {
		log("no PG* enviromment variables")
	}

	db, err = sql.Open(
		"postgres",
		"sslmode=disable",
	)
	if err != nil {
		panic(err)
	}

	log("preparing %d queries", len(qset))
	for n, q := range qset {
		log("	%s", n)
		q.stmt, err = db.Prepare(q.sql_text)
		if err != nil {
			ERROR("sql prepare failed:\n%s", q.sql_text)
			die("%s", err)
		}
	}
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
		q.WARN("add empty section to elimate this warning")
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
		case "smallint":
		case "int":
		case "int2":
		case "int4":
		case "int8":
		default:
			q.die("unknown pgtype: %s", qa.PGType)
		}
	}
	log("    }")
	q.qargv = make([]*SQLQueryArg, len(q.SQLQueryArgSet))
	q.parse_pgsql(in)

	//  build argv table
	for _, qa := range q.SQLQueryArgSet {
		q.qargv[qa.position - 1] = qa
	}
	//  Note: build mapping for http args
}

// Reply to an sql query request

func (q *SQLQuery) handle(w http.ResponseWriter, r *http.Request, cf *Config) {

	url := r.URL

	if r.Method != http.MethodGet {
		herror(
			w,
			http.StatusMethodNotAllowed,
			"unknown method: %s",
			r.Method,
		)
		return
	}
	path := url.Path

	fmt.Fprintf(w, "Path: %s", path)

	us := html.EscapeString(url.String())
	log("%s: %s: %s", r.RemoteAddr, r.Method, us)
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
