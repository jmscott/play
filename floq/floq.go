package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var floq_trace_next_op bool
var floq_trace_compile bool

var usage = "usage: floq [pass1|pass2|compile|frisk|server] path/to/prog.floq\n"

var caught_sig os.Signal

func exit(status int) {
	os.Exit(status)
}

//  die() during boot up, before cli args parsed, etc

func croak(format string, args ...interface{}) {

	fmt.Fprintf(
		os.Stderr,
		"floq: ERROR: %s\n",
		fmt.Sprintf(format, args...),
	)
	fmt.Fprintf(os.Stderr, usage)
	exit(16)
}

func rcaller(frame int) string {

	pc, _, _, ok := runtime.Caller(frame)
	if !ok {
		die("runtime.Caller(%d) failed", frame)
	}
	cn := runtime.FuncForPC(pc).Name()
	if strings.HasPrefix(cn, "main.(*flow).") == true {
		cn = cn[13:]
	}
	return cn
}

//  write stack trace of all running goroutines into file floq.trace

func stacktrace() {
	buf := make([]byte, 1<<20)
	len := runtime.Stack(buf, true)

	fmt.Fprintf(os.Stderr, "\nstack trace in floq.trace\n")
	os.WriteFile(
		"floq.trace",
		[]byte(fmt.Sprintf("\n=== Stack Trace ===\n%s\n", buf[:len])),
		0640,
	)
}

func env_bool(evar string) bool {

env, exists := os.LookupEnv(evar)
if exists {
	b, err := strconv.ParseBool(env)
	if err != nil {
		croak("can not parse env var: %s", os.Getenv(env))
	}
	return b 
}
return false
}

//  set yacc yyDebug level for dumping parsing details into y.output

func env_yydebug() {

	yyd, exists := os.LookupEnv("FLOQ_YYDEBUG")
	if !exists {
		return
	}

	var err error

	i, err := strconv.ParseUint(yyd, 10, 8)
	if err != nil {
		croak("can not parse env FLOQ_YYDEBUG: %s", yyd)
	}
	yyDebug = int(i)
	if yyDebug > 0 {
		yyErrorVerbose = true
	}
}

func main() {

	argv := os.Args[1:]
	argc := len(argv)

	if argc >= 1 {
		switch argv[0] {
		case "help", "--help":
			argc--
			help(argc, argv[1:])
		case "build", "--build":
			os.Stdout.WriteString(build + "\n")
			os.Exit(0)
		}
	}

	if argc != 2 {
		croak("wrong number of cli args: expected 2, got %d", argc)
	}
	action := argv[0]

	switch action {
		case "frisk",
		     "pass1",
		     "pass2",
		     "compile",
		     "server":
		default:
			croak("unknown action: %s", action)
	}

	//  set up tracing environment variables

	floq_trace_compile = env_bool("FLOQ_TRACE_COMPILE")
	floq_trace_next_op = env_bool("FLOQ_TRACE_NEXT_OP")
	env_yydebug()
	
	//  open and parse the floq file

	floq_path := argv[1]
	floq, err := os.OpenFile(floq_path, os.O_RDONLY, 0)
	if err != nil {
		if os.IsNotExist(err) {
			croak("file does not exist: %s", floq_path)
		}
		croak("OpenFile(%s) failed: %s", floq_path, err)
	}

	//  parse the floq file into a flow of goops

	root, err := parse(bufio.NewReader(floq))
	if err != nil {
		croak("parse(%s) failed: %s", floq_path, err)
	}
	floq.Close()

	//  set up signal handlers, doing a stacktrace on SIGQUIT

	go func() {
		c := make(chan os.Signal)
		signal.Notify(
			c,
			syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGINT,
		)
		caught_sig = <-c

		//  dump stack trace in file floq.trace

		if caught_sig == syscall.SIGQUIT {
			stacktrace()
		}
		exit(0)
	}()

	//  setup tracing FLOQ_TRACE_*

	switch action {
	case "pass1":
		root.print()
	case "pass2":
		if err = xpass2(root);  err != nil {
			croak("xpass2(%s) failed: %s", floq_path, err)
		}
		root.print()
	case "frisk":
		if err = xpass2(root);  err != nil {
			croak("frisk: xpass2(%s) failed: %s", floq_path, err)
		}
	case "compile":
		if err := xpass2(root);  err != nil {
			croak("compile/pass2(%s) failed: %s", floq_path, err)
		}
		compile(root)	//  any error is a panic()
	case "server":
		if err := xpass2(root);  err != nil {
			croak("server/pass2(%s) failed: %s", floq_path, err)
		}
		server(root)
		/* NOTREACHED */
	default:
		croak("unknown action: %s", action)
	}
	exit(0)
}

var die_mux sync.Mutex

func die(format string, args ...interface{}) {
	die_mux.Lock()
	os.Stderr.Write([]byte("\nfloq: good bye, cruel world\n"))
	panic(fmt.Sprintf(format, args...))
}

//  truncate a string to slen chars and conditionally append an ellipse

func string_brief(str string, clen int, ellipse bool) string {
	slen := len(str)
	if slen <= clen {
		return str
	}
	str = str[:clen]
	if ellipse == false {
		return str
	}
	return str + "..."
}

func trace(format string, args ...interface{}) {

	var caller string

	//  get name of calling function
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		fname := "unknown"

		f := runtime.FuncForPC(pc)
		if f != nil {
			fname = f.Name()
		}
		caller = fname
		fld := strings.Split(fname, ".")
		flen := len(fld)
		if flen > 1 {
			caller = fld[flen-1]
			switch caller {
			case "func1", "func2", "func3", "func4":
				caller = fld[flen-2] + "." + caller
			}
		} else if flen == 1 {
			caller = fld[0]
		} else {
			caller = fname
		}
		if caller == "1" {		//  nested anonymous goroutine
			caller = fname
		}
	}
	format = caller + ": " + format
	os.Stderr.WriteString(fmt.Sprintf("TRACE: " + format, args...) + "\n")
}
