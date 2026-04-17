package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

var usage = "usage: floq [pass1|pass2|compile|frisk|server] path/to/prog.floq\n"

var caught_sig os.Signal

func exit(status int) {
	os.Exit(status)
}

//  die() during boot up

func croak(format string, args ...interface{}) {

	fmt.Fprintf(
		os.Stderr,
		"floq: ERROR: %s\n",
		fmt.Sprintf(format, args...),
	)
	fmt.Fprintf(os.Stderr, usage)
	exit(16)
}

//  write stack trace of all running goroutines into file floq.trace

func tracedump() {
	buf := make([]byte, 1<<20)
	len := runtime.Stack(buf, true)

	fmt.Fprintf(os.Stderr, "\ntrace in floq.trace\n")
	os.WriteFile(
		"floq.trace",
		[]byte(fmt.Sprintf("\n=== Stack Trace ===\n%s\n", buf[:len])),
		0640,
	)
}

func main() {

	argv := os.Args[1:]
	argc := len(argv)
	if argc != 2 {
		croak("wrong number of arguments: expected 2, got %d", argc)
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
	floq_path := argv[1]

	floq, err := os.OpenFile(floq_path, os.O_RDONLY, 0)
	if err != nil {
		if os.IsNotExist(err) {
			croak("file does not exist: %s", floq_path)
		}
		croak("OpenFile(%s) failed: %s", floq_path, err)
	}
	defer floq.Close()

	root, err := parse(bufio.NewReader(floq))
	if err != nil {
		croak("parse(%s) failed: %s", floq_path, err)
	}

	go func() {
		c := make(chan os.Signal)
		signal.Notify(
			c,
			syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGINT,
		)
		caught_sig = <-c
		if caught_sig == syscall.SIGQUIT {
			tracedump()
		}
		os.Exit(0)
	}()

	switch action {
	case "pass1":
		root.walk_print(0, nil)
	case "pass2":
		if err = xpass2(root);  err != nil {
			croak("xpass2(%s) failed: %s", floq_path, err)
		}
		root.walk_print(0, nil)
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
		if err := server(root);  err != nil {
			croak("server(%s) failed: %s", floq_path, err) 
		}
	default:
		croak("unknown action: %s", action)
	}
	exit(0)
}

func corrupt(format string, args ...interface{}) {
	panic(fmt.Sprintf("corrupt: " + format, args...))
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

//  debug with attitude 

func WTF(format string, args ...interface{}) {

	if format == "" {
		os.Stderr.WriteString("\n")
		return
	}
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
	os.Stderr.WriteString(fmt.Sprintf("WTF: " + format, args...) + "\n")
}
