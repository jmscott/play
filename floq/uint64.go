package main

import (
	"strconv"
)

type uint64_value struct {
	uint64

	is_null		bool

	*flow
}

type uint64_chan chan *uint64_value

type relop_uint64_func func (*flow, uint64_chan, uint64_chan) bool_chan
var relop_uint64 = map[int]relop_uint64_func{
		GT:	gt_ui64,
		GTE:	gte_ui64,
		EQ:	eq_ui64,
		NEQ:	neq_ui64,
		LTE:	lte_ui64,
		LT:	lt_ui64,
	}


//  wait for left and right hand uint64 of any binary operator

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

func (flo *flow) eq_ui64(left, right uint64_chan) (out bool_chan) {

	out = make(bool_chan)
	out.frisk_ui64(left, right)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
				flow:		flo,
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

func eq_ui64(flo *flow, left, right uint64_chan) (out bool_chan) {
	return flo.eq_ui64(left, right)
}

//  compare two uint64 for inequality

func (flo *flow) neq_ui64(left, right uint64_chan) (out bool_chan) {

	out = make(bool_chan)
	out.frisk_ui64(left, right)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
				flow:		flo,
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

func neq_ui64(flo *flow, left, right uint64_chan) (out bool_chan) {
	return flo.neq_ui64(left, right)
}

//  compare two uint64s for left lexically greater than right

func (flo *flow) gt_ui64(left, right uint64_chan) (out bool_chan) {

	out = make(bool_chan)
	out.frisk_ui64(left, right)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
				flow:		flo,
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

func gt_ui64(flo *flow, left, right uint64_chan) (out bool_chan) {
	return flo.gt_ui64(left, right)
}

//  compare two uint64s for left greater than or equal to right

func (flo *flow) gte_ui64(left, right uint64_chan) (out bool_chan) {

	out = make(bool_chan)
	out.frisk_ui64(left, right)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
				flow:		flo,
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

func gte_ui64(flo *flow, left, right uint64_chan) (out bool_chan) {
	return flo.gte_ui64(left, right)
}

//  compare two uint64s for left lexically less than right

func (flo *flow) lt_ui64(left, right uint64_chan) (out bool_chan) {

	out = make(bool_chan)
	out.frisk_ui64(left, right)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
				flow:		flo,
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

func lt_ui64(flo *flow, left, right uint64_chan) (out bool_chan) {
	return flo.lt_ui64(left, right)
}

//  compare two uint64s for left lexically less than or equal to right

func (flo *flow) lte_ui64(left, right uint64_chan) (out bool_chan) {

	out = make(bool_chan)
	out.frisk_ui64(left, right)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
				flow:		flo,
			}
			if !bv.is_null {
				bv.bool = lv.uint64 <= rv.uint64
			}
			out <- bv

			flo = flo.get()
		}
	}()

	return out
}
func lte_ui64(flo *flow, left, right uint64_chan) (out bool_chan) {
	return flo.lte_ui64(left, right)
}

func (out uint64_chan) frisk(left, right uint64_chan) {
	if left == nil {
		corrupt("left uint64 chan is nil")
	}
	if right == nil {
		corrupt("eight uint64 chan is nil")
	}
}

func (flo *flow) add_ui64(left, right uint64_chan) (out uint64_chan) {

	out = make(uint64_chan)
	out.frisk(left, right)

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
				flow:		flo,
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

func (flo *flow) mul_ui64(left, right uint64_chan) (out uint64_chan) {

	out = make(uint64_chan)
	out.frisk(left, right)

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
				flow:		flo,
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

func (flo *flow) sub_ui64(left, right uint64_chan) (out uint64_chan) {

	out = make(uint64_chan)
	out.frisk(left, right)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			uiv := &uint64_value {
				is_null:	lv.is_null || rv.is_null,
				flow:		flo,
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

func (flo *flow) const_ui64(u64 uint64) (out uint64_chan) {

	out = make(uint64_chan)
	go func() {
		for {
			out <- &uint64_value{
				uint64:	u64,
				is_null: false,
				flow:	flo,
			}
			flo = flo.get()
		}
	}()

	return out
}

func (flo *flow) cast_uint64(in uint64_chan) (out string_chan) {

	out = make(string_chan)
	go func() {
		for {
			uiv := <- in
			out <- &string_value{
				string:	strconv.FormatUint(uiv.uint64, 10),
				is_null:uiv.is_null,
			}
			flo = flo.get()
		}
	}()

	return out
}

func (a *ast) is_uint64() bool {

	switch a.yy_tok {
	case UINT64:
		return true
	case COMMAND_SYSATT:
		return a.command_ref.is_sysatt_uint64(a.name)
	}
	return false
}

func (flo *flow) uint64_string(in uint64_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		for {
			u64 := <- in
			if u64 == nil {
				return
			}
			sv := &string_value{}
			if u64.is_null == false {
				sv.is_null = false
				sv.string = strconv.FormatUint(u64.uint64, 10)
			}
			out <- sv

			flo = flo.get()
		}
	}()

	return out
}
