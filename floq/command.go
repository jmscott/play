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

	for an := al.left;  an != nil;  an = an.next {
		switch an.left.string {

		//  Note: no attribute can be given twice, so collapse
		case "path":
			if cmd.path != "" {
				return "attribute more than once: path"
			}
			cmd.path = an.right.string
		case "argv":
			if cmd.args != nil {
				return "att more than once: argv"
			}
			cmd.args = an.right.array_ref
		case "env":
			if cmd.env != nil {
				return "att more than once: env"
			}
			cmd.env = an.right.array_ref
		default:
			return "unknown attribute: " + an.left.string
		}
	}
	if cmd.path == "" {
		return "missing attribute: \"path\""
	}
	return ""
}
