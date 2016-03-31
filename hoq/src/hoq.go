//  main() for command line hoq interpreter

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
	stdout = os.NewFile(uintptr(syscall.Stdout), "/dev/stdout")
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

	//  open the hoq source file

	src, err := os.Open(source_path)
	if err != nil {
		die("%s", err)
	}

	//  let YACC parse the standard input into an abstract syntax tree

	ast, depend_order, err := parse(src)
	if err != nil {
		die("%s: %s", source_path, err)
		os.Exit(1)
	}
	src.Close()

	//  rewrite nodes in the tree for type casts and trivial optimizations.

	ast.rewrite()

	//  are we just dumping the syntax tree for debugging?

	if dump {
		ast.dump()

		fmt.Printf("\nDepend Order of %d Calls:\n", len(depend_order))
		for i, n := range depend_order {
			fmt.Printf("	#%d: %s\n", i, n)
		}
		os.Exit(0)
	}

	//  set up first flow to compile all ast nodes into a single flow graph.
	//  each flow terminates by sending a count of the fired unix processes.

	flowA := &flow{
		next:     make(chan flow_chan),
		resolved: make(chan struct{}),
	}
	uc := flowA.compile(ast, depend_order)
	close(flowA.resolved)

	//  start pumping standard input to the flow graph of nodes

	in := bufio.NewReader(os.Stdin)
	for {
		line, err := in.ReadString('\n')
		if err != nil {
			if err == io.EOF { //  has flowB resolved?
				break
			}
			panic(err)
		}

		//  trim and split the input line of text

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

		//  wait for flowB to finish, exiting on nil

		if <-uc == nil {
			break
		}

		//  broadcast to all waiting nodes in flowb by closing
		//  the resolved channel, which all nodes listen on.

		close(flowB.resolved)

		//  and so the wheel turns.

		flowA = flowB
	}

	os.Exit(0)
}
