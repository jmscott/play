package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type SQLQueryArg struct {
	name	string
	PGType	string	`json:"type"'`
}

type SQLQueryArgs map[string]SQLQueryArg

type SQLQuery struct {
	name              string
	SourcePath        string `json:"source-path"`
	SQLQueryArgs	SQLQueryArgs
}

func (q *SQLQuery) die(format string, args ...interface{}) {

	die("sql query: %s: %s", q.SourcePath, fmt.Sprintf(format, args...))
}

func (q *SQLQuery) WARN(format string, args ...interface{}) {

	log("WARN: sql query: %s: %s", q.SourcePath,
		fmt.Sprintf(format, args...))
}

func (q *SQLQuery) load(conf *Config) {

	log("	%s", q.SourcePath)

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
		log("		no command line arguments")
		return
	}
	log("		%d arguments: {", len(q.SQLQueryArgs))
	for n := range q.SQLQueryArgs {
		qa := q.SQLQueryArgs[n]
		qa.name = n
		log("			%s:{pgtype:%s}", qa.name, qa.PGType)
		
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
	log("		}")
}
