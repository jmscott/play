package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	_ "github.com/lib/pq"
)

type SQLQueryArg struct {
	//  Note: should "path" be "name"?
	path      string
	pg_type   string
	pgtype_re *regexp.Regexp
	gokind    reflect.Kind
	position  uint8
	http_arg  *HTTPQueryArg
}

type SQLQueryArgSet map[string]*SQLQueryArg

type SQLQuery struct {
	name           string
	synopsis       string
	description    string
	SourcePath     string `json:"source-path"`
	SQLQueryArgSet `json:"query-arg-set"`
	sql_text       string
	stmt           *sql.Stmt
	argv           []*SQLQueryArg
}

var (
	db                      *sql.DB
	tab                     []byte
	newline                 []byte
	pgsql_command_prefix_re = regexp.MustCompile(`^[ \t]*\\`)
	pgsql_colon_var         = regexp.MustCompile(`(?:[^:]|\A):[\w]+`)
	trim_re                 = regexp.MustCompile(`^[ \t\n]+|[ \t\n]+$`)
	psql_clv_re             = regexp.MustCompile(
					`^\s*(\w{1,63})\s+(\w{1,63})\s*$`)

	//  map pgtypes to regular expressions that matches domain

	pgtype2re = map[string]*regexp.Regexp{
		//  Note: what about null in the string?
		//        1000 is a limit imposed by package regexp
		"text": regexp.MustCompile(`^.{0,1000}$`),

		//  0 - 65535
		"uint16": regexp.MustCompile(
			`^(?:6553[0-5]|655[0-2][0-9]|65[0-4][0-9]{2}|6[0-4][0-9]{3}|[1-5][0-9]{4}|[1-9][0-9]{0,3}|0)$`),

		//  0 - 4294967295
		"uint32": regexp.MustCompile(
			`^(?:429496729[0-5]|42949672[0-8][0-9]|4294967[01][0-9]{2}|429496[0-6][0-9]{3}|42949[0-5][0-9]{4}|4294[0-8][0-9]{5}|429[0-3][0-9]{6}|42[0-8][0-9]{7}|4[01][0-9]{8}|[1-3][0-9]{9}|[1-9][0-9]{0,8}|0)$`),

		//  0 - 9223372036854775807
		"ubigint": regexp.MustCompile(
			`^(?:922337203685477580[0-7]|9223372036854775[0-7][0-9]{2}|922337203685477[0-4][0-9]{3}|92233720368547[0-6][0-9]{4}|9223372036854[0-6][0-9]{5}|922337203685[0-3][0-9]{6}|92233720368[0-4][0-9]{7}|9223372036[0-7][0-9]{8}|922337203[0-5][0-9]{9}|92233720[0-2][0-9]{10}|922337[01][0-9]{12}|92233[0-6][0-9]{13}|9223[0-2][0-9]{14}|922[0-2][0-9]{15}|92[01][0-9]{16}|9[01][0-9]{17}|[1-8][0-9]{18}|[1-9][0-9]{0,17}|0)$`),
	}
)

