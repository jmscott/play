package main

import (
	"strings"
	"sync"
)

//  string values used by flow operations

type string_value struct {
	string

	is_null		bool
}

//  let the strings flow

type string_chan chan *string_value

//  table of boolean, relational operations on strings

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
//  Note: how does performance of passing *string_value compare to string_value?

func (left string_chan) wait2(right string_chan) (lv, rv *string_value) {
	for lv == nil || rv == nil {
		select {
			case lv = <- left:
			case rv = <- right:
		}
	}
	return
}

//  op: "left" || "right"

func (flo *flow) concat(left, right string_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		<-compiling

		for {
			lv, rv := left.wait2(right)

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

			flo = flo.next()
		}
	}()

	return out
}

//  op: "left" == "right"

func (flo *flow) eq_string(left, right string_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			lv, rv := left.wait2(right)

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !bv.is_null {
				bv.bool = lv.string == rv.string
			}
			out <- bv

			flo = flo.next()
		}
	}()

	return out
}

//  unbound version of flow.eq_string() for global init table

func eq_string(flo *flow, left, right string_chan) (out bool_chan) {
	return flo.eq_string(left, right)
}

//  op: "left" != "right"

func (flo *flow) neq_string(left, right string_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			lv, rv := left.wait2(right)

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !bv.is_null {
				bv.bool = lv.string != rv.string
			}
			out <- bv
			flo = flo.next()
		}
	}()

	return out
}

//  neq_string() is an unbound version of flow.neq_string(),
//  for global table.  see init()
func neq_string(flo *flow, left, right string_chan) (out bool_chan) {
	return flo.neq_string(left, right)
}

//  op: "left" > "right", lexically

func (flo *flow) gt_string(left, right string_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			lv, rv := left.wait2(right)

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !bv.is_null {
				bv.bool = lv.string > rv.string
			}
			out <- bv

			flo = flo.next()
		}
	}()

	return out
}

func gt_string(flo *flow, left, right string_chan) (out bool_chan) {
	return flo.gt_string(left, right)
}

//  op: "left" >= "right"

func (flo *flow) gte_string(left, right string_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			lv, rv := left.wait2(right)

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !bv.is_null {
				bv.bool = lv.string >= rv.string
			}
			out <- bv

			flo = flo.next()
		}
	}()

	return out
}

func gte_string(flo *flow, left, right string_chan) (out bool_chan) {
	return flo.gte_string(left, right)
}

//  op: "left" < "right"

func (flo *flow) lt_string(left, right string_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			lv, rv := left.wait2(right)

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !bv.is_null {
				bv.bool = lv.string < rv.string
			}
			out <- bv

			flo = flo.next()
		}
	}()

	return out
}

func lt_string(flo *flow, left, right string_chan) (out bool_chan) {
	return flo.lt_string(left, right)
}

//  op: "left" <= "right"

func (flo *flow) lte_string(left, right string_chan) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			lv, rv := left.wait2(right)

			bv := &bool_value {
				is_null:	lv.is_null || rv.is_null,
			}
			if !bv.is_null {
				bv.bool = lv.string <= rv.string
			}
			out <- bv

			flo = flo.next()
		}
	}()

	return out
}

func lte_string(flo *flow, left, right string_chan) (out bool_chan) {
	return flo.lte_string(left, right)
}

//  op: send constant string value, never null

func (flo *flow) const_string(s string) (out string_chan) {

	out = make(string_chan)

	go func() {
		<-compiling

		for {
			out <- &string_value{
				string:	s,
				is_null: false,
			}
			flo = flo.next()
		}
	}()

	return out
}

//  is the ast node a string value?

func (a *ast) is_string() bool {

	switch a.yy_tok {
	case STRING, CONCAT, EXPAND_ENV,
	     PROJECT_OSX_START_TIME,
	     PROJECT_OSX_STDOUT,
	     PROJECT_OSX_TUPLE_TSV,
	     PROJECT_OSX_TUPLE_TSV_N,
	     PROJECT_OSX_STDERR,
	     PROJECT_TSV:
		return true
	case CAST, CAST_UINT64, CAST_BOOL, CAST_STRING:
		if a.right.yy_tok == yy_STRING {
			return true
		}
	}
	return false
}

//  no-op to cast a string to itself.
//
//  Note: eventually will optimize out this cast, in pass2.

func (flo *flow) cast_string(in string_chan) (out string_chan) {

	out = make(string_chan)

	go func() {
		<-compiling

		for {
			out <-<-in
			flo = flo.next()
		}
	}()
	return out
}

//  op: is a string value null?

func (flo *flow) is_null_string(in string_chan) (out bool_chan) {

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

//  op: is a string value not null?

func (flo *flow) is_not_null_string(in string_chan) (out bool_chan) {

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

//  Stringifier that handles nill and null.

func (sv *string_value) String() string {

	if sv == nil {
		return "string_value(nil)"
	}
	if sv.is_null {
		return "NULL"
	}
	return sv.string
}

//  op: write string value to null channel

func (flo *flow) string_null(in string_chan) {

	go func() {
		<-compiling

		for {
			<- in

			flo = flo.next()
		}
	}()
}

//  Project a string inside tab separated line of fields.
//  The field is referenced by field number, offset from 1.
//  A field out of bounds send a null string.

func (flo *flow) project_tsv(in string_chan, field uint8) (out string_chan) {
	
	out = make(string_chan)

	go func() {
		<-compiling

		for {
			sv := <- in

			var str string

			is_null := sv.is_null
			if is_null == false {
				fld := strings.Split(
						strings.Trim(
							sv.string,
							"\n"),
						"\t",
					)
				if int(field) <= len(fld) {
					str = fld[field-1]
				} else {
					is_null = true
				}
			}
			out <- &string_value{
				string:		str,
				is_null:	is_null,
			}

			flo = flo.next()
		}
	}()
	return out
}

//  op: fanout a single string value to multiple channels

func (flo *flow) string_fo(in string_chan, count uint8) (out []string_chan) {

	out = make([]string_chan, count)
	for i := uint8(0); i < count; i++ {
		out[i] = make(string_chan)
	}

	go func() {
		<-compiling

		for {
			sv := <-in

			//  broadcast to channels in output slice

			var wg sync.WaitGroup

			wg.Add(int(count))
			for _, sc := range out {
				go func() {
					sc <- sv
					wg.Done()
				}()
			}
			wg.Wait()

			flo = flo.next()
		}
	}()
	return out
}
