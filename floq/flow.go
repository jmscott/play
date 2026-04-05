package main

import (
	"os/exec"
	"bufio"
	"io"
	"os"
)

//  a river to my people ...
type flow_chan chan *flow

type flow struct {

        resolved	chan struct{}

	next chan	flow_chan
}

type osx_start struct {

	*command

	stdin		io.WriteCloser
	stdout		*bufio.Reader
	stderr		*bufio.Reader

	process		*os.Process
}

func (flo *flow) get() *flow {

        <-flo.resolved

        //  next active flow arrives on this channel
        reply := make(flow_chan)

        //  request another flow, sending reply channel to mother
        flo.next <- reply

        //  return next flow
        return <-reply

	return &flow{
			resolved:       make(chan struct{}),
			next:           make(chan flow_chan),
	}
}

func (flo *flow) start(cmd *command) (pro *osx_start) {

	var err error

	pro = &osx_start{
		command:	cmd,
	}

	cx := exec.Command(cmd.look_path)
	path, err := exec.LookPath(cmd.path)
	if err != nil {
		corrupt("%s: LookPath(%s) failed: %s", cmd, path, err)
	}
	name := cmd.name

	cx.Path = path
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

	go func() {
		err := cx.Wait()
		if err == nil {
			corrupt("Wait(%s) returned", cmd)
		} else {
			corrupt("Wait(%s) failed: %s", cmd, err)
		}
	}()

	return pro
}

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
			out <- &string_value{
				string:	str,
			}
			flo = flo.get()
		}
	}()
	return
}
