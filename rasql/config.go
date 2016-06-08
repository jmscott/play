package main

import (
	"strings"
	"io"
	"io/ioutil"
	"encoding/json"
)

type Config struct {
	source_path      string

	Synopsis       string `json:"synopsis"`
	HTTPListen     string `json:"http-listen"`
	RESTPathPrefix string `json:"rest-path-prefix"`
	SQLQueries     `json:"sql-queries"`
	HTTPQueryArgs	`json:"http-query-args"`
}

func (cf *Config) load(path string) {

	log("loading config file: %s", path)

	cf.source_path = path

	//  slurp config file into string

	b, err := ioutil.ReadFile(cf.source_path)
	if err != nil {
		die("config read failed: %s", err)
	}

	//  decode json in config file

	dec := json.NewDecoder(strings.NewReader(string(b)))
	err = dec.Decode(&cf)
	if err != nil && err != io.EOF {
		die("config json decoding failed: %s", err)
	}

	log("rest path prefix: %s", cf.RESTPathPrefix)
	log("http listen: %s", cf.HTTPListen)

	//  summarize sql queries
	//  Note: why not load queries from file here?

	cf.SQLQueries.load()
	cf.HTTPQueryArgs.load()

	log("%s: done", cf.source_path)
}
