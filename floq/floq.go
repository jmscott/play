package main

import (
	"bufio"
	"fmt"
	"os"
	"syscall"
)

var (
	stderr = os.NewFile(uintptr(syscall.Stderr), "/dev/stderr")
)

//  temporary die() used only during boot up
func croak(format string, args ...interface{}) {

	fmt.Fprintf(stderr, "floq: ERROR: %s\n", fmt.Sprintf(format, args...))
	fmt.Fprintf(stderr, "usage: floq [parse|ast] server.floq\n")
	os.Exit(16)
}

// flowd [parse|ast] <schema.flow>
func main() {

	argv := os.Args[1:]
	argc := len(argv)
	if argc != 2 {
		croak("wrong number of arguments: expected 2, got %d", argc)
	}
	action := argv[0]

	switch action {
		case "parse": 
		case "ast": 
		default:
			croak("unknown action: %s", action)
	}
	floq_path := argv[1]

	floq, err := os.OpenFile(floq_path, os.O_RDONLY, 0)
	if err != nil {
		if os.IsNotExist(err) {
			croak("file does not exist: %s", floq_path)
		}
		croak("OpenFile(%s) failed: %s", floq_path, err)
	}
	defer floq.Close()

	root, err := parse(bufio.NewReader(floq))
	if err != nil {
		croak("parse() failed: %s", err)
	}

	switch action {
	case "parse":
	case "ast":
		root.walk_print(0)
	default:
		croak("unknown action: %s", action)
	}
	os.Exit(0)
}
