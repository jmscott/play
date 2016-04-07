//  execute commands in (unix) file system and wait for output and exit status

package main

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

type command struct {

	//  name in command{} declaration in hoq source code

	name string

	//  path to program in file system, typicall relative

	path string

	//  full path to program in $PATH list, resolved at runtime

	full_path string

	//  count of dependencies on exit status in compiled hoq code

	depend_ref_count uint8

	//  initial static argument vector for command line

	init_argv []string
}

//  exec() a unix command, wait for exit status and write output to
//  standard  out and standard error.

func (cmd *command) exec(argv []string) (exit_status uint8) {

	argc := len(cmd.init_argv)
	xargv := make([]string, 1+argc+len(argv))

	//  the first argument must be the command path
	xargv[0] = cmd.path

	copy(xargv[1:], cmd.init_argv[:])
	copy(xargv[1+argc:], argv)

	//  the first argument must be the command path

	ex := &exec.Cmd{
		Path: cmd.full_path,
		Args: xargv,
	}

	//  run the command

	output, err := ex.Output()

	if err != nil {

		//  Ignore wierd err upon of non-zero exit.
		//  Signaled process panics(), which is not clean

		if !strings.HasPrefix(err.Error(), "exit status ") {
			panic(err)
		}
		if ee, ok := err.(*exec.ExitError); ok {
			stderr.Write(ee.Stderr)
		}
		err = nil
	}

	//  write standard output

	if len(output) > 0 {
		_, err = stdout.Write(output)
		if err != nil {
			panic(err)
		}
	}

	//  assemble exit status of process

	ps := ex.ProcessState
	if ps == nil {
		panic(fmt.Sprintf("%s: exec.Cmd.ProcessState is nil", cmd.name))
	}

	wait := ps.Sys().(syscall.WaitStatus)
	ws := uint16(wait)

	// already caught signal above in wierd error status

	return uint8(ws >> 8)
}

//  map a relative path of the executable file to the full path
//  in $PATH variable

func (cmd *command) lookup_full_path() {

	fp, err := exec.LookPath(cmd.path)
	if err != nil {
		panic(err)
	}
	cmd.full_path = fp
}
