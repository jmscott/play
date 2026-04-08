package main

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
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
	wall_duration  := time.Since(start_time)

	/*
	 *  golang exec considers any non-zero exit_code to be the error.
	 *  "exit status <code>".  determine if error is real error.  
	 */
	if err != nil {
		if strings.HasPrefix(err.Error(), "exit status ") == false {
			croak("osx_run(%s) failed: %s", cmd.name, err)
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
		croak("osx_run: %s: process state is null", cmd.name)
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

//  run a process with no argv nor "when" predicate

func (flo *flow) osx_run_0(cmd *command) (out osx_chan) {

	out = make(osx_chan)

	go func() {
		for {
			flo.osx_run(cmd, nil, out)
			flo = flo.get()
		}
	}()

	return out
}

//  run a process with an argv and no "when" predicate

func (flo *flow) osx_run_a(cmd *command, in argv_chan) (out osx_chan) {

	out = make(osx_chan)

	null_osx := &osx_value{
			is_null:	true,
			command:	cmd,
	}
	go func() {
		for {
			av := <-in
			if av == nil {
				return
			}

			//  Note: huh.  when is argv[] null?

			if av.is_null == false {
				flo.osx_run(cmd, av.argv, out)
			} else {
				out <- null_osx
			}

			flo = flo.get()
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
		for {
			bv := <- when
			if bv.bool {
				flo.osx_run(cmd, nil, out)
			} else {
				out <- null_osx
			}

			flo = flo.get()
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

	null_osx := &osx_value{
			is_null: true,
			command: cmd,
		    }
	go func() {
		for {
			var bv *bool_value
			var av *argv_value

			//  wait for both argv[] and when clause to finish
			for bv == nil || av == nil {
				select {
				case bv = <-when:
				case av = <-args:
				}
			}

			//  Note:  when is argv null!

			if bv.bool == true && av.is_null == false {
				flo.osx_run(cmd, av.argv, out)
			} else {
				out <- null_osx
			}
			flo = flo.get()
		}
	}()

	return out
}

//  read strings from multiple input channels and write assembled argv[]

func (flo *flow) argv(in_args []string_chan) (out argv_chan) {

	out = make(argv_chan)
	argc := len(in_args)

	//  called RUN has arguments, so wait on args via string channels
	//  before sending assembled argv[]

	go func() {

		defer close(out)

		for {
			var wg sync.WaitGroup
			wg.Add(int(argc))
			
			argv := make([]string, argc)

			for i := 0;  i < argc;  i++ {
				go func(int) {
					//  Note: not handling null!!
					argv[i] = (<- in_args[i]).string
					wg.Done()
				}(i)
			}
			wg.Wait()
			out <- &argv_value{
				argv:    argv,
			}

			flo = flo.get()
		}
	}()

	return out
}

func (flo *flow) osx_null(in osx_chan) {

	go func() {
		for {
			<- in

			flo = flo.get()
		}
	}()
}

func (cmd *command) is_sysatt_uint64(name string) bool {
	switch name {
	case "exit_code", "wall_duration":
		return true
	}
	return false
}

func (cmd *command) String() string {
	return cmd.name
}

/*
 *  project a particular attribute of a tab separated, new line terminated
 *  set of tuples as osx tuples
 */
func (flo *flow) osx_proj_tuple_tsv(
	in osx_chan,
	cmd *command,
	att *attribute,
  ) (out string_chan) {

	out = make(string_chan)
	tsv_field := att.tsv_field-1

	_die := func (format string, args ...interface{}) {
		fmt := fmt.Sprintf(
				"%s: %s#%d: xv.Stdout: ",
				cmd,
				att,
				att.tsv_field,
			)
		corrupt(fmt + format, args...)
	}

	go func() {
		for {
			xv := <- in
			if xv == nil {
				return
			}
			var str string
			
			str = strings.TrimRight(xv.Stdout, "\n")
			if strings.Count(str, "\n") > 0 {
				_die("more than one newline")
			}

			fld := strings.Split(str, "\t")
			if len(fld) != len(att.tuple_ref.atts) {
				_die("not %d fields", len(att.tuple_ref.atts))
			}

			str = fld[tsv_field]
			if att.matches.MatchString(str) == false {
				_die(
					"matches fails: %s !~ %s",
					att.matches.String(),
					str,
				)
			}

			out <- &string_value{
				string:		str,
				is_null:	xv.is_null,
			}

			flo = flo.get()
		}
	}()
	return out
}
/*
 *  project a particular field via offset of a tab separated, new line
 *  terminated set of tuples as osx tuples
 */
func (flo *flow) osx_proj_tuple_tsv_n(
	in osx_chan,
	field uint8,
  ) (out string_chan) {

	out = make(string_chan)

	go func() {
		for {
			xv := <- in
			if xv == nil {
				return
			}

			var str string
			
			is_null := xv.is_null
			if xv.is_null == false {
				fld := strings.Split(
					strings.TrimRight(
						xv.Stdout,
						"\n",
					),
					"\t",
				)
				if int(field) <= len(fld) {
					str = fld[field-1]
					is_null = false
				} else {
					is_null = true
				}
			}

			out <- &string_value{
				string:		str,
				is_null:	is_null,
			}

			flo = flo.get()
		}
	}()
	return out
}

//  project the command$exit_code from an osx_record

func (flo *flow) osx_proj_exit_code(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		for {
			xv := <- in
			if xv == nil {
				return
			}

			out <- &uint64_value{
				uint64:		uint64(xv.exit_code),
				is_null:	xv.is_null,
			}

			flo = flo.get()
		}
	}()

	return out
}

//  project the command$Stdout from an osx_record

func (flo *flow) osx_proj_Stdout(in osx_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		for {
			xv := <- in
			if xv == nil {
				return
			}

			out <- &string_value{
				string:		xv.Stdout,
				is_null:	xv.is_null,
			}

			flo = flo.get()
		}
	}()

	return out
}

//  project the command$Stdout from an osx_record

func (flo *flow) osx_proj_Stderr(in osx_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		for {
			xv := <- in
			if xv == nil {
				return
			}

			out <- &string_value{
				string:		xv.Stderr,
				is_null:	xv.is_null,
			}

			flo = flo.get()
		}
	}()

	return out
}

//  project the command$pid from an osx_record

func (flo *flow) osx_proj_pid(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		for {
			xv := <- in
			if xv == nil {
				return
			}

			out <- &uint64_value{
				uint64:		uint64(xv.pid),
				is_null:	xv.is_null,
			}

			flo = flo.get()
		}
	}()

	return out
}

func (flo *flow) osx_proj_start_time(in osx_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		for {
			xv := <- in
			if xv == nil {
				return
			}

			out <- &string_value{
				string:		xv.start_time.Format(
							time.RFC3339Nano,
						),
				is_null:	xv.is_null,
			}

			flo = flo.get()
		}
	}()

	return out
}

