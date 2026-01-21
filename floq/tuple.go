package main

import (
	"errors"
	"fmt"
	"regexp"
)

type attribute struct {
	name		string
	matches		*regexp.Regexp
	tuple_ref	*tuple
}

type tuple struct {
	name		string
			/*
			 *  Note:
			 *	we can not use attribute{}.  get wierd error
			 *	
			 *		"struct containing regexp.Regexp
			 *		cannot be compared"
			 *
			 */
	atts		map[string]*attribute
	tsv_line	[]*attribute
}

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
		tup.atts[a.name] = &attribute{
					name:		a.name,
					tuple_ref:	tup,
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
			return ea("can not compile re: %s", mat.string)
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
	if tsv_line.count != atts.count {
		return et(
			"attributes counts do not match: %d != %d", 
			atts.count,
			tsv_line.count,
		)
	}
	tup.tsv_line = make([]*attribute, tsv_line.count)
	for t := tsv_line.left;  t != nil;  t = t.next {
		if t.yy_tok != STRING {
			return et(
				"array element not string: %s",
				yy_name(t.yy_tok),
			)
		}
	}

	return tup, nil
}
