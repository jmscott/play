package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type flow struct {

	//  flow sequence, driven by "flow <command>()"

	seq		uint64

	start_time	time.Time

	//  synchronize all "op" goroutines in this flow

	wg_op		*sync.WaitGroup

	//  total number of go routine "operators" per compiled flow.
	//  sets the operator WaitGroup().
	//  no need for atomic, since only synchronous compile() changes.

	op_count	uint8

	//  the next flow, fetched by each op goroutine

	next_flow	*flow
}

//  a river to my people ...

type flow_chan chan *flow

func (flo *flow) new() *flow {

	seq := uint64(1)
	op_count := uint8(0)		//  compile incremenets first flow 

	if flo != nil {
		seq = flo.seq + 1
		op_count = flo.op_count
	}

	f := &flow{
		seq:		seq,
		start_time:	time.Now(),
		op_count:	op_count,
	}

	var wg sync.WaitGroup
	wg.Add(int(op_count))
	f.wg_op = &wg

	return f
}

//   increment operator count for a flow operation by +1
func (flo *flow) incr() {
	flo.wg_op.Add(1)
	flo.op_count++
}

//  start an os process as part of the "flow <command>" statement

type osx_start struct {

	*command

	stdin		io.WriteCloser
	stdout		*bufio.Reader
	stderr		*bufio.Reader

	process		*os.Process
}

//  Note: why global?

var next_mux sync.Mutex

func (flo *flow) next() *flow {

	var nm string

	if floq_trace_next_op {
		nm = fmt.Sprintf("%s#%d", rcaller(2), flo.seq)

		trace("%s: hello", nm)
	}

	flo.wg_op.Done()

	//  wait for all operators in this flow to finish

	flo.wg_op.Wait()

	//  allocate a new flow ... only once.

	next_mux.Lock()
	defer func() {
		next_mux.Unlock()
	}()
	
	if flo.next_flow == nil {
		flo.next_flow = flo.new()
	}

	if floq_trace_next_op {
		trace("next flow: %d", flo.next_flow.seq)
	}

	return flo.next_flow
}

func (flo *flow) String() string {

	if flo == nil {
		return "(*flow)(nil)"
	}
	return fmt.Sprintf("flow#%d", flo.seq)
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

		if err != nil {
			die("Wait(%s) failed: %s", cmd, err)
		}
		die("Wait(%s) exit (no error)", cmd)
	}()

	return pro
}

//  start the comand of a "flow <command>();" statement and perptually feed
//  lines of strings downstream

func (flo *flow) osx_flow(cmd *command) (out string_chan) {

	out = make(string_chan)

	go func() {

		<-compiling

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
	return out
}

//  project the sequence number of current flow

func (flo *flow) proj_flow_seq(in string_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {

		<-compiling

		for {
			<- in		//  only send if flow record

			out <- &uint64_value{
				uint64:		flo.seq,
			}

			flo = flo.next()
		}
	}()
	return
}

//  project the tab separated field associated with a particular attribute

func (flo *flow) proj_flow_tsv_att(
	proj *projection,
	in string_chan,
  ) (out string_chan) {

	out = make(string_chan)

	go func() {

		<-compiling

		idx := proj.att_ref.tab_field-1

		for {
			sv := <- in

			if sv.is_null == false {
				fld := strings.Split(sv.string, "\t")
				if int(idx) >= len(fld) {
					die(
						"flow#%d: " +
						"index >= field count: %s: %s",
						flo.seq,
						proj,
						sv.string) 

				}

				sv.string = fld[idx]
			}
			out <- sv

			flo = flo.next()
		}
	}()
	return
}
