package main

import (
	"os"
)


var help_tuple = `

define tuple <name> as {
	attributes: {
		interface: {
			matches: "^[a-z][a-z0-9_]{0,7}$",
			tab_field: 1
		},
		byte_count: {
			#  64bit unsigned
			matches: "^[1-9][0-9]{19}$",
			tab_field: 2
		}
	}
};
`

var help_env = `

export FLOQ_YYDEBUG
	Synopsis:
		Debug level of yacc parsing states, written to stdout
	Usage:
		export FLOQ_YYDEBUG=4
		floq frisk bug.conf

export FLOQ_TRACE_COMPILE
	Synopsis:
		Compilation tracing, written to stdout
	Usage:
		export FLOQ_TRACE_COMPILE=true
		floq compile bug.floq

export FLOQ_TRACE_NEXT_OP
	Synopsis:
		Trace transition from current to next goop, written to stdout
	Usage:
		export FLOQ_TRACE_NEXT_OP=true		
		floq server bug.floq
`

var help_command = `

define command <name>[.<tuple>] as {
	path:	"seq",
	args:	["--equal-width"],
	env:	["PATH=/usr/bin:/bin", "HOME=/home/blobio"]
}
};
`

func help(argc int, argv []string) {

	if argc == 0 {
		os.Stdout.WriteString(usage)
		os.Stdout.WriteString("help: floq help [tuple|command|env]\n")
		os.Exit(0)
	}

	switch argv[0] {
	case "env":
		os.Stdout.WriteString(help_env)
	case "tuple":
		os.Stdout.WriteString(help_tuple)
	case "command":
		os.Stdout.WriteString(help_command)
	default:
		croak("unknown help option: %s", argv[0])
	}
	os.Exit(0)
}
