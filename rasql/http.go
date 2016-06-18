package main

import (
	"fmt"
	"net/http"
	"regexp"
)

type HTTPQueryArg struct {
	name       string
	Default    string `json:"default"`
	Matches    string `json:"matches"`
	matches_re *regexp.Regexp
	SQLAlias   string `json:"sql-alias"`
}

type HTTPQueryArgSet map[string]*HTTPQueryArg

func (qa HTTPQueryArgSet) load() {
	var err error

	log("http query args: %d args", len(qa))

	alog := func(what, value string) {
		if value == "" {
			return
		}
		log("    %s: %s", what, value)
	}

	for n := range qa {
		a := qa[n]
		a.name = n

		log("  %s: {", a.name)

		alog("default", a.Default)
		alog("matches", a.Matches)
		alog("sql-alias", a.SQLAlias)
		log("  }")

		if a.Matches == "" {
			if a.SQLAlias != "" {
				continue
			}
			die("query arg: missing \"matches\" regexp: %s", a.name)
		}
		a.matches_re, err = regexp.Compile(a.Matches)

		if err != nil {
			ERROR("query-arg: %s: Compile(matches) failed",
				a.name,
			)
			die("query-arg %s: %s",
				a.name,
				err,
			)
		}
	}
}

func herror(
	w http.ResponseWriter,
	status int,
	format string,
	args ...interface{},
) {
	msg := fmt.Sprintf(format, args...)
	ERROR("%s", msg)
	http.Error(w, msg, status)
}
