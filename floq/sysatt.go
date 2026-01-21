package main

import (
	"fmt"
	"strings"
)

type sysatt struct {

	name		string
	command_ref	*command
	call_order	uint8
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
%s          @: %p
%s}`,
		tab, cmd, cmd,
		tab, sa.call_order,
		tab, sa,
		strings.Repeat("\t", indent),
	)
}
