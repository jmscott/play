//  convert "traditional" syslog format to json
//  roughly follows rfc5424
//  https://docs.ruckuswireless.com/fastiron/08.0.60/fastiron-08060-monitoringguide/GUID-88F338BA-B7BF-485C-B1DE-7418710452A6.html
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
const log_time_RE =
		`^((?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) ` +
		`(?:(?: [1-9])|(?:[123][0-9])) ` +
		`[0-9]{2}:[0-9]{2}:[0-9]{2}) `
const host_name_RE = `^([a-zA-Z0-9_-]{1,64}) `
const time_template = `Jan _2 15:04:05 2006`

const process_RE = `^postfix/([a-zA-Z][a-zA-Z0-9_-]{0,31})\[\d{1,20}]: `
const queue_id_RE = `^([A-Z0-9]{12}): `

/*
(?:(warning|statistics|fatal|[A-Z0-9]{12}): )|` +
                     `(daemon started)|(refreshing) `
*/

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
	HostName		map[string]uint64	`json:"host_name"`
	Process			map[string]uint64	`json:"process"`
	QueueId			map[string]uint64	`json:"queue_id"`

	xx512x1			[20]byte
	time_location		*time.Location
}
var run *Run;

var
	log_time_re,
	host_name_re,
	process_re,
	queue_id_re	*regexp.Regexp

func init() {
	log_time_re = regexp.MustCompile(log_time_RE)
	host_name_re = regexp.MustCompile(host_name_RE)
	process_re = regexp.MustCompile(process_RE)
	queue_id_re = regexp.MustCompile(queue_id_RE)
	run = &Run{}
	run.HostName = make(map[string]uint64)
	run.Process = make(map[string]uint64)
	run.QueueId = make(map[string]uint64)
}

func die(format string, args ...interface{}) {

        fmt.Fprintf(os.Stderr, "ERROR: " + format + "\n", args...);
        leave(1)
}

func fdie(what string, err error) {
	die("%s failed: %s", what, err)
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

//  match and extract leading time stamp in log stream: "^Mon DD HH:MM:SS "

func (run *Run) bust_log_time(line []byte) int {

	_die := func(format string, args ...interface{}) {
		die("%s", fmt.Sprintf(
				"bust_logtime: line %d: %s",
				run.LineCount,
				fmt.Sprintf(format, args...),
		))
	}

	//  match and extract "^Mon DD HH:MM:SS "
	midx := log_time_re.FindAllSubmatchIndex(line, -1)
	if midx == nil {
		_die("does not match regexp")
	}

	l := len(midx)
	if l != 1 {
		_die("unexpected len of match idx: got %d, want 1", l)
	}

	//  parse the leading log time

	offset := midx[0]
	l = len(offset)
	if l != 4 {
		_die("unexpected len of match offset: got %d, want 4", l)
	}
	date := string(line[offset[2]:offset[3]])

	tm, err := time.ParseInLocation(
			time_template,
			fmt.Sprintf("%s %d", date, run.Year),
			run.time_location,
	)
	if err != nil {
		_die("time.ParseInLocation(log)", err)
	}
	rfc3339 := tm.Format(time.RFC3339)
	if run.StartTime == "" {
		if run.EndTime != "" {
			panic("EndTime parsed before StartTime")
		}
		run.StartTime = rfc3339
	}

	//  Note: incorrectly assume times totally ordered
	run.EndTime = rfc3339

	return offset[1]
}

//  match and extract host name following leading log timestamp

func (run *Run) bust_host_name(line []byte) int {

	_die := func(format string, args ...interface{}) {
		die("%s", fmt.Sprintf(
				"bust_host_name: line %d: %s",
				run.LineCount,
				fmt.Sprintf(format, args...),
		))
	}

	midx := host_name_re.FindAllSubmatchIndex(line, -1)
	if midx == nil {
		_die("does not match regex")
	}

	var l int
	if l = len(midx);  l != 1 {
		_die("unexpected length of match idx: got %d, want 1", l)
	}

	//  parse the host name after log time

	offset := midx[0]
	if l = len(offset);  l != 4 {
		_die("unexpected length of match offset: got %d, want 4", l)
	}
	host := string(line[offset[2]:offset[3]])
	run.HostName[host]++

	return offset[1]
}

//  match and extract leading process[pid]

func (run *Run) bust_process(line []byte) int {

	_die := func(format string, args ...interface{}) {
		die("%s", fmt.Sprintf(
				"bust_process: line %d: %s",
				run.LineCount,
				fmt.Sprintf(format, args...),
		))
	}

	midx := process_re.FindAllSubmatchIndex(line, -1)
	if midx == nil {
		_die("does not match regex")
	}
	var l int

	if l = len(midx);  l != 1 {
		_die("unexpected len of match idx: got %d, want 1", l)
	}

	//  parse the 'postfix/<process>[pid]: ' after the host name

	offset := midx[0]
	if l = len(offset);  l != 4 {
		_die("unexpected l of match offset: got %d, want 4", l)
	}
	process := string(line[offset[2]:offset[3]])
	run.Process[process]++

	return offset[1]
}

//  bust exception where a queuid was expected

func (run *Run) bust_queue_ex(line []byte) int {
	return 0
}
	
func (run *Run) bust_queue_id(line []byte) int {

	_die := func(format string, args ...interface{}) {
		die("%s", fmt.Sprintf(
				"bust_queue_id: line %d: %s",
				run.LineCount,
				fmt.Sprintf(format, args...),
		))
	}

	midx := queue_id_re.FindAllSubmatchIndex(line, -1)
	if midx == nil {
		return run.bust_queue_ex(line)
	}
	var l int

	if l := len(midx);  l != 1 {
		_die("unexpected length of match idx: got %d, want 1", l)
	}

	//  parse the '[A-Z0-9]{12}: ' after the process[pid]

	offset := midx[0]
	if l = len(offset);  l != 4 {
		_die("unexpected len of match offset: got %d, want 4", l)
	}
	queue_id := string(line[offset[2]:offset[3]])	// matches queueid
	run.QueueId[queue_id]++
	return offset[1]
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
		die("wrong number of cli args: got %d, expected 4", argc)
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
				fdie("strconv.ParseUint(time)", err)
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
				fdie("time.LoadLocation(--time-location)", err)
			}
			run.time_location = loc
		} else {
			die("unknown cli arg: %s", arg)
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
			fdie("bufio.ReadBytes(Stdin)", err)
		}
		run.ByteCount += int64(len(bytes))
		h512.Write(bytes)			//  digest input
		run.LineCount++

		i := run.bust_log_time(bytes)

		bytes = bytes[i:]
		i = run.bust_host_name(bytes)

		bytes = bytes[i:]
		i = run.bust_process(bytes)

		bytes = bytes[i:]
		run.bust_queue_id(bytes)

		run.KnownLineCount++
	}
	run.xx512x1 = xx512x1(h512.Sum(nil))
	run.InputDigest = fmt.Sprintf("%x", run.xx512x1)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "	")
	err := enc.Encode(&run)
	if err != nil {
		fdie("enc.Encode(json)", err) 
	}

	leave(0)
}
