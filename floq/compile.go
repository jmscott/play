package main

type compilation struct {

	root	*ast

	flo		*flow

	//  boolean logical comparison, constants and the "when" predicate.
	a2bool		map[*ast]bool_chan

	//  string concatenation, comparison, constants and projections
	//  of tuples
	a2str		map[*ast]string_chan

	//  unsigned 64bit int comparison, constants and projection 
	a2ui64		map[*ast]uint64_chan

	//  variations of the "run <command(...)", with or without "when"
	//  predicate, with or without (...) arguments.
	a2osx		map[*ast]osx_chan

	//  argument vector for "run <commmand>" statements  
	a2argv		map[*ast]argv_chan

	//  fanout targets for specific osx_chan records, like
	//  PROJECT_OSX_EXIT_CODE
	a2osxfo		map[*ast][]osx_chan		//  fanout osx records
	
	//  fanout targets for all projections/consumers
	cmd2fo		map[*command][]osx_chan
}

func compile(root *ast) (*flow) {

	cmp := &compilation{
			root:	root,
			flo:	&flow{
				resolved:       make(chan struct{}),
				next:           make(chan flow_chan),
			},
			a2bool:		make(map[*ast]bool_chan),
			a2str:		make(map[*ast]string_chan),
			a2ui64:		make(map[*ast]uint64_chan),
			a2osx:		make(map[*ast]osx_chan),
			a2argv:		make(map[*ast]argv_chan),
			a2osxfo:	make(map[*ast][]osx_chan),
			cmd2fo:		make(map[*command][]osx_chan),
	}
	cmp.compile(root)
	return cmp.flo
}

func (cmp *compilation) relop(a *ast) {

	flo := cmp.flo
	a2bool := cmp.a2bool
	a2str := cmp.a2str
	a2ui64 := cmp.a2ui64

	l := a.left
	r := a.right
	switch {
	case l.is_string() && r.is_string():
		a2bool[a] = relop_string[a.yy_tok](
				flo,
				a2str[a.left],
				a2str[a.right],
		)

	case l.is_uint64() && r.is_uint64():
		a2bool[a] = relop_uint64[a.yy_tok](
				flo,
				a2ui64[a.left],
				a2ui64[a.right],
		)
	case l.is_bool() && r.is_bool():
		a2bool[a] = relop_bool[a.yy_tok](
				flo,
				a2bool[a.left],
				a2bool[a.right],
		)
	default:
		nm := a.yy_name()
		a.corrupt("relop: %s: can not compile %s %s %s", nm, l, nm, r)
	}
}

//  compile a root abstract syntax tree into channels connecting the nodes.

