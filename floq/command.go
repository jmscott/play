package main

import (
	"os/exec"
)

type command struct {
	
	name	string
	cmd	*exec.Cmd
}
