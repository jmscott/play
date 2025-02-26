package main

import (
	"bufio"
)

type scanner struct {

	name		string
	split		bufio.SplitFunc
	scanner		*bufio.Scanner
}
