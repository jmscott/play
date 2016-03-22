package main

type compile struct {
	*ast
	*flow
}

func (cmpl *compile) compile() fdr_chan {

	type call_output struct {

		//  fanout channels listening for exit_status of a process

		out_chans []uint8_chan

		//  next free channel

		next_chan int
	}

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
			//  argv() needs a string_chan slice
			in := make([]string_chan, a.uint64)

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
			cmd := a.call.command
			a2x[a] = flo.call(
				cmd,
				cmpl.os_exec_chan,
				a2a[a.left],
				a2b[a.right],
			)

			//  broadcast to all dependent calls
			command2uint8[cmd.name] =
				&call_output{
					out_chans: flo.fanout_xdr(
						a2x[a],
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
			cx := command2uint8[a.command.name]

			//  Note: zap this test and add a sanity test at the
			//	  end of the compile that verifys that
			//	  cx.next_chan == len(cx.out_chans)

			if cx == nil {
				panic("missing command -> uint8 map for " +
						a.command.name)
			}
			a2u[a] = flo.project_uint8(cx.out_chans[cx.next_chan])
			cx.next_chan++

		case EQ_UINT64:
			a2b[a] = flo.eq_uint64(a.uint64, a2u[a.left])
		case NEQ_UINT64:
			a2b[a] = flo.neq_uint64(a.uint64, a2u[a.left])
		case EQ_STRING:
			a2b[a] = flo.eq_string(a.string, a2s[a.left])
		case NEQ_STRING:
			a2b[a] = flo.neq_string(a.string, a2s[a.left])
		case EQ_BOOL:
			a2b[a] = flo.eq_bool(a.bool, a2b[a.left])
		case OR:
			a2b[a] = flo.bool2(or, a2b[a.left], a2b[a.right])
		case AND:
			a2b[a] = flo.bool2(and, a2b[a.left], a2b[a.right])
		default:
			panic(Sprintf("impossible yy_tok in ast: %d", a.yy_tok))
		}
		flo.confluent_count += cc
	}

	//  compile nodes from least dependent to most dependent order
	for _, n := range par.depend_order {

		//  skip tail dependency
		if n == conf.tail.name {
			continue
		}

		var root *ast
		if root = par.call2ast[n]; root == nil {
			root = par.query2ast[n]
		}
		if root == nil {
			panic(Sprintf("map to abstract syntax tree: %s", n))
		}
		compile(root)
	}

	//  Wait for all xdr to flow in before reducing the whole set
	//  into a single fdr record

	xdr_out := make([]xdr_chan, len(command2xdr))
	i := 0
	for n, cx := range command2xdr {

		//  cheap sanity test that all output channels have consumers
		if cx.next_chan != len(cx.out_chans) {
			panic(Sprintf(
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

		xdr_out[i] = flo.log_xdr_error(
			cmpl.info_log_chan,
			flo.log_xdr(
				cmpl.xdr_log_chan,
				cx.out_chans[0],
			))
		i++
	}
	flo.confluent_count += i

	//  Wait for all qdr to flow in before reducing the whole set
	//  into a single fdr record

	i = 0
	qdr_out := make([]qdr_chan, len(query2qdr))
	for n, qq := range query2qdr {

		//  cheap sanity test that all output channels have consumers
		if qq.next_chan != len(qq.out_chans) {
			panic(Sprintf(
				"%s: expected %d consumed chans, got %d",
				n,
				len(qq.out_chans),
				qq.next_chan,
			))
		}

		//  wait for the qdr log entry to be written.
		//
		//  Note:
		//	why make log_qdr_error() wait on log_qdr()?

		qdr_out[i] = flo.log_qdr_error(
			cmpl.info_log_chan,
			flo.log_qdr(
				cmpl.qdr_log_chan,
				qq.out_chans[0],
			))
		i++
	}
	flo.confluent_count += i

	flo.confluent_count += 2
	return flo.log_fdr(cmpl.fdr_log_chan, flo.reduce(xdr_out, qdr_out))
}