//  project the command$wall_duration from an osx_record

func (flo *flow) osx_proj_wall_duration(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		for {
			xv := <- in
			if xv == nil {
				return
			}

			out <- &uint64_value{
				uint64:		uint64(xv.wall_duration),
				is_null:	xv.is_null,
			}

			flo = flo.get()
		}
	}()

	return out
}

//  project the command$user_sec from an osx_record

func (flo *flow) osx_proj_user_sec(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		for {
			xv := <- in
			if xv == nil {
				return
			}

			out <- &uint64_value{
				uint64:		uint64(xv.user_sec),
				is_null:	xv.is_null,
			}

			flo = flo.get()
		}
	}()

	return out
}

//  project the command$user_usec from an osx_record

func (flo *flow) osx_proj_user_usec(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		for {
			xv := <- in
			if xv == nil {
				return
			}

			out <- &uint64_value{
				uint64:		uint64(xv.user_usec),
				is_null:	xv.is_null,
			}

			flo = flo.get()
		}
	}()

	return out
}

//  project the command$sys_usec from an osx_record

func (flo *flow) osx_proj_sys_usec(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		for {
			xv := <- in
			if xv == nil {
				return
			}

			out <- &uint64_value{
				uint64:		uint64(xv.sys_usec),
				is_null:	xv.is_null,
			}

			flo = flo.get()
		}
	}()

	return out
}

//  project the command$sys_sec from an osx_record

func (flo *flow) osx_proj_sys_sec(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		for {
			xv := <- in
			if xv == nil {
				return
			}

			out <- &uint64_value{
				uint64:		uint64(xv.sys_sec),
				is_null:	xv.is_null,
			}

			flo = flo.get()
		}
	}()

	return out
}

func (flo *flow) osx_fanout(in osx_chan, count uint8) (out []osx_chan) {

	out = make([]osx_chan, count)
	for i := uint8(0); i < count; i++ {
		out[i] = make(osx_chan)
	}

	go func() {

		defer func() {
			for _, a := range out {
				close(a)
			}
		}()

		for {
			xv := <-in
			if xv == nil {
				return
			}

			//  broadcast to channels in output slice

			for _, xc := range out {
				go func() {
					xc <- xv
				}()
			}
			flo = flo.get()
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
%s       env: %s
%s         @: %p
%s}`,		
		tab, cmd.name,
		tab, tn, cmd.tuple_ref,
		tab, cmd.path,
		tab, cmd.look_path,
		tab, strings.Join(cmd.args, ", "),
		tab, strings.Join(cmd.env, ", "),
		tab, cmd,
		strings.Repeat("\t", indent),
	)
}
