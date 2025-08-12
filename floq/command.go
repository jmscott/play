package main

import (
	"os/exec"
	"syscall"
	"time"
)

type command struct {
	
	name	string
	cmd	*exec.Cmd
	path	string
	args	[]string
	env	[]string
}

//  result of waiting on an executing command
type osx_value struct {
	*command
	argv		[]string
	err		error
	pid		int
	start_time	time.Time
	wall_duration	time.Duration
	user_sec	int64
	user_usec	int32
	sys_sec		int64
	sys_usec	int32
}

type osx_chan chan *osx_value

//  exec an os command process and write description of run

func (flo *flow) osx_run(cmd *command, out osx_chan) {
	cx := exec.Command(
			cmd.path,
	)
	cx.Args = cmd.args
	cx.Env = cmd.env

	st := time.Now()
	err := cx.Run() 
	ru := cx.ProcessState.SysUsage().(*syscall.Rusage)
	out <- &osx_value{
			command:	cmd,
			err:		err,	
			pid:		cx.Process.Pid,	
			user_sec:	ru.Utime.Sec,
			user_usec:	ru.Utime.Usec,
			sys_sec:	ru.Stime.Sec,
			sys_usec:	ru.Stime.Usec,
			start_time:	st,
			wall_duration:	time.Since(st),
	}
}

//  unconditionally run a process with no argv
func (flo *flow) osx0(cmd *command) (out osx_chan) {

	out = make(osx_chan)

	go func() {
		for {
			flo = flo.get()

			flo.osx_run(cmd, out)
		}
	}()

	return out
}

//  conditionally run a command process with no arguments
func (flo *flow) osx0when(cmd *command, when bool_chan) (out osx_chan) {

	out = make(osx_chan)

	go func() {

		for {
			flo = flo.get()

			bv := <- when
			if bv.bool {
				flo.osx_run(cmd, out)
			}
		}
	}()

	return out
}
