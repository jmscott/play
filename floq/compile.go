package main

func (flo *flow) compile(root *ast) (flow_chan, error) {

	if root == nil {
		return nil, nil
	}
	if root.yy_tok != FLOW {
		impossible("root node not FLOW: %s", root.name())
	}
	if root.left == nil {
		impossible("root.left is nil")
	}
	if root.left.yy_tok != STMT_LIST {
		impossible("root.left not STMT_LIST: %s", root.left.name)
	}

	//  map asts to their output channels

	a2bool := make(map[*ast]bool_chan)
	a2str := make(map[*ast]string_chan)
	a2ui := make(map[*ast]uint64_chan)

	var compile func(a *ast)

	//  bottom up compilation of abstract syntax tree into network of 
	//  channels.  consistency checks need cause tree may be change
	//  significally from the yacc generated tree.
	//  

	compile = func(a *ast) {

		ckparent := func(a *ast, expect_tok int) {
			if a.parent == nil {
				a.impossible("parent is nil")
			}
			if a.parent.yy_tok != expect_tok {
				a.impossible(
					"parent not %s: %s",
					yy_name(expect_tok),
					a.parent.name(),
				)
			}
		}
		ckleft := func(a *ast, expect_tok int) {
			if a.left == nil {
				a.impossible("left is nil")
			}
			if a.left.yy_tok != expect_tok {
				a.impossible(
					"left not %s: %s",
					yy_name(expect_tok),
					a.parent.name(),
				)
			}
		}

		if a == nil {
			return
		}

		compile(a.left)
		compile(a.right)

		switch a.yy_tok {
		case NAME:
		case ATT:
			ckparent(a, ATT_TUPLE)
		case ATT_TUPLE:
			if a.parent == nil {
				a.impossible("parent is nill")
			}
		case yy_TRUE:
			a2bool[a] = flo.const_true()
		case yy_FALSE:
			a2bool[a] = flo.const_false()
		case STRING:
			a2str[a] = flo.const_string(a.string)
		case UINT64:
			a2ui[a] = flo.const_uint64(a.uint64)
		case SCANNER_REF:
			ckleft(a, ATT_TUPLE)
		case CREATE:
			ckparent(a, STMT)
		case STMT:
			ckparent(a, STMT_LIST)
		case STMT_LIST:
			ckparent(a, FLOW)
		default:
			a.impossible("can not compile ast")
		}
	}

	compile(root.left)

	return make(flow_chan), nil
}
