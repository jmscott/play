package main

type compiler struct {

	flo	*flow

	root	*ast
	active	*ast		//  node being parsed

	//  map ast nodes to output channels of bool, uint4, string 
	a2bool	map[*ast]bool_chan
	a2ui64	map[*ast]ui64_chan
	a2str	map[*ast]string_chan
}

func compile(root *ast) (*flow, flow_chan, error) {

	cmpl := &compiler{}
	return cmpl.compile(root)
}

func (cmpl *compiler) corrupt(format string, args...interface{}) {

	cmpl.active.corrupt("compile: " + format, args...)
}

func (cmpl *compiler) ckparent(expect ...int) {

	a := cmpl.active

	for _, tok := range expect {
		if a.parent.yy_tok == tok {
			return
		}
	}
	cmpl.corrupt("unexpected parent node: %s", a.name())
}

func (cmpl *compiler) ckleft(expect ...int) {

	a := cmpl.active

	if a.left == nil {
		cmpl.corrupt("left is nil")
	}
	for _, tok := range expect {
		if a.left.yy_tok == tok {
			return
		}
	}
	cmpl.corrupt("unexpected left node: %s", a.left.name())
}

func (cmpl *compiler) ckrelop() {

	a := cmpl.active

	tok := a.left.yy_tok

	switch tok {
	case STRING, UINT64, yy_TRUE, yy_FALSE:
	default:
		cmpl.corrupt("bad relop type: %s", yy_name(tok))
	}
}

func (cmpl *compiler) relop() {

	a := cmpl.active

	cmpl.ckrelop()
	op_type := a.left.yy_type()
	switch op_type { 
	case "string":
		cmpl.a2bool[a] = (relop_string[a.yy_tok]) (
				cmpl.flo,
				cmpl.a2str[a.left],
				cmpl.a2str[a.right],
		)
	case "uint64":
		cmpl.a2bool[a] = (relop_ui64[a.yy_tok]) (
				cmpl.flo,
				cmpl.a2ui64[a.left],
				cmpl.a2ui64[a.right],
		)
	case "bool":
	default:
		cmpl.corrupt("can not compile %s", op_type)
	}
}

//  compile a root abstract syntax tree into connected channels. data
//  flows from least dependent leaves to most dependent "flow_stmt".

func (cmpl *compiler) compile(root *ast) (*flow, flow_chan, error) {

	if root == nil {
		cmpl.corrupt("root is nil")
	}
	cmpl.root = root
	cmpl.active = root

	if root.yy_tok != FLOW {
		cmpl.corrupt("root node not FLOW: %s", root.name())
	}
	if root.left == nil {
		cmpl.corrupt("root.left is nil")
	}
	if root.left.yy_tok != STMT_LIST {
		cmpl.corrupt("root.left not STMT_LIST: %s", root.left.name)
	}

	//  map asts to their output channels

	cmpl.root = root
	cmpl.flo = &flow{}

	cmpl.a2bool = make(map[*ast]bool_chan)
	cmpl.a2str = make(map[*ast]string_chan)
	cmpl.a2ui64 = make(map[*ast]ui64_chan)

	var compile func(*ast)

	//  bottom up compilation of abstract syntax tree into network of 
	//  channels.  consistency rechecks needed cause tree may have been
	//  rewritten significally from the original yacc generated tree.

	compile = func(a *ast) {

		if a == nil {
			return
		}

		_corrupt := func(format string, args...interface{}) {
			cmpl.active = a
			//cmpl.corrupt(format string, args...interface{})
			panic("WTF")
		}

		compile(a.left)
		compile(a.right)

		cmpl.active = a

		if a.is_binary() {
			if a.left == nil {
				_corrupt("left child is nil")
			}
			if a.right == nil {
				_corrupt("right child is nil")
			}
		} else if a.is_unary() {
			if a.right != nil {
				_corrupt("right exists for unary op")
			}
		}
		if a.parent == nil {
			_corrupt("parent is nil")
		}

		switch a.yy_tok {
		case NAME:
		case ATT:
			cmpl.ckparent(ATT_TUPLE)
		case ATT_TUPLE:
			cmpl.ckparent(COMMAND_REF, SCANNER_REF, TRACER_REF)
		case yy_TRUE:
			cmpl.a2bool[a] = cmpl.flo.const_true()
		case yy_FALSE:
			cmpl.a2bool[a] = cmpl.flo.const_false()
		case STRING:
			cmpl.a2str[a] = cmpl.flo.const_string(a.string)
		case UINT64:
			cmpl.a2ui64[a] = cmpl.flo.const_ui64(a.uint64)
		case SCANNER_REF:
			cmpl.ckleft(ATT_TUPLE)
		case CREATE:
			cmpl.ckparent(STMT)
		case STMT:
			cmpl.ckparent(STMT_LIST)
		case STMT_LIST:
			_corrupt("unexpected STMT_LIST")
		case TRACER_REF, COMMAND_REF:
			cmpl.ckleft(ATT_TUPLE)
		case ARG_LIST:
		case RUN:
		case LT, LTE, EQ, NEQ, GTE, GT:
			cmpl.relop()
		case yy_OR:
			cmpl.a2bool[a] = cmpl.flo.bool2(
					or,
					cmpl.a2bool[a.left],
					cmpl.a2bool[a.right],
			)
		case yy_AND:
			cmpl.a2bool[a] = cmpl.flo.bool2(
						and,
						cmpl.a2bool[a.left],
						cmpl.a2bool[a.right],
			)
		case NOT:
			cmpl.a2bool[a] = cmpl.flo.not(cmpl.a2bool[a.left])
		case WHEN:
			if a.parent.left.is_flowable() == false {
				_corrupt("parent of WHEN not flowable")
			}
			if a.left.is_bool() == false {
				_corrupt("qualification not bool")
			}
			cmpl.a2bool[a] = cmpl.a2bool[a.left] 
		default:
			_corrupt("can not compile ast")
		}
	}

	for stmt := root.left.left;  stmt != nil;  stmt = stmt.next {
		if stmt.yy_tok != STMT {
			cmpl.corrupt("root.left.left not yy_tok STMT")
		}
		compile(stmt)
	}

	return cmpl.flo, make(flow_chan), nil
}
