package main

import (
	"fmt"
)

func (flo *flow) compile(ast_head *ast, depend_order []string) {

	type call_output struct {

		//  fanout channels listening for exit_status of a process

		out_chans []uint8_chan

		//  next free channel

		next_chan int
	}

	//  map the name of the CALL nodes to their abstract syntax tree node
	//  later, we will compile the nodes in DAG order

	call2ast := make(map[string]*ast)
	var find_call func(*ast)
	find_call = func(a *ast) {

		if a == nil {
			return
		}
		if a.yy_tok == CALL {
			call2ast[a.command.name] = a
		}
		find_call(a.left)
		find_call(a.right)
		find_call(a.next)
	}
	find_call(ast_head)

	//  map command output onto list of fanout channels

	command2uint8 := make(map[string]*call_output)

	//  map abstract syntax tree nodes to compiled channels

	a2b := make(map[*ast]bool_chan)
	a2u := make(map[*ast]uint8_chan)
	a2s := make(map[*ast]string_chan)
	a2a := make(map[*ast]argv_chan)

	var compile func(a *ast)
	compile = func(a *ast) {
		if a == nil {
			return
		}

		//  track number of confluing go routines compiled by each node
		cc := 1

		//  compile kids first

		compile(a.left)
		compile(a.right)

		switch a.yy_tok {
		case UINT8:
			a2u[a] = flo.const_uint8(a.uint8)
		case STRING:
			a2s[a] = flo.const_string(a.string)
		case ARGV0:
			a2a[a] = flo.argv0()
		case ARGV1:
			a2a[a] = flo.argv1(a2s[a.left])
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
			a2a[a] = flo.argv(in)

		//  call an os executable
		case CALL:
			cmd := a.command
			a2u[a] = flo.call(cmd, a2a[a.left], a2b[a.right])

			//  broadcast exit_status to all dependent calls

			command2uint8[cmd.name] =
				&call_output{
					out_chans: flo.fanout_uint8(
						a2u[a],

						//  each exit_status in
						//  qualification plus
						//  the fanin channel that
						//  terminates the flow

						cmd.depend_ref_count+1,
					),
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
			//  CALL must occur before exit_status reference

			cx := command2uint8[a.command.name]

			//  cheap sanity test

			if cx == nil {
				panic("missing command -> uint8 map for " +
					a.command.name)
			}
			a2u[a] = cx.out_chans[cx.next_chan]
			cx.next_chan++

		case EQ_UINT8:
			a2b[a] = flo.uint8_rel2(
				a2u[a.left],
				a2u[a.right],
				uint8_eq,
			)
		case NEQ_UINT8:
			a2b[a] = flo.uint8_rel2(
				a2u[a.left],
				a2u[a.right],
				uint8_neq,
			)
		case EQ_STRING:
			a2b[a] = flo.string_rel2(
				a2s[a.left],
				a2s[a.right],
				string_eq,
			)
		case NEQ_STRING:
			a2b[a] = flo.string_rel2(
				a2s[a.left],
				a2s[a.right],
				string_neq,
			)
		case OR:
			a2b[a] = flo.bool_rel2(
				a2b[a.left],
				a2b[a.right],
				or,
			)
		case AND:
			a2b[a] = flo.bool_rel2(
				a2b[a.left],
				a2b[a.right],
				and,
			)
		case TO_STRING_UINT8:
			a2s[a] = flo.to_string_uint8(
				a2u[a.left],
			)
		default:
			panic(fmt.Sprintf(
				"impossible yy_tok in ast: %d", a.yy_tok))
		}
		flo.confluent_count += cc
	}

	//  compile CALL nodes from least dependent to most dependent order

	for _, n := range depend_order {
		compile(call2ast[n])
	}

	uint8_out := make([]uint8_chan, len(command2uint8))
	i := 0
	for n, cx := range command2uint8 {

		//  cheap sanity test that all output channels have consumers
		if cx.next_chan != len(cx.out_chans) {
			panic(fmt.Sprintf(
				"%s: expected %d consumed chans, got %d",
				n,
				len(cx.out_chans),
				cx.next_chan,
			))
		}

		//  wait for the xdr log entry to be written.
		//
		//  Note:
		//	why make log_xdr_error() wait on log_xdr()?

		uint8_out[i] = cx.out_chans[0]
		i++
	}
	flo.confluent_count += i
}
