package main

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"
)

// bool_value is result of AND, OR and binary relational operations

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

type uint8_value struct {
	uint8
	is_null bool

	*flow
}
type uint8_chan chan *uint8_value

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

	//  the whole line of input with trailing new line removed

	line string

	//  tab separated fields split out from the line read from
	//  standard input

	fields []string

	//  count of go routines still flowing expressions
	confluent_count int
}

type flow_chan chan *flow

//  wait for all go routines to resolve, then request and return another flow

func (flo *flow) get() *flow {

	<-flo.resolved

	//  next active flow arrives on this channel

	reply := make(flow_chan)

	//  request another flow, sending reply channel to scheduler

	flo.next <- reply

	//  return next flow

	return <-reply
}

//  wait for two boolean input channels to resolve to either true, false or null

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
	//  someday the qualification tree will be pruned.

	if lv == nil {
		<-in_left
	} else if rv == nil {
		<-in_right
	}
	return next
}

//  compare two strings read from left and right input channels
//  and send boolean answer upstream.
//
//  if either string value is null (in SQL sense) then the boolean answer is
//  null.

func (flo *flow) string_rel2(
	rel2 func(left, right string) bool,
	in_left, in_right string_chan,
) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {

			var left, right *string_value

			for left == nil || right == nil {

				//  wait for either left or right hand value
				//  to arrive

				select {

				//  wait for left hand string to arrive

				case lv := <-in_left:
					if lv == nil {
						return
					}
					if left != nil {
						panic("string2: left again")
					}
					left = lv

				//  wait for right hand string to arrive

				case rv := <-in_right:
					if rv == nil {
						return
					}
					if right != nil {
						panic("string2: right again")
					}
					right = rv
				}
			}

			bv := &bool_value{
				flow:    flo,
				is_null: left.is_null || right.is_null,
			}

			//  invoke the string operator on non-null values

			if bv.is_null == false {
				bv.bool = rel2(left.string, right.string)
			}
			out <- bv
		}
	}()

	return out
}

//  map a uint8Xuint8 relation onto a boolean fault, honoring SQL null sematics.

func (flo *flow) uint8_rel2(
	rel2 [65536]bool,
	in_left, in_right uint8_chan,
) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {

			var left, right *uint8_value

			for left == nil || right == nil {

				//  wait for either left or right hand value
				//  to arrive

				select {

				//  wait for left hand value of operator

				case lv := <-in_left:
					if lv == nil {
						return
					}
					if left != nil {
						panic("uint8_rel: left again")
					}
					left = lv

				//  wait for right hand value of operator

				case rv := <-in_right:
					if rv == nil {
						return
					}
					if right != nil {
						panic("uint8_rel: right again")
					}
					right = rv
				}
			}

			bv := &bool_value{
				flow:    flo,
				is_null: left.is_null || right.is_null,
			}

			//  invoke the uint8 binary operator on non-null values

			if bv.is_null == false {

				bv.bool = rel2[(uint16(left.uint8)<<8) |
						uint16(right.uint8)]
			}
			out <- bv
		}
	}()

	return out
}

//  implement logical AND, logical OR, boolean comparison ==, !=.
//  using state table.

func (flo *flow) bool_rel2(
	op [137]rummy,
	in_left, in_right bool_chan,
) (out bool_chan) {

	out = make(bool_chan)

	//  logical bool binary operator

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {

			var b, is_null bool

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
				string: s,
				flow:   flo,
			}
		}
	}()

	return out
}

//  send a uint8 constant

func (flo *flow) const_uint8(ui uint8) (out uint8_chan) {

	out = make(uint8_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {

			out <- &uint8_value{
				uint8: ui,
				flow:  flo,
			}
		}
	}()

	return out
}

//  send a bool constant

func (flo *flow) const_bool(b bool) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {

			out <- &bool_value{
				bool: b,
				flow: flo,
			}
		}
	}()

	return out
}

//  convert an uint8 to a string

func (flo *flow) to_string_uint8(in uint8_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {
			ui := <-in
			if ui == nil {
				return
			}

			sv := &string_value{
				flow:    flo,
				is_null: ui.is_null,
			}
			if ui.is_null == false {
				sv.string = strconv.FormatUint(
					uint64(ui.uint8),
					10,
				)
			}
			out <- sv
		}
	}()
	return out
}

//  convert an uint8 to a string

func (flo *flow) to_string_bool(in bool_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {
			bv := <-in
			if bv == nil {
				return
			}

			sv := &string_value{
				flow:    flo,
				is_null: bv.is_null,
			}
			if bv.is_null == false {
				if bv.bool {
					sv.string = "true"
				} else {
					sv.string = "false"
				}
			}
			out <- sv
		}
	}()
	return out
}

//  send $I, the i'th tab separated field of the input line, upstream

func (flo *flow) dollar(i uint8) (out string_chan) {

	out = make(string_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {

			sv := &string_value{
				flow: flo,
			}
			if int(i) < len(flo.fields) {
				sv.string = flo.fields[i]
			} else {
				sv.is_null = true
			}
			out <- sv
		}
	}()

	return out
}

//  send the $0, the entire line, upstream

