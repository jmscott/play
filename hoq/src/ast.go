//  abstract syntax tree, the output of yacc in parser.y

package main

import (
	"fmt"
	"reflect"
)

//  abstract syntax tree that represents the parsed program.
//
//  unfortunatly since go has no union type all possible node values must be
//  explicitly defined.

type ast struct {

	//  lexical token, automatically defined by yacc.  see parser.go

	yy_tok int

	go_type reflect.Kind

	//  approximate line number in the source file

	line_no uint64

	string
	uint8
	bool

	//  a unix command{} declaration

	*command

	//  child nodes.  shoulld be {left, right}_child

	left  *ast
	right *ast

	//  ought to next_sibling

	next *ast
}

//  dump a node in the abstract syntax tree

func (a *ast) String() string {

	switch a.yy_tok {
	case COMMAND:
		cmd := a.command
		var argv string

		for i, s := range cmd.init_argv {
			if i > 0 {
				argv += ", "
			}
			argv += "\"" + s + "\""
		}
		if len(argv) > 0 {
			argv = "(" + argv + ")"
		}
		if cmd.full_path == "" {
			return fmt.Sprintf("COMMAND.%s%s->%s",
				cmd.name,
				argv,
				cmd.path,
			)
		}
		return fmt.Sprintf("COMMAND.%s%s->%s->%s",
			cmd.name,
			argv,
			cmd.path,
			cmd.full_path,
		)

	case EXEC:
		return fmt.Sprintf("EXEC.%s", a.command.name)
	case STRING:
		return fmt.Sprintf("\"%s\"", a.string)
	case UINT8:
		return fmt.Sprintf("UINT8=%d", a.uint8)
	case DOLLAR:
		return fmt.Sprintf("$%d", a.uint8)
	case ARGV:
		return fmt.Sprintf("ARGV#%d", a.uint8)
	case EXIT_STATUS:
		return fmt.Sprintf("%s.exit_status", a.command.name)
	}

	offset := a.yy_tok - __MIN_YYTOK + 3
	if a.yy_tok > __MIN_YYTOK {
		return yyToknames[offset]
	}
	return fmt.Sprintf("UNKNOWN_TOKEN(%d)", a.yy_tok)
}

//  recursivly dump indented nodes of the abstract syntax tree

func (a *ast) dump_tree(indent int, is_first_sibling bool) {

	if a == nil {
		return
	}

	//  indent, two space for each level

	for i := 0; i < indent; i++ {
		fmt.Print("  ")
	}

	//  print the node

	fmt.Println(a.String())

	//  recusively print the kids

	a.left.dump_tree(indent+1, true)
	a.right.dump_tree(indent+1, true)

	//  dump siblings if we are

	if is_first_sibling {
		for as := a.next; as != nil; as = as.next {
			as.dump_tree(indent, false)
		}
	}
}

func (a *ast) dump() {
	a.dump_tree(0, true)
}

//  rewrite argument vector with no arguments into node ARGV0.

func (a *ast) rewrite_ARGV0() {

	if a == nil {
		return
	}
	if a.yy_tok == EXEC && a.left == nil {
		a.left = &ast{
			yy_tok: ARGV0,
		}
	}
	a.left.rewrite_ARGV0()
	a.right.rewrite_ARGV0()
	a.next.rewrite_ARGV0()
}

//  rewrite argument vector with single argument into node ARGV1.

func (a *ast) rewrite_ARGV1() {

	if a == nil {
		return
	}
	if a.yy_tok == ARGV && a.left != nil && a.left.next == nil {
		a.yy_tok = ARGV1
		a.uint8 = 1
	}
	a.left.rewrite_ARGV1()
	a.right.rewrite_ARGV1()
	a.next.rewrite_ARGV1()
}

//  since unix commands require strings in argv[], we first transform the
//  argument nodes to EXEC that are uint8 or bool into strings:
//
//	EXEC func(123) to EXEC func(to_string_uint8(123))

func (a *ast) rewrite_EXEC_ARGV_UINT8() {

	if a == nil {
		return
	}

	if a.yy_tok == EXEC {

		//  walk through argv of exec, looking for scalar uint8
		//  or exec.exit_status nodes.

		argv := a.left
		prev := (*ast)(nil)

		for arg := argv.left; arg != nil; arg = arg.next {

			var a *ast

			a = arg
			switch arg.go_type {
			case reflect.String:
				prev = arg
				continue
			case reflect.Uint8:
				arg = &ast{
					yy_tok:  TO_STRING_UINT8,
					go_type: reflect.String,
				}
			case reflect.Bool:
				arg = &ast{
					yy_tok:  TO_STRING_BOOL,
					go_type: reflect.String,
				}
			default:
				panic("impossible ast.go_type")
			}
			arg.left = a
			arg.next = a.next
			a.next = nil

			if argv.left == a {
				argv.left = arg
			} else {
				prev.next = arg
			}
			prev = arg
		}
	}
	a.next.rewrite_EXEC_ARGV_UINT8()
}

//  Change generic binary operator nodes to type specific version

func (a *ast) rewrite_binop() {

	if a == nil {
		return
	}

	switch a.yy_tok {

	case EQ:
		switch a.left.go_type {
		case reflect.Bool:
			a.yy_tok = EQ_BOOL

		case reflect.String:
			a.yy_tok = EQ_STRING

		case reflect.Uint8:
			a.yy_tok = EQ_UINT8

		default:
			panic("EQ: impossible go_type")
		}
	case NEQ:
		switch a.left.go_type {

		case reflect.Bool:
			a.yy_tok = NEQ_BOOL

		case reflect.String:
			a.yy_tok = NEQ_STRING

		case reflect.Uint8:
			a.yy_tok = NEQ_UINT8

		default:
			panic("NEQ: impossible go_type")
		}
	}
	a.left.rewrite_binop()
	a.right.rewrite_binop()
	a.next.rewrite_binop()
}

//  rewrite an empty qualification in an exec() statement to always send true.

func (a *ast) rewrite_EXEC_NO_QUAL() {

	if a == nil {
		return
	}
	if a.yy_tok == EXEC && a.right == nil {
		a.right = &ast{
			yy_tok: TRUE,
		}
	}
	a.next.rewrite_EXEC_NO_QUAL()
}

func (a *ast) rewrite_DOLLAR0() {

	if a == nil {
		return
	}
	if a.yy_tok == DOLLAR && a.uint8 == 0 {
		a.yy_tok = DOLLAR0
	}
	a.left.rewrite_DOLLAR0()
	a.right.rewrite_DOLLAR0()
	a.next.rewrite_DOLLAR0()
}

//  expand the relative Path to the COMMAND into the full path to the
//  executable file

func (a *ast) lookup_command_PATH() {

	if a == nil {
		return
	}
	if a.yy_tok == COMMAND {
		a.command.lookup_full_path()
	}
	a.next.lookup_command_PATH()
}

func (root *ast) rewrite(dumping bool) {

	root.rewrite_binop()
	root.rewrite_ARGV0()
	root.rewrite_ARGV1()
	root.rewrite_EXEC_ARGV_UINT8()
	root.rewrite_EXEC_NO_QUAL()
	root.rewrite_DOLLAR0()
	if !dumping {
		root.lookup_command_PATH()
	}
}
