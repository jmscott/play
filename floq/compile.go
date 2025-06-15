package main

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

		ckparent := func(expect ...int) {
			for _, tok := range expect {
				if a.parent.yy_tok == tok {
					return
				}
			}
			_die("parent node: %s", a.parent.name())
		}

		ckleft := func(expect ...int) {
			if a.left == nil {
				_die("left is nil")
			}
			for _, tok := range expect {
				if a.left.yy_tok == tok {
					return
				}
			}
			_die("left node: %s", a.left.name())
		}

		relop := func() {

			tok := a.left.yy_tok

			switch tok {
			case STRING:
			case UINT64:
			default:
				_die("relop: bad type: %s", yy_name(tok))
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
			ckparent(ATT_TUPLE)
		case ATT_TUPLE:
			ckparent(COMMAND_REF, SCANNER_REF, TRACER_REF)
		case yy_TRUE:
			a2bool[a] = flo.const_true()
		case yy_FALSE:
			a2bool[a] = flo.const_false()
		case STRING:
			a2str[a] = flo.const_string(a.string)
		case UINT64:
			a2ui[a] = flo.const_uint64(a.uint64)
		case SCANNER_REF:
			ckleft(ATT_TUPLE)
		case CREATE:
			ckparent(STMT)
		case STMT:
			ckparent(STMT_LIST)
		case STMT_LIST:
			a.corrupt("unexpected STMT_LIST")
		case TRACER_REF, COMMAND_REF:
			ckleft(ATT_TUPLE)
		case ARG_LIST:
		case RUN:
		case LT, LTE, EQ, NEQ, GTE, GT:
			relop()
			switch a.left.yy_tok {
			case STRING:
				//a2bool[a] = relop_string[a.yy_tok]

			/*
			case UINT64:
				a2bool[a] = flo.eq_uint64(
						a2ui[a.left],
						a2ui[a.right],
				)
			*/
			default:
				_die("relop: %s", yy_name)
			}			
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
			if a.parent.left.is_flowable() == false {
				_die("parent of 'when' not flowable")
			}
			if a.left.is_bool() == false {
				_die("qualification not bool")
			}
			a2bool[a] = a2bool[a.left] 
		default:
			_die("can not compile ast")
		}
	}

	for stmt := root.left.left;  stmt != nil;  stmt = stmt.next {
		if stmt.yy_tok != STMT {
			stmt.corrupt("root.left.left not yy_tok STMT")
		}
		compile(stmt)
	}

	return make(flow_chan), nil
}
