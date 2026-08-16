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
	case 
		PROJ_FLOW_TSV_ATT,
		PROJ_FLOW_TSV_N,
		PROJ_OSX_START_TIME_RFC3339,
		PROJ_OSX_STDERR,
		PROJ_OSX_STDOUT_TSV_N,
		PROJ_OSX_STDOUT,
		PROJ_OSX_TSV,
		STRING, CONCAT,
		PROJ_OSX_WALL_DURATION_SECONDS,
		PROJ_OSX_USER_SECONDS,
		PROJ_OSX_SYS_SECONDS,
		EXPAND_ENV:
		return true
	case CAST, CAST_UINT64, CAST_BOOL, CAST_STRING:
		if a.right.yy_tok == yy_STRING {
			return true
		}
	case CONDITIONAL:
		return a.right.next.is_string()
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

//  ternary conditional operator (bool ? string: string)
func (flo *flow) cond3_string(
	in_test bool_chan,
	in_if_true,
	in_if_false string_chan,
) (out string_chan) {

	out = make(string_chan)

	go func() {
		<- compiling

		for {
			var bv *bool_value
			var sv_t, sv_f *string_value

			//  wait for test, if true and if false values
			for bv == nil || sv_t == nil || sv_f == nil {
				select {
					case bv = <- in_test:
					case sv_t = <- in_if_true:
					case sv_f = <- in_if_false:
				}
			}

			//  if test is null then ?: expression is null;
			//  otherwise send true or false strings.
			if bv.is_null {
				out <- &string_value{is_null:true}
			} else {
				if bv.bool {
					out <- sv_t
				} else {
					out <- sv_f
				}
			}
			flo = flo.next()
		}
	}()
	return out
}

/*
 *  Project the idx'th field of a tab separated string.
 *
 *  Note:
 *	why is this specific to a flow() statement?
 */
func (flo *flow) proj_tsv_n(
	in_str string_chan,
	in_idx uint64_chan,
  ) (out string_chan) {
	out = make(string_chan)

	go func() {
		<-compiling

		for {
			var sv *string_value
			var iv *uint64_value

			//  wait for both the string to project and the field
			//  index

			for sv == nil || iv == nil {
				select {
				case s := <-in_str:
					if sv != nil {
						die("input out of sync: string")
					}
					sv = s
				case i := <- in_idx:
					if iv != nil {
						die("input out of sync: index")
					}
					if i.is_null == false && i.uint64 == 0 {
						die("tsv field is 0")
					}
					iv = i
				}
			}

			if sv.is_null == false {
				idx := int(iv.uint64) - 1

				fld := strings.Split(sv.string, "\t")
				if idx < len(fld) {
					sv.string = fld[idx]
				} else {
					sv.is_null = true
				}
			}
			out <- sv

			flo = flo.next()
		}
	}()
	return
}
