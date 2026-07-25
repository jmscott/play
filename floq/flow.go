package main

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type flow struct {
	//  flow sequence, unique while floq running
	seq		uint64

	//  when particular flow started
	start_time	time.Time

	//  synchronize all oproutines in this flow
	wg_op		*sync.WaitGroup

	//  number of operators in a single flow
	//
	//  Note:  is this not global, like rest of next_* variables?
	op_count	uint8
}

//  a river to my people ...
type flow_chan chan *flow

func (flo *flow) new(op_count uint8) *flow {

	seq := uint64(1)

	if flo != nil {
		seq = flo.seq + 1
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

//   increment operator count for a flow by 1
func (flo *flow) inc() {
	flo.wg_op.Add(1)
	flo.op_count++
}

//   decrement operator count for a flow by 1
func (flo *flow) decr() {
	
	//  cheap sanity test
	if flo.op_count < 1 {
		die("op_count < 1")
	}
	flo.wg_op.Add(-1)
	flo.op_count--
}

//   add opcounts
func (flo *flow) add(delta uint8) {

	
	flo.wg_op.Add(int(delta))
	flo.op_count += delta
}

//  start an os process as part of the "flow <command>" statement

type osx_start struct {

	*command

	stdin		io.WriteCloser
	stdout		*bufio.Reader
	stderr		*bufio.Reader

	process		*os.Process
}

var next_mux sync.Mutex
var next_flow *flow

var next_lead_op_seen bool

//  number of oproutines seen in current flow
var next_op_count uint8

//  get the next flow for an operator to crunch

func (flo *flow) next() *flow {

	//  help trace a panic: how expensive??

	caller := rcaller(2)
	if strings.HasSuffix(caller, ".func1") {
		slen := len(caller)
		caller = caller[:slen-6]
	}

	flo.wg_op.Done()

	//  wait for all operators in this flow to finish
	flo.wg_op.Wait()

	next_mux.Lock()
	defer func() {
		next_mux.Unlock()
	}()

	//  count the number of operators processed
	next_op_count++

	//  begining of new flow
	if next_lead_op_seen == false {
		next_flow = flo.new(flo.op_count)
		next_lead_op_seen = true
	}

	//  on final op of this flow
	if next_op_count == flo.op_count {
		next_op_count = 0
		next_lead_op_seen = false
	}

	//  cheap sanity test
	if next_flow == flo {
		die("%s#%d: next_flow==flo: %p", caller, flo.seq,  flo)
	}

	return next_flow
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

//  start the comand in a "flow <command>();" and perptually feed the single
//  output to a string channel.
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

/*
 *  Project the idx'th field of a tab separated record from a flow process.
 *
 *  Note:
 *	why is this specific to a flow() statement?
 */
func (flo *flow) project_flow_tsv_n(
	cmd *command,
	in_str string_chan,
	in_idx uint64_chan,
  ) (out string_chan) {
	out = make(string_chan)

	go func() {
		<-compiling

		for {
			var sv *string_value
			var iv *uint64_value


			//  wait for both string to project and the field index
			for sv == nil || iv == nil {
				select {
				case s := <-in_str:
					if sv != nil {
						die("input out of sync: string")
					}
					sv = s
				case i := <- in_idx:
					if iv != nil {
						die("input out of sync: index")
					}
					if i.is_null == false {
						if i.uint64 == 0 {
							die("tsv field is 0")
						}
						//  why limit to 255?
						if i.uint64 > 255 {
							die("tsv field > 255")
						}
					}
					iv = i
				}
			}


			//  the projected i'th tsv field
			fv := &string_value{
				is_null: sv.is_null || iv.is_null,
			}

			if fv.is_null == false {
				idx := int(iv.uint64) - 1

				fld := strings.Split(sv.string, "\t")
				if len(fld) <= idx {
					fv.string = fld[idx]
				} else {
					fv.is_null = true
				}
			}
			out <- fv
		}
	}()
	return
}

func (flo *flow) project_flow_seq() (out uint64_chan) {

	out = make(uint64_chan)

	go func() {

		<-compiling

		for {
			out <- &uint64_value{
					uint64: flo.seq,
			}

			flo = flo.next()
		}
	}()


	return out
}
