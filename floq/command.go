package main

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"strconv"
	"syscall"
	"time"
)

type command struct {
	
	name		string
	cmd		*exec.Cmd
	path		string
	
	//  the resolved executable pa
	look_path	string

	//  static array of strings, prepended before dynamic argv[] in call
	args		[]string

	//  array of env vars defined in "env" attribute of set
	env		[]string

	//  references to system "$<attribute>"
	sref_count	uint8

	//  references to defined ".<attribuet>"
	ref_count	uint8

	//  possible tuple bound to command
	tuple_ref	*tuple
}

//  result of waiting on an executing command
type osx_value struct {
	*command
	argv		[]string
	err		error
	exit_code	int
	pid		int
	start_time	time.Time
	wall_duration	time.Duration
	user_sec	int64
	user_usec	int32
	sys_sec		int64
	sys_usec	int32
	Stdout		string
	Stderr		string

	is_null		bool
}

type osx_chan chan *osx_value

//  argv_value represents a function or query string argument vector
type argv_value struct {
	argv    []string
	is_null bool
}

//  argv_chan is channel of *argv_values;  nil indicates closure
type argv_chan chan *argv_value

/*
 *  exec an os command process and build osx_value from process Stdout, Stderr,
 *  pid, and several rusage fields.  
 *
 *  Note:
 *	
 *	signals not handled correctly!
 */
func (flo *flow) osx_run(cmd *command, argv []string, out osx_chan) {

	cx := exec.Command(
			cmd.path,
	)
	cx.Args = cmd.args
	cx.Args = append(cx.Args, argv...)

	cx.Env = cmd.env

	var stdout, stderr strings.Builder

	cx.Stdout = &stdout
	cx.Stderr = &stderr

	start_time := time.Now()

	err := cx.Run()

	wall_duration := time.Since(start_time)

	/*
	 *  golang exec considers any non-zero exit_code to be the error.
	 *  "exit status <code>".  determine if error is real error.  
	 */
	if err != nil {
		if strings.HasPrefix(err.Error(), "exit status ") == false {
			die("Run(%s) failed: %s", cmd.name, err)
		}
	}
	if out == nil {		//  caller does not want osx_value
		return
	}

	val := &osx_value{
			command:	cmd,
			start_time:	start_time,
			wall_duration:	wall_duration,
			pid:		cx.Process.Pid,
		}

	ps := cx.ProcessState
	if ps == nil {
		die("process state is null: %s", cmd.name)
	}

	val.exit_code = ps.ExitCode()

	ru := ps.SysUsage().(*syscall.Rusage)
	val.user_sec = ru.Utime.Sec
	val.user_usec = ru.Utime.Usec
	val.sys_sec = ru.Stime.Sec
	val.sys_usec = ru.Stime.Usec
	val.Stdout = stdout.String()
	val.Stderr = stderr.String()

	out <- val
}

//  run a process with neither argv[] nor "when" predicate

func (flo *flow) osx_run_0(cmd *command) (out osx_chan) {

	out = make(osx_chan)

	go func() {
		<-compiling

		for {
			flo.osx_run(cmd, nil, out)

			flo = flo.next()
		}
	}()

	return out
}

//  run a process with an argv and no "when" predicate

func (flo *flow) osx_run_a(cmd *command, in argv_chan) (out osx_chan) {

	out = make(osx_chan)

	go func() {
		<-compiling

		for {
			flo.osx_run(cmd, (<-in).argv, out)

			flo = flo.next()
		}
	}()

	return out
}

//  conditionally run a command process with no argv

func (flo *flow) osx_run_w(cmd *command, when bool_chan) (out osx_chan) {

	out = make(osx_chan)

	null_osx := &osx_value{
			is_null:	true,
			command:	cmd,
	}
	go func() {
		<-compiling

		for {
			bv := <- when
			if bv.bool {
				flo.osx_run(cmd, nil, out)
			} else {
				out <- null_osx
			}

			flo = flo.next()
		}
	}()

	return out
}

//  run a command process with argv and "when" predicate

