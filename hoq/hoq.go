package main

import (
	// "bufio"
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

	floA := &flow{
		next:     make(chan flow_chan),
		resolved: make(chan struct{}),
	}
	floA.compile(ast, depend_order)
	close(floA.resolved)

	/*
	in := bufio.NewReader(os.Stdin)
	for {
		line, err := in.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		line = TrimRight(line, "\n")
		floB = &flow{
			line:     line,
			fields:	  strings.Split(line, "\t"),
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
		fv := <-fc
		if fv == nil {
			break
		}
		close(fv.resolved)

		//  cheap sanity test

		if fv.flow.seq != flowB.seq {
			panic("fdr out of sync with flowB")
		}

		//  send stats to server goroutine

		sam.ok_count = uint64(fv.fdr.ok_count)
		sam.fault_count = uint64(fv.fdr.fault_count)
		sam.wall_duration = fv.fdr.wall_duration
		work.flow_sample_chan <- sam

		flowA = flowB
	}
	*/

	os.Exit(0)
}