func (cmp *compilation) compile(a *ast) {

	if a == nil {
		return
	}

	//  skip "define command/tuple" statements.
	if a.yy_tok == DEFINE {
		cmp.compile(a.next)
		return
	}

	_corrupt := func(format string, args...interface{}) {
		a.corrupt("compile: " + format, args...)
	}

	flo := cmp.flo
	a2osx := cmp.a2osx
	a2str := cmp.a2str
	a2ui64 := cmp.a2ui64
	a2bool := cmp.a2bool
	a2argv := cmp.a2argv
	a2osxfo := cmp.a2osxfo
	cmd2fo := cmp.cmd2fo

	//  compile from leaves to root

	cmp.compile(a.left)
	cmp.compile(a.right)

	switch a.yy_tok {
	case CAST_UINT64:
		a2str[a] = flo.cast_uint64(a2ui64[a.left])
	case CAST_BOOL:
		a2str[a] = flo.cast_bool(a2bool[a.left])
	case CAST_STRING:
		a2str[a] = flo.cast_string(a2str[a.left])
	case yy_TRUE:
		a2bool[a] = flo.const_true()
	case yy_FALSE:
		a2bool[a] = flo.const_false()
	case yy_STRING:
		//  in a CAST ::string
	case STRING:
		a2str[a] = flo.const_string(a.string)
	case UINT64:
		a2ui64[a] = flo.const_ui64(a.uint64)
	case ARGV:
		in := make([]string_chan, a.count)
		for n := a.left;  n != nil;  n = n.next {
			in[n.order-1] = a2str[n]
		}
		a2argv[a] = flo.argv(in)
	case LT, LTE, EQ, NEQ, GTE, GT:
		cmp.relop(a)
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
		cmd := a.command_ref

		if argv == nil {
			if when == nil {
				a2osx[a] = flo.osx0(cmd)
			} else {
				a2osx[a] = flo.osx0w(cmd, a2bool[when])
			}
		} else {
			if when == nil {
				a2osx[a] = flo.osx(cmd, a2argv[argv])
			} else {
				a2osx[a] = flo.osxw(
						cmd,
						a2argv[argv],
						a2bool[when],
					)
			}
		}
		if cmd.ref_count == 0 {
			flo.osx_null(a2osx[a])
		} else {
			//  map command to fanout 
			if cmd2fo[cmd] != nil {
				_corrupt("command %s: fanout exists", cmd.name)
			}

			//  fanout osx record
			a2osxfo[a] = flo.osx_fanout(a2osx[a], cmd.ref_count)
			cmd2fo[cmd] = a2osxfo[a]
		}
	case PROJECT_OSX_EXIT_CODE:
		proj := a.proj_ref
		cmd := proj.sysatt_ref.command_ref
		fo := cmd2fo[cmd]
		a2ui64[a] = flo.osx_proj_exit_code(fo[proj.call_order-1])
	case PROJECT_OSX_PID:
		proj := a.proj_ref
		cmd := proj.sysatt_ref.command_ref
		fo := cmd2fo[cmd]
		a2ui64[a] = flo.osx_proj_pid(fo[proj.call_order-1])
	case PROJECT_OSX_USER_SEC:
		proj := a.proj_ref
		cmd := proj.sysatt_ref.command_ref
		fo := cmd2fo[cmd]
		a2ui64[a] = flo.osx_proj_user_sec(fo[proj.call_order-1])
	case PROJECT_OSX_USER_USEC:
		proj := a.proj_ref
		cmd := proj.sysatt_ref.command_ref
		fo := cmd2fo[cmd]
		a2ui64[a] = flo.osx_proj_user_usec(fo[proj.call_order-1])
	case PROJECT_OSX_SYS_SEC:
		proj := a.proj_ref
		cmd := proj.sysatt_ref.command_ref
		fo := cmd2fo[cmd]
		a2ui64[a] = flo.osx_proj_sys_sec(fo[proj.call_order-1])
	case PROJECT_OSX_SYS_USEC:
		proj := a.proj_ref
		cmd := proj.sysatt_ref.command_ref
		fo := cmd2fo[cmd]
		a2ui64[a] = flo.osx_proj_sys_usec(fo[proj.call_order-1])
	case PROJECT_OSX_START_TIME:
		proj := a.proj_ref
		cmd := proj.sysatt_ref.command_ref
		fo := cmd2fo[cmd]
		a2str[a] = flo.osx_proj_start_time(fo[proj.call_order-1])
	case PROJECT_OSX_WALL_DURATION:
		proj := a.proj_ref
		cmd := proj.sysatt_ref.command_ref
		fo := cmd2fo[cmd]
		a2ui64[a] = flo.osx_proj_wall_duration(fo[proj.call_order-1])
	case PROJECT_OSX_STDOUT:
		proj := a.proj_ref
		cmd := proj.sysatt_ref.command_ref
		fo := cmd2fo[cmd]
		a2str[a] = flo.osx_proj_Stdout(fo[proj.call_order-1])
	case PROJECT_OSX_STDERR:
		proj := a.proj_ref
		cmd := proj.sysatt_ref.command_ref
		fo := cmd2fo[cmd]
		a2str[a] = flo.osx_proj_Stderr(fo[proj.call_order-1])
	case PROJECT_OSX_TUPLE_TSV:
		proj := a.proj_ref
		cmd := proj.sysatt_ref.command_ref
		fo := cmd2fo[cmd]
		a2str[a] = flo.osx_proj_tuple_tsv(
				fo[proj.call_order-1],
				proj.att_ref,
			)
	case IS_NULL_STRING:
		a2bool[a] = flo.is_null_string(a2str[a.left])
	case IS_NULL_UINT64:
		a2bool[a] = flo.is_null_uint64(a2ui64[a.left])
	case IS_NULL_BOOL:
		a2bool[a] = flo.is_null_bool(a2bool[a.left])
	case IS_NOT_NULL_STRING:
		a2bool[a] = flo.is_not_null_string(a2str[a.left])
	case IS_NOT_NULL_UINT64:
		a2bool[a] = flo.is_not_null_uint64(a2ui64[a.left])
	case IS_NOT_NULL_BOOL:
		a2bool[a] = flo.is_not_null_bool(a2bool[a.left])
	case FLOW, STMT_LIST, DEFINE:
	default:
		_corrupt("can not compile ast")
	}
	cmp.compile(a.next)
}
