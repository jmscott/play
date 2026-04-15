package main

import (
	"fmt"
	"regexp"
)

//  an attribute object in tuple.attributes set in "define tuple"
//  statement

type attribute struct {
	name		string
	matches		*regexp.Regexp
	tsv_field	uint8			//  field offset in tsv row
	tuple_ref	*tuple			//  points to "define tuple"
}

type tuple struct {
	name		string
	atts		map[string]*attribute
}

//  project value of a particular command

type projection struct {
	command_ref	*command
	att_ref		*attribute
	sysatt_ref	*sysatt
	field		uint8

	call_order	uint8
}

//  build a tuple struct from a "DEFINE TUPLE" abstract syntax tree.
//  the "set" has already been frisked.
//
//  Note: no test for duplicate attributes in the define set!

func new_tuple(name string, define *ast) (*tuple, error) {

	tup := &tuple{
		name:	name,
	}

	_e := func(format string, args ...interface{}) (*tuple, error) {
		return nil, fmt.Errorf(format, args...)
	}

	tup.atts = make(map[string]*attribute)

	//  find ast node "attributes"

	atts := define.left

	//  Note: assume no empty set!
	if atts.name != "attributes" {
		return _e("unknown element: %s", define.left.name)
	}
	if atts.next != nil {
		return _e("too many elements: %s", atts.next.name)
	}

	prev_fld := uint8(0)	// Note: variable disappears inside for{}!

	//  build attribute set from element "attributes"

	for as := atts.left;  as != nil;  as = as.next {

		_et := func(fmt string, args ...interface{}) (*tuple, error) {
			fmt = as.name + ": tsv_field: " + fmt
			return _e(fmt, args...)
		}

		_em := func(fmt string, args ...interface{}) (*tuple, error) {
			fmt = as.name + ": matches: " + fmt
			return _e(fmt, args...)
		}

		at := &attribute{
			name:		as.name,
			tuple_ref:	tup,
		}

		has_tsv_field := false
		has_matches := false

		//  only elements named "matches" and "tsv_field" are valid
		for a := as.left;  a != nil;  a = a.next {
			var err error

			switch a.name {
			case "matches":

				at.matches, err = regexp.Compile(a.string)
				if err != nil {
					return _em("%s", err)
				}
				has_matches = true
			case "tsv_field":
				if a.uint64 == 0 {
					return _et("cannot be 0")
				}
				if a.uint64 > 255 {
					return _et("%d > 255", a.uint64)
				}
				
				//  Note:  tsv_field must be unique
				fld := uint8(a.uint64)
				for _, at2 := range tup.atts {
					if at2.tsv_field == fld {
						return _et(
							"%s and %s: same %d",
							at.name,
							at2.name,
							fld,
						)
					}
				}

				//  insure tsv_fields are sequential

				if prev_fld == 0 {
					if fld != 1 {
						return _et(
							"%s: must equal 1: %d",
							at.name,
							fld,
						)
					}
				} else {
					if fld != prev_fld + 1 {
						return _et(
							"%s: out of order: %d",
							at.name,
							fld,
						)
					}
				}
				at.tsv_field = fld
				prev_fld = fld
				has_tsv_field = true
			default:
				return _e("unknown element: \"%s\"", a.name)
			}
		}
		if has_matches == false {
			return _em("missing element")
		}
		if has_tsv_field == false {
			return _et("missing element")
		}
		tup.atts[as.name] = at
	}
	return tup, nil
}

func (tup *tuple) String() string {

	return fmt.Sprintf("%s%#v", tup.name, tup.atts)
}

func (att *attribute) String() string {
	if att == nil {
		return "<nil attribute>"
	}
	if att.tuple_ref == nil {
		return fmt.Sprintf("%s (nil tuple)", att.name)
	}
	return fmt.Sprintf(
		"%s.%s",
		att.tuple_ref.name,
		att.name,
	)
}

func (proj *projection) String() string {

	var what string

	if proj.att_ref != nil {
		what = proj.att_ref.String()
	} else if proj.sysatt_ref != nil {
		what = proj.sysatt_ref.String()
	} else {
		what = "att/sys ref both nil"
	}
	return fmt.Sprintf(
			"%s: (cord=%d)",
			what,
			proj.call_order,
	)
}