func init() {
	tab = make([]byte, 1)
	tab[0] = 0x09
	newline = make([]byte, 1)
	newline[0] = 0x0a
}

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

	q.synopsis = trim_re.ReplaceAllLiteralString(pre["Synopsis"], "")
	q.description = trim_re.ReplaceAllLiteralString(pre["Description"], "")

	//  parse the declaration of the command line variables in the section
	//
	//  Command Line Variables:
	//
	//	name1	pgtype
	//	name2	pgtype

	clv, exists := pre["Command Line Variables"]
	if !exists {
		q.WARN("no \"Command Line Variables\" section")
		q.WARN("add empty section to eliminate this warning")
	}

	q.SQLQueryArgSet = make(map[string]*SQLQueryArg, 0)
	vars := bufio.NewReader(strings.NewReader(clv))
	for {
		line, err := vars.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			q.die("Command Line Variables: %s", err)
		}

		//  extract the variable declaration that matches
		//
		//	name	pgtype

		matches := psql_clv_re.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}

		q.SQLQueryArgSet[matches[1]] = &SQLQueryArg{
			path: matches[1],
			pg_type: matches[2],
		}
	}

	if len(q.SQLQueryArgSet) > 255 {
		q.die("> 255 sql query arguments")
	}

	//  verify pg sql types

	log("    %d arguments:", len(q.SQLQueryArgSet))
	for n, qa := range q.SQLQueryArgSet {
		qa.path = n
		log("      %s:{pgtype:%s}",qa.path, qa.pg_type)

		// verify PostgreSQL types
		// Note: replace with table lookup

		switch qa.pg_type {
		case "text":
			qa.gokind = reflect.String
		case "uint16":
			qa.gokind = reflect.Uint16
		case "uint32":
			qa.gokind = reflect.Uint32
		case "ubigint":
			qa.gokind = reflect.Uint64
		default:
			q.die("unknown pgtype: %s", qa.pg_type)
		}
		qa.pgtype_re = pgtype2re[qa.pg_type]
	}

	q.argv = make([]*SQLQueryArg, len(q.SQLQueryArgSet))
	q.parse_pgsql(in)

	//  build query argument argument vector

	for _, qa := range q.SQLQueryArgSet {
		if qa.position == 0 {
			q.WARN("unused sql command line variable: %s", qa.path)
			q.WARN("remove %s declaration to eliminate warning")
			q.argv = q.argv[0:len(q.argv) - 1]
		} else {
			qa.position--
			q.argv[qa.position] = qa
		}
	}
}

//  run an sql query for a get request

