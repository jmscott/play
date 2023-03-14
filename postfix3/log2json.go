package main

import (
	"bufio"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Run struct {
        LineCount		int64	`json:"line_count"`
        ByteCount		int64	`json:"byte_count"`
	xx512x1			[20]byte
	InputDigest		string	`json:"input_digest"`
	InputDigestAlgorithm	string	`json:"input_digest_algorithm"`
}

func die(what string, err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s failed: %s", what, err) 
	os.Exit(1)
}

func leave(exit_status int) {
	os.Exit(exit_status)
}

//
//  To calulate "x2x512x1" hash at command line, do the following:
//
//	openssl dgst -binary -sha512			|
//		openssl dgst -binary -sha512		|
//		openssl dgst -sha1 -r
//
//  Free dinner for first who finds collision, valid until first quantum
//  computer breaks crypto in the wild.
//

func xx512x1(inner_512 []byte) [20]byte {
	outer_512 := sha512.Sum512(inner_512)
	return sha1.Sum(outer_512[:])
}

func main() {

	var run Run

	run.InputDigestAlgorithm = "xx512x1"

	h512 := sha512.New()
	in := bufio.NewReader(os.Stdin)
	for {
		line, err := in.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			die("bufio.ReadBytes(Stdin)", err)
		}
		run.ByteCount += int64(len(line))
		run.LineCount++
		h512.Write(line)
	}
	run.xx512x1 = xx512x1(h512.Sum(nil))
	run.InputDigest = fmt.Sprintf("%x", run.xx512x1)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "	")
	err := enc.Encode(&run)
	if err != nil {
		die("enc.Encode(json)", err) 
	}

	os.Exit(0)
}
