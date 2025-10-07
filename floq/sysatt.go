package main

import (
	"fmt"
	"regexp"
	"strings"
)

type sysatt struct {

	name		string
	command_ref	*command
	call_order	uint8
}

var re_sysatt_indent *regexp.Regexp

func init() {

	re_sysatt_indent = regexp.MustCompile("(?m)\t([a-z])")
}

func (sa *sysatt) is_uint64() bool {

	if sa.command_ref != nil {
		return sa.command_ref.is_sysatt_uint64(sa.name)
	}
	return false
}

func (sa *sysatt) String() string {

	return sa.command_ref.name + "$" + sa.name
}

func (sa *sysatt) string(indent int) string {

	if sa == nil {
		return "nil sysatt"
	}
	cmd := sa.command_ref
	tab := strings.Repeat("\t", indent)
	return fmt.Sprintf(`{
%scommand_ref: %s@%p,
%s call_order: %d
%s  }`,
		tab,
		cmd,
		cmd,
		tab,
		sa.call_order,
		strings.Repeat("\t", indent - 1),
	)
}
