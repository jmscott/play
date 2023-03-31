//  convert "traditional" syslog format for mail to json,
//  roughly following rfc3164 and rfc5424
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
const source_host_RE = `^([a-zA-Z0-9_][a-zA-Z0-9_-]{0,63}) `
const log_time_template = `Jan _2 15:04:05 2006`

const process_RE = `^postfix/([a-zA-Z][a-zA-Z0-9_-]{0,31})\[\d{1,20}]: `
const queue_id_RE = `^([A-Z0-9]{10,12}): `
const warning_RE = `^warning: `
const statistics_RE = `^statistics: `
const fatal_RE = `^fatal: `
const arg_custom_RE = `^([a-z][a-z0-9_-]{0,16}):(.{1,64})$`
const backwards_compat_RE = `^using backwards-compatible `

//  Note: investigate why $ pattern fails in FindSubmatchIndex()
const refresh_postfix_RE = `^refreshing the Postfix mail system`

const reload_RE = `^reload -- version 3`	//  force postfix3
const daemon_started_RE = `^daemon started -- version `
const connect_from_RE = `^connect from `
const lost_connect_RE = `^lost connection after `
const disconnect_from_RE = `^disconnect from `
const connect_to_RE = `^connect to `
const message_repeated_RE = `^message repeated [1-9]{1,20} times: `

const start_postfix_RE = `^starting the Postfix mail system`

type CustomRE struct {
	Tag		string	`json:"tag"`
	Field		string	`json:"field"`
	RegExp		string	`json:"regexp"`
	regexp		*regexp.Regexp
}

type SourceCount struct {
	ProcessCount		map[string]int64	`json:"process_count"`

        UnknownLineCount	int64	`json:"unknown_line_count"`
	WarningCount		int64	`json:"warning_count"`
	StatisticsCount		int64	`json:"statistics_count"`
	FatalCount		int64	`json:"fatal_count"`
	DaemonStartedCount	int64	`json:"daemon_started_count"`
	RefreshPostfixCount	int64	`json:"refresh_postfix_count"`
	ReloadCount		int64	`json:"reload_count"`
	ConnectFromCount	int64	`json:"connect_from_count"`
	LostConnectCount	int64	`json:"lost_connect_count"`
	DisconnectFromCount	int64	`json:"disconnect_from_count"`
	ConnectToCount		int64	`json:"connect_to_count"`
	BackwardsCompatCount	int64	`json:"backwards_compat_count"`
	MessageRepeatedCount	int64	`json:"message_repeated_count"`
	StartPostfixCount	int64	`json:"start_postfix_count"`
}

type SourceHost struct {
	run			*Run

	HostName		string		`json:"host_name"`
	CountStat		SourceCount	`json:"count_stat"`		
	CustomRECount		map[string]int64`json:"custom_re_count"`
}

type Run struct {
	CLICustomRE		map[string]*CustomRE
        LineCount		int64	`json:"line_count"`
        ByteCount		int64	`json:"byte_count"`
	InputDigest		string	`json:"input_digest"`
	InputDigestAlgo		string	`json:"input_digest_algo"`
	MinTime			time.Time`json:"min_time"`
	MaxTime			time.Time`json:"max_time"`
	TimeLocation		string	`json:"time_location"`
	Year			uint16	`json:"year"`
	SourceHost		map[string]*SourceHost	`json:"source_host"`


	xx512x1			[20]byte
	time_location		*time.Location
}

var
	log_time_re,
	source_host_re,
	process_re,
	warning_re,
	statistics_re,
	fatal_re,
	daemon_started_re,
	refresh_postfix_re,
	reload_re,
	connect_from_re,
	lost_connect_re,
	disconnect_from_re,
	connect_to_re,
	arg_custom_re,
	backwards_compat_re,
	message_repeated_re,
	start_postfix_re,
	queue_id_re	*regexp.Regexp

