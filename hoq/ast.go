package main

import (
	"fmt"
)

//  abstract syntax tree that represents the parsed program.
//  all possible node values must be explicitly defined since go has no union
//  types.

type ast struct {

	//  lexical token automatically defined by yacc
	yy_tok	int

	//  a string value associated with the node
	string

	//  a unix command declaration
	*command

	//  child nodes
	left	*ast
	right	*ast

	//  siblings
	next	*ast
}

//  dump a node in the abstract syntax tree

func (a *ast) String() string {

	switch a.yy_tok {
	case COMMAND:
		return fmt.Sprintf("COMMAND(%s)", a.command.name)
	case STRING:
		return fmt.Sprintf("STRING(\"%s\")", a.string)
	case NAME:
		return fmt.Sprintf("NAME(\"%s\")", a.string)
	}

	offset := a.yy_tok - __MIN_YYTOK + 3
	if (a.yy_tok > __MIN_YYTOK) {
		return yyToknames[offset]
	}
	return fmt.Sprintf("UNKNOWN_TOKEN(%d)", a.yy_tok)
}

//  recursivly print indented nodes of the abstract syntax tree

func (a *ast) print_tree(indent int, is_first_sibling bool) {

	if a == nil {
		return
	}

	//  indent
	for i := 0;  i < indent;  i++ {
		fmt.Print("  ")
	}

	//  print the node

	fmt.Println(a.String())

	//  recusively print the kids

	a.left.print_tree(indent + 1, true)
	a.right.print_tree(indent + 1, true)

	//  print siblings if we are

	if is_first_sibling {
		for as := a.next;  as != nil;  as = as.next {
			as.print_tree(indent, false)
		}
	}
}

func (a *ast) print() {
	a.print_tree(0, true)
}
