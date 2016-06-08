package main

import (
	"strings"
	"io"
	"io/ioutil"
	"encoding/json"
)

type Config struct {
	file_path      string
	Synopsis       string `json:"synopsis"`
	HTTPListen     string `json:"http-listen"`
	RESTPathPrefix string `json:"rest-path-prefix"`
	SQLQueries     `json:"sql-queries"`
	HTTPQueryArgs	`json:"http-query-args"`
}

func (cf *Config) load(path string) {

	cf.file_path = path
	log("loading config file: %s", cf.file_path)

	//  slurp config file into string

	b, err := ioutil.ReadFile(cf.file_path)
	if err != nil {
		die("config load failed: %s", err)
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
}
