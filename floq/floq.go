package main

import (
	"bufio"
	"fmt"
	"os"
)

var usage = "usage: floq [ast|compile|parse|server] path/to/prog.floq\n"

func exit(status int) {
	os.Exit(status)
}

//  die() during boot up

func croak(format string, args ...interface{}) {

	fmt.Fprintf(
		os.Stderr,
		"floq: ERROR: %s\n",
		fmt.Sprintf(format, args...),
	)
	fmt.Fprintf(os.Stderr, usage)
	exit(16)
}

// usage: floq [server|parse|ast|compile] <schema.flow>

func main() {

	argv := os.Args[1:]
	argc := len(argv)
	if argc != 2 {
		croak("wrong number of arguments: expected 2, got %d", argc)
	}
	action := argv[0]

	switch action {
		case "parse", "ast", "compile", "server":
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
		croak("parse(%s) failed: %s", floq_path, err)
	}

	switch action {
	case "parse":

	case "ast":
		root.walk_print(0, nil)

	case "compile":
		_, err :=  compile(root)
		if err != nil {
			croak("compile(%s) failed: %s", floq_path, err)
		}
	case "server":
		err := server(root)
		if err != nil {
			croak("server(%s) failed: %s", floq_path, err) 
		}
	default:
		croak("unknown action: %s", action)
	}
	exit(0)
}

func corrupt(format string, args ...interface{}) {
	panic(fmt.Sprintf("corrupt: " + format, args...))
}

func WTF(format string, args ...interface{}) {

	os.Stderr.Write([]byte(fmt.Sprintf("WTF: " + format, args...) + "\n"))
}
