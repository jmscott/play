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
	full_path        string
	depend_ref_count uint8
}

func (cmd *command) exec(argv []string) uint8 {

	//  the first argument must be the command path

	exec := &exec.Cmd{
		Path: cmd.full_path,
		Args: argv,
	}

	//  run the command

	output_256, err := exec.CombinedOutput()

	//  any output from the process is a panicy error

	if output_256 != nil {
		stderr.Write(output_256)
		panic(fmt.Sprintf("%s: unexpected output", cmd.name))
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
