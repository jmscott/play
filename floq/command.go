package main

import (
	"os"
	"os/exec"
)

type command struct {
	
	name	string
	cmd	*exec.Cmd
	path	string
	args	[]string
	env	[]string
}

type osx_value struct {
	*command
	argv	[]string
	err	error

	*flow
}

type osx_chan chan osx_value

//  run a command process with no arguments
func (flo *flow) osx0(cmd *command, when bool_chan) (out osx_chan) {

	out = make(osx_chan)

	go func() {

		//  check existence of executable iin "path" variable

		path, err := exec.LookPath(cmd.path)
		if err != nil {
			croak("exec.LookPath(%s) failed: %s", err)
		}

		//  build process environment
		//
		//  Note:
		//	what about dups in env?
		//

		env := os.Environ()
		for _, e := range cmd.env {
			env = append(env, e)
		}

		for {
			flo := flo.get()

			bv := <- when
			if bv.bool == false {
				continue
			}
			cx := exec.Command(
					path,
			)
			cx.Args = cmd.args
			cx.Env = env

			err := cx.Run() 
			out <- osx_value{
					flow:	flo,
					err:	err,	
				}
		}
	}()

	return out
}

func (cmd *command) frisk_att(atup *ast) (err error) {

	err = atup.frisk_att2(
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
		atup.impossible("can not find ATT 'path'")
	}
	if ap.right == nil {
		ap.impossible("path ATT has not right")
	}
	cmd.path, err = exec.LookPath(ap.right.string)
	return err
}
