package main

import (
	"regexp"
	"fmt"
)

type HTTPQueryArg struct {
	name	string
	Default	string	`json:"default"`
	Matches string `json:"matches"`
	matches_re	regexp.Regexp
	Required bool `json:"required"`
}

type HTTPQueryArgs map[string]*HTTPQueryArg

func (qa HTTPQueryArgs) load() {
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
		alog("required", fmt.Sprintf("%t", a.Required))
		log("  }")

	}
}
