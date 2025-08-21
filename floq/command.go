package main

import (
	"os/exec"
	"syscall"
	"sync"
	"time"
)

var osx_wg	sync.WaitGroup

type command struct {
	
	name		string
	cmd		*exec.Cmd
	path		string
	look_path	string
	args		[]string
	env		[]string
}

//  result of waiting on an executing command
type osx_value struct {
	*command
	argv		[]string
	err		error
	exit_status	int
	pid		int
	start_time	time.Time
	wall_duration	time.Duration
	user_sec	int64
	user_usec	int32
	sys_sec		int64
	sys_usec	int32
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

//  exec an os command process

func (flo *flow) osx_run(cmd *command, argv []string, out osx_chan) {
	cx := exec.Command(
			cmd.look_path,
	)
	cx.Args = cmd.args
	cx.Args = append(cx.Args, argv...)
	cx.Env = cmd.env

	val := &osx_value{
			command:	cmd,
			start_time:	time.Now(),
	}

	val.err = cx.Run()
	if out == nil {
		return
	}
	if val.err == nil {		//  process failed to start
		ru := cx.ProcessState.SysUsage().(*syscall.Rusage)

		val.pid = cx.Process.Pid
		val.exit_status = cx.ProcessState.ExitCode()
		val.user_sec = ru.Utime.Sec
		val.user_usec = ru.Utime.Usec
		val.sys_sec = ru.Stime.Sec
		val.sys_usec = ru.Stime.Usec
		val.wall_duration = time.Since(val.start_time)
	}
	out <- val
}

//  run a process with no argv

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

//  run a process with argv

func (flo *flow) osx(cmd *command, in argv_chan) (out osx_chan) {

	out = make(osx_chan)

	go func() {
		for {
			av := <-in
			if av == nil {
				return
			}
			if av.is_null == false {
				flo.osx_run(cmd, av.argv, out)
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

	go func() {
		for {
			bv := <- when
			if bv.bool {
				flo.osx_run(cmd, nil, out)
			}
			osx_wg.Done()
			flo = flo.get()
		}
	}()

	return out
}

//  conditionally run a command process with argv

func (flo *flow) osxw(
	cmd *command,
	args argv_chan,
	when bool_chan,
) (out osx_chan) {

	out = make(osx_chan)

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
			if bv.bool == true && av.is_null == false {
				flo.osx_run(cmd, av.argv, out)
			}
			osx_wg.Done()
			flo = flo.get()
		}
	}()

	return out
}

//  read strings from multiple input channels and write assmbled argv[]
//  any null value renders the whole argv[] null

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
			return
		}()

		for {

			av := make([]string, argc)
			ac := uint8(0)
			is_null := false

			//  read until we have an argv[] for which all elements
			//  are also non-null.  any null argv[] element makes
			//  the whole argv[] null

			for ac < argc {

				a := <-merge

				//  Note: compile generates error for
				//        arg_value{}

				if a == (arg_value{}) {
					return
				}

				sv := a.string_value
				pos := a.position

				//  any null element forces entire argv[]
				//  to be null.  not sure this is reasonable.

				if a.is_null {
					is_null = true
				}

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
