package main

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

	//  will eventually resolve to true, false or null
	rum_WAIT = rummy(0x1)

	//  known to be null in the sql sense
	rum_NULL = rummy(0x2)

	//  known to be false
	rum_FALSE = rummy(0x4)

	//  known to be true
	rum_TRUE = rummy(0x8)
)

//  state tables for logical and/or
var and = [137]rummy{}
var or = [137]rummy{}

func init() {

	//  some shifted constants for left hand bits

	const lw = rummy(rum_WAIT << 4)
	const ln = rummy(rum_NULL << 4)
	const lf = rummy(rum_FALSE << 4)
	const lt = rummy(rum_TRUE << 4)

	//  PostgrSQL logical 'and' semantics for discovered values,
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

	//  PostgrSQL logical 'or' semantics for discovered values,
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
	
	*flow
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

func (flo *flow) wait_bool2(
	op [137]rummy,
	left, right bool_chan,
) (
	next rummy,
) {
	var lv, rv *bool_value

	next = rum_WAIT
	for next == rum_WAIT {

		select {
		case l := <-left:
			if l == nil {
				return rum_NIL
			}

			// cheap sanity test.  will go away soon
			if lv != nil {
				corrupt("left hand value out of sync")
			}
			lv = l

		case r := <-right:
			if r == nil {
				return rum_NIL
			}

			// cheap sanity test.  will go away soon
			if rv != nil {
				corrupt("right hand value out of sync")
			}
			rv = r
		}
		next = op[(lv.rummy()<<4)|rv.rummy()]
	}

	//  drain unread channel.
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

func (flo *flow) bool2(
	op [137]rummy,
	left, right bool_chan,
) (out bool_chan) {

	out = make(bool_chan)

	go func() {

		defer close(out)

		for {

			flo = flo.get()

			var b, is_null bool

			//  Note: op can go away
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
				flow:    flo,
			}
			
		}
	}()

	return
}

func (flo *flow) const_true() (out bool_chan) {

	out = make(bool_chan)

	go func() {
		for {
			flo = flo.get()
			out <- &bool_value{
				bool:	true,
				flow:	flo,
			}
		}
	}()

	return out
}

func (flo *flow) const_false() (out bool_chan) {

	out = make(bool_chan)

	go func() {
		for {
			flo = flo.get()
			out <- &bool_value{
				bool:	false,
				flow:	flo,
			}
		}
	}()

	return out
}

func (flo *flow) not(in bool_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		for {
			flo = flo.get()

			bv := <- in

			out <- &bool_value{
					bool:		!bv.bool,
					is_null:	bv.is_null,
					flow:		flo,
			}
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
	     MATCH, NOMATCH:
		return true
	}
	return false
}

//  wait for left and right hand bools of any binary operator
//
//  Note: how does passing *bool_value compare to bool_value?

func (left bool_chan) wait2(right bool_chan) (
	lv, rv *bool_value, closed bool,
) {
	for lv == nil || rv == nil {
		select {
		case lv = <- left:
			closed = lv == nil
			if closed {
				return
			}
		case rv = <- right:
			closed = rv == nil
			if rv == nil {
				return
			}
		}
	}
	return
}

//  compare two bools for equality

func (flo *flow) eq_bool(left, right bool_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {

		for {
			defer close(out)

			flo = flo.get()

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
				flow:		flo,
			}
			if !bv.is_null {
				bv.bool = lv.bool == rv.bool
			}
			out <- bv
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

		for {
			defer close(out)

			flo = flo.get()

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
				flow:		flo,
			}
			if !bv.is_null {
				bv.bool = lv.bool != rv.bool
			}
			out <- bv
		}
	}()

	return out
}

func neq_bool(flo *flow, left, right bool_chan) (out bool_chan) {
	return flo.neq_bool(left, right)
}
