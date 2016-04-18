//  execute commands in (unix) file system and wait for output and exit status

package main

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

type command struct {

	//  name used by hoq qualifications

	name string

	//  full path to program in $PATH list, resolved at runtime

	full_path string

	//  count of dependencies on exit status in compiled hoq code
	//  Note: is depend_ref_count really needed?

	depend_ref_count uint8

	//  initial argument vector, before appending argv from exec()

	argv []string
}

//  exec() a unix command, wait for exit status and write output to
//  standard out and standard error.

func (cmd *command) exec(argv []string) (exit_status uint8) {

	argc := len(cmd.argv)
	xargv := make([]string, argc+len(argv))

	//  a copy of argv[] per exec()

	copy(xargv[:], cmd.argv[:])
	copy(xargv[argc:], argv)

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

	fp, err := exec.LookPath(cmd.argv[0])
	if err != nil {
		panic(err)
	}
	cmd.full_path = fp
}

func (cmd *command) newc(name string, argv []string) *command {

	var path string

	if len(argv) == 0 {
		path = name
		argv = make([]string, 1)
		argv[0] = path
	} else {
		path = argv[0]
	}
	return &command{
		name: name,
		argv: argv,
	}
}
