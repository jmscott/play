package main

import (
	"os/exec"
)

type command struct {
	
	name	string
	cmd	*exec.Cmd
	path	string
}

func (cmd *command) yy_frisk(al *ast) string {

	c := 0
	var an *ast
	for an = al.left;  an.next != nil;  an = an.next {
		c++
		if c > 2 {
			return "more than two attributes"
		}
	}
	return ""
}
