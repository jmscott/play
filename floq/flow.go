package main

import (
	"sync"
	"sync/atomic"
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

//  a river to my people ...
type flow_chan chan *flow

type concurrent_group struct {

	//  number of active waiters
	waiter_count	int64

	//  number of exited done
	done_count	int64

	flow_seq	uint64

	//  lock for updating counts
	mux		sync.Mutex

	//  the wait group tracking concurrent operator goroutines.
	gmux		sync.WaitGroup
}

var con_group concurrent_group

func (cg *concurrent_group) next(count uint8) *flow {
	cg.mux.Lock()
	defer cg.mux.Unlock()

	//  cheap sanity test to insure all waiters done
	c := atomic.LoadInt64(&cg.waiter_count)
	if c != 0  {
		die("wait count not 0, %d instead", c)
	}

	cg.gmux.Add(int(count))
	f := (*flow)(nil).new(atomic.AddUint64(&cg.flow_seq, 1))
	return f
}

func (flo *flow) new(seq uint64) *flow {

	return &flow{
		seq:		seq,
		start_time:	time.Now(),
		done:		make(chan bool),
	}
}

func (cg *concurrent_group) wait(caller string) {

	//  increase waiter count and do cheap sanity test.
	if atomic.AddInt64(&cg.waiter_count, 1) < 1 {
		die("enter: waiter_count < 1: %s", caller)
	}
	cg.gmux.Wait()

	//  decrement waiter count and do cheap sanity test
	if atomic.AddInt64(&cg.waiter_count, -1) < 0 {
		die("exit: waiter_count < 0: %s", caller)
	}
}

func (cg *concurrent_group) done() {
	cg.mux.Lock()
	defer cg.mux.Unlock()

	cg.gmux.Done()
	atomic.AddInt64(&cg.done_count, -1)
}

func (cg *concurrent_group) wait_count() uint8 {

	return uint8(atomic.LoadInt64(&cg.waiter_count))
}

var flow_cop_get	chan flow_chan

type flow struct {
	//  flow sequence, unique while floq running
	seq		uint64

	//  when flow started
	start_time	time.Time

	//  wait for all op routies to finish
	done		chan(bool)
}

//  start an os process as part of the "flow <command>" statement

type osx_start struct {

	*command

	stdin		io.WriteCloser
	stdout		*bufio.Reader
	stderr		*bufio.Reader

	process		*os.Process
}

func (flo *flow) next() *flow {
	caller := rcaller(2)


	con_group.done()

	//  wait for other oproutines to finish
	<-flo.done


	var f *flow

	//  request the new active flow by sending channel
	fc := make(flow_chan)
	flow_cop_get <- fc

	//  read the new flow
	f = <- fc

	//  cheap sanity tests
	switch {
	case f == nil:
		die("next flow == <nil>: %s", caller)

	case f == flo:
		die("next flow == current flow: %s", caller)
	}

	return f
}

//  start a process that runs perpetually.
//  fatal error if process exits.

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
		die("cmd.StdinPipe(%s) failed: %s", name, err)
	}

	var r io.ReadCloser

	r, err = cx.StdoutPipe()
	if err != nil {
		die("cmd.StdoutPipe(%s) failed: %s", name, err)
	}
	pro.stdout = bufio.NewReader(r)

	r, err = cx.StderrPipe()
	if err != nil {
		die("cmd.StderrPipe(%s) failed: %s", name, err)
	}
	pro.stderr = bufio.NewReader(r)

	err = cx.Start()
	if err != nil {
		die("cmd.Start(%s) failed: %s", name, err)
	}
	pro.process = cx.Process

	//  Wait() on a process that should never terminate

	go func() {
		err := cx.Wait()

		//  floq process exiting due to user signal
		if caught_sig != nil {
			return
		}

		if err == nil {
			die("Wait(%s) exit (no error)", cmd)
		} else {
			die("Wait(%s) failed: %s", cmd, err)
		}
	}()

	return pro
}

// The traffic cop goroutine who coordinates flow of operators.

func (flo *flow) cop(goroutine_count uint8) {

	active_flow := flo

	flow_cop_get = make(chan flow_chan)

	for {
		if con_group.wait_count() == 0 {
			done := active_flow.done
			active_flow = con_group.next(goroutine_count)
			close(done)
		}

		fc := <-flow_cop_get
		fc <- active_flow
	}
}

//  execute the statement "flow <command>();"

func (flo *flow) osx_flow(cmd *command) (out string_chan) {

	out = make(string_chan)

	go func() {
		stdout := flo.start(cmd).stdout

		for {
			str, err := stdout.ReadString('\n')
			if err != nil {
				die("%s: Read(stdout) failed: %s", cmd, err)
			}

			out <- &string_value{
				string:		str,
			}
		}
	}()
	return
}

func (flo *flow) String() string {

	return fmt.Sprintf("%p", flo)
}
