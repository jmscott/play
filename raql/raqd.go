//  rest service damon built from postgresql query files and raml configs

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var listen string = ":8080";
var path_prefix = "/rest/pdfbox2/";

var (
	stderr = os.NewFile(uintptr(syscall.Stderr), "/dev/stderr")
	stdout = os.NewFile(uintptr(syscall.Stdout), "/dev/stdout")
)

type query_cli_arg struct {
	name	string
	pgtype	string
}

type query_file struct {
	query_path	string
	source_path	string
	query_cli_arg	map[string]query_cli_arg
	in		*bufio.Reader
	line_no		int
	query_args	map[string]query_cli_arg
}

var rest_queries = []query_file {

	{"query/keyword",	"lib/keyword.sql", nil, nil, 0, nil},
}

var	preamble_begin_re *regexp.Regexp	
var	preamble_end_re *regexp.Regexp	

var	sql_json_begin_re *regexp.Regexp
var	sql_json_end_re *regexp.Regexp
var	sql_json_prefix_re *regexp.Regexp

func init() {
	
	for _, q := range rest_queries {
		q.query_cli_arg = make(map[string]query_cli_arg)
	}

	preamble_begin_re = regexp.MustCompile(`^\s*/[*]\s*$`)
	preamble_end_re = regexp.MustCompile(`^\s*[*]/\s*$`)

	//  regular expression for parsing comment preamble

	sql_json_begin_re = regexp.MustCompile(
	      `^\s*[*]\s*Command\s+Line\s+Arguments:\s*{\s*$`)
	sql_json_prefix_re = regexp.MustCompile(`^\s*[*] *(.*)`)
	sql_json_end_re = regexp.MustCompile(`^ [*]  }\s*$`)
}

func usage() {
	fmt.Fprintf(stderr, "usage: rest-pdfbox2\n")
}

func ERROR(format string, args ...interface{}) {

	fmt.Fprintf(
		stderr,
		"%s: rest-pdfbox2: ERROR: %s\n",
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
	os.Exit(status);
}

func die(format string, args ...interface{}) {

	ERROR(format, args...)
	leave(2)
}

//  Load the very first comment in the file and extract the
//  json in the "Command Line Arguments:" section.

func (q *query_file) load_preamble() {

	const clia_redefined string =
		`section "Command Line Arguments:" redefined`

	var qjson bytes.Buffer

	seen_clia_section := false
	in_clia_section := false

	_die := func(format string, args ...interface{}) {
		die("%s: preamble: %s near line %d",
			q.source_path,
			fmt.Sprintf(format, args...),
			q.line_no,
		)
	}

	for {
		line, err := q.in.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		q.line_no++

 		//  Command Line Arguments:  {

		if sql_json_begin_re.MatchString(line) {
			if seen_clia_section {
				_die(clia_redefined)
			}
			seen_clia_section = true
			in_clia_section = true
			qjson.WriteString("{")
			continue
		}

		if !in_clia_section {
			continue
		}

		//  At end of json declation?

		if sql_json_end_re.MatchString(line) {
			in_clia_section = false
			qjson.WriteString("}")
			continue
		}

		//  Extract json line

		matches := sql_json_prefix_re.FindStringSubmatch(line)
		if len(matches) != 2 {
			_die("unexpected prefix in json declaration")
		}
		_, _ = qjson.WriteString(matches[1])
	}

	//  verify json command line args exist
	js := qjson.String()
	if js == "" {
		_die("missing json command section")
	}
	dec := json.NewDecoder(strings.NewReader(js))
	err := dec.Decode(&q.query_args)
	if err != nil && err != io.EOF {
		log("	json: %s", js)
		_die("failed to decode json: %s", err.Error())
	}
}

func (q *query_file) load() {
	
	log("loading sql rest query: %s", q.query_path)
	log("	sql source file: %s", q.source_path)

	inf, err := os.Open(q.source_path)
	if err != nil {
		die("%s", err)
	}
	defer inf.Close()

	q.in = bufio.NewReader(inf)

	for {
		line, err := q.in.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		line = strings.TrimSuffix(line, "\n")
		q.line_no++

		//  look for comment preamble at start of file

		if preamble_begin_re.MatchString(line) {
			if q.line_no == 1 {
				q.load_preamble()
			}
			continue
		}
	}
}

func main() {

	log("hello, world");

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
		url :=  html.EscapeString(r.URL.String())
		fmt.Fprintf(w, "Rest: %s: %s", r.Method, url)
		log("%s: %s: %s", r.RemoteAddr, r.Method, url)
	})

	err := http.ListenAndServe(listen, nil)
	if err != nil {
		die("%s", err)
	}
	leave(0)
}
