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
	fmt.Fprintf(stderr, "usage: floq [server|parse|ast] server.floq\n")
	os.Exit(16)
}

// flowd [server|parse|ast] <schema.flow>
func main() {

	argv := os.Args[1:]
	argc := len(argv)
	if argc != 2 {
		croak("wrong number of arguments: expected 2, got %d", argc)
	}
	action := argv[0]

	switch action {
		case "server":
		case "ast":
		case "depend":
		case "parse": 
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

	_, err = parse(bufio.NewReader(floq))
	if err != nil {
		croak("parse() failed: %s", err)
	}

	switch action {
	case "parse":
	case "server":
	case "ast":
	default:
		croak("unknown action: %s", action)
	}
	os.Exit(0)
}
