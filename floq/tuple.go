package main

import (
	"errors"
	"fmt"
	"regexp"
)

//  an attribute object in tuple.attributes set in "define tuple"
//  statement

type attribute struct {
	name		string
	matches		*regexp.Regexp
	tuple_ref	*tuple
	call_order	uint8
	tsv_field	uint8
}

type tuple struct {
	name		string
	atts		map[string]*attribute
	tsv_line	[]*attribute
}


//  build a tuple struct from a "DEFINE TUPLE" abstract syntax tree.

func new_tuple(name string, define *ast) (*tuple, error) {

	tup := &tuple{
		name:	name,
	}

	e := func(format string, args ...interface{}) (*tuple, error) {
		return nil, errors.New(fmt.Sprintf(format, args...))
	}

	ea := func(format string, args ...interface{}) (*tuple, error) {
		return nil, errors.New(
			fmt.Sprintf(
				"attributes set: " + format,
				args...,
			),
		)
	}

	et := func(format string, args ...interface{}) (*tuple, error) {
		return nil, errors.New(
			fmt.Sprintf(
				"\"tsv_line\" array: " + format,
				args...,
			),
		)
	}

	atts := define.set_element("attributes")
	if atts == nil {
		return e("can not find SET element \"attributes\"")
	}
	if atts.count == 0 {
		return ea("set is empty")
	}
	if atts.count > 255 {
		return ea("too many elements in \"attributes\" set")
	}

	/*
	 *  Build the attributes map for the tuple.
	 */
	tup.atts = make(map[string]*attribute)
	for a := atts.left;  a != nil;  a = a.next {
		if a.yy_tok != yy_SET {
			return ea("not a SET: %s", yy_name(a.yy_tok))
		}
		if a.count == 0 {
			return ea("has no elements", a.name)
		}
		if a.count != 1 {
			return ea("element count != 1")
		}
		if tup.atts[a.name] != nil {
			return ea("defined more than once: %s", a.name)
		}
		mat := a.left
		if mat.yy_tok != STRING {
			return ea(
				"%s.matches: element not string: %s",
				a.name,
				yy_name(mat.yy_tok),
			)
		}
		re, err := regexp.Compile(mat.string)
		if err != nil {
			return ea("can not compile matches re: %s", mat.string)
		}
		tup.atts[a.name] = &attribute{
					name:		a.name,
					tuple_ref:	tup,
				}
		tup.atts[a.name].matches = re
	}

	tsv_line := define.array_element("tsv_line")
	if tsv_line == nil {
		return tup, nil
	}
	if tsv_line.count == 0 {
		return et("array is empty")
	}

	//  "tsv_line" must  contain all attributes.  why?
	if tsv_line.count != atts.count {
		return et(
			"\"attributes\" and \"tsv_line\" counts: %d != %d", 
			atts.count,
			tsv_line.count,
		)
	}

	//  all members of "tsv_line" must be strings and no dups

	seen := make(map[string]bool)
	tup.tsv_line = make([]*attribute, tsv_line.count)
	for t := tsv_line.left;  t != nil;  t = t.next {
		if t.yy_tok != STRING {
			return et(
				"array element not string: %s",
				yy_name(t.yy_tok),
			)
		}
		anm := t.string
		if tup.atts[anm] == nil {
			return et("unknown attribute: %s", anm)
		}
		if seen[anm] {
			return et("duplicate attibute: %s", anm) 
		}
		fld := t.order-1
		tup.tsv_line[fld] = tup.atts[anm]
		tup.atts[anm].tsv_field = uint8(fld)
		seen[anm] = true
	}

	return tup, nil
}

func (tup *tuple) String() string {

	return fmt.Sprintf("%s%#v", tup.name, tup.atts)
}

func (att *attribute) String() string {
	return fmt.Sprintf(
		"%s.%s: cord=%d",
		att.tuple_ref.name,
		att.name,
		att.call_order,
	)
}
