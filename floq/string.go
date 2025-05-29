package main

import "strings"

type string_value struct {
	string

	is_null		bool

	*flow
}

type string_chan chan *string_value

func (flo *flow) strcat(left, right string_chan) (out string_chan) {

	out = make(string_chan)

	go func() {

		for {
			defer close(out)

			flo = flo.get()

			var lv, rv *string_value

			// wait for left and right sides to arrive

			for lv == nil || rv == nil {
				select {
				case lv = <- left:
					if lv == nil {
						return
					}
				case rv = <- right:
					if rv == nil {
						return
					}
				}
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
