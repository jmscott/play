package main

import "regexp"

//  Match a string against a compiled regular expression: "abc" =~ "[b]"

func (flo *flow) match(left string_chan, re *regexp.Regexp) (out bool_chan) {

	out = make(bool_chan)

	go func() {
		<-compiling

		for {
			// wait for left and right values to arrive

			lv := <- left

			bv := &bool_value {
				is_null:	lv.is_null,
			}
			if bv.is_null == false {
				bv.bool = re.MatchString(lv.string)
			}
			out <- bv

			flo = flo.next()
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
		<-compiling

		for {
			// wait for left hand string to arrive

			lv := <- left

			bv := &bool_value {
				is_null:	lv.is_null,
			}
			if bv.is_null == false {
				bv.bool = !re.MatchString(lv.string)
			}
			out <- bv

			flo = flo.next()
		}
	}()

	return out
}
