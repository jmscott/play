package main

import (
	"errors"
	"fmt"
	"os"
)

//  abstract syntax tree of floq program defined in yacc grammer in parser.y

type ast struct {

	//  generated token in yyToknames[]

	yy_tok		int

	//  approximate line number in *.floq file

	line_no		uint32

	//  natural name associated with ast node, for debugging.
	//  for example, a COMMAND_REF node would be the command name

	name		string

	//  parsing order of individual ast nodes of particular list structure,
	//  like statements or call argument vector

	order		uint32
	
	//  total count of child ast nodes of particular structure. 

	count		uint32

	//  children of ast node.

	left		*ast
	right		*ast

	//  siblings ast nodes

	next		*ast
	prev		*ast

	//  parent ast node, never null except for FLOQ node

	parent		*ast

	//  track various datum built during parsing.  explcite structure
	//  members prefered over single interface{} for readabilty.
	//
	//  too bad golang has no union type.

	proj_ref	*projection
	tuple_ref	*tuple
	command_ref	*command
	att_ref		*attribute
	uint64
	string
	bool
	set_ref		*set
	array_ref	[]string
}

//  name made by yacc grammar of token associated with ast node

func (a *ast) yy_name() string {
	if a == nil {
		return "nil"
	}
	return yy_name(a.yy_tok)
}

//  overly complex String() for an ast node.
//
//  Note: rename to ast.dump() and simplyfy String()

func (a *ast) String() string {

	var what string

	if a == nil {
		return "*ast=nil"
	}

	//  colon separator indicates existence of left/right kids

	var colon string
	if a.left == nil {
		if a.right == nil {
			colon = ": "
		} else {
			colon = "\\ "
		}
	} else {
		if a.right != nil {
			colon = "/\\ "
		} else {
			colon = "/ "
		}
	}

	//  format details of specific ast nodes

	switch a.yy_tok {
	case 0:
		a.corrupt("ast has yy_tok == 0")
	case ARGV:
		what = fmt.Sprintf("ARGV%s (cnt=%d)", colon, a.count)
	case ARRAY:
		if a.name == "" {
			what = fmt.Sprintf(
				"ARRAY%s(cnt=%d) @%p",
				colon,
				a.count,
				len(a.array_ref),
				cap(a.array_ref),
				a.array_ref,
			)
		} else {
			what = fmt.Sprintf(
				"ARRAY%s%s (cnt=%d) (len=%d,cap=%d) @%p",
				colon,
				a.name,
				a.count,
				len(a.array_ref),
				cap(a.array_ref),
				a.array_ref,
			)
		}

	case DEFINE:
		what = fmt.Sprintf(
			"DEFINE%s(ord=%d, lno=%d)",
			colon,
			a.order,
			a.line_no,
		)
	case RUN:
		nm := a.command_ref.name
		tup := a.command_ref.tuple_ref
		if tup != nil {
			nm = fmt.Sprintf("%s.%s", nm, tup.name)
		}
		what = fmt.Sprintf(
				"RUN%s%s (ord=%d,lno=%d) ",
				colon,
				nm,
				a.order,
				a.line_no,
			)
		what += a.command_ref.detail(2)
	case FLOW:
		cmd := a.command_ref
		what = fmt.Sprintf(
				"FLOW%s%s (ord=%d,lno=%d) ",
				colon,
				cmd.name,
				a.order,
				a.line_no,
			)
		what += cmd.detail(2)
	case STMT_LIST:
		what = fmt.Sprintf("STMT_LIST%s(cnt=%d)", colon, a.count)
	case yy_SET:
		var crc64 string
		if a.set_ref != nil {
			crc64 = fmt.Sprintf(
					"ec=%d crc=%s",
					a.set_ref.count(),
					a.set_ref.crc64_brief(5, true),
				)
		}
		if a.name == "" {
			what = fmt.Sprintf(
					"SET%s%s@%p",
						colon,
						crc64,
						a.set_ref,
					)
		} else {
			what = fmt.Sprintf(
					"SET%s%s %s@%p",
					colon,
					a.name,
					crc64,
					a.set_ref,
			)
		}
	case STRING:
		if a.name == "" {
			what = fmt.Sprintf(
					"STRING \"%s\"",
					a.string,
			)
		} else {
			what = fmt.Sprintf(
					"STRING:%s \"%s\"",
					a.name,
					a.string,
			)
		}
	case UINT64:
		if a.name == "" {
			what = fmt.Sprintf("UINT64 %d", a.uint64)
		} else {
			what = fmt.Sprintf("UINT64:%s %d", a.name, a.uint64)
		}
	case yy_AND:
		what = "AND"
	case yy_OR:
		what = "OR"
	case yy_FALSE:
		what = "FALSE"
	case yy_TRUE:
		what = "TRUE"
	case PROJECT_OSX_EXIT_CODE,
	     PROJECT_OSX_PID,
	     PROJECT_OSX_START_TIME,
	     PROJECT_OSX_WALL_DURATION,
	     PROJECT_OSX_USER_SEC,
	     PROJECT_OSX_USER_USEC,
	     PROJECT_OSX_SYS_SEC,
	     PROJECT_OSX_SYS_USEC:
		what = fmt.Sprintf(
			"%s: %s",
			a.yy_name(),
			a.proj_ref,
		)
	case PROJECT_OSX_TUPLE_TSV:
		what = "PROJECT_OSX_TUPLE_TSV: " + a.proj_ref.String()
	case PROJECT_OSX_TUPLE_TSV_N:
		what = fmt.Sprintf(
				"PROJECT_OSX_TUPLE_TSV_N: %s[%d]",
				a.command_ref,
				a.proj_ref.field,
			)
	default:
		what = fmt.Sprintf("%s%s%s", a.yy_name(), colon, a.name)
	}
	return what
}

