package main

// Note: change uint64 to uint63!  for compatiblebilit with signed uint63!

import (
	"strconv"
)

//  uint64 value passed between flow operators

type uint64_value struct {
	uint64

	is_null		bool
}

//  a river of uint64s

type uint64_chan chan *uint64_value

//  table of boolean, relational operations on uint64

type relop_uint64_func func (*flow, uint64_chan, uint64_chan) bool_chan
var relop_uint64 = map[int]relop_uint64_func{
		GT:			gt_ui64,
		GTE:			gte_ui64,
		EQ:			eq_ui64,
		NEQ:			neq_ui64,
		LTE:			lte_ui64,
		LT:			lt_ui64,
	}


//  wait for both left and right hand alues uint64 of any binary operator,
//  indicating when either channels closes.

func (left uint64_chan) wait2(right uint64_chan) (
	lv, rv *uint64_value, closed bool,
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

//  op: left_ui64 == right_ui64

func (flo *flow) eq_ui64(left, right uint64_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !bv.is_null {
				bv.bool = lv.uint64 == rv.uint64
			}
			out <- bv

			flo = flo.get()
		}
	}()

	return out
}

//  unbound version of flo.eq_ui64(), in operator table

func eq_ui64(flo *flow, left, right uint64_chan) (out bool_chan) {
	return flo.eq_ui64(left, right)
}

//  op: left_ui64 != right_ui64

func (flo *flow) neq_ui64(left, right uint64_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !bv.is_null {
				bv.bool = lv.uint64 != rv.uint64
			}
			out <- bv

			flo = flo.get()
		}
	}()

	return out
}

//  unbound version of flo.neq_ui64(), in operator table

func neq_ui64(flo *flow, left, right uint64_chan) (out bool_chan) {
	return flo.neq_ui64(left, right)
}

//  op: left_ui64 > right_ui64

func (flo *flow) gt_ui64(left, right uint64_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !bv.is_null {
				bv.bool = lv.uint64 > rv.uint64
			}
			out <- bv

			flo = flo.get()
		}
	}()

	return out
}

//  unbound version of flo.gt_ui64(), in operator table

func gt_ui64(flo *flow, left, right uint64_chan) (out bool_chan) {
	return flo.gt_ui64(left, right)
}

//  op: left_ui64 >= right_ui64

func (flo *flow) gte_ui64(left, right uint64_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !bv.is_null {
				bv.bool = lv.uint64 >= rv.uint64
			}
			out <- bv

			flo = flo.get()
		}
	}()

	return out
}

//  unbound version of flo.gte_ui64(), in operator table

func gte_ui64(flo *flow, left, right uint64_chan) (out bool_chan) {
	return flo.gte_ui64(left, right)
}

//  op: left_ui64 < right_ui64

func (flo *flow) lt_ui64(left, right uint64_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !bv.is_null {
				bv.bool = lv.uint64 < rv.uint64
			}
			out <- bv

			flo = flo.get()
		}
	}()

	return out
}

//  unbound version of flo.lt_ui64(), in operator table

func lt_ui64(flo *flow, left, right uint64_chan) (out bool_chan) {
	return flo.lt_ui64(left, right)
}

//  op: left_ui64 <= fight_ui64

func (flo *flow) lte_ui64(left, right uint64_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if bv.is_null == false {
				bv.bool = lv.uint64 <= rv.uint64
			}
			out <- bv

			flo = flo.get()
		}
	}()

	return out
}

//  unbound version of flo.lte_ui64(), in operator table

func lte_ui64(flo *flow, left, right uint64_chan) (out bool_chan) {
	return flo.lte_ui64(left, right)
}

//  op: left_ui64 + right_ui64
//
//  Note: no overflow!

func (flo *flow) add_ui64(left, right uint64_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {

		for {
			defer close(out)

			flo = flo.get()

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			uiv := &uint64_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !uiv.is_null {
				uiv.uint64 = lv.uint64 + rv.uint64
			}
			out <- uiv

			flo = flo.get()
		}
	}()

	return out
}

//  op: left_ui64 * right_ui64
//  Note: no overflow!

func (flo *flow) mul_ui64(left, right uint64_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {

		for {
			defer close(out)

			flo = flo.get()

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			uiv := &uint64_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !uiv.is_null {
				uiv.uint64 = lv.uint64 * rv.uint64
			}
			out <- uiv

			flo = flo.get()
		}
	}()

	return out
}

//  op: left_ui64 - right_ui64
//  Note: no underflow!  should l <= r be enforced!

func (flo *flow) sub_ui64(left, right uint64_chan) (out uint64_chan) {

	out = make(uint64_chan)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			uiv := &uint64_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !uiv.is_null {
				uiv.uint64 = lv.uint64 - rv.uint64
			}
			out <- uiv

			flo = flo.get()
		}
	}()

	return out
}

//  op: send constant ui64

func (flo *flow) const_ui64(u64 uint64) (out uint64_chan) {

	out = make(uint64_chan)
	go func() {
		for {
			out <- &uint64_value{
				uint64:	u64,
				is_null: false,
			}
			flo = flo.get()
		}
	}()

	return out
}

//  op: cast ui64 to string

func (flo *flow) cast_uint64(in uint64_chan) (out string_chan) {

	out = make(string_chan)
	go func() {
		for {
			var s string

			uiv := <- in
			if uiv.is_null == false { 
				s = strconv.FormatUint(uiv.uint64, 10)
			}
			out <- &string_value{
				string:	s,
				is_null:uiv.is_null,
			}
			flo = flo.get()
		}
	}()

	return out
}

//  op: is a ui64 value null?

func (flo *flow) is_null_uint64(in uint64_chan) (out bool_chan) {

	out = make(bool_chan)
	go func() {
		for {
			out <- &bool_value{
				bool:	(<-in).is_null,
			}

			flo = flo.get()
		}
	}()

	return out
}

//  op: is ui64 value not null?

func (flo *flow) is_not_null_uint64(in uint64_chan) (out bool_chan) {

	out = make(bool_chan)
	go func() {
		for {
			out <- &bool_value{
				bool:	(<-in).is_null == false,
			}

			flo = flo.get()
		}
	}()

	return out
}

//  is a ast node ui64 value?

func (a *ast) is_uint64() bool {

	switch a.yy_tok {
	case UINT64,
		PROJECT_OSX_EXIT_CODE,
		PROJECT_OSX_PID,
		PROJECT_OSX_WALL_DURATION,
		PROJECT_OSX_USER_SEC,
		PROJECT_OSX_USER_USEC,
		PROJECT_OSX_SYS_SEC,
		PROJECT_OSX_SYS_USEC:
		return true
	}
	return false
}

//  Stringfy a ui64 value, possibly null

func (uv *uint64_value) String() string {

	if uv == nil {
		return "uint64_value(nil)"
	}
	if uv.is_null {
		return "NULL"
	}
	return strconv.FormatUint(uv.uint64, 10)
}
