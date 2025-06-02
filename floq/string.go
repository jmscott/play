package main

import "strings"

type string_value struct {
	string

	is_null		bool

	*flow
}

type string_chan chan *string_value


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

func (flo *flow) strcat(left, right string_chan) (out string_chan) {

	out = make(string_chan)

	go func() {

		for {
			defer close(out)

			flo = flo.get()

			lv, rv, done := left.wait2(right)
			if done {
				return
			}

			sv := &string_value {
				is_null:	lv.is_null || rv.is_null,
				flow:		flo,
			}
			if !sv.is_null {
				var b strings.Builder

				b.WriteString(lv.string)
				b.WriteString(rv.string)
				sv.string = b.String()
			}
			out <- sv
		}
	}()

	return out
}

//  compare two strings for equality

func (flo *flow) eq_string(left, right string_chan) (out bool_chan) {

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
				bv.bool = lv.string == rv.string
			}
			out <- bv
		}
	}()

	return out
}

//  compare two strings for equality

func (flo *flow) neq_string(left, right string_chan) (out bool_chan) {

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
				bv.bool = lv.string != rv.string
			}
			out <- bv
		}
	}()

	return out
}

//  compare two strings for left lexically greater than right

func (flo *flow) gt_string(left, right string_chan) (out bool_chan) {

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
				bv.bool = lv.string > rv.string
			}
			out <- bv
		}
	}()

	return out
}

//  compare two strings for left lexically greater than or equal to right

func (flo *flow) gte_string(left, right string_chan) (out bool_chan) {

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
				bv.bool = lv.string >= rv.string
			}
			out <- bv
		}
	}()

	return out
}

//  compare two strings for left lexically less than right

func (flo *flow) lt_string(left, right string_chan) (out bool_chan) {

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
				bv.bool = lv.string < rv.string
			}
			out <- bv
		}
	}()

	return out
}

//  compare two strings for left lexically less than or equal to right

func (flo *flow) lte_string(left, right string_chan) (out bool_chan) {

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
				bv.bool = lv.string <= rv.string
			}
			out <- bv
		}
	}()

	return out
}

func (flo *flow) const_string(s string) (out string_chan) {

	out = make(string_chan)
	go func() {
		for {
			flo = flo.get()

			out <- &string_value{
				string:	s,
				is_null: false,
				flow:	flo,
			}
		}
	}()

	return out
}

func (a *ast) is_string() bool {

	if a.yy_tok == STRING || a.yy_tok == CONCAT {
		return true
	}
	return false
}
