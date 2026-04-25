package main

import (
	"fmt"
)

//  A rummy records for temporal + tri-state logic: true, false, null, waiting
//  There are known knowns, there are known unknowns ...
//
//   Note:
//	Rummy devised from different project and may be more complex than
//	needed.  orginal idea was to map process timeouts onto null

type rummy uint8

const (
	//  end of stream
	rum_NIL = rummy(0)

	//  waiting for resolutipn to true, false or null
	rum_WAIT = rummy(0x1)

	//  known to be null in the sql sense
	rum_NULL = rummy(0x2)

	//  known to be false
	rum_FALSE = rummy(0x4)

	//  known to be true
	rum_TRUE = rummy(0x8)
)

func (rum rummy) String() string {
	switch rum {
	case rum_NIL:
		return "NIL"
	case rum_WAIT:
		return "WAIT"
	case rum_NULL:
		return "NULL"
	case rum_FALSE:
		return "FALSE"
	case rum_TRUE:
		return "TRUE"
	default:
		return fmt.Sprintf("0x%02x", rum)
	}
}

//  state tables for logical and/or
var and = [137]rummy{}
var or = [137]rummy{}

func init() {

	//  some shifted constants for left hand bits

	const lw = rummy(rum_WAIT << 4)
	const ln = rummy(rum_NULL << 4)
	const lf = rummy(rum_FALSE << 4)
	const lt = rummy(rum_TRUE << 4)

	//  SQL logical 'and' semantics for discovered values,
	//  applied sequentially
	//
	//  false and *	=>	false
	//  * and false	=>	false
	//  null and *	=>	null
	//  * and null	=>	null
	//  *		=>	true

	//  left value is false

	and[lf|rum_WAIT] = rum_FALSE
	and[lf|rum_NULL] = rum_FALSE
	and[lf|rum_FALSE] = rum_FALSE
	and[lf|rum_TRUE] = rum_FALSE

	//  right value is false

	and[lw|rum_FALSE] = rum_FALSE
	and[ln|rum_FALSE] = rum_FALSE
	and[lt|rum_FALSE] = rum_FALSE
	and[lf|rum_FALSE] = rum_FALSE

	//  left value is true

	and[lt|rum_WAIT] = rum_WAIT
	and[lt|rum_NULL] = rum_NULL
	and[lt|rum_FALSE] = rum_FALSE
	and[lt|rum_TRUE] = rum_TRUE

	//  right value is true

	and[lw|rum_TRUE] = rum_WAIT
	and[ln|rum_TRUE] = rum_NULL
	and[lt|rum_TRUE] = rum_TRUE
	and[lf|rum_TRUE] = rum_FALSE

	//  left value is null

	and[ln|rum_NULL] = rum_NULL
	and[ln|rum_TRUE] = rum_NULL
	and[ln|rum_FALSE] = rum_FALSE
	and[ln|rum_WAIT] = rum_WAIT

	//  right value is null

	and[lt|rum_NULL] = rum_NULL
	and[lf|rum_NULL] = rum_FALSE
	and[ln|rum_NULL] = rum_NULL
	and[lw|rum_NULL] = rum_WAIT

	//  left value is waiting

	and[lw|rum_NULL] = rum_WAIT
	and[lw|rum_TRUE] = rum_WAIT
	and[lw|rum_FALSE] = rum_FALSE
	and[lw|rum_WAIT] = rum_WAIT

	//  right value is waiting

	and[lt|rum_WAIT] = rum_WAIT
	and[lf|rum_WAIT] = rum_FALSE
	and[ln|rum_WAIT] = rum_WAIT
	and[lw|rum_WAIT] = rum_WAIT

	//  SQL logical 'or' semantics for discovered values,
	//  applied sequentially
	//
	//  true or *	=>	true
	//  * or true	=>	true
	//  null or *	=>	null
	//  * or null	=>	null
	//  *		=>	false

	//  left value is true

	or[lt|rum_WAIT] = rum_TRUE
	or[lt|rum_NULL] = rum_TRUE
	or[lt|rum_FALSE] = rum_TRUE
	or[lt|rum_TRUE] = rum_TRUE

	//  right value is true

	or[lw|rum_TRUE] = rum_TRUE
	or[ln|rum_TRUE] = rum_TRUE
	or[lf|rum_TRUE] = rum_TRUE
	or[lt|rum_TRUE] = rum_TRUE

	//  left value is false

	or[lf|rum_WAIT] = rum_WAIT
	or[lf|rum_NULL] = rum_NULL
	or[lf|rum_FALSE] = rum_FALSE
	or[lf|rum_TRUE] = rum_TRUE

	//  right value is false

	or[lw|rum_FALSE] = rum_WAIT
	or[ln|rum_FALSE] = rum_NULL
	or[lf|rum_FALSE] = rum_FALSE
	or[lt|rum_FALSE] = rum_TRUE

	//  left value is null

	or[ln|rum_WAIT] = rum_WAIT
	or[ln|rum_NULL] = rum_NULL
	or[ln|rum_FALSE] = rum_NULL
	or[ln|rum_TRUE] = rum_TRUE

	//  right value is null

	or[ln|rum_NULL] = rum_NULL
	or[lt|rum_NULL] = rum_TRUE
	or[lf|rum_NULL] = rum_NULL
	or[lw|rum_NULL] = rum_WAIT

	//  left value is waiting

	or[lw|rum_WAIT] = rum_WAIT
	or[lw|rum_NULL] = rum_WAIT
	or[lw|rum_TRUE] = rum_TRUE
	or[lw|rum_FALSE] = rum_WAIT

	//  right value is waiting
	or[lw|rum_WAIT] = rum_WAIT
	or[lt|rum_WAIT] = rum_TRUE
	or[lf|rum_WAIT] = rum_WAIT
	or[ln|rum_WAIT] = rum_WAIT
}

