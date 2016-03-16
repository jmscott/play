package main

import (
	"fmt"
	"sync"
)

// bool_value is result of AND, OR and relational operations
type bool_value struct {
	bool
	is_null bool

	*flow
}
type bool_chan chan *bool_value

func (bv *bool_value) String() string {
	if bv == nil {
		return "NIL"
	}
	if bv.is_null {
		return "<NULL>"
	}
	return fmt.Sprintf("%t", bv.bool)
}

type string_value struct {
	string
	is_null bool

	*flow
}
type string_chan chan *string_value

type exit_status_value struct {
	uint8
	is_null bool
	called bool

	*flow
}
type exit_status_value_chan chan *exit_status_value

type argv_value struct {
	argv    []string
	is_null bool

	*flow
}
type argv_chan chan *argv_value

//  a flow tracks the firing of rules over a single line of input text.

type flow struct {

	//  request a new flow from this channel,
	//  reading reply on sent side-channel

	next chan flow_chan

	//  channel is closed when all call()s make no further progress

	resolved chan struct{}

	//  tab separated fields split out from the line read from
	//  standard input

	fields	[]string

	//  count of go routines still flowing expressions
	confluent_count int
}

type flow_chan chan *flow

//  wait for all go routines to resolve

func (flo *flow) get() *flow {

	<-flo.resolved

	//  next active flow arrives on this channel

	reply := make(flow_chan)

	//  request another flow, sending reply channel to scheduler

	flo.next <- reply

	//  return next flow

	return <-reply
}

//  wait for two boolean input channels to resolve

func (flo *flow) wait_bool2(
	op [137]rummy,
	in_left, in_right bool_chan,
) (
	next rummy,
) {
	var lv, rv *bool_value

	next = rum_WAIT
	for next == rum_WAIT {

		select {
		case l := <-in_left:
			if l == nil {
				return rum_NIL
			}

			// cheap sanity test.  will go away soon
			if lv != nil {
				panic("left hand value out of sync")
			}
			lv = l

		case r := <-in_right:
			if r == nil {
				return rum_NIL
			}

			// cheap sanity test.  will go away soon
			if rv != nil {
				panic("right hand value out of sync")
			}
			rv = r
		}
		next = op[(lv.rummy()<<4)|rv.rummy()]
	}

	//  drain unread channel.
	//
	//  Note: reading in the background causes a mutiple read of
	//        same left/right hand side.  Why?  Shouldn't the flow
	//	  block on current sequence until all qualfications converge?

	if lv == nil {
		<-in_left
	} else if rv == nil {
		<-in_right
	}
	return next
}

/*
 *  Execute either logical AND or logical OR
 */
func (flo *flow) bool2(
	op [137]rummy,
	in_left, in_right bool_chan,
) (out bool_chan) {

	out = make(bool_chan)

	//  logical bool binary operator

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {

			var b, is_null bool

			//  Note: op can go away
			rum := flo.wait_bool2(op, in_left, in_right)
			switch rum {
			case rum_NIL:
				return
			case rum_NULL:
				is_null = true
			case rum_TRUE:
				b = true
			}
			out <- &bool_value{
				bool:    b,
				is_null: is_null,
				flow:    flo,
			}
		}
	}()
	return out
}

//  send a string constant

func (flo *flow) const_string(s string) (out string_chan) {

	out = make(string_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {
			out <- &string_value{
				string:  s,
				is_null: false,
				flow:    flo,
			}
		}
	}()

	return out
}

//  send the i'th tab separated field of the input line

func (flo *flow) dollar(i int) (out string_chan) {

	out = make(string_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {
			
			sv := &string_value{
				flow:	flo,
			}
			if i < len(flo.fields) {
				sv.string = flo.fields[i]
			} else {
				sv.is_null = true
			}
			out <- sv
		}
	}()

	return out
}
//  a rule fires if and only if both the argv[] exists and
//  the when clause is true.

func (flo *flow) wait_fire(
	in_argv argv_chan,
	in_when bool_chan,
) (
	argv *argv_value,
	when *bool_value,
) {

	//  wait for both an argv[] and resolution of the when clause
	for argv == nil || when == nil {
		select {
		case argv = <-in_argv:
			if argv == nil {
				return nil, nil
			}

		case when = <-in_when:
			if when == nil {
				return nil, nil
			}
		}
	}
	return
}

