package main

type uint64_value struct {
	uint64

	is_null		bool

	*flow
}

type uint64_chan chan *uint64_value


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

func (flo *flow) eq_uint64(left, right uint64_chan) (out bool_chan) {

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
				bv.bool = lv.uint64 == rv.uint64
			}
			out <- bv
		}
	}()

	return out
}

//  compare two uinnt64 for equality

func (flo *flow) neq_uint64(left, right uint64_chan) (out bool_chan) {

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
				bv.bool = lv.uint64 != rv.uint64
			}
			out <- bv
		}
	}()

	return out
}

//  compare two uint64s for left lexically greater than right

func (flo *flow) gt_uint64(left, right uint64_chan) (out bool_chan) {

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
				bv.bool = lv.uint64 > rv.uint64
			}
			out <- bv
		}
	}()

	return out
}

//  compare two uint64s for left lexically greater than or equal to right

func (flo *flow) gte_uint64(left, right uint64_chan) (out bool_chan) {

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
				bv.bool = lv.uint64 >= rv.uint64
			}
			out <- bv
		}
	}()

	return out
}

//  compare two uint64s for left lexically less than right

func (flo *flow) lt_uint64(left, right uint64_chan) (out bool_chan) {

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
				bv.bool = lv.uint64 < rv.uint64
			}
			out <- bv
		}
	}()

	return out
}

//  compare two uint64s for left lexically less than or equal to right

func (flo *flow) lte_uint64(left, right uint64_chan) (out bool_chan) {

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
				bv.bool = lv.uint64 <= rv.uint64
			}
			out <- bv
		}
	}()

	return out
}

func (flo *flow) add_uint64(left, right uint64_chan) (out uint64_chan) {

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
				flow:		flo,
			}
			if !uiv.is_null {
				uiv.uint64 = lv.uint64 + rv.uint64
			}
			out <- uiv
		}
	}()

	return out
}

func (flo *flow) mul_uint64(left, right uint64_chan) (out uint64_chan) {

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
				flow:		flo,
			}
			if !uiv.is_null {
				uiv.uint64 = lv.uint64 * rv.uint64
			}
			out <- uiv
		}
	}()

	return out
}

func (flo *flow) sub_uint64(left, right uint64_chan) (out uint64_chan) {

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
				flow:		flo,
			}
			if !uiv.is_null {
				uiv.uint64 = lv.uint64 - rv.uint64
			}
			out <- uiv
		}
	}()

	return out
}

func (flo *flow) const_uint64(u64 uint64) (out uint64_chan) {

	out = make(uint64_chan)
	go func() {
		for {
			flo = flo.get()

			out <- &uint64_value{
				uint64:	u64,
				is_null: false,
				flow:	flo,
			}
		}
	}()

	return out
}

func (a *ast) is_uint64() bool {

	return a.yy_tok == UINT64
}
