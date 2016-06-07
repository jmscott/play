//  rest service damon built from postgresql query files and raml configs

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	stderr = os.NewFile(uintptr(syscall.Stderr), "/dev/stderr")
	stdout = os.NewFile(uintptr(syscall.Stdout), "/dev/stdout")
)

type Config struct {
	file_path      string
	Synopsis       string               `json:"synopsis"`
	HTTPListen     string               `json:"http-listen"`
	RESTPathPrefix string               `json:"rest-path-prefix"`
	SQLQueries     map[string]*SQLQuery `json:"sql-queries"`
}

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

func (conf *Config) load(path string) {

	conf.file_path = path
	log("loading config file: %s", conf.file_path)

	//  slurp config file into string

	b, err := ioutil.ReadFile(conf.file_path)
	if err != nil {
		die("config load failed: %s", err)
	}

	//  decode json in config file

	dec := json.NewDecoder(strings.NewReader(string(b)))
	err = dec.Decode(&conf)
	if err != nil && err != io.EOF {
		die("config json decoding failed: %s", err)
	}

	log("rest path prefix: %s", conf.RESTPathPrefix)
	log("http listen: %s", conf.HTTPListen)

	//  summarize sql queries
	//  Note: why not load queries from file here?

	log("%d sql query files {", len(conf.SQLQueries))
	for n := range conf.SQLQueries {
		q := conf.SQLQueries[n]
		q.name = n
		log("	%s: {", q.name)
		log("		source-path: %s", q.SourcePath)
		log("	}")
	}
	log("}")

	//  load sql queries from external files

	log("loading sql queries from %d files", len(conf.SQLQueries))
	for n := range conf.SQLQueries {
		q := conf.SQLQueries[n]
		q.load(conf)
	}
}

func usage() {
	fmt.Fprintf(stderr, "usage: rasqld <config.json>\n")
}

func ERROR(format string, args ...interface{}) {

	fmt.Fprintf(
		stderr,
		"%s: rasqld: ERROR: %s\n",
		time.Now().Format("2006/01/02 15:04:05"),
		fmt.Sprintf(format, args...),
	)
}

func log(format string, args ...interface{}) {

	fmt.Fprintf(stdout, "%s: %s\n",
		time.Now().Format("2006/01/02 15:04:05"),
		fmt.Sprintf(format, args...),
	)
}

func die(format string, args ...interface{}) {

	ERROR(format, args...)
	os.Exit(2)
}

func main() {

	log("hello, world")
	defer log("good bye, cruel world")

	if len(os.Args) != 2 {
		die(
			"wrong number of arguments: got %d, expected 1",
			len(os.Args),
		)
	}

	//  catch signals

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGTERM)
		signal.Notify(c, syscall.SIGQUIT)
		signal.Notify(c, syscall.SIGINT)
		s := <-c
		log("caught signal: %s", s)
		os.Exit(0)
	}()

	var conf Config
	conf.load(os.Args[1])

	log("process id: %d", os.Getpid())
	log("go version: %s", runtime.Version())

	http.HandleFunc(
		conf.RESTPathPrefix,
		func(w http.ResponseWriter, r *http.Request,
		) {
			url := html.EscapeString(r.URL.String())
			fmt.Fprintf(w, "Rest: %s: %s", r.Method, url)
			log("%s: %s: %s", r.RemoteAddr, r.Method, url)
		})

	err := http.ListenAndServe(conf.HTTPListen, nil)
	if err != nil {
		die("%s", err)
	}
	os.Exit(0)
}
