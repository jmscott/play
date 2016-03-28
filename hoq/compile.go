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

	//  map the COMMAND name onto their EXEC ast nodes.
	//
	//  later, we compile the nodes in order of directed acyclic graph,
	//  using tsort command

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
		case TRUE:
			a2b[a] = flo.const_bool(true)
		case FALSE:
			a2b[a] = flo.const_bool(false)
		case WHEN:
			a2b[a] = a2b[a.left]
			cc = 0
		case EXIT_STATUS:
			cx := cmd2out[a.command.name]

			//  cheap sanity test

			if cx == nil {
				panic("missing command -> uint8 map for " +
					a.command.name)
			}
			a2u8[a] = cx.out_chans[cx.next_chan]
			cx.next_chan++
			cc = 0

		case EQ_UINT8:
			a2b[a] = flo.uint8_rel2(
				uint8_eq,
				a2u8[a.left],
				a2u8[a.right],
			)
		case NEQ_UINT8:
			a2b[a] = flo.uint8_rel2(
				uint8_neq,
				a2u8[a.left],
				a2u8[a.right],
			)
		case EQ_STRING:
			a2b[a] = flo.string_rel2(
				string_eq,
				a2s[a.left],
				a2s[a.right],
			)
		case NEQ_STRING:
			a2b[a] = flo.string_rel2(
				string_neq,
				a2s[a.left],
				a2s[a.right],
			)
		case OR:
			a2b[a] = flo.bool_rel2(
				or,
				a2b[a.left],
				a2b[a.right],
			)
		case AND:
			a2b[a] = flo.bool_rel2(
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
			a2b[a] = flo.string_rel2(
				re_match,
				a2s[a.left],
				a2s[a.right],
			)
		case RE_NMATCH:
			a2b[a] = flo.string_rel2(
				re_nmatch,
				a2s[a.left],
				a2s[a.right],
			)
		case TO_STRING_BOOL:
			a2s[a] = flo.to_string_bool(
				a2b[a.left],
			)
		default:
			panic(fmt.Sprintf(
				"impossible yy_tok in ast: %d", a.yy_tok))
		}
		flo.confluent_count += cc
	}

	//  compile EXEC nodes from least dependent to most dependent order

	for _, n := range depend_order {
		compile(exec2ast[n])
	}

	//  map output of each exec.exit_status onto a fanin channel.
	//  out_chan[0] is reserved for the fanin channel

	uint8_out := make([]uint8_chan, len(cmd2out))
	i := 0
	for n, cx := range cmd2out {

		//  cheap sanity test that all output channels have consumers

		if cx.next_chan != len(cx.out_chans) {
			panic(fmt.Sprintf(
				"%s: expected %d consumed chans, got %d",
					n, len(cx.out_chans), cx.next_chan,
			))
		}

		uint8_out[i] = cx.out_chans[0]
		i++
	}

	//  fanin counts as one confluent_count

	flo.confluent_count++

	return flo.fanin_uint8(uint8_out)
}