func (q *SQLQuery) db_query(
	w http.ResponseWriter,
	r *http.Request,
	cf *Config,
) (
	duration float64,
	columns []string,
	rows *sql.Rows,
	rowv []interface{},
) {
	var err error

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

	argv := make([]interface{}, len(q.argv))
	req_qa := url.Query()
	for _, qa := range q.argv {

		bada := func(format string, args ...interface{}) {
			msg := fmt.Sprintf(
				"query arg: %s: %s",
				qa.path,
				fmt.Sprintf(format, args...),
			)
			http.Error(w, msg, http.StatusBadRequest)
			ERROR("%s", msg)
		}

		var an string

		//  does the sql query arg have an http alias?

		ha := qa.http_arg
		if ha == nil {
			ha = cf.HTTPQueryArgSet[qa.path]
			an = qa.path
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
	rows, err = q.stmt.Query(argv...)
	if err != nil {
		panic(err)
	}

	//  grumble about slow queries.

	duration = time.Since(start_time).Seconds()
	if duration > cf.WarnSlowSQLQueryDuration {
		WARN("slow query: %s: %.9fs: %s", q.name, duration, url)
	}

	columns, err = rows.Columns()
	if err != nil {
		panic(err)
	}

	//  make the row string vector

	rowv = make([]interface{}, len(columns))
	for i := range rowv {
		rowv[i] = new(sql.NullString)
	}
	return
}

//  JSON reply to an sql query request from a url

func (q *SQLQuery) handle_query_json(
	w http.ResponseWriter,
	r *http.Request,
	cf *Config,
) {
	duration, columns, rows, rowv := q.db_query(w, r, cf)
	if rowv == nil {
		return
	}
	defer rows.Close()

	putf := func(format string, args ...interface{}) {
		fmt.Fprintf(w, format, args...)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	//  write the reply with the query duration

	putf(`[
    "duration,colums,rows",
    %.9f,
    `,
		duration,
	)

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

	//  write the columns

	puta(columns)
	puts(",\n\n    [\n")

	count := uint64(0)
	for rows.Next() {

		if count > 0 {
			puts(",\n")
		}
		count++

		err := rows.Scan(rowv...)
		if err != nil {
			panic(err)
		}
		puts("      [")
		for i, si := range rowv {
			if i > 0 {
				puts(",")
			}
			s := si.(*sql.NullString)
			if s.Valid {
				putjs(s.String)
			} else {
				puts("null")
			}
		}
		puts("]")
	}
	puts("\n    ]\n]\n")
}

//  Tab separated reply to an sql query request from a url.
//  Any tabs or newline in sql data are replaced with a space.
//  See: https://www.iana.org/assignments/media-types/text/tab-separated-values

func (q *SQLQuery) handle_query_tsv(
	w http.ResponseWriter,
	r *http.Request,
	cf *Config,
) {
	_, columns, rows, rowv := q.db_query(w, r, cf)
	if rowv == nil {
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type",
		"text/tab-separated-values; charset=utf-8")

	//  write bytes string to client

	putb := func(b []byte) {
		_, err := w.Write(b)
		if err != nil {
			panic(err)
		}
	}

	//  write a string to client, replace tab and newline with space.
	//  Note: would \r be a better replacement for \n ?

	puts := func(s string) {
		putb([]byte(
			strings.Replace(
				strings.Replace(
					s,
					"\t",
					" ",
					-1,
				),
				"\n",
				" ",
				-1,
			),
		))
	}

	//  write the columns to the client

	for i, s := range columns {
		if i > 0 {
			putb(tab)
		}
		puts(s)
	}
	putb(newline)

	//  write the rows to the client.  null is empty string

	for rows.Next() {

		err := rows.Scan(rowv...)
		if err != nil {
			panic(err)
		}
		for i, si := range rowv {
			if i > 0 {
				putb(tab)
			}
			s := si.(*sql.NullString)
			if s.Valid {
				puts(s.String)
			}
		}
		putb(newline)
	}
}

//  Comma separated spreadsheet reply to an sql query request from a url.

func (q *SQLQuery) handle_query_csv(
	w http.ResponseWriter,
	r *http.Request,
	cf *Config,
) {
	_, columns, rows, rowv := q.db_query(w, r, cf)

	if rowv == nil {
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "text/csv;  charset=utf-8")

	out := csv.NewWriter(w)

	put := func(r []string) {
		err := out.Write(r)
		if err != nil {
			panic(err)
		}
	}

	//  write the column headers
	put(columns)

	//  write the rows

	row := make([]string, len(rowv))
	for rows.Next() {

		err := rows.Scan(rowv...)
		if err != nil {
			panic(err)
		}
		for i, si := range rowv {
			s := si.(*sql.NullString)
			if s.Valid {
				row[i] = s.String
			} else {
				row[i] = ""
			}
		}
		put(row)
	}

	out.Flush()
	err := out.Error()
	if err != nil {
		panic(err)
	}
}
func (q *SQLQuery) handle_query_html(
	w http.ResponseWriter,
	r *http.Request,
	cf *Config,
) {
	_, columns, rows, rowv := q.db_query(w, r, cf)

	if rowv == nil {
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "text/html;  charset=utf-8")

	put := func(s string) {
		w.Write([]byte(s))
	}

	put_row := func(element string, row []string) {

		put("\n <tr>\n")
		for _, s := range row {
			put(`  <`)
			put(element)
			put(`>`)
			put(html.EscapeString(s))
			put(`</`)
			put(element)
			put(`>`)
		}
		put("\n </tr>\n")
	}

	put("<table>\n")

	put_row(`th`, columns)

	//  write the rows

	row := make([]string, len(rowv))
	for rows.Next() {

		err := rows.Scan(rowv...)
		if err != nil {
			panic(err)
		}
		for i, si := range rowv {
			s := si.(*sql.NullString)
			if s.Valid {
				row[i] = s.String
			} else {
				row[i] = ""
			}
		}
		put_row(`td`, row)
	}
	put(`</table>`)
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
