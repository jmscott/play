package main

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

//  wait for all processes to stop then floq exits.
//
//  Note: why var osx_wf global?  why not in flow struct?

var osx_wg	sync.WaitGroup

type command struct {
	
	name		string
	cmd		*exec.Cmd
	path		string
	look_path	string
	args		[]string
	env		[]string
	ref_count	uint8
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
	output		string

	is_null		bool
}

type osx_chan chan *osx_value

//  argv_value represents a function or query string argument vector
type argv_value struct {
	argv    []string
	is_null bool

	*flow
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
			cmd.look_path,
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
	if out == nil {		//  caller does not want osx_val7ue
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
	val.wall_duration = time.Since(val.start_time)

	out <- val
}

//  run a process with no argv nor "when" predicate

func (flo *flow) osx0(cmd *command) (out osx_chan) {

	out = make(osx_chan)

	go func() {
		for {
			flo.osx_run(cmd, nil, out)
			osx_wg.Done()
			flo = flo.get()
		}
	}()

	return out
}

//  run a process with an argv and no "when" predicate

func (flo *flow) osx(cmd *command, in argv_chan) (out osx_chan) {

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
			osx_wg.Done()

			flo = flo.get()
		}
	}()

	return out
}


//  conditionally run a command process with no argv

func (flo *flow) osx0w(cmd *command, when bool_chan) (out osx_chan) {

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
			osx_wg.Done()

			flo = flo.get()
		}
	}()

	return out
}

//  run a command process with argv and "when" predicate

func (flo *flow) osxw(
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
			osx_wg.Done()
			flo = flo.get()
		}
	}()

	return out
}

//  read strings from multiple input channels and write assembled argv[]

func (flo *flow) argv(in_args []string_chan) (out argv_chan) {

	//  track a received string and position in argv[]
	type arg_value struct {
		*string_value
		position uint8
	}

	out = make(argv_chan)
	argc := uint8(len(in_args))

	//  called RUN has arguments, so wait on args via string channels
	//  before sending assembled argv[]

	go func() {

		defer close(out)

		//  merge() output of string channels onto a single channel of
		//  []string values.

		merge := func() (mout chan arg_value) {

			var wg sync.WaitGroup
			mout = make(chan arg_value)

			io := func(sc string_chan, p uint8) {
				for sv := range sc {
					mout <- arg_value{
						string_value: sv,
						position:     p,
					}
				}
				wg.Done()
			}

			wg.Add(len(in_args))
			for i, sc := range in_args {
				go io(sc, uint8(i))
			}

			//  Start a goroutine to close 'mout' channel
			//  once all the output goroutines are done.

			go func() {
				wg.Wait()
				close(mout)
			}()
			return mout
		}()

		for {

			av := make([]string, argc)
			ac := uint8(0)
			is_null := false

			//  read until we have an argv[] for which all elements
			//  are also non-null.  any null argv[] element makes
			//  the whole argv[] null

			for ac < argc {

				arg := <-merge

				//  merge channel closed
				//
				//  Note: golang compile generates error for
				//        arg_value{} without parens

				if arg == (arg_value{}) {
					return
				}

				sv := arg.string_value
				pos := arg.position

				//  any null element forces entire argv[]
				//  to be null.  not sure this is reasonable.
				//
				//  Note: why, seems wrong!

				sv.is_null = arg.is_null

				//  cheap sanity test to insure we don't
				//  see the same argument twice

				if av[pos] != "" {
					croak("argv[%d] element not \"\"", pos)
				}
				av[pos] = sv.string
				ac++
			}

			out <- &argv_value{
				argv:    av,
				is_null: is_null,
				flow:    flo,
			}
			
			flo = flo.get()
		}
	}()

	return out
}

func (flo *flow) osx_null(in osx_chan) {

	go func() {
		<- in

		flo = flo.get()
	}()
}

func (cmd *command) is_sysatt(name string) bool {

	switch name {
	case "exit_code",
		"pid",
		"start_time",
		"wall_duration",
		"user_sec",
		"user_usec",
		"sys_sec",
		"sys_usec",
		"StdoutPipe":
		return true
	}
	return false
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

//  project the command$exit_code from an osx_record

func (flo *flow) osx_proj_exit_code(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		xv := <- in
		if xv == nil {
			return
		}

		out <- &uint64_value{
			uint64:		uint64(xv.exit_code),
			is_null:	xv.is_null,
		}

		flo = flo.get()
	}()

	return out
}

//  project the command$Output from an osx_record

func (flo *flow) osx_proj_output(in osx_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		xv := <- in
		if xv == nil {
			return
		}

		out <- &string_value{
			string:		xv.output,
			is_null:	xv.is_null,
		}

		flo = flo.get()
	}()

	return out
}

//  project the command$pid from an osx_record

func (flo *flow) osx_proj_pid(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		xv := <- in
		if xv == nil {
			return
		}

		out <- &uint64_value{
			uint64:		uint64(xv.pid),
			is_null:	xv.is_null,
		}

		flo = flo.get()
	}()

	return out
}

func (flo *flow) osx_proj_start_time(in osx_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
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
	}()

	return out
}

//  project the command$wall_duration from an osx_record

func (flo *flow) osx_proj_wall_duration(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		xv := <- in
		if xv == nil {
			return
		}

		out <- &uint64_value{
			uint64:		uint64(xv.wall_duration),
			is_null:	xv.is_null,
		}

		flo = flo.get()
	}()

	return out
}

//  project the command$user_sec from an osx_record

func (flo *flow) osx_proj_user_sec(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		xv := <- in
		if xv == nil {
			return
		}

		out <- &uint64_value{
			uint64:		uint64(xv.user_sec),
			is_null:	xv.is_null,
		}

		flo = flo.get()
	}()

	return out
}

//  project the command$user_usec from an osx_record

func (flo *flow) osx_proj_user_usec(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		xv := <- in
		if xv == nil {
			return
		}

		out <- &uint64_value{
			uint64:		uint64(xv.user_usec),
			is_null:	xv.is_null,
		}

		flo = flo.get()
	}()

	return out
}

//  project the command$sys_usec from an osx_record

func (flo *flow) osx_proj_sys_usec(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		xv := <- in
		if xv == nil {
			return
		}

		out <- &uint64_value{
			uint64:		uint64(xv.sys_usec),
			is_null:	xv.is_null,
		}

		flo = flo.get()
	}()

	return out
}

//  project the command$sys_sec from an osx_record

func (flo *flow) osx_proj_sys_sec(in osx_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {
		xv := <- in
		if xv == nil {
			return
		}

		out <- &uint64_value{
			uint64:		uint64(xv.sys_sec),
			is_null:	xv.is_null,
		}

		flo = flo.get()
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
func (cmd *command) string(indent int) string {

	if cmd == nil {
		return "nil command"
	}
	tab := strings.Repeat("\t", indent)
	return fmt.Sprintf(`%s: {
%s      path: %s
%s look_path: %s
%s      args: %s
%s       env: %s
%s ref_count: %d
%s         @: %p
%s}`,		
		cmd.name,
		tab, cmd.path,
		tab, cmd.look_path,
		tab, strings.Join(cmd.args, ", "),
		tab, strings.Join(cmd.env, ", "),
		tab, cmd.ref_count,
		tab, cmd,
		strings.Repeat("\t", indent),
	)
}