func init() {
	log_time_re = regexp.MustCompile(log_time_RE)
	source_host_re = regexp.MustCompile(source_host_RE)
	process_re = regexp.MustCompile(process_RE)
	queue_id_re = regexp.MustCompile(queue_id_RE)
	warning_re = regexp.MustCompile(warning_RE)
	statistics_re = regexp.MustCompile(statistics_RE)
	fatal_re = regexp.MustCompile(fatal_RE)
	daemon_started_re = regexp.MustCompile(daemon_started_RE)
	refresh_postfix_re = regexp.MustCompile(refresh_postfix_RE)
	reload_re = regexp.MustCompile(reload_RE)
	connect_from_re = regexp.MustCompile(connect_from_RE)
	lost_connect_re = regexp.MustCompile(lost_connect_RE)
	disconnect_from_re = regexp.MustCompile(disconnect_from_RE)
	connect_to_re = regexp.MustCompile(connect_to_RE)
	arg_custom_re = regexp.MustCompile(arg_custom_RE)
	backwards_compat_re = regexp.MustCompile(backwards_compat_RE)
	message_repeated_re = regexp.MustCompile(message_repeated_RE)
	start_postfix_re = regexp.MustCompile(start_postfix_RE)
}

func die(format string, args ...interface{}) {

        fmt.Fprintf(os.Stderr, "ERROR: " + format + "\n", args...);
        leave(1)
}

func fdie(what string, err error) {
	die("%s failed: %s", what, err)
}

func panic(msg string) {
	die("PANIC: %s", msg)
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

//  match, extract and set min/max log time

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
			log_time_template,
			fmt.Sprintf("%s %d", date, run.Year),
			run.time_location,
	)
	if err != nil {
		_die("time.ParseInLocation(log)", err)
	}
	if run.MinTime.IsZero() {
		if !run.MaxTime.IsZero() {
			panic("MaxTime parsed before MinTime")
		}
		run.MinTime = tm
	}
	if run.MaxTime.Before(tm) {
		run.MaxTime = tm
	}

	return offset[1]
}

//  match and extract host name following leading log timestamp

func (run *Run) bust_source_host(line []byte) (int, *SourceHost) {

	_die := func(format string, args ...interface{}) {
		die("%s", fmt.Sprintf(
				"bust_source_host: line %d: %s",
				run.LineCount,
				fmt.Sprintf(format, args...),
		))
	}

	midx := source_host_re.FindAllSubmatchIndex(line, -1)
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

	shost := run.SourceHost[host]
	if shost == nil {
		shost = &SourceHost{
			run:		run,
			HostName: 	host,
		}
		shost.CountStat.ProcessCount = make(map[string]int64)
		shost.CustomRECount = make(map[string]int64)
		run.SourceHost[host] = shost
	}
	return offset[1], run.SourceHost[host]
}

func (shost *SourceHost) bust_custom_re(line []byte) int {

	run := shost.run
	for _, cre := range run.CLICustomRE {
		if cre.regexp.Find(line) != nil {
			shost.CustomRECount[cre.Tag]++
			return -1
		}
	}
	die("bust_custom_re: line %d: no match for custom re", run.LineCount)
	return -2
}

//  match and extract leading process[pid]

