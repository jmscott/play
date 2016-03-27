package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
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

	src, err := os.Open(source_path)
	if err != nil {
		die("%s", err)
	}

	//  parse the standard input into an abstract syntax tree

	ast, depend_order, err := parse(src)
	if err != nil {
		die("%s: %s", source_path, err)
		os.Exit(1)
	}
	src.Close()

	ast.optimize()

	if dump {
		ast.dump()

		fmt.Printf("\nDepend Order of %d Calls:\n", len(depend_order))
		for i, n := range depend_order {
			fmt.Printf("	#%d: %s\n", i, n)
		}
		os.Exit(0)
	}

	flowA := &flow{
		next:     make(chan flow_chan),
		resolved: make(chan struct{}),
	}

	uc := flowA.compile(ast, depend_order)
	close(flowA.resolved)

	in := bufio.NewReader(os.Stdin)
	for {
		line, err := in.ReadString('\n')
		if err != nil {
			if err == io.EOF { //  has flowB resolved?
				break
			}
			panic(err)
		}
		line = strings.TrimRight(line, "\n")
		flowB := &flow{
			line:     line,
			fields:   strings.SplitN(line, "\t", 255),
			next:     make(chan flow_chan),
			resolved: make(chan struct{}),
		}

		//  push flowA to flowB

		for flowA.confluent_count > 0 {

			reply := <-flowA.next
			flowA.confluent_count--

			reply <- flowB
			flowB.confluent_count++
		}

		//  wait for flowB to finish
		uv := <-uc
		if uv == nil {
			break
		}
		close(flowB.resolved)
		flowA = flowB
	}

	os.Exit(0)
}
