package main

import "regexp"

//  match string against compiled regular expression
//
//	"abc" =~ "[b]"

func (flo *flow) match(left string_chan, re *regexp.Regexp) (out bool_chan) {

	out = make(bool_chan)

	go func() {

		for {
			defer close(out)

			var lv *string_value

			// wait for left and right values to arrive

			lv = <- left
			if lv == nil {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null,
				flow:		flo,
			}
			if bv.is_null == false {
				bv.bool = re.MatchString(lv.string)
			}
			out <- bv
			flo = flo.get()
		}
	}()

	return out
}

//  negate match string against compiled regular expression
//
//	"abc" =~ "[b]"

func (flo *flow) nomatch(left string_chan, re *regexp.Regexp) (out bool_chan) {

	out = make(bool_chan)

	go func() {

		for {
			defer close(out)

			var lv *string_value

			// wait for left and right values to arrive

			lv = <- left
			if lv == nil {
				return
			}

			bv := &bool_value {
				is_null:	lv.is_null,
				flow:		flo,
			}
			if bv.is_null == false {
				bv.bool = !re.MatchString(lv.string)
			}
			out <- bv

			flo = flo.get()
		}
	}()

	return out
}
