package main

import (
	"bufio"
)

type scanner struct {

	name		string
	path		string
	split		bufio.SplitFunc
	scanner		*bufio.Scanner
}

func (scan *scanner) frisk_att(al *ast) string {

	for an := al.left;  an != nil;  an = an.next {
		switch an.left.string {
		case "Path":
			if scan.path != "" {
				return "attribute more than once: path"
			}
			scan.path = an.right.string
		default:
			return "unknown attribute: " + an.left.string
		}
	}
	if scan.path == "" {
		return "missing required attribute: path"
	}
	return ""
}
