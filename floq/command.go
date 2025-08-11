package main

import (
	"os/exec"
)

type command struct {
	
	name	string
	cmd	*exec.Cmd
	path	string
	args	[]string
	env	[]string
}

/*

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

//  unconditionally exec an os command process with no dynamoic
//  arguments

func (flo *flow) osx(cmd *command, out osx_chan) {
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

func (flo *flow) osx0(cmd *command) (out osx_chan) {

	out = make(osx_chan)
	//  check existence of executable in "path" variable

	go func() {
		for {
			flo = flo.get()

			flo.osx(cmd, out)
		}
	}()

	return out
}
//  exec an os command process with no arguments
func (flo *flow) osx0when(cmd *command, when bool_chan) (out osx_chan) {

	out = make(osx_chan)

	go func() {

		for {
			flo = flo.get()

			bv := <- when
			if bv.bool {
				flo.osx(cmd, out)
			}
		}
	}()

	return out
}

func (cmd *command) frisk_att(atup *ast) (err error) {

	err = atup.frisk_att(
		"command." + cmd.name,
		ast{
			string:		"path",
			yy_tok:		STRING,
			uint64:		1,
		},
		ast{
			string:		"argv",
			yy_tok:		ATT_ARRAY,
			uint64:		0,
		},
		ast{
			string:		"env",
			yy_tok:		ATT_ARRAY,
			uint64:		0,
		},
		ast{
			string:		"dir",
			yy_tok:		STRING,
			uint64:		0,
		},
	)
	if err != nil {
		return err
	}
	ap := atup.find_ATT("path")
	if ap == nil {
		atup.corrupt("can not find ATT 'path'")
	}
	if ap.right == nil {
		ap.corrupt("path ATT has not right")
	}

	env := os.Environ()
	for _, e := range cmd.env {
		env = append(env, e)
	}
	cmd.env = env

	cmd.path, err = exec.LookPath(ap.right.string)
	return err
}
*/