func (shost *SourceHost) bust_process(line []byte) int {

	run := shost.run
	_die := func(format string, args ...interface{}) {
		die("%s", fmt.Sprintf(
				"bust_process: line %d: %s",
				run.LineCount,
				fmt.Sprintf(format, args...),
		))
	}

	midx := process_re.FindAllSubmatchIndex(line, -1)
	if midx == nil {
		if len(run.CLICustomRE) > 0 {
			return shost.bust_custom_re(line)
		}
		_die("no match of regexp")
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
	shost.CountStat.ProcessCount[process]++

	return offset[1]
}

func (shost *SourceHost) bust_warning(line []byte, midx []int) int {
	shost.CountStat.WarningCount++
	return 0
}

func (shost *SourceHost) bust_statistics(line []byte, midx []int) int {
	shost.CountStat.StatisticsCount++
	return 0
}

func (shost *SourceHost) bust_fatal(line []byte, midx []int) int {
	shost.CountStat.FatalCount++
	return 0
}

func (shost *SourceHost) bust_daemon_started(line []byte, midx []int) int {
	shost.CountStat.DaemonStartedCount++
	return 0
}

func (shost *SourceHost) bust_refresh_postfix(line []byte, midx []int) int {
	shost.CountStat.RefreshPostfixCount++
	return 0
}

func (shost *SourceHost) bust_reload(line []byte, midx []int) int {
	shost.CountStat.ReloadCount++
	return 0
}

func (shost *SourceHost) bust_connect_from(line []byte, midx []int) int {
	shost.CountStat.ConnectFromCount++
	return 0
}

func (shost *SourceHost) bust_lost_connect(line []byte, midx []int) int {
	shost.CountStat.LostConnectCount++
	return 0
}

func (shost *SourceHost) bust_disconnect_from(line []byte, midx []int) int {
	shost.CountStat.DisconnectFromCount++
	return 0
}

func (shost *SourceHost) bust_connect_to(line []byte, midx []int) int {
	shost.CountStat.ConnectToCount++
	return 0
}

func (shost *SourceHost) bust_backwards_compat(line []byte, midx []int) int {
	shost.CountStat.BackwardsCompatCount++
	return 0
}

func (shost *SourceHost) bust_message_repeated(line []byte, midx []int) int {
	shost.CountStat.MessageRepeatedCount++
	return 0
}

func (shost *SourceHost) bust_start_postfix(line []byte, midx []int) int {
	shost.CountStat.StartPostfixCount++
	return 0
}

//  bust exception parsing queue id

func (shost *SourceHost) bust_queue_ex(line []byte) int {

	midx := warning_re.FindSubmatchIndex(line)
	if midx != nil {
		return shost.bust_warning(line, midx)
	}

	midx = statistics_re.FindSubmatchIndex(line)
	if midx != nil {
		return shost.bust_statistics(line, midx)
	}

	midx = fatal_re.FindSubmatchIndex(line)
	if midx != nil {
		return shost.bust_fatal(line, midx)
	}

	midx = daemon_started_re.FindSubmatchIndex(line)
	if midx != nil {
		return shost.bust_daemon_started(line, midx)
	}

	midx = refresh_postfix_re.FindSubmatchIndex(line)
	if midx != nil {
		return shost.bust_refresh_postfix(line, midx)
	}

	midx = reload_re.FindSubmatchIndex(line)
	if midx != nil {
		return shost.bust_reload(line, midx)
	}

	midx = connect_from_re.FindSubmatchIndex(line)
	if midx != nil {
		return shost.bust_connect_from(line, midx)
	}

	midx = lost_connect_re.FindSubmatchIndex(line)
	if midx != nil {
		return shost.bust_lost_connect(line, midx)
	}

	midx = disconnect_from_re.FindSubmatchIndex(line)
	if midx != nil {
		return shost.bust_disconnect_from(line, midx)
	}

	midx = connect_to_re.FindSubmatchIndex(line)
	if midx != nil {
		return shost.bust_connect_to(line, midx)
	}

	midx = backwards_compat_re.FindSubmatchIndex(line)
	if midx != nil {
		return shost.bust_backwards_compat(line, midx)
	}

	midx = message_repeated_re.FindSubmatchIndex(line)
	if midx != nil {
		return shost.bust_message_repeated(line, midx)
	}

	midx = start_postfix_re.FindSubmatchIndex(line)
	if midx != nil {
		return shost.bust_start_postfix(line, midx)
	}

	die("bust_queue_ex: can not match line %d", shost.run.LineCount)

	shost.CountStat.UnknownLineCount++
	return 0
}

func (shost *SourceHost) bust_queue_id(line []byte) int {

	_die := func(format string, args ...interface{}) {
		die("%s", fmt.Sprintf(
				"bust_queue_id: line %d: %s",
				shost.run.LineCount,
				fmt.Sprintf(format, args...),
		))
	}

	midx := queue_id_re.FindAllSubmatchIndex(line, -1)
	if midx == nil {
		return shost.bust_queue_ex(line)
	}
	var l int

	if l := len(midx);  l != 1 {
		_die("unexpected length of match idx: got %d, want 1", l)
	}

	//  parse the '[A-Z0-9]{10,12}: ' after the process[pid]

	offset := midx[0]
	if l = len(offset);  l != 4 {
		_die("unexpected len of match offset: got %d, want 4", l)
	}

	/*
	queue_id := string(line[offset[2]:offset[3]])	// matches queueid
	run.QueueId[queue_id]++
	*/
	return offset[1]
}

func a2die(opt string) {
	die("option given twice: --%s", opt)
}

func axdie(opt string) {
	die("no required option: --%s", opt)
}

func noarg(opt, what string) {
	die("option missing arg: --%s: %s", opt, what)
}

func (run *Run) push_custom_re(tag_re string) {
	_die := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		die(`--custom-re: %s`, msg)
	}

	bytes := []byte(tag_re)

	offset := arg_custom_re.FindSubmatchIndex(bytes)
	if offset == nil {
		_die("no match tag:regexp: %s", tag_re)
	}
	l := len(offset)
	if l != 6 {
		_die("unexpected length of re offsets: got %d, expected 6", l)
	}
	tag := string(bytes[offset[2]:offset[3]])
	if _, ok := run.CLICustomRE[tag];  ok == true {
		_die("tag already exists: %s", tag)
	}

	re := string(bytes[offset[4]:offset[5]])
	regexp, err := regexp.Compile(re)
	if err != nil {
		_die("can not compile regexp: %s: %s", err, re)
	}
	run.CLICustomRE[tag] = &CustomRE{
		Tag:		tag,
		RegExp:		re,
		regexp:		regexp,
	}
}