//  convert an abstract syntax tree to a set defined in set.go

func (aset *ast) parse_set() (*set, error) {

	_e := func(format string, args...interface{}) (*set, error) {
		return nil, fmt.Errorf("frisk_set: " + format, args...)
	}

	if aset.yy_tok != yy_SET {
		return _e("root not a set: %s", yy_name(aset.yy_tok))
	}

	//  find duplicate elements starting at left branch

	set := new_set()
	for ele := aset.left;  ele != nil;  ele = ele.next {
		switch ele.yy_tok {
		case yy_TRUE, yy_FALSE:
		case UINT64:
		case STRING:
		case yy_SET:
			if _, err := ele.parse_set();  err != nil {
				return nil, err
			}
		default:
			aset.corrupt("impossible set element: %s", ele)
		}
	}
	 
	return set, nil
}

//  recursively print arbitrary ast nodes and descendents,
//  tracking node depth for indentation

func (a *ast) walk_print(indent int, parent *ast) {

	if a == nil {
		return
	}
	if parent != nil && a.parent != parent {
		if a.parent == nil {
			a.corrupt("unexpected nil ast parent")
		}
		a.corrupt(
			"call parent(%s) not ast parent: %s",
			parent.yy_name(),
			a.parent.yy_name(),
		)
	}
	if indent == 0 {
		os.Stderr.WriteString("")
	} else {
		if a.parent == nil {
			a.corrupt("indent > 0: ast parent is nil")
		}
		for i := 0;  i < indent;  i++ {
			os.Stderr.WriteString("  ")
		}
	}

	os.Stderr.WriteString(a.String() + "\n")

	//  print kids

	a.left.walk_print(indent + 1, a)
	a.right.walk_print(indent + 1, a)

	//  print siblings

	if a.prev == nil {
		for as := a.next;  as != nil;  as = as.next {
			as.walk_print(indent, parent)
		}
	}
}

//  recursively print arbitrary ast nodes and descendents, starting with
//  no indentation.

func (a *ast) print() {
	a.walk_print(0, nil)
}

//  panic on a impossible ast node, typically in cheap sanity test.
//
//  Note: why not rename to ast.panic()?

func (a *ast) corrupt(format string, args...interface{}) {

	msg := fmt.Sprintf(format, args...)
	die("%s: node \"%s\", near line %d", msg, a.yy_name(), a.line_no)
	//  NOTREACHED*/
}

//  is a yy token of an ast node in vararg list of yy tokens?

func (a *ast) in_tok_set(expect ...int) bool {

	for _, tok := range expect {
		if tok == a.yy_tok {
			return true
		}
	}
	return false
}

//  error specific to an ast node

func (a *ast) error(format string, args...interface{}) error {

	emsg := fmt.Sprintf(format, args...)
	if a.line_no == 0 {
		return errors.New(emsg)
	}
	return fmt.Errorf("%s, near line %d", emsg, a.line_no)
}

//  recursively count ast nodes of particular types in
//  a vararg list of yy_tokens.  for example,
//
//	root.yy_count(BOOL, UINT64, STRING)
//
//  counts all descendent constants of type bool, uint64 or string

func (a *ast) yy_count(tokens ...int) int {
	
	if a == nil {
		return 0
	}
	count := 0
	if a.in_tok_set(tokens...) {
		count++
	}
	count += a.left.yy_count(tokens...)
	count += a.right.yy_count(tokens...)

	if a.prev == nil {
		for kid := a.next;  kid != nil;  kid = kid.next {
			count += kid.yy_count(tokens...)
		}
	}
	return count
}

//  new left node for an ast node

func (parent *ast) push_left(kid *ast) {

	parent.push_lr(&parent.left, kid)
}

//  new right node for an ast node

func (parent *ast) push_right(kid *ast) {

	parent.push_lr(&parent.right, kid)
}

//  new left or right node for an ast node

func (parent *ast) push_lr(lr **ast, kid *ast) {

	var k *ast

	kid.parent = parent
	if *lr == nil {
		*lr = kid
		kid.order = 1
		parent.count = 1
		return
	}
	for k = *lr;  k.next != nil;  k = k.next {}
	k.next = kid
	kid.order = k.order + 1
	parent.count++
	kid.prev = k
}

//  find a named STRING element in an ast SET

func (set *ast) string_element(name string) string { 
	if set.yy_tok != yy_SET {
		set.corrupt("ast: expected SET, got %s", set.yy_name())
	}

	for kid := set.left;  kid != nil;  kid = kid.next {
		if kid.name != name {
			continue
		}
		if kid.yy_tok == STRING {
			return kid.string
		}
	}
	return ""
}

//  get a named []string from ARRAY element in a SET

func (set *ast) array_string_element(name string) []string { 

	if set.yy_tok != yy_SET {
		set.corrupt("expected SET, got %s", set.yy_name())
	}

	var kid *ast

	for kid = set.left;  kid != nil;  kid = kid.next {
		if kid.name == name && kid.yy_tok == ARRAY {
			break
		}
	}
	if kid == nil {
		return nil
	}
	ar := make([]string, kid.count)
	var cnt int
	for kid = kid.left;  kid != nil;  kid = kid.next {
		if kid.yy_tok == STRING {
			ar[cnt] = kid.string
			cnt++
		}
	}
	return ar[:cnt]
}

//  find first ancestor of particular type

func (a *ast) yy_ancestor(yy_tok int) *ast {

	if a == nil {
		return nil
	}
	if a.yy_tok == yy_tok {
		return a
	}
	if a.parent == nil {
		return nil
	}
	return a.parent.yy_ancestor(yy_tok)
}
