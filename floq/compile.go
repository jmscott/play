package main

func compile(root *ast) (*flow, error) {

	flo := new_flow()
	err := flo.compile(root)
	return flo, err
}

func new_flow() *flow {
	return &flow{
		resolved:       make(chan struct{}),
		next:           make(chan flow_chan),
	}
}

//  compile a root abstract syntax tree into connected channels. data
//  flows from least dependent leaves to most dependent "flow_stmt".

func (flo *flow) compile(root *ast) error {

	if root == nil {
		return nil
	}
	if root.yy_tok != FLOW {
		corrupt("root node not FLOW: %s", root.yy_name())
	}
	if root.left == nil {
		corrupt("root.left is nil")
	}
	if root.left.yy_tok != STMT_LIST {
		corrupt("root.left not STMT_LIST: %s", root.left.name)
	}

	//  map asts to their output channels

	a2bool := make(map[*ast]bool_chan)
	a2str := make(map[*ast]string_chan)
	a2ui := make(map[*ast]uint64_chan)
	a2osx := make(map[*ast]osx_chan)
	a2argv := make(map[*ast]argv_chan)

	var compile1 func(a *ast)

	//  bottom up compilation of abstract syntax tree into network of 
	//  channels.  consistency rechecks needed cause tree may have been
	//  rewritten significally from the original yacc generated tree.

	compile1 = func(a *ast) {

		_corrupt := func(format string, args...interface{}) {
			a.corrupt("compile1: " + format, args...)
		}

		relop := func() {

			tok := a.left.yy_tok
			switch tok {
			case STRING:
				a2bool[a] = relop_string[a.yy_tok](
						flo,
						a2str[a.left],
						a2str[a.right],
				)

			case UINT64:
				a2bool[a] = relop_uint64[a.yy_tok](
						flo,
						a2ui[a.left],
						a2ui[a.right],
				)
			default:
				a2bool[a] = relop_bool[a.yy_tok](
						flo,
						a2bool[a.left],
						a2bool[a.right],
				)
			}
		}

		if a == nil {
			return
		}

		//  compile from leaves to branches

		compile1(a.left)
		compile1(a.right)

		switch a.yy_tok {
		case yy_TRUE:
			a2bool[a] = flo.const_true()
		case yy_FALSE:
			a2bool[a] = flo.const_false()
		case STRING:
			a2str[a] = flo.const_string(a.string)
		case UINT64:
			a2ui[a] = flo.const_ui64(a.uint64)
		case ARGV:
			in := make([]string_chan, a.count)
			for n := a.left;  n != nil;  n = n.next {
				in[n.order-1] = a2str[n]
			}
			a2argv[a] = flo.argv(in)
		case LT, LTE, EQ, NEQ, GTE, GT:
			relop()
		case yy_OR:
			a2bool[a] = flo.bool2(
					or,
					a2bool[a.left],
					a2bool[a.right],
			)
		case yy_AND:
			a2bool[a] = flo.bool2(
					and,
					a2bool[a.left],
					a2bool[a.right],
			)
		case NOT:
			a2bool[a] = flo.not(a2bool[a.left])
		case CONCAT:
			a2str[a] = flo.concat(a2str[a.left], a2str[a.right])
		case WHEN:
			a2bool[a] = a2bool[a.left]
		case RUN:
			argv := a.left
			when := a.right

			if argv == nil {
				if when == nil {
					a2osx[a] = flo.osx0(a.command_ref)
				} else {
					a2osx[a] = flo.osx0w(
							a.command_ref,
							a2bool[when],
					)
				}
			} else {
				if when == nil {
					a2osx[a] = flo.osx(
						a.command_ref,
						a2argv[argv],
					)
				} else {
					a2osx[a] = flo.osxw(
						a.command_ref,
						a2argv[argv],
						a2bool[when],
					)
				}
			}
		default:
			_corrupt("can not compile ast")
		}
		compile1(a.next)
	}

	//  compile each statement, skipping  nodes, which are handled
	//  in the parser

	for stmt := root.left.left;  stmt != nil;  stmt = stmt.next {
		if stmt.yy_tok == DEFINE {
			continue
		}
		stmt.frisk()
		compile1(stmt)
	}

	return nil
}
