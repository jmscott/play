package main

func compile(root *ast) (*flow, flow_chan, error) {

	flo := &flow{}
	fc, err := flo.compile(root)
	return flo, fc, err
}

//  compile a root abstract syntax tree into connected channels. data
//  flows from least dependent leaves to most dependent "flow_stmt".

func (flo *flow) compile(root *ast) (flow_chan, error) {

	if root == nil {
		return nil, nil
	}
	if root.yy_tok != FLOW {
		corrupt("root node not FLOW: %s", root.name())
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

	var compile func(a *ast)

	//  bottom up compilation of abstract syntax tree into network of 
	//  channels.  consistency rechecks needed cause tree may have been
	//  rewritten significally from the original yacc generated tree.

	compile = func(a *ast) {

		_die := func(format string, args...interface{}) {
			a.corrupt("compile: " + format, args...)
		}

		relop := func() {

			tok := a.left.yy_tok
			switch tok {
			case STRING:
				if relop_string[a.yy_tok] == nil {
					_die("no flow operator")
				}
				a2bool[a] = relop_string[a.yy_tok](
						flo,
						a2str[a.left],
						a2str[a.right],
				)

			case UINT64:
				/*
				a2bool[a] = flo.eq_ui64(
						a2ui[a.left],
						a2ui[a.right],
				)
				*/
			default:
				_die("relop: compile left %s", yy_name(tok))
			}
		}

		if a == nil {
			return
		}

		compile(a.left)
		compile(a.right)

		if a.is_binary() {
			if a.left == nil {
				_die("left child is nil")
			}
			if a.right == nil {
				_die("right child is nil")
			}
		} else if a.is_unary() {
			if a.right != nil {
				_die("right exists for unary op")
			}
		}
		if a.parent == nil {
			_die("parent is nill")
		}
		switch a.yy_tok {
		case NAME:
		case ATT:
		case ATT_TUPLE:
		case yy_TRUE:
			a2bool[a] = flo.const_true()
		case yy_FALSE:
			a2bool[a] = flo.const_false()
		case STRING:
			a2str[a] = flo.const_string(a.string)
		case UINT64:
			a2ui[a] = flo.const_ui64(a.uint64)
		case SCANNER_REF:
		case CREATE:
		case STMT:
		case STMT_LIST:
			a.corrupt("unexpected STMT_LIST")
		case TRACER_REF, COMMAND_REF:
		case ARG_LIST:
		case RUN:
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
		case WHEN:
			a2bool[a] = a2bool[a.left] 
		case CONCAT:
			a2str[a] = flo.concat(a2str[a.left], a2str[a.right])
		default:
			_die("can not compile ast")
		}
	}

	for stmt := root.left.left;  stmt != nil;  stmt = stmt.next {
		if stmt.yy_tok != STMT {
			stmt.corrupt("root.left.left not yy_tok STMT")
		}
		stmt.frisk()
		if stmt.left.yy_tok != CREATE {
			compile(stmt)
		}
	}

	return make(flow_chan), nil
}
