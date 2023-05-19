package main

import (
	"regexp"
)

type HTTPQueryArg struct {
	name       string
	Default    string `json:"default"`
	Matches    string `json:"matches"`
	matches_re *regexp.Regexp
	SQLAlias   string `json:"sql_alias"`
}

type HTTPQueryArgSet map[string]*HTTPQueryArg

func (qa HTTPQueryArgSet) load() {
	var err error

	INFO("http query args: %d args", len(qa))

	aINFO := func(what, value string) {
		if value == "" {
			return
		}
		INFO("    %s: %s", what, value)
	}

	for n := range qa {
		a := qa[n]
		a.name = n

		INFO("  %s: {", a.name)

		aINFO("default", a.Default)
		aINFO("matches", a.Matches)
		aINFO("sql-alias", a.SQLAlias)
		INFO("  }")

		if a.Matches == "" {
			if a.SQLAlias != "" {
				continue
			}
			die("query arg: missing \"matches\" regexp: %s", a.name)
		}
		a.matches_re, err = regexp.Compile(a.Matches)

		if err != nil {
			die("query arg: %s: Compile(matches) failed: %s",
				a.name,
				err,
			)
		}
	}
}
