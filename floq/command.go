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

func (cmd *command) frisk_att(al *ast) string {

	for an := al.left;  an.next != nil;  an = an.next {
		switch an.left.string {
		case "path":
			if cmd.path != "" {
				return "attribute more than once: path"
			}
			cmd.path = an.right.string
		case "args":
			if cmd.args != nil {
				return "attribute more than once: args"
			}
			cmd.args = an.right.array_ref
		case "env":
			if cmd.env != nil {
				return "attribute more than once: env"
			}
			cmd.env = an.right.array_ref
		default:
			return "unknown attribute: " + an.left.string
		}
	}
	if cmd.path == "" {
		return "missing required attribute: path"
	}
	return ""
}
