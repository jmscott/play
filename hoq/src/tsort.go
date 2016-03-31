//  topologically sort dependency pairs beteen processes in hoq qualifications
//  use gnu 'tsort' program to sort.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

//  invoke program tsort on an array of label pairs.

func tsort(pairs []string) (depend_order []string) {

	tmp, err := ioutil.TempFile("", "hoq-tsort.out")
	if err != nil {
		panic(err)
	}
	defer func() {
		tmp.Close()
		syscall.Unlink(tmp.Name())
	}()

	out := bufio.NewWriter(tmp)

	//  Write the vertices of the dependency graph built by the attribute
	//  references in either the argv or the 'when' clause.

	for _, v := range pairs {
		fmt.Fprintf(out, "%s\n", v)
	}

	/*
	 *  Note:
	 *	I (jmscott) can't find the stdio equivalent of a close on
	 *	for buffered output in package bufio.  Seems odd to have
	 *	to actually flush before closing.
	 */
	err = out.Flush()
	if err != nil {
		panic(err)
	}
	tmp.Close()

	tsort := os.Getenv("HOQ_TSORT_PATH")
	if tsort == "" {
		tsort = "/usr/bin/tsort"
	}
	argv := make([]string, 2)
	argv[0] = tsort
	argv[1] = tmp.Name()
	gcmd := &exec.Cmd{
		Path: tsort,
		Args: argv[:],
	}

	//  run gnu tsort command, check after ps is fetched, since any
	//  non-zero exit returns err.  see comments in os_exec.go.

	output, _ := gcmd.CombinedOutput()

	ps := gcmd.ProcessState
	if ps == nil {
		panic("os/exec.CombinedOutput(): nil ProcessState")
	}
	sys := ps.Sys()
	if sys.(syscall.WaitStatus).Signaled() {
		panic("os/exec.CombinedOutput(): terminated with a signal")
	}

	//  non-zero error indicates failure to sort graph topologically.
	//  man pages for gnu tsort say nothing about meaning of non-zero
	//  exit status, so we can't do much smart error handling.

	ex := uint8((uint16(sys.(syscall.WaitStatus))) >> 8)
	if ex != 0 {
		//  Note:
		//	tsort writes the cycle, so why not capture?
		panic(errors.New("dependency graph has cycles"))
	}

	//  build the dependency order for the calls/queries

	depend_order = make([]string, 0)
	in := bufio.NewReader(strings.NewReader(string(output)))
	for i := 0; ; i++ {
		var name string

		name, err = in.ReadString(byte('\n'))
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return
		}

		name = strings.TrimSpace(name)
		depend_order = append(depend_order, name)
	}

	//  reverse depend order so roots of the graph are first

	for i, j := 0, len(depend_order)-1; i < j; i, j = i+1, j-1 {
		depend_order[i], depend_order[j] =
			depend_order[j], depend_order[i]
	}

	return
}