type bool_value struct {
	
	bool
	is_null	bool
}

type bool_chan chan *bool_value

type relop_bool_func func (*flow, bool_chan, bool_chan) bool_chan
var relop_bool = map[int]relop_bool_func{
		EQ:	eq_bool,
		NEQ:	neq_bool,
	}

func (bv *bool_value) rummy() rummy {

        switch {
        case bv == nil:
                return rum_WAIT
        case bv.is_null:
                return rum_NULL
        case bv.bool:
                return rum_TRUE
        }
        return rum_FALSE
}

// 
func (flo *flow) wait_bool2(op [137]rummy, left, right bool_chan) (next rummy) {
	var lv, rv *bool_value

	next = rum_WAIT

	//  read left or right bools until not waiting state
	for next == rum_WAIT {
		select {
			case lv = <-left:
			case rv = <-right:
		}

		//  build a rummy.  nil ok
		next = op[(lv.rummy()<<4)|rv.rummy()]
	}

	//  drain unread left or right channel.
	//
	//  Note: reading in the background causes a multiple read of
	//        same left/right hand side.  Why?  Shouldn't the flow
	//	  block on current sequence until all qualfications converge?

	if lv == nil {
		<-left
	} else if rv == nil {
		<-right
	}
	return next
}

func (bv *bool_value) String() string {

	if bv == nil {
		return "bool_value(nil)"
	}
	if bv.is_null {
		return "NULL"
	}
	if bv.bool {
		return "TRUE"
	}
	return "FALSE"
}

func (flo *flow) bool2(
	op [137]rummy,
	left, right bool_chan,
) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {

			var b, is_null bool

			rum := flo.wait_bool2(op, left, right)
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
			}

			flo = flo.next()
		}
	}()

	return
}

//  send a constant "true" value
func (flo *flow) const_true() (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			out <- &bool_value{
				bool:	true,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  send a constant "null" value
func (flo *flow) const_false() (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			out <- &bool_value{
				bool:	false,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  Negate a read boolean value
func (flo *flow) not(in bool_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			bv := <- in

			out <- &bool_value{
					bool:		!bv.bool,
					is_null:	bv.is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

func (a *ast) is_bool() bool {
	switch a.yy_tok {
	case yy_AND, yy_OR,
	     yy_TRUE, yy_FALSE,
	     EQ, NEQ,
	     GT, GTE,
	     LT, LTE,
	     NOT,
	     IS_NULL,
	       IS_NULL_BOOL, IS_NULL_UINT64,  IS_NULL_STRING,
	     IS_NOT_NULL,
	       IS_NOT_NULL_BOOL, IS_NOT_NULL_UINT64, IS_NOT_NULL_STRING,
	     MATCH, NOMATCH:
		return true
	}
	return false
}

//  Wait for both left and right hand bools of any binary operator

func (left bool_chan) wait2(right bool_chan) (lv, rv *bool_value) {
	for lv == nil || rv == nil {
		select {
			case lv = <- left:
			case rv = <- right:
		}
	}
	return
}

//  compare two bools for equality

func (flo *flow) eq_bool(left, right bool_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			lv, rv := left.wait2(right)

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !bv.is_null {
				bv.bool = lv.bool == rv.bool
			}
			out <- bv
			flo = flo.next()
		}
	}()

	return out
}

func eq_bool(flo *flow, left, right bool_chan) (out bool_chan) {
	return flo.eq_bool(left, right)
}

//  compare two bools for equality

func (flo *flow) neq_bool(left, right bool_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			lv, rv := left.wait2(right)

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !bv.is_null {
				bv.bool = lv.bool != rv.bool
			}
			out <- bv
			flo = flo.next()
		}
	}()

	return out
}

func neq_bool(flo *flow, left, right bool_chan) (out bool_chan) {
	return flo.neq_bool(left, right)
}

//  Cast a read boolean value to a string: "true" or "false"
func (flo *flow) cast_bool(in bool_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		<-compiling

		for {
			bv := <-in
			
			sv := &string_value{
					is_null:	bv.is_null,
			}
			if sv.is_null == false {
				if bv.bool == true {
					sv.string = "true"
				} else {
					sv.string = "false"
				}
			}
			out <-sv

			flo = flo.next()
		}
	}()
	return out
}

//  Is a boolean value null in an sql sense?
func (flo *flow) is_null_bool(in bool_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			out <- &bool_value{
				bool:	(<-in).is_null,
			}

			flo = flo.next()
		}
	}()

	return out
}

//  Is a boolean value not null in an sql sense?
func (flo *flow) is_not_null_bool(in bool_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			out <- &bool_value{
				bool:	(<-in).is_null == false,
			}

			flo = flo.next()
		}
	}()

	return out
}
