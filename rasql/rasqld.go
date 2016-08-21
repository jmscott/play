//  rest service damon built from postgresql query files and raml configs

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"syscall"
	"time"
)

var stderr = os.NewFile(uintptr(syscall.Stderr), "/dev/stderr")

func usage() {
	fmt.Fprintf(stderr, "usage: rasqld <config.json>\n")
}

func reply_ERROR(
	status int,
	w http.ResponseWriter,
	r *http.Request,
	format string, args ...interface{}) {

	msg := fmt.Sprintf(format, args...)

	ERROR("%s: %s", r.RemoteAddr, msg)
	http.Error(w, msg, status)
}

func ERROR(format string, args ...interface{}) {

	fmt.Fprintf(
		stderr,
		"%s: ERROR: %s\n",
		time.Now().Format("2006/01/02 15:04:05"),
		fmt.Sprintf(format, args...),
	)
}

func WARN(format string, args ...interface{}) {

	fmt.Fprintf(
		stderr,
		"%s: WARN: %s\n",
		time.Now().Format("2006/01/02 15:04:05"),
		fmt.Sprintf(format, args...),
	)
}

func log(format string, args ...interface{}) {

	fmt.Fprintf(stderr, "%s: %s\n",
		time.Now().Format("2006/01/02 15:04:05"),
		fmt.Sprintf(format, args...),
	)
}

func die(format string, args ...interface{}) {

	ERROR(format, args...)
	leave(2)
}

func leave(exit_status int) {
	log("good bye, cruel world")
	os.Exit(exit_status)
}

func main() {

	log("hello, world")

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
		leave(0)
	}()

	var cf Config

	log("process id: %d", os.Getpid())
	log("go version: %s", runtime.Version())
	log("process environment ...")

	//  dump the process environment

	env := os.Environ()
	sort.Strings(env)
	for _, e := range env {
		log("  %s", e)
	}

	cf.load(os.Args[1])
	cf.SQLQuerySet.open()
	defer db.Close()

	log("path sql query index: %s", cf.RESTPathPrefix)
	http.HandleFunc(
		cf.RESTPathPrefix,
		cf.handle_query_index_json,
	)

	//  install sql query handlers
	//
	//	/<rest-path-prefix>/<sql-query>
	//	/<rest-path-prefix>/csv/<sql-query>
	//	/<rest-path-prefix>/tsv/<sql-query>
	//	/<rest-path-prefix>/html/<sql-query>
	//

	for n, q := range cf.SQLQuerySet {

		//  json handler, the default

		http.HandleFunc(
			fmt.Sprintf("%s/%s", cf.RESTPathPrefix, n),
			cf.new_handler_query_json(q),
		)

		//  tsv handler

		http.HandleFunc(
			fmt.Sprintf("%s/tsv/%s", cf.RESTPathPrefix, n),
			cf.new_handler_query_tsv(q),
		)

		//  csv handler

		http.HandleFunc(
			fmt.Sprintf("%s/csv/%s", cf.RESTPathPrefix, n),
			cf.new_handler_query_csv(q),
		)

		//  html handler

		http.HandleFunc(
			fmt.Sprintf("%s/html/%s", cf.RESTPathPrefix, n),
			cf.new_handler_query_html(q),
		)
	}

	if cf.HTTPListen != "" {
		log("listening: %s%s", cf.HTTPListen, cf.RESTPathPrefix)
		go func() {
			err := http.ListenAndServe(cf.HTTPListen, nil)
			die("http listen error: %s", err)
		}()
	}
	if cf.TLSHTTPListen != "" {
		if cf.TLSCertPath == "" {
			die("http listen tls: missing tls-cert-path")
		}
		if cf.TLSKeyPath == "" {
			die("http listen tls: missing tls-key-path")
		}
		log("tls listening: %s%s", cf.TLSHTTPListen, cf.RESTPathPrefix)
		go func() {
			err := http.ListenAndServeTLS(
				cf.TLSHTTPListen,
				cf.TLSCertPath,
				cf.TLSKeyPath,
				nil,
			)
			die("http listen tls: %s", err)
		}()
	}
	pause := make(chan interface{})
	<-pause
}
