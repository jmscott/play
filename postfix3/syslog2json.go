//  convert "traditional" syslog format to json
package main

import (
	"time"
	"bufio"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
)

//  typical syslog timestamp for mail logging
const time_RE =
		`^((?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) ` +
		`(?:(?: |[1-9])|(?:[123][0-9])) ` +
		`[0-9]{2}:[0-9]{2}:[0-9]{2}) `
const host_RE = ` [a-zA-Z0-9_-]{1,64} `

type Run struct {
        LineCount		int64	`json:"line_count"`
        ByteCount		int64	`json:"byte_count"`
        KnownLineCount		int64	`json:"known_line_count"`
        UnknownLineCount	int64	`json:"unknown_line_count"`
	InputDigest		string	`json:"input_digest"`
	InputDigestAlgo		string	`json:"input_digest_algo"`
	StartTime		string	`json:"start_time"`
	EndTime			string	`json:"end_time"`
	TimeLocation		string	`json:"time_location"`
	Year			uint16	`json:"year"`

	xx512x1			[20]byte
	time_location		*time.Location
}
var run Run

var line_re, time_re *regexp.Regexp

func init() {
	time_re = regexp.MustCompile(time_RE)
	line_re = regexp.MustCompile(host_RE)
}

func die(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s failed: %s", msg, err) 
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: %s failed", msg)
	}
	os.Exit(1)
}

func panic(msg string) {
	die("PANIC: " + msg, nil)
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

//  convert typical syslog time stamp to RFC3339

func syslog_time(date string) {
	tm, err := time.ParseInLocation(
			"Jan _2 15:04:05 2006",
			fmt.Sprintf("%s %d", date, run.Year),
			run.time_location,
	)
	if err != nil {
		die("time.ParseInLocation", err)
	}
	rfc3339 := tm.Format(time.RFC3339)
	if run.StartTime == "" {
		if run.EndTime != "" {
			panic("EndTime parsed before StartTime")
		}
		run.StartTime = rfc3339
	}
	run.EndTime = rfc3339
}

func a2die(option string) {
	die("option given twice: --" + option, nil)
}

func axdie(option string) {
	die("no required option: --" + option, nil)
}

func main() {

	argc := len(os.Args) - 1
	if argc != 4 {
		msg := "wrong number of cli args"
		msg = fmt.Sprintf("%s: got %d, expected 2 or 4", argc)
		die(msg, nil)
	}

	for i := 1;  i <= argc;  i++  {
		arg := os.Args[i]
		if arg == "--year" {
			if run.Year > 0 {
				a2die("year")
			}
			i++
			u, err := strconv.ParseUint(os.Args[i], 10, 12)
			if err != nil {
				die("strconv.ParseUint(time)", err)
			}
			run.Year = uint16(u)
		} else if arg == "--time-location" {
			if run.TimeLocation != "" {
				a2die("time-location")
			}
			i++
			run.TimeLocation = os.Args[i]
			loc, err := time.LoadLocation(run.TimeLocation)
			if err != nil {
				die("time.LoadLocation()", err)
			}
			run.time_location = loc
		} else {
			die("unknown cli arg: " + arg, nil)
		}
	}
	if run.TimeLocation == "" {
		axdie("time-location")
	}
	if run.Year == 0 {
		axdie("year")
	}

	run.InputDigestAlgo = "xx512x1"
	h512 := sha512.New()
	in := bufio.NewReader(os.Stdin)
	for {
		bytes, err := in.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			die("bufio.ReadBytes(Stdin)", err)
		}
		run.ByteCount += int64(len(bytes))
		run.LineCount++

		h512.Write(bytes)	//  digest of input

		midx := time_re.FindAllSubmatchIndex(bytes, -1)
		//fmt.Println("WTF: midx:", midx)

		offset := midx[0]
		//fmt.Println("WTF: offset[0]:", offset)

		syslog_time(string(bytes[offset[2]:offset[3]]))

		run.KnownLineCount++
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
