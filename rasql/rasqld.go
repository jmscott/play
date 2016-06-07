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
	Synopsis       string               `synopsis`
	HTTPListen     string               `json:"http-listen"`
	RESTPathPrefix string               `json:"rest-path-prefix"`
	SQLQueries     map[string]*SQLQuery `json:"sql-queries"`
}

type SQLQueryArguments struct {
	name		string
	pgtype		string
}

type SQLQuery struct {
	name       string
	SourcePath string `json:"source-path"`
	SQLQueryArguments	map[string]string
}

func (q *SQLQuery) die(format string, args ...interface{}) {

	die("sql query: %s: %s", q.SourcePath, fmt.Sprintf(format, args...))
}

func (q *SQLQuery) log(format string, args ...interface{}) {

	log("sql query: %s: %s", q.SourcePath, fmt.Sprintf(format, args...))
}

func (q *SQLQuery) WARN(format string, args ...interface{}) {

	log("WARN: sql query: %s: %s", q.SourcePath,
		fmt.Sprintf(format, args...))
}

func (q *SQLQuery) load(conf *Config) {

	log("loading sql file: %s", q.SourcePath)

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
	if pre["Command Line Arguments"] == "" {
		q.WARN("no Command Line Arguments section")
	}
	q.log("preamble: %d/%d sections/lines", len(pre), line_count)
}

func (conf *Config) load(config_path string) {

	conf.file_path = config_path
	log("loading config file: %s", conf.file_path)

	b, err := ioutil.ReadFile(conf.file_path)
	if err != nil {
		die("config load failed: %s", err)
	}

	dec := json.NewDecoder(strings.NewReader(string(b)))

	err = dec.Decode(&conf)
	if err != nil && err != io.EOF {
		die("config json decoding failed: %s", err)
	}
	log("rest path prefix: %s", conf.RESTPathPrefix)
	log("http listen: %s", conf.HTTPListen)

	log("%d sql query files {", len(conf.SQLQueries))
	for n := range conf.SQLQueries {
		q := conf.SQLQueries[n]
		q.name = n
		log("	%s: {", q.name)
		log("		source-path: %s", q.SourcePath)
		log("	}")
	}
	log("}")

	log("loading sql queries from files")
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