func (flo *flow) osx_run_aw(
	cmd *command,
	args argv_chan,
	when bool_chan,
) (out osx_chan) {


	out = make(osx_chan)

	//  null_osx sent when command not run

	null_osx := &osx_value{
			is_null: true,
			command: cmd,
		    }

	go func() {

		<-compiling

		for {
			var bv *bool_value
			var av *argv_value

			//  wait for both "when" clause (bv) and "argv (av)"
			//  to finish.  also do cheap sanity tests

			for bv == nil || av == nil {
				select {

				//  wait for "when" expression resolve

				case b := <-when:
					//  cheap sanity test
					if bv != nil {
						die("bv seen twice")
					}
					bv = b

				//  wait for "argv" expression resolve

				case a := <-args:

					//  cheap sanity test
					if av != nil {
						die("av seen twice")
					}
					av = a
				}
			}

			//  "when" clause is true, so run command.
			//  osx_run send value

			if bv.bool == true {
				flo.osx_run(cmd, av.argv, out)
			} else {
				out <- null_osx
			}

			flo = flo.next()
		}
	}()

	return out
}

//  Read strings from multiple input channels to assemble an []string
//  argv to pass to exec via ("run <command>(argv)".

func (flo *flow) argv(in_args []string_chan) (out argv_chan) {

	out = make(argv_chan)
	argc := len(in_args)

	//  called RUN has arguments, so wait on args via string channels
	//  before sending assembled argv[]

	go func() {
		<-compiling

		for {
			argv := make([]string, argc)

			//  wait for string arguments to arrive

			var wg sync.WaitGroup
			wg.Add(int(argc))
			for i := 0;  i < argc;  i++ {
				go func(int) {
					argv[i] = (<- in_args[i]).string
					wg.Done()
				}(i)
			}
			wg.Wait()

			out <- &argv_value{
				argv:    argv,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  drain an osx record, for when "run <command>" has no target
func (flo *flow) osx_null(in osx_chan) {

	go func() {
		<-compiling

		for {
			<- in

			flo = flo.next()
		}
	}()
}

func (cmd *command) String() string {
	return cmd.name
}

/*
 *  project a particular attribute of a tab separated, new line terminated
 *  set of tuples as osx tuples
 */
func (flo *flow) osx_proj_tsv(
	in osx_chan,
	cmd *command,
	att *attribute,
  ) (out string_chan) {

	out = make(string_chan)
	tab_field := att.tab_field-1

	_die := func (format string, args ...interface{}) {
		fmt := fmt.Sprintf(
				"%s: %s#%d: xv.Stdout: ",
				cmd,
				att,
				att.tab_field,
			)
		die(fmt + format, args...)
	}

	go func() {
		<-compiling

		for {
			xv := <- in

			var str string

			if !xv.is_null {
				str = strings.TrimRight(xv.Stdout, "\n")
				if str == "" {
					_die("empty string")
				}

				if strings.Count(str, "\n") > 0 {
					_die("more than one newline")
				}

				//  Note: is panic correct action?
				//        maybe send null.

				fld := strings.Split(str, "\t")
				tupa := att.tuple_ref.atts
				if len(fld) != len(tupa) {
					_die("not %d fields", len(tupa))
				}

				str = fld[tab_field]
				if att.matches.MatchString(str) == false {
					_die(
						"matches fails: %s !~ %s",
						att.matches.String(),
						str,
					)
				}
			}

			out <- &string_value{
				string:		str,
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()
	return out
}
//  project the command$exit_code from an osx_record

func (flo *flow) osx_proj_exit_code(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		<-compiling

		for {
			xv := <- in

			out <- &uint64_value{
				uint64:		uint64(xv.exit_code),
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  project the command$Stdout from an osx_record

func (flo *flow) osx_proj_Stdout(in osx_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		<-compiling

		for {
			xv := <- in

			out <- &string_value{
				string:		xv.Stdout,
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  project the command$Stdout from an osx_record

func (flo *flow) osx_proj_Stderr(in osx_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		<-compiling

		for {
			xv := <- in

			out <- &string_value{
				string:		xv.Stderr,
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  project the command$pid from an osx_record

func (flo *flow) osx_proj_pid(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		<-compiling

		for {
			xv := <- in

			out <- &uint64_value{
				uint64:		uint64(xv.pid),
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

func (flo *flow) osx_proj_start_time(in osx_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		<-compiling

		for {
			xv := <- in

			out <- &string_value{
				string:		xv.start_time.Format(
							time.RFC3339Nano,
						),
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  project the command$wall_duration from an osx_record

func (flo *flow) osx_proj_wall_duration(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		<-compiling

		for {
			xv := <- in

			out <- &uint64_value{
				uint64:		uint64(xv.wall_duration),
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  project osx command$wall_duration_seconds as sec.msec

func (flo *flow) osx_proj_wall_duration_seconds(in osx_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		<-compiling

		for {
			xv := <- in

			//  format as <sec>.<msec>

			wd := strconv.FormatFloat(
					xv.wall_duration.Seconds(),
					'f',
					-1,
					64,
			)
			out <- &string_value{
				string:		wd,
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  project the command$user_sec from an osx_record

func (flo *flow) osx_proj_user_sec(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		<-compiling

		for {
			xv := <- in

			out <- &uint64_value{
				uint64:		uint64(xv.user_sec),
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  project the command$user_usec from an osx_record

func (flo *flow) osx_proj_user_usec(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		<-compiling

		for {
			xv := <- in

			out <- &uint64_value{
				uint64:		uint64(xv.user_usec),
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  project the command$user_seconds, derived from an osx_record

func (flo *flow) osx_proj_user_seconds(in osx_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		<-compiling

		for {
			xv := <- in

			sec := strconv.FormatFloat(
				float64(xv.user_sec) +
					float64(xv.user_usec)/1000000.0,
				'f',
				-1,
				64,
			)

			out <- &string_value{
				string:		sec,
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  project the command$sys_usec from an osx_record

func (flo *flow) osx_proj_sys_usec(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		<-compiling

		for {
			xv := <- in

			out <- &uint64_value{
				uint64:		uint64(xv.sys_usec),
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  project the command$sys_sec from an osx_record

func (flo *flow) osx_proj_sys_sec(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		<-compiling

		for {
			xv := <- in

			out <- &uint64_value{
				uint64:		uint64(xv.sys_sec),
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  project the command$sys_seconds, derived from an osx.{sys_sec, sys_usec}

func (flo *flow) osx_proj_sys_seconds(in osx_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		<-compiling

		for {
			xv := <- in

			sec := strconv.FormatFloat(
				float64(xv.sys_sec) +
					float64(xv.sys_usec)/1000000.0,
				'f',
				-1,
				64,
			)

			out <- &string_value{
				string:		sec,
				is_null:	xv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  fanout an osx_value to listeners

func (flo *flow) osx_fo(in osx_chan, count uint8) (out []osx_chan) {

	out = make([]osx_chan, count)
	for i := uint8(0); i < count; i++ {
		out[i] = make(osx_chan)
	}

	go func() {
		<-compiling

		for {
			xv := <-in

			//  broadcast to channels in output slice

			var wg sync.WaitGroup

			wg.Add(int(count))
			for _, xc := range out {
				go func() {
					xc <- xv
					wg.Done()
				}()
			}
			wg.Wait()

			flo = flo.next()
		}
	}()
	return out
}

func (cmd *command) detail(indent int) string {

	if cmd == nil {
		return "nil command"
	}
	tab := strings.Repeat("\t", indent)
	var tn string
	if cmd.tuple_ref == nil {
		tn = "<nil>"
	}
	return fmt.Sprintf(`{
%s      name: %s
%s     tuple: %s@%p
%s      path: %s
%s      args: %s
%s look_path: %s
%s ref_count: %d
%ssref_count: %d
%s       env: %s
%s         @: %p
%s}`,		
		tab, cmd.name,
		tab, tn, cmd.tuple_ref,
		tab, cmd.path,
		tab, cmd.look_path,
		tab, strings.Join(cmd.args, ", "),
		tab, cmd.ref_count,
		tab, cmd.sref_count,
		tab, strings.Join(cmd.env, ", "),
		tab, cmd,
		strings.Repeat("\t", indent),
	)
}
