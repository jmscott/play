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

func die(format string, args ...interface{}) {

	msg := fmt.Sprintf("hoq: ERROR: %s\n", fmt.Sprintf(format, args...))
	stderr.Write([]byte(msg))
	os.Exit(1)
}

func main() {

	argc := len(os.Args)
	if argc > 2 {
		die("wrong number of command line arguments")
	}
	ast, err := parse(bufio.NewReader(os.Stdin))
	if err != nil {
		die("parser: %s", err)
		os.Exit(1)
	}
	if argc == 2 {
		if os.Args[1] != "ast" {
			die("unknown action: %s", os.Args[1])
		}
		ast.print()
	}
	os.Exit(0)
}
