package main

import "strings"

type string_value struct {
	string

	is_null		bool
}

type string_chan chan *string_value

type relop_str_func func (*flow, string_chan, string_chan) bool_chan
var relop_string = map[int]relop_str_func{
		GT:	gt_string,
		GTE:	gte_string,
		EQ:	eq_string,
		NEQ:	neq_string,
		LTE:	lte_string,
		LT:	lt_string,
	}

//  wait for left and right hand strings of any binary operator
//
//  Note: how does passing *string_value compare to string_value?

func (left string_chan) wait2(right string_chan) (
	lv, rv *string_value, closed bool,
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

//  cheap sanity test

func (out string_chan) frisk(left, right string_chan) {
	if left == nil {
		corrupt("left string_chan is nil")
	}
	if right == nil {
		corrupt("right string_chan is nil")
	}
}

func (flo *flow) concat(left, right string_chan) (out string_chan) {

	out = make(string_chan)
	out.frisk(left, right)

	go func() {

		for {
			defer close(out)

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			sv := &string_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !sv.is_null {
				var b strings.Builder

				b.WriteString(lv.string)
				b.WriteString(rv.string)
				sv.string = b.String()
			}
			out <- sv
			flo = flo.get()
		}
	}()

	return out
}


//  compare two strings for equality

func (flo *flow) eq_string(left, right string_chan) (out bool_chan) {

	out = make(bool_chan)
	out.frisk_str(left, right)

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
				bv.bool = lv.string == rv.string
			}
			out <- bv

			flo = flo.get()
		}
	}()

	return out
}

func eq_string(flo *flow, left, right string_chan) (out bool_chan) {
	return flo.eq_string(left, right)
}

//  compare two strings for equality

func (flo *flow) neq_string(left, right string_chan) (out bool_chan) {

	out = make(bool_chan)
	out.frisk_str(left, right)

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
				bv.bool = lv.string != rv.string
			}
			out <- bv
			flo = flo.get()
		}
	}()

	return out
}

func neq_string(flo *flow, left, right string_chan) (out bool_chan) {
	return flo.neq_string(left, right)
}

//  compare two strings for left lexically greater than right

func (flo *flow) gt_string(left, right string_chan) (out bool_chan) {

	out = make(bool_chan)
	out.frisk_str(left, right)

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
				bv.bool = lv.string > rv.string
			}
			out <- bv

			flo = flo.get()
		}
	}()

	return out
}

func gt_string(flo *flow, left, right string_chan) (out bool_chan) {
	return flo.gt_string(left, right)
}

//  compare two strings for left lexically greater than or equal to right

func (flo *flow) gte_string(left, right string_chan) (out bool_chan) {

	out = make(bool_chan)
	out.frisk_str(left, right)

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
				bv.bool = lv.string >= rv.string
			}
			out <- bv

			flo = flo.get()
		}
	}()

	return out
}

func gte_string(flo *flow, left, right string_chan) (out bool_chan) {
	return flo.gte_string(left, right)
}

//  compare two strings for left lexically less than right

func (flo *flow) lt_string(left, right string_chan) (out bool_chan) {

	out = make(bool_chan)
	out.frisk_str(left, right)

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
				bv.bool = lv.string < rv.string
			}
			out <- bv

			flo = flo.get()
		}
	}()

	return out
}

func lt_string(flo *flow, left, right string_chan) (out bool_chan) {
	return flo.lt_string(left, right)
}

//  compare two strings for left lexically less than or equal to right

func (flo *flow) lte_string(left, right string_chan) (out bool_chan) {

	out = make(bool_chan)
	out.frisk_str(left, right)

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
				bv.bool = lv.string <= rv.string
			}
			out <- bv

			flo = flo.get()
		}
	}()

	return out
}

func lte_string(flo *flow, left, right string_chan) (out bool_chan) {
	return flo.lte_string(left, right)
}

func (flo *flow) const_string(s string) (out string_chan) {

	out = make(string_chan)

	go func() {
		for {
			out <- &string_value{
				string:	s,
				is_null: false,
			}
			flo = flo.get()
		}
	}()

	return out
}

func (a *ast) is_string() bool {

	switch a.yy_tok {
	case STRING, CONCAT, EXPAND_ENV, CAST_UINT64:
		return true
	}
	return false
}
