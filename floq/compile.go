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
	//a2osx := make(map[*ast]osx_chan)

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
				if relop_string[a.yy_tok] == nil {
					_corrupt("no string flow op")
				}
				a2bool[a] = relop_string[a.yy_tok](
						flo,
						a2str[a.left],
						a2str[a.right],
				)

			case UINT64:
				if relop_uint64[a.yy_tok] == nil {
					_corrupt("no uint64 flow op")
				}
				a2bool[a] = relop_uint64[a.yy_tok](
						flo,
						a2ui[a.left],
						a2ui[a.right],
				)
			default:
				nm := yy_name(tok)
				_corrupt("relop: can not compile left (%s)", nm)
			}
		}

		if a == nil {
			return
		}

		compile1(a.left)
		compile1(a.right)

		if a.is_binary() {
			if a.left == nil {
				_corrupt("left child is nil for binary op")
			}
			if a.right == nil {
				_corrupt("right child is nil for binary op")
			}
		} else if a.is_unary() {
			if a.right != nil {
				_corrupt("right exists for unary op")
			}
		}
		if a.parent == nil {
			_corrupt("parent is nill")
		}
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
			if a.left != nil {
				_corrupt("ARGV has left child")
			}
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
			
		default:
			_corrupt("can not compile ast")
		}
	}

	//  compile each statement, skipping  nodes, which are handled
	//  in the parser

	for stmt := root.left.left;  stmt != nil;  stmt = stmt.next {
		stmt.frisk()
		if stmt.left.yy_tok != DEFINE {
			compile1(stmt)
		}
	}

	return nil
}
