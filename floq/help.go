package main

import (
	"os"
)


var help_tuple = `
define tuple blob_request_record as {
	attributes: {
		att1: {
			matches:
	"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})",
			tab_field: 1
		},
		att2: {
			matches: "[a-z][a-z0-9_]{0,7}~[[:graph:]]{1,128}",
			tab_field: 2
		}
	}
};
`

func help(argc int, argv []string) {

	if argc == 0 {
		os.Stdout.WriteString(usage)
		os.Exit(0)
	}

	switch argv[0] {
	case "tuple":
		os.Stdout.WriteString(help_tuple)
	default:
		croak("unknown help option: %s", argv[0])
	}
	os.Exit(0)
}