func main() {

	argc := len(os.Args) - 1
	if argc < 4  {
		die("wrong number of cli args: got %d, expected >= 4", argc)
	}
	if argc % 2 == 1 {
		die("--option missing second argument")
	}

	run := &Run{
		CLICustomRE: make(map[string]*CustomRE),
		SourceHost: make(map[string]*SourceHost),
	}
	for i := 1;  i <= argc;  i++  {
		arg := os.Args[i]
		i++
		if arg == "--year" {
			if i > argc {
				noarg("year", "missing year")
			}
			if run.Year > 0 {
				a2die("year")
			}
			u, err := strconv.ParseUint(os.Args[i], 10, 12)
			if err != nil {
				fdie("strconv.ParseUint(time)", err)
			}
			run.Year = uint16(u)
		} else if arg == "--time-location" {
			if i > argc {
				noarg("time-location", "missing time zone")
			}
			if run.TimeLocation != "" {
				a2die("time-location")
			}
			run.TimeLocation = os.Args[i]
			loc, err := time.LoadLocation(run.TimeLocation)
			if err != nil {
				fdie("time.LoadLocation(--time-location)", err)
			}
			run.time_location = loc
		} else if arg == "--custom-re" {
			if i > argc {
				noarg("custom-re", "missing tag:regexg")
			}
			run.push_custom_re(os.Args[i])
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
		l := len(bytes)
		if l == 0 {
			panic("impossible read of empty line")
		}
		run.ByteCount += int64(l)
		h512.Write(bytes)			//  digest input
		run.LineCount++
		
		//  zap terminating newline for $ in regex matches
		if bytes[l - 1] != '\n' {
			panic("line not terminated by newline")
		}
		bytes[l - 1] = 0

		i := run.bust_log_time(bytes)

		var shost *SourceHost
		bytes = bytes[i:]
		i, shost = run.bust_source_host(bytes)

		bytes = bytes[i:]
		i = shost.bust_process(bytes)
		if i > -1 {
			bytes = bytes[i:]
			shost.bust_queue_id(bytes)
		}
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
