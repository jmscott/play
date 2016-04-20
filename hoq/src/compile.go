//  compile an abstract syntax tree into a flow graph of nodes

package main

import (
	"fmt"
)

func (flo *flow) compile(
	ast_head *ast,
	depend_order []string,
) uint8_chan {

	type exec_output struct {

		//  fanout channels listening for exit_status of a process

		out_chans []uint8_chan

		//  next free channel

		next_chan int
	}

	//  map the command name to EXEC ast node

	exec2ast := make(map[string]*ast)
	var find_EXEC func(*ast)
	find_EXEC = func(a *ast) {

		if a == nil {
			return
		}
		if a.yy_tok == EXEC {
			exec2ast[a.command.name] = a
		}
		find_EXEC(a.next)
	}
	find_EXEC(ast_head)

	//  map command output onto list of fanout channels

	cmd2out := make(map[string]*exec_output)

	type predicate_output struct {

		//  fanout channels listening for exit_status of a process

		out_chans []bool_chan

		//  next free channel

		next_chan int
	}

	//  map the predicate name to PREDICATE ast node

	pred2ast := make(map[string]*ast)
	var find_PREDICATE func(*ast)
	find_PREDICATE = func(a *ast) {

		if a == nil {
			return
		}
		if a.yy_tok == PREDICATE {
			pred2ast[a.predicate.name] = a
		}
		find_PREDICATE(a.next)
	}
	find_PREDICATE(ast_head)

	//  map command output onto list of fanout channels

	pred2out := make(map[string]*predicate_output)

	//  map abstract syntax tree nodes to compiled channels

	// ast node to bool channel

	a2b := make(map[*ast]bool_chan)

	// ast node to uint8 channel

	a2u8 := make(map[*ast]uint8_chan)

	//  ast node to string channel

	a2s := make(map[*ast]string_chan)

	//  ast node to argv channel

	a2av := make(map[*ast]argv_chan)

	var compile func(a *ast)
	compile = func(a *ast) {
		if a == nil {
			return
		}

		//  track number of confluing go routines compiled for
		//  each node
		cc := 1

		//  compile kids first

		compile(a.left)
		compile(a.right)

		switch a.yy_tok {
		case UINT8:
			a2u8[a] = flo.const_uint8(a.uint8)
		case STRING:
			a2s[a] = flo.const_string(a.string)
		case ARGV0:
			a2av[a] = flo.argv0()
		case ARGV1:
			a2av[a] = flo.argv1(a2s[a.left])
		case ARGV:
			in := make([]string_chan, a.uint8)

			//  first arg node already compiled
			aa := a.left
			in[0] = a2s[aa]
			aa = aa.next

			//  compile arg nodes 2 ... n

			for i := 1; aa != nil; aa = aa.next {
				compile(aa)
				in[i] = a2s[aa]
				i++
			}
			a2av[a] = flo.argv(in)

		//  execute a program in the file system

		case EXEC:
			cmd := a.command
			a2u8[a] = flo.exec(cmd, a2av[a.left], a2b[a.right])

			//  broadcast exit_status to interested go routines

			cmd2out[cmd.name] =
				&exec_output{
					out_chans: flo.fanout_uint8(
						a2u8[a],

						//  each exit_status in
						//  qualification plus
						//  the fan-in channel that
						//  terminates the flow

						cmd.depend_ref_count+1,
					),

					//  slot 0 is the fan-in channel

					next_chan: 1,
				}

			//  for the extra fanout_uint8()
			cc++

		//  execute a program in the file system

		case PREDICATE:
			pred := a.predicate

			//  broadcast boolean qualification to interested
			//  go routines

			pred2out[pred.name] = &predicate_output{
					out_chans: flo.fanout_bool(
						a2b[a.left],

						//  each bool in
						//  qualification plus
						//  the fan-in channel that
						//  terminates the flow

						pred.depend_ref_count+1,
					),

					//  slot 0 is the fan-in channel

					next_chan: 1,
				}
		case TRUE:
			a2b[a] = flo.const_bool(true)
		case FALSE:
			a2b[a] = flo.const_bool(false)
		case WHEN:
			a2b[a] = a2b[a.left]
			cc = 0
		case EXIT_STATUS:
			co := cmd2out[a.command.name]
			a2u8[a] = co.out_chans[co.next_chan]
			co.next_chan++
			cc = 0
		case XPREDICATE:
			po := pred2out[a.predicate.name]
			a2b[a] = po.out_chans[po.next_chan]
			po.next_chan++
			cc = 0

		case EQ_UINT8:
			a2b[a] = flo.rel2_uint8(
				uint8_eq,
				a2u8[a.left],
				a2u8[a.right],
			)
		case NEQ_UINT8:
			a2b[a] = flo.rel2_uint8(
				uint8_neq,
				a2u8[a.left],
				a2u8[a.right],
			)
		case EQ_STRING:
			a2b[a] = flo.rel2_string(
				string_eq,
				a2s[a.left],
				a2s[a.right],
			)
		case NEQ_STRING:
			a2b[a] = flo.rel2_string(
				string_neq,
				a2s[a.left],
				a2s[a.right],
			)
		case OR:
			a2b[a] = flo.rel2_bool(
				or,
				a2b[a.left],
				a2b[a.right],
			)
		case AND:
			a2b[a] = flo.rel2_bool(
				and,
				a2b[a.left],
				a2b[a.right],
			)
		case TO_STRING_UINT8:
			a2s[a] = flo.to_string_uint8(
				a2u8[a.left],
			)
		case DOLLAR:
			a2s[a] = flo.dollar(a.uint8 - 1)
		case DOLLAR0:
			a2s[a] = flo.dollar0()
		case RE_MATCH:
			a2b[a] = flo.rel2_string(
				re_match,
				a2s[a.left],
				a2s[a.right],
			)
		case RE_NMATCH:
			a2b[a] = flo.rel2_string(
				re_nmatch,
				a2s[a.left],
				a2s[a.right],
			)
		case TO_STRING_BOOL:
			a2s[a] = flo.to_string_bool(
				a2b[a.left],
			)
		case EQ_BOOL:
			a2b[a] = flo.rel2_bool(
				bool_eq,
				a2b[a.left],
				a2b[a.right],
			)
		case NEQ_BOOL:
			a2b[a] = flo.rel2_bool(
				bool_neq,
				a2b[a.left],
				a2b[a.right],
			)
		case NOT:
			a2b[a] = flo.not(a2b[a.left])
		default:
			panic(fmt.Sprintf(
				"impossible yy_tok in ast: %d, near line %d",
					a.yy_tok,
					a.line_no,
					))
		}
		flo.confluent_count += cc
	}

	//  compile EXEC/PREDICATe nodes from least dependent to most
	//  dependent order

	for _, n := range depend_order {
		switch {
		case exec2ast[n] != nil:
			compile(exec2ast[n])
		case pred2ast[n] != nil:
			compile(pred2ast[n])
		default:
			panic("unknown type in depend order: " + n)
		}
	}

	//  map uint8 output of each exec.exit_status onto a fanin channel.
	//  uint8_chan[0] is reserved for the fanin channel

	uint8_out := make([]uint8_chan, len(cmd2out))
	i := 0
	for n, ox := range cmd2out {

		//  cheap sanity test that all output channels have consumers

		if ox.next_chan != len(ox.out_chans) {
			panic(fmt.Sprintf(
				"exec: %s: expected %d consumed chans, got %d",
				n, len(ox.out_chans), ox.next_chan,
			))
		}

		uint8_out[i] = ox.out_chans[0]
		i++
	}
	flo.confluent_count++

	//  map bool output of each predicate onto a fanin channel.
	//  bool_chan[0] is reserved for the fanin channel

	bool_out := make([]bool_chan, len(pred2out))
	i = 0
	for n, op := range pred2out {

		//  cheap sanity test that all output channels have consumers

		if op.next_chan != len(op.out_chans) {
			panic(fmt.Sprintf(
				"pred %s: expected %d consumed chans, got %d",
				n, len(op.out_chans), op.next_chan,
			))
		}

		bool_out[i] = op.out_chans[0]
		i++
	}
	flo.confluent_count++

	//  reduce() counts as one a conflowing go routine

	flo.confluent_count++

	return flo.reduce(
			flo.reduce_uint8(uint8_out),
			flo.reduce_bool(bool_out),
	)
}