//  empty non-null argv sends immediatly

func (flo *flow) argv0() (out argv_chan) {

	out = make(argv_chan)

	go func() {
		defer close(out)

		var argv [0]string

		for flo = flo.get(); flo != nil; flo = flo.get() {

			out <- &argv_value{
				is_null: false,
				argv:    argv[:],
				flow:    flo,
			}
		}
	}()
	return out
}

//  direct answer of argv with single string

func (flo *flow) argv1(in string_chan) (out argv_chan) {

	out = make(argv_chan)

	go func() {
		defer close(out)

		var argv [1]string

		for flo = flo.get(); flo != nil; flo = flo.get() {

			sv := <-in
			if sv == nil {
				return
			}
			argv[0] = sv.string
			out <- &argv_value{
				is_null: sv.is_null,
				argv:    argv[:],
				flow:    flo,
			}
		}
	}()
	return out
}

//  read strings from multiple input channels and write assmbled argv[]
//  any null value renders the whole argv[] null

func (flo *flow) argv(in_args []string_chan) (out argv_chan) {

	//  track a received string and position in argv[]
	type arg_value struct {
		*string_value
		position uint8
	}

	out = make(argv_chan)
	argc := uint8(len(in_args))

	//  called func has arguments, so wait on multple string channels
	//  before sending assembled argv[]

	go func() {

		defer close(out)

		//  merge() many string channels onto a single channel of
		//  argument values.

		merge := func() (mout chan arg_value) {

			var wg sync.WaitGroup
			mout = make(chan arg_value)

			io := func(sc string_chan, p uint8) {
				for sv := range sc {
					mout <- arg_value{
						string_value: sv,
						position:     p,
					}
				}
				wg.Done()
			}

			wg.Add(len(in_args))
			for i, sc := range in_args {
				go io(sc, uint8(i))
			}

			//  Start a goroutine to close 'mout' channel
			//  once all the output goroutines are done.

			go func() {
				wg.Wait()
				close(mout)
			}()
			return
		}()

		for flo = flo.get(); flo != nil; flo = flo.get() {

			av := make([]string, argc)
			ac := uint8(0)
			is_null := false

			//  read until we have an argv[] for which all elements
			//  are also non-null.  any null argv[] element makes
			//  the whole argv[] null

			for ac < argc {

				a := <-merge

				//  Note: compile generates error for
				//        arg_value{}

				if a == (arg_value{}) {
					return
				}

				sv := a.string_value
				pos := a.position

				//  any null element forces entire argv[]
				//  to be null

				if a.is_null {
					is_null = true
				}

				//  cheap sanity test tp insure we don't
				//  see the same argument twice
				//
				//  Note:
				//	technically this implies an empty
				//	string is not allowed which is probably
				//	unreasonable

				if av[pos] != "" {
					panic("argv[] element not \"\"")
				}
				av[pos] = sv.string
				ac++
			}

			//  feed the hungry world our new, boundless argv[]
			out <- &argv_value{
				argv:    av,
				is_null: is_null,
				flow:    flo,
			}
		}
	}()

	return out
}

func (flo *flow) call(
	cmd *command,
	in_argv argv_chan,
	in_when bool_chan,
) (out exit_status_value_chan) {

	out = make(exit_status_value_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {

			argv, when := flo.wait_fire(in_argv, in_when)
			if argv == nil {
				return
			}

			//  don't fire command if either argv is null or
			//  when is false or null

			exv := &exit_status_value{
				flow: flo,
			}

			switch {

			//  uint8 is null when either argv or when is null

			case argv.is_null || when.is_null:
				exv.is_null = true

			//  when is true and argv exists, so fire command

			case when.bool:
				//  synchronous call to command{} process
				//  always returns partially built xdr value

				exv.called = true
				//xv = cmd.call(argv.argv, osx_q)
			}

			//  the flow never resolves until all call()s
			//  generate an xdr record.

			out <- exv
		}
	}()
	return out
}
