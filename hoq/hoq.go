package main

import (
	"fmt"
	"os"
	"syscall"
)

var (
	stderr = os.NewFile(uintptr(syscall.Stderr), "/dev/stderr")
)

func die(format string, args ...interface{}) {

	msg := fmt.Sprintf("hoq: ERROR: %s\n", fmt.Sprintf(format, args...))
	stderr.Write([]byte(msg))
	stderr.Write([]byte("usage: hoq [--dump] <source.hoq>\n"))
	os.Exit(1)
}

func main() {

	source_path := ""
	dump := false

	//  parse command line arguments

	switch len(os.Args) {
	case 2:
		source_path = os.Args[1]
	case 3:
		if os.Args[1] == "--dump" {
			dump = true
			source_path = os.Args[2]
		} else if os.Args[2] == "--dump" {
			dump = true
			source_path = os.Args[1]
		} else {
			die("expected --dump option")
		}
	default:
		die("wrong number of arguments: %d", len(os.Args))
	}

	//  open the source file

	in, err := os.Open(source_path)
	if err != nil {
		die("open failed: %s: %s", source_path, err)
	}

	//  parse the standard input into an abstract syntax tree

	ast, err := parse(in)
	if err != nil {
		die("parser: %s", err)
		os.Exit(1)
	}

	if dump {
		ast.dump()
		os.Exit(0)
	}
	os.Exit(0)
}