func (flo *flow) dollar0() (out string_chan) {

	out = make(string_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {

			out <- &string_value{
				flow:   flo,
				string: flo.line,
			}
		}
	}()

	return out
}

//  a rule fires if and only if both the argv[] exist and
//  the "when" clause is true.

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

//  optimized opcode for single string argument

func (flo *flow) argv1(in string_chan) (out argv_chan) {

	out = make(argv_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {

			var argv [1]string

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

//  concurrently read strings from multiple input channels to assemble
//  argument vector of a call() statement.  a single null string renders
//  the entire vector null.

func (flo *flow) argv(in_args []string_chan) (out argv_chan) {

	//  track a received string and position in argv[]
	type arg_value struct {
		*string_value
		position uint8
	}

	out = make(argv_chan)

	//  called func has arguments, so wait on multple string channels
	//  before sending assembled argv[]

	go func() {

		defer close(out)

		argc := uint8(len(in_args))

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

				if a == (arg_value{}) { // stream closed
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

func (flo *flow) exec(
	cmd *command,
	in_argv argv_chan,
	in_when bool_chan,

) (out uint8_chan) {

	out = make(uint8_chan)

	go func() {
		defer close(out)

		for flo = flo.get(); flo != nil; flo = flo.get() {

			//  wait for resolution of both the argument
			//  vector and the boolean "when" qualification.

			argv, when := flo.wait_fire(in_argv, in_when)
			if argv == nil {
				return
			}

			uv := &uint8_value{
				flow: flo,
			}

			switch {

			//  exit value is null when either argumenr vector or
			//  the "when" qualification is null

			case argv.is_null || when.is_null:
				uv.is_null = true

			//  when is true and argv exists, so fire the
			//  associated command

			case when.bool:
				uv.uint8 = cmd.exec(argv.argv)

			//  when clause is false

			default:
				uv.is_null = true
			}

			out <- uv
		}
	}()
	return out
}

func re_match(sample, re string) bool {

	matched, err := regexp.MatchString(re, sample)
	if err != nil {
		panic(err)
	}
	return matched
}

func re_nmatch(sample, re string) bool {

	matched, err := regexp.MatchString(re, sample)
	if err != nil {
		panic(err)
	}
	return !matched
}

func string_eq(s1, s2 string) bool {

	return s1 == s2
}

func string_neq(s1, s2 string) bool {

	return s1 == s2
}

//  broadcast a uint8 to many uint8 listeners
//
//  Note:
//	would be nice to randomize writes to the output channels

func (flo *flow) fanout_uint8(
	in uint8_chan,
	count uint8,
) (out []uint8_chan) {

	out = make([]uint8_chan, count)
	for i := uint8(0); i < count; i++ {
		out[i] = make(uint8_chan)
	}

	put := func(uv *uint8_value, uc uint8_chan) {

		uc <- uv
	}

	go func() {

		defer func() {
			for _, a := range out {
				close(a)
			}
		}()

		for flo = flo.get(); flo != nil; flo = flo.get() {

			uv := <-in
			if uv == nil {
				return
			}

			//  broadcast to channels in slice
			for _, uc := range out {
				go put(uv, uc)
			}
		}
	}()
	return out
}

var uint8_eq = [256 * 256]bool{}
var uint8_neq = [256 * 256]bool{}

//  build the state tables for boolean '==' and '!-' with sql semantics for null

func init() {

	//  initialize diagonal of '==' operator to true.

	for i := uint16(0);  i <= 255;  i++ {
		uint8_eq[i << 8 | i] = true
	}

	//  initialize all entries of uint8 "!=" operator as true.

	for i := range uint8_neq {
		uint8_neq[i] = true
	}

	//  initialize diagonal of '!=' operator to false.

	for i := uint16(0);  i <= 255;  i++ {
		uint8_neq[i << 8 | i] = false
	}
}

//  Reduce all the CALL statements into single uint8, which is the count
//  of the programs that actuall fired

func (flo *flow) fanin_uint8(inx []uint8_chan) (out uint8_chan) {

	if len(inx) == 0 {
		panic("no channels for input to fanin_uint8")
	}
	out = make(uint8_chan)

	go func() {
		defer close(out)

		inx_count := len(inx)

		//  merge many uint8 channels onto a single uint8

		uint8_merge := func() (merged uint8_chan) {

			var wg sync.WaitGroup
			merged = make(uint8_chan)

			io := func(in uint8_chan) {
				for uv := range in {
					merged <- uv
				}

				//  decrement active go routine count

				wg.Done()
			}
			wg.Add(len(inx))

			//  start a worker for each input channel

			for _, in := range inx {
				go io(in)
			}

			//  Start a goroutine to wait for all merge workers
			//  to exit, then close the merged channel.

			go func() {
				wg.Wait()
				close(merged)
			}()
			return
		}()

		for flo = flo.get(); flo != nil; flo = flo.get() {

			//  wait for len(inx) uint8 to arrive

			exec_count := uint8(0)
			for i := 0; i < inx_count; i++ {
				uv := <-uint8_merge
				if uv == nil {
					return
				}
				if uv.is_null == false {
					exec_count++
				}
			}
			out <- &uint8_value{
				uint8: exec_count,
				flow:  flo,
			}
		}
	}()
	return out
}