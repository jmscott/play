package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strings"
)

type SQLQueryArg struct {
	name   string
	PGType string `json:"type"`
	order  uint8
}

type SQLQueryArgs map[string]*SQLQueryArg

type SQLQuery struct {
	name         string
	SourcePath   string `json:"source-path"`
	SQLQueryArgs SQLQueryArgs
}

type SQLQuerySet map[string]*SQLQuery

func (queries SQLQuerySet) load() {

	log("%d sql query files in config {", len(queries))
	for n := range queries {
		q := queries[n]
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
	err = dec.Decode(&q.SQLQueryArgs)
	if err != nil && err != io.EOF {
		q.die("failed to decode json in command line arguments")
	}

	if len(q.SQLQueryArgs) == 0 {
		log("    no command line arguments")
		return
	}
	if len(q.SQLQueryArgs) > 255 {
		q.die("> 255 sql query arguments")
	}

	//  verify pg sql types

	log("    %d arguments: {", len(q.SQLQueryArgs))
	for n := range q.SQLQueryArgs {
		qa := q.SQLQueryArgs[n]
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
