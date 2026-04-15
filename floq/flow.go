package main

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"sync"
)

//  a river to my people ...
type flow_chan chan *flow

//  number of "run <command>" statements.
//
//  Note: why global?  can not be scoped to "flow

var run_count	uint8

type flow struct {

	run_group	sync.WaitGroup		
}

//  start an os process as part of the "flow <command>" statement
type osx_start struct {

	*command

	stdin		io.WriteCloser
	stdout		*bufio.Reader
	stderr		*bufio.Reader

	process		*os.Process
}

func (flo *flow) get() *flow {

	flo.run_group.Wait()

	f := &flow{}
	f.run_group.Add(int(run_count))
WTF("run_count=%d", run_count)

	return f
}

//  start a process that runs perpetually.
//  fatal error is process exits.

func (flo *flow) start(cmd *command) (pro *osx_start) {

	var err error

	pro = &osx_start{
		command:	cmd,
	}

	cx := exec.Command(cmd.look_path)
	name := cmd.name

	cx.Path = cmd.look_path
	cx.Args = cmd.args
	cx.Env = cmd.env

	pro.stdin, err = cx.StdinPipe()
	if err != nil {
		corrupt("cmd.StdinPipe(%s) failed: %s", name, err)
	}

	var r io.ReadCloser

	r, err = cx.StdoutPipe()
	if err != nil {
		corrupt("cmd.StdoutPipe(%s) failed: %s", name, err)
	}
	pro.stdout = bufio.NewReader(r)

	r, err = cx.StderrPipe()
	if err != nil {
		corrupt("cmd.StderrPipe(%s) failed: %s", name, err)
	}
	pro.stderr = bufio.NewReader(r)

	err = cx.Start()
	if err != nil {
		corrupt("cmd.Start(%s) failed: %s", name, err)
	}
	pro.process = cx.Process

	//  Wait() on a process that should never terminate

	go func() {
		err := cx.Wait()
		if err == nil {
			corrupt("Wait(%s) exit (no error)", cmd)
		} else {
			corrupt("Wait(%s) failed: %s", cmd, err)
		}
	}()

	return pro
}

//  execute the statement "flow <command>();", no arguments, no when condition.
//  feed a single line of text per flow down stream.

func (flo *flow) osx_flow_0(cmd *command) (out string_chan) {

	out = make(string_chan)
	pro := flo.start(cmd)

	go func() {
		defer close(out)

		for {
			str, err := pro.stdout.ReadString('\n')
			if err != nil {
				croak("%s: Read(stdout) failed: %s", cmd, err)
			}
WTF("str=%s", str)
			out <- &string_value{
				string:		str,
			}
WTF("read: pre: flow=%#v", flo)

			flo = flo.get()
WTF("read: post: flow=%#v", flo)
		}
	}()
	return
}
