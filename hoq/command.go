package main

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

type command struct {
	name             string
	path             string
	depend_ref_count uint8

	//  static command line arguments
	argv []string
}

func (cmd *command) exec(argv []string) uint8 {

	argc := len(cmd.argv)
	xargv := make([]string, 1+argc+len(argv))

	//  the first argument must be the command path
	xargv[0] = cmd.path

	copy(xargv[1:], cmd.argv[:])
	copy(xargv[1+argc:], argv)

	//  the first argument must be the command path

	exec := &exec.Cmd{
		Path: cmd.path,
		Args: xargv,
	}

	//  run the command

	output, err := exec.CombinedOutput()

	//  need to segregate stout from stderr

	if len(output) > 0 {
		_, err = stderr.Write(output)
		if err != nil {
			panic(err)
		}
	}

	//  Ignore wierd err upon of non-zero exit codes or signal

	if err != nil {
		if !strings.HasPrefix(err.Error(), "exit status") &&
			!strings.HasPrefix(err.Error(), "signal") {
			panic(err)
		}
		err = nil
	}

	ps := exec.ProcessState
	if ps == nil {
		panic(fmt.Sprintf("%s: exec.Cmd.ProcessState is nil", cmd.name))
	}

	wait := ps.Sys().(syscall.WaitStatus)
	ws := uint16(wait)
	if wait.Signaled() {
		panic(fmt.Sprintf("%s: process signaled: #%d", uint8(ws)))
	}
	return uint8(ws >> 8)
}
