//  rest service damon built from postgresql query files and json config

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var stderr = os.NewFile(uintptr(syscall.Stderr), "/dev/stderr")

func usage() {
	fmt.Fprintf(stderr, "usage: rasqld <config.json>\n")
}

func reply_ERROR(
	status int,
	w http.ResponseWriter,
	r *http.Request,
	format string,
	args ...interface{},
) {
	ERROR(r.RemoteAddr+": "+format, args...)
	http.Error(w, fmt.Sprintf(format, args...), status)
}

func boot() {

	var cf Config

	INFO("process id: %d", os.Getpid())
	INFO("go version: %s", runtime.Version())

	// load the config from json file
	cf.load(os.Args[1])

	// parse the sql queries
	cf.SQLQuerySet.open()

	// install the url to list all sql queries.
	// the path is /<rest-path-prefix.
	INFO("path sql query index: %s", cf.RESTPathPrefix)
	http.HandleFunc(
		cf.RESTPathPrefix + "/json",
		cf.handle_query_index_json,
	)
	http.HandleFunc(
		cf.RESTPathPrefix + "/tsv",
		cf.handle_query_index_tsv,
	)
	http.HandleFunc(
		cf.RESTPathPrefix + "/csv",
		cf.handle_query_index_csv,
	)
	http.HandleFunc(
		cf.RESTPathPrefix + "/html",
		cf.handle_query_index_html,
	)
	http.HandleFunc(
		cf.RESTPathPrefix,
		cf.handle_query_index_html,
	)

	//  for each sql query install four urls to handle the query
	//
	//	/<rest-path-prefix>/<sql-query>/json
	//	/<rest-path-prefix>/<sql-query>/csv
	//	/<rest-path-prefix>/<sql-query>/tsv
	//	/<rest-path-prefix>/<sql-query>/html
	//
	for n, q := range cf.SQLQuerySet {

		// json handler
		http.HandleFunc(
			cf.RESTPathPrefix +
				string(os.PathSeparator) +
				n +
				string(os.PathSeparator) +
				"json",
			cf.new_handler_query_json(q),
		)

		//  tab separated data handler
		http.HandleFunc(
			cf.RESTPathPrefix +
				string(os.PathSeparator) +
				n +
				string(os.PathSeparator) +
				"tsv",
			cf.new_handler_query_tsv(q),
		)

		//  comma separated handler
		http.HandleFunc(
			cf.RESTPathPrefix +
				string(os.PathSeparator) +
				n +
				string(os.PathSeparator) +
				"csv",
			cf.new_handler_query_csv(q),
		)

		//  html table handler
		http.HandleFunc(
			cf.RESTPathPrefix +
				string(os.PathSeparator) +
				n +
				string(os.PathSeparator) +
				"html",
			cf.new_handler_query_html(q),
		)

		//  Note: default ought to be html index into qhole query!
	}

	if cf.HTTPListen == "" && cf.TLSHTTPListen == "" {
		WARN("no listener on either http or https")
		leave(0)
	}

	//  start clear text http server
	if cf.HTTPListen != "" {
		INFO("listening: %s%s", cf.HTTPListen, cf.RESTPathPrefix)
		go func() {
			err := http.ListenAndServe(cf.HTTPListen, nil)
			die("http listen error: %s", err)
		}()
	}

	//  start ssl http server
	if cf.TLSHTTPListen != "" {
		if cf.TLSCertPath == "" {
			die("http listen tls: missing tls-cert-path")
		}
		if cf.TLSKeyPath == "" {
			die("http listen tls: missing tls-key-path")
		}
		INFO("tls listening: %s%s", cf.TLSHTTPListen, cf.RESTPathPrefix)
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
}

func main() {

	if len(os.Args) != 2 {
		die("wrong number of arguments: got %d, expected 1",
			len(os.Args),
		)
	}
	log_init("rasqld")

	boot()

	//  wait for signals
	//  Note: see SIGTERM ignored on mac 10.13.2 from time to time.

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGQUIT)
	signal.Notify(c, syscall.SIGINT)
	s := <-c
	INFO("caught signal: %s", s)
	leave(0)
}
