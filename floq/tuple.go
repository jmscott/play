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
	name		string

	att_ref		*attribute
	sysatt_ref	*sysatt
	call_order	uint8
}

//  build a tuple struct from a "DEFINE TUPLE" abstract syntax tree.

func new_tuple(name string, define *ast) (*tuple, error) {

	tup := &tuple{
		name:	name,
	}

	tup.atts = make(map[string]*attribute)
	for a := define.left;  a != nil;  a = a.next {
		switch a.yy_tok {
		case yy_SET:
		case STRING:
		case ARRAY:
		default:
			a.corrupt("new_tuple: node not set element: %s", a)
		}
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
		what = "unknown att type"
	}
	return fmt.Sprintf(
			"projection: %s (cord=%d): %s",
			proj.name, 
			proj.call_order,
			what,
	)
}
