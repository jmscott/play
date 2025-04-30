package main

import (
	"bufio"
	"os"
	"fmt"
)

func main() {
	
	var seen map[string]bool
	var buf[4096 * 4096]byte

	in := bufio.NewScanner(os.Stdin)
	in.Buffer(buf[:], len(buf))
	seen = make(map[string]bool, 4096 * 4096)
	for in.Scan() {
		txt := in.Text()
		if !seen[txt] {
			fmt.Println(txt)
			seen[txt] = true
		}
	}
	if err := in.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
