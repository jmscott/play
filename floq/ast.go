//  abstract syntax tree generated by yacc grammar
package main

import (
	"fmt"
)

type ast struct {

	yy_tok		int
	line_no		int

	//  children
	left		*ast
	right		*ast

	//  siblings
	previous	*ast
	next		*ast

	parent		*ast

	tracer_ref	*tracer
	scanner_ref	*scanner
	command_ref	*command
	uint64
	string
	array_ref	[]string
}

func (a *ast) String() string {

	var what string

	switch a.yy_tok {
	case STATEMENT:
		what = fmt.Sprintf("STATEMENT#%d", a.line_no)
	case SCANNER_REF:
		what = fmt.Sprintf("SCANNER_REF(%s)", a.scanner_ref.name)
	case COMMAND_REF:
		what = fmt.Sprintf("COMMAND_REF(%s)", a.command_ref.name)
	case TRACER_REF:
		what = fmt.Sprintf("TRACER_REF(%s)", a.tracer_ref.name)
	case STRING:
		what = fmt.Sprintf("STRING(%s)", a.string)
	case NAME:
		what = fmt.Sprintf("NAME(%s)", a.string)
	case UINT64:
		what = fmt.Sprintf("UINT64(%d)", a.uint64)
	case ATT_ARRAY:
		ar := a.array_ref
		what = fmt.Sprintf("ATT_ARRAY(l=%d,c=%d)", len(ar), cap(ar))
	default:
		//  print token name or int value of yy token
		offset := a.yy_tok - __MIN_YYTOK + 3
		if (a.yy_tok > __MIN_YYTOK) {
			what = yyToknames[offset]
		} else {
			what = fmt.Sprintf( "UNKNOWN(%d)", a.yy_tok)
		}
	}
	return what
}

func (a *ast) walk_print(indent int) {

	if a == nil {
		return
	}
	if indent == 0 {
		fmt.Println("")
	} else {
		if a.parent == nil {
			panic("ast: parent is nil")
		}
		for i := 0;  i < indent;  i++ {
			fmt.Print("  ")
		}
	}
	fmt.Println(a.String())

	//  print kids

	a.left.walk_print(indent + 1)
	a.right.walk_print(indent + 1)

	//  print siblings

	if a.previous == nil {
		for as := a.next;  as != nil;  as = as.next {
			as.walk_print(indent)
		}
	}
}

func (a *ast) print() {
	a.walk_print(0)
}
