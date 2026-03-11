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
	tsv_line	[]*attribute
}

type projection struct {
	att_ref		*attribute
	sysatt_ref	*sysatt

	call_order	uint8
}

//  build a tuple struct from a "DEFINE TUPLE" abstract syntax tree.
//  Note: no test for duplicate attributes in the define set!

func new_tuple(name string, define *ast) (*tuple, error) {

	tup := &tuple{
		name:	name,
	}

	_e := func(format string, args ...interface{}) (*tuple, error) {
		return nil, fmt.Errorf(format, args...)
	}

	//  errors for tsv_line element
	_et := func(format string, args ...interface{}) (*tuple, error) {
		return _e("element \"tsv_line\": " + format, args...)
	}

	tup.atts = make(map[string]*attribute)

	//  find ast node "attributes"

	var atts *ast
	for a := define.left;  a != nil;  a = a.next {
		if a.name == "attributes" {
			atts = a
			break
		}
	}
	if atts == nil {
		return _e("\"attributes\" not defined")
	}

	//  build attribute set from element "attributes"

	for as := atts.left;  as != nil;  as = as.next {
		if tup.atts[as.name] != nil {
			return _e("duplicate \"%s\"", as.name)
		}

		at := &attribute{
			name:		as.name,
			tuple_ref:	tup,
		}
		for a := as.left;  a != nil;  a = a.next {
			var err error

			switch a.name {
			case "matches":

				at.matches, err = regexp.Compile(a.string)
				if err != nil {
					return _e("\"matches\": %s", err)
				}
			default:
				return _e("unknown element: \"%s\"", a.name)
			}
		}
		tup.atts[as.name] = at
	}

	//  assign order for tab separated tuple

	var atsv *ast
	for a := define.left;  a != nil;  a = a.next {
		if a.name == "tsv_line" {
			atsv = a
			break
		}
	}
	if atsv == nil {
		return _et("not defined")
	}
	if atsv.yy_tok != ARRAY {
		return _et("not array")
	}
	if int(atsv.count) < len(tup.atts) {
		return _et("missing attrbutes")
	}
	if int(atsv.count) > len(tup.atts) {
		return _et("too many attributes")
	}

	//  find field offset for each tab separated attribute in the tuple

	for a, i := atsv.left, 0;  a != nil;  a, i = a.next, i + 1 {
		nm := a.string
		at := tup.atts[nm]
		if at == nil {
			return _et("unknown attribute: %s", nm)
		}
		if at.tsv_field > 0 {
			return _et("duplicate attribute: %s", nm)
		}
		at.tsv_field = uint8(i)
	}

	return tup, nil
}

func (tup *tuple) String() string {

	return fmt.Sprintf("%s%#v", tup.name, tup.atts)
}

func (att *attribute) String() string {
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
