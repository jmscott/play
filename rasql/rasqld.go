//  rest service damon built from postgresql query files and raml configs

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var listen string = ":8080"
var path_prefix = "/"

var (
	stderr = os.NewFile(uintptr(syscall.Stderr), "/dev/stderr")
	stdout = os.NewFile(uintptr(syscall.Stdout), "/dev/stdout")
)

type query_cli_arg struct {
	name   string
	pgtype string
}

type query_file struct {
	query_path    string
	source_path   string
	query_cli_arg map[string]query_cli_arg
	in            *bufio.Reader

	line_no    int
	query_args map[string]query_cli_arg
}

var rest_queries = []query_file{

	{"query/keyword", "lib/keyword.sql", nil, nil, 0, nil},
}

func init() {

	for _, q := range rest_queries {
		q.query_cli_arg = make(map[string]query_cli_arg)
	}
}

func usage() {
	fmt.Fprintf(stderr, "usage: raqd <config.json>\n")
}

func ERROR(format string, args ...interface{}) {

	fmt.Fprintf(
		stderr,
		"%s: raqd: ERROR: %s\n",
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

func leave(status int) {
	log("good bye, cruel world")
	os.Exit(status)
}

func die(format string, args ...interface{}) {

	ERROR(format, args...)
	leave(2)
}

//  Load the very first comment in the file and extract the
//  json in the "Command Line Arguments:" section.

func (q *query_file) load_preamble() {

	pre, _, err := parse_Ccomment_preamble(q.in)
	if err != nil {
		q.die("error parsing preamble: %s", err)
	}

	//  section "Command Line Arguments" is json descriptions of args

	js := pre["Command Line Arguments"]
	if js == "" {
		q.die("missing preamble section: Command Line Arguments")
	}

	//  insure the cli json is well formed

	dec := json.NewDecoder(strings.NewReader(js))
	err = dec.Decode(&q.query_args)
	if err != nil && err != io.EOF {
		log("ERROR:	json: %s", js)
		q.die("failed to decode cli json: %s", err.Error())
	}
}

func (q *query_file) die(format string, args ...interface{}) {

	msg := fmt.Sprintf(format, args...)
	if q.line_no > 0 {
		msg += fmt.Sprintf(" near line %d", q.line_no)
	}
	die("%s: %s", q.source_path, msg)
}

func (q *query_file) load() {

	log("loading sql rest query: %s", q.query_path)
	log("	sql source file: %s", q.source_path)

	_die := func(format string, args ...interface{}) {
		q.die("load: %s", fmt.Sprintf(format, args...))
	}

	inf, err := os.Open(q.source_path)
	if err != nil {
		_die("%s", err)
	}
	defer inf.Close()

	q.in = bufio.NewReader(inf)

	//  first line of sql file must be "/*"

	line, err := q.in.ReadString('\n')
	if err != nil {
		_die(err.Error())
	}
	if line != "/*" {
		_die("first line is not \"/*\"")
	}
	q.load_preamble()
}

func main() {

	log("hello, world")

	if len(os.Args) != 1 {
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
		leave(0)
	}()

	log("process id: %d", os.Getpid())
	log("go version: %s", runtime.Version())
	log("listen service: %s", listen)

	log("loading sql files ...")
	c := 0
	for _, q := range rest_queries {
		q.load()
		c++
	}
	log("loaded %d sql files", c)

	http.HandleFunc(
		path_prefix,
		func(w http.ResponseWriter, r *http.Request,
		) {
			url := html.EscapeString(r.URL.String())
			fmt.Fprintf(w, "Rest: %s: %s", r.Method, url)
			log("%s: %s: %s", r.RemoteAddr, r.Method, url)
		})

	err := http.ListenAndServe(listen, nil)
	if err != nil {
		die("%s", err)
	}
	leave(0)
}
