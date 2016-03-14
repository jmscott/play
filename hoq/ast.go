package main

import (
	"fmt"
)

func (a *ast) to_string() string {

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

func (a *ast) String() string {

	return a.to_string()

}

func (a *ast) walk_print(indent int, is_first_sibling bool) {

	if a == nil {
		return
	}
	if indent == 0 {
		fmt.Println("")
	} else {
		for i := 0;  i < indent;  i++ {
			fmt.Print("  ")
		}
	}
	fmt.Println(a.to_string())

	//  print kids
	a.left.walk_print(indent + 1, true)
	a.right.walk_print(indent + 1, true)

	//  print siblings
	if is_first_sibling {
		for as := a.next;  as != nil;  as = as.next {
			as.walk_print(indent, false)
		}
	}
}

func (a *ast) print() {
	a.walk_print(0, true)
}
