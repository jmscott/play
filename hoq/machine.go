package main

import (
	"fmt"
)

// bool_value is result of AND, OR and relational operations
type bool_value struct {
	bool
	is_null bool

	*flow
}

func (bv *bool_value) String() string {
	if bv == nil {
		return "NIL"
	}
	if bv.is_null {
		return "<NULL>"
	}
	return fmt.Sprintf("%t", bv.bool)
}

type bool_chan chan *bool_value

type string_value struct {
	string
	is_null bool

	*flow
}

type string_chan chan *string_value

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

	fields	[]int

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

//  send a field from the input line

func (flo *flow) dollar(field int) (out string_chan) {

	out = make(string_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {
			
			if field < len(flo.fields)

			out <- &string_value{
				string:  s,
				is_null: false,
				flow:    flo,
			}
		}
	}()

	return out
}
