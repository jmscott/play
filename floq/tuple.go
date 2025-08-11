package main

import (
	"regexp"
)

type attribute struct {
	name		string
	match		regexp.Regexp
}

type tuple struct {
	name		string
	attributes	map[string]attribute
}
