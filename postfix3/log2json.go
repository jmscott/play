package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func
die(what string, err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s failed: %s", what, err) 
	os.Exit(1)
}

func
leave(exit_status int) {
	os.Exit(exit_status)
}

func
main() {

	line_count := uint64(0)
	in := bufio.NewReader(os.Stdin)
	for {
		_, err := in.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			die("bufio.ReadString(Stdin)", err)
		}
		line_count++
	}
	fmt.Printf("line_count: %d\n", line_count)

}
