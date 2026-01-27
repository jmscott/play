package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var usage = "usage: floq [pass1|pass2|compile|frisk|server] path/to/prog.floq\n"

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

func main() {

	argv := os.Args[1:]
	argc := len(argv)
	if argc != 2 {
		croak("wrong number of arguments: expected 2, got %d", argc)
	}
	action := argv[0]

	switch action {
		case "frisk",
		     "pass1",
		     "pass2",
		     "compile",
		     "server":
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

	go func() {
		c := make(chan os.Signal)
		signal.Notify(
			c,
			syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGINT,
		)
		s := <-c
		fmt.Fprintf(os.Stderr, "\nfloq: caught signal: %s\n", s)
		os.Exit(0)
	}()

	switch action {
	case "pass1":
		root.walk_print(0, nil)
	case "pass2":
		if err = xpass2(root);  err != nil {
			croak("xpass2(%s) failed: %s", floq_path, err)
		}
		root.walk_print(0, nil)

	case "frisk":
		if err = xpass2(root);  err != nil {
			croak("frisk: xpass2(%s) failed: %s", floq_path, err)
		}
	case "compile":
		if err := xpass2(root);  err != nil {
			croak("compile/pass2(%s) failed: %s", floq_path, err)
		}
		compile(root)	//  any error is a panic()
	case "server":
		if err := xpass2(root);  err != nil {
			croak("server/pass2(%s) failed: %s", floq_path, err)
		}
		if err := server(root);  err != nil {
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
