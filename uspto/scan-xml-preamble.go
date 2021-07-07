package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

var offset int
var line_no int

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func scanPreamble(
	data []byte,
	atEOF bool,
) (
	advance int,
	token []byte,
	err error,
) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		
		line_no++
		offset += i + 1
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	if atEOF {
		line_no++
		offset += len(data)
		return len(data), dropCR(data), nil
	}
	return 0, nil, nil
}

func main() {

	in := os.Stdin
	scan := bufio.NewScanner(in)
	scan.Split(scanPreamble)
	for {
		pre_offset := offset
		if !scan.Scan() {
			break
		}
		if scan.Text() == `<?xml version="1.0" encoding="UTF-8"?>` {
			fmt.Printf("%d	%d\n", pre_offset, line_no)
		}
	}
}
