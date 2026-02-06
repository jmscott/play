package main

/*
 *  Synopsis:
 *	Validate abstract syntax tree after correctly parse by pass1.
 */

import (
	"errors"
	"os/exec"
	"slices"
	"fmt"
)

type pass2 struct {

	root	*ast

	run		map[string]*ast
	depends		map[string]string

	//  track sysatts ast nodes referenced in "run <command>" statements.
	//
	//  references can be in either the "when" clause or the argument
	//  vector "run command(args...)"
	run_sysatt	map[*command][]*ast

	//  track PROJECT_OSX_TUPLE_TSV ast nodes referenced in "run <command>"
	//  statements.
	//
	//  references can be in either the "when" clause or the argument
	//  vector "run command(args...)"
	run_att		map[*command][]*ast

	//  "run <command>"  statements
	run_call	map[*command]*ast
}

//  depth first check of node pointers

func (p2 *pass2) plumb(a *ast) {

	if a == nil {
		return
	}
	if a.parent == nil {
		a.corrupt("parent is nil")
	}

	plumb_kid := func(what string, kid *ast) {
		if kid == nil {
			return
		}
		what = fmt.Sprintf("%s: %s", what, kid)
		if kid.prev != nil {
			kid.corrupt("%s: prev %s not nil", what, kid.prev)
		}
		if kid.parent != a {
			kid.corrupt("%s: parent not %s", what, a)
		}
		p2.plumb(kid)
	}

	plumb_kid("left", a.left)
	plumb_kid("right", a.right)

	if a.prev != nil {	//  in middle of sibling list

		//  Note: can we test order?
		return
	}

	if a.prev != nil {
		p := a.prev
		if p.next == nil {
			p.corrupt("next is nil in sib list")
		}
		if p.next != a {
			p.corrupt("next is wrong sib: %s", a)
		}
	}

	for sib, prev := a.next, (*ast)(nil);  sib != nil;  sib = sib.next {
		if prev != nil {
			if sib.prev == nil {
				sib.corrupt("sib.prev sibling is nil")
			}
			if sib.prev != prev {
				sib.corrupt("prev sibling not %s", prev)
			}
		}
		p2.plumb(sib)
		prev = sib
	}
}

//  build map of "run command(...)" nodes, indexed by command name.
//  each command can have only one "run" command.

func (p2 *pass2) map_run() {
	if p2.root.left == nil {
		return
	}

	for stmt := p2.root.left.left;  stmt != nil;  stmt = stmt.next {
		if stmt.yy_tok == RUN {
			p2.run[stmt.command_ref.name] = stmt
		}
	}
}

func (p2 *pass2) xrun(a *ast) error {
	if a == nil {
		return nil
	}
	if err := p2.xrun(a.left);  err != nil {
		return err
	}
	if err := p2.xrun(a.right);  err != nil {
		return err
	}

	if a.yy_tok == RUN {
		cmd := a.command_ref

		if cmd == nil {
			croak("command_ref is nil")
		}
		if p2.run_call[cmd] != nil {
			return a.error("run more than once: %s", cmd.name)
		}
		p2.run_call[cmd] = a
	}

	return p2.xrun(a.next)
}

//  verify all projections of system attributes of a <command> occur
//  after "run <command> ..."

func (p2 *pass2) xrun_sysatt(a *ast) error {

	if a ==  nil {
		return nil
	}

	if err := p2.xrun_sysatt(a.left);  err != nil {
		return err
	}
	if err := p2.xrun_sysatt(a.right);  err != nil {
		return err
	}

	_c := func(format string, args...interface{}) {
		a.corrupt(format, args...)
	}

	_e := func(format string, args...interface{}) error {
		return a.error(format, args...)
	}

	switch a.yy_tok {

	//  add sysatt to list of what references this "run".
	case PROJECT_OSX_EXIT_CODE,
	     PROJECT_OSX_PID,
	     PROJECT_OSX_START_TIME,
	     PROJECT_OSX_WALL_DURATION,
	     PROJECT_OSX_USER_SEC,  PROJECT_OSX_USER_USEC,
	     PROJECT_OSX_SYS_SEC,  PROJECT_OSX_SYS_USEC,
	     PROJECT_OSX_STDOUT,
	     PROJECT_OSX_STDERR:
		sa := a.sysatt_ref
		if sa == nil {
			_c("sysatt_ref is nil")
		}
		cmd := sa.command_ref
		if cmd == nil {
			_c("sysatt_ref.command_ref is nil")
		}

		ar := p2.run_call[cmd]
		if ar == nil {
			return _e("command for sysatt never run: %s", sa)
		}
		if ar.line_no >= a.line_no {
			return _e("run call after sysatt: %s", sa)
		}

		if len(p2.run_sysatt[cmd]) == 255 {
			return _e(
				"too many sysatt ref to command: %s",
				cmd.name,
			)
		}

		//  append PROJECT_OSX... ast node to array of sysatt
		//  references.
		p2.run_sysatt[cmd] = append(p2.run_sysatt[cmd], a)
		a.sysatt_ref.call_order = uint8(len(p2.run_sysatt[cmd]))
	}
	return p2.xrun_sysatt(a.next)
}

//  find all sysatt *ast nodes to "run <command>" statement

func (p2 *pass2) xrun_att(a *ast) error {

	if a ==  nil {
		return nil
	}

	if err := p2.xrun_att(a.left);  err != nil {
		return err
	}
	if err := p2.xrun_att(a.right);  err != nil {
		return err
	}
	
	_c := func(format string, args...interface{}) {
		a.corrupt("pass2: xrun_att: " + format, args...)
	}

	_e := func(format string, args...interface{}) error {
		return a.error(format, args...)
	}

	if a.yy_tok == PROJECT_OSX_TUPLE_TSV {

		att := a.att_ref
		if att == nil {
			_c("att_ref is nil")
		}
		cmd := a.command_ref
		if cmd == nil {
			_c("att_ref.command_ref is nil")
		}

		ar := p2.run_call[cmd]
		if ar == nil {
			return _e("command for att never run: %s", att)
		}
		if ar.line_no >= a.line_no {
			return _e("run call after att: %s", att)
		}

		if len(p2.run_sysatt[cmd]) == 255 {
			return _e(
				"too many att ref to command: %s",
				cmd.name,
			)
		}

		//  append PROJECT_OSX... ast node to array of sysatt
		//  references.
		p2.run_att[cmd] = append(p2.run_att[cmd], a)
		a.att_ref.call_order = uint8(len(p2.run_att[cmd]))
	}
	return p2.xrun_att(a.next)
}

func (p2 *pass2) run_depends(a *ast) error {

	if a == nil {
		return nil
	}

	_e := func(format string, args...interface{}) error {
		return a.error(format, args...)
	}

	_c := func(format string, args...interface{}) {
		a.corrupt(format, args...)
	}

	switch a.yy_tok {
	case RUN:
		rn := p2.run[a.name]
		if rn == nil {
			_c("node is nil in map p2.run")
		}
		if rn != a {
			_c("node in p2.run unexpected: %s", rn)
		}
	case PROJECT_OSX_EXIT_CODE,
	     PROJECT_OSX_PID,
	     PROJECT_OSX_START_TIME,
	     PROJECT_OSX_WALL_DURATION,
	     PROJECT_OSX_USER_SEC,
	     PROJECT_OSX_USER_USEC,
	     PROJECT_OSX_SYS_SEC,
	     PROJECT_OSX_SYS_USEC,
	     PROJECT_OSX_STDOUT,
	     PROJECT_OSX_STDERR:
		sa := a.sysatt_ref
		if sa == nil {
			_c("sysatt_ref is nil")
		}
		cmd := sa.command_ref
		if cmd == nil {
			_c("sysatt_ref.command_ref is nil")
		}

		//  verify "run <command(...)" statement occurs before
		//  an reference to its projected values.
		if p2.run[cmd.name] == nil {
			return _e("command never run: %s", cmd.name)
		}

		//  track cyclic references
		//p2.depends[a.name] = cmd.name

	case PROJECT_OSX_TUPLE_TSV:
		ar := a.att_ref
		if ar == nil {
			_c("att_ref is nil")
		}

		tup := ar.tuple_ref
		if tup == nil {
			_c("tuple_ref is nil")
		}

		cmd := a.command_ref
		if cmd == nil {
			_c("cmd is nil")
		}

		//p2.depends[run.name] = cmd.name
	}
	if err := p2.run_depends(a.left);  err != nil {
		return err
	}
	if err := p2.run_depends(a.right);  err != nil {
		return err
	}
	if a.prev != nil {	//  in middle of sibling list
		return nil
	}

	for sib := a.next;  sib != nil;  sib = sib.next {
		if err := p2.run_depends(sib);  err != nil {
			return err
		}
	}
	return nil
}

func (p2 *pass2) error(format string, args...interface{}) error {

	return errors.New(fmt.Sprintf("pass2: " + format, args...)) 
}

//  find cyclic dependencies.

func (p2 *pass2) cycle() error {

	for stmt := p2.root.left.left;  stmt != nil;  stmt = stmt.next {
		
		if stmt.yy_tok != RUN {
			continue
		}

		if err := p2.run_depends(stmt.left);  err != nil {
			return err
		}

		if err := p2.run_depends(stmt.right);  err != nil {
			return err
		}
	}

	//  check for cyclic dependencies
	var depends []string
	for key, val := range p2.depends {
		depends = append(depends, key + " " + val)
	}
	if len(depends) > 0 && tsort(depends) == nil {
		return p2.error("cyclic dependncy")
	}
	return nil
}

func (p2 *pass2) look_path(a *ast) error {

	if a == nil {
		return nil
	}
	if err := p2.look_path(a.left);  err != nil {
		return err
	}
	if err := p2.look_path(a.right);  err != nil {
		return err
	}
	if a.yy_tok == RUN {
		cmd := a.command_ref
		look_path, err := exec.LookPath(cmd.path)
		if err != nil {
			return p2.error(
				"LookPath(%s) failed: %s",
				cmd.path,
				err,
			)
		}
		cmd.look_path = look_path
		//  set argv[0] == path to executable
		cmd.args = slices.Insert(cmd.args, 0, look_path)
	}
	return p2.look_path(a.next)
}

func (p2 *pass2) run_parent_argv(a *ast) {

	if a == nil {
		return
	}
	p2.run_parent_argv(a.left)
	p2.run_parent_argv(a.right)
	if a.yy_tok == ARGV {
		p := a.parent
		if p.yy_tok != RUN {
			a.corrupt("parent not RUN: %s", a.parent)
		}
		cmd := p.command_ref
		if cmd == nil {
			a.corrupt("parent command_ref is nil: %s", p)
		}
	}
	p2.run_parent_argv(a.next)
}

func (p2 *pass2) argv_is_string(a *ast) error {
	if a == nil {
		return nil
	}
	if err := p2.argv_is_string(a.left);  err != nil {
		return err
	}
	if err := p2.argv_is_string(a.right);  err != nil {
		return err
	}
	if a.yy_tok == ARGV {
		for arg := a.left;  arg != nil;  arg = arg.next {
			if arg.is_string() == false {
				return a.error(
					"arg #%d not string: %s",
					arg.order,
					arg,
				)
			}
		}
	}
	return p2.argv_is_string(a.next)
}

//  rewrite "::<type>" nodes to particular type: bool, uint64, string
func (p2 *pass2) cast(a *ast) error {

	if a == nil {
		return nil
	}
	if err := p2.cast(a.left);  err != nil {
		return err
	}
	if err := p2.cast(a.right);  err != nil {
		return err
	}
	if a.yy_tok == CAST {
		l := a.left
		switch {
		case l.is_uint64():
			a.yy_tok = CAST_UINT64
		case l.is_bool():
			a.yy_tok = CAST_BOOL
		case l.is_string():
			a.yy_tok = CAST_STRING
		default:
			a.corrupt("CAST: unknown left type: %s", l)
		}
	}
	return p2.cast(a.next)
}

//  rewrite "IS NULL" nodes for particular type: string, uint64, bool
func (p2 *pass2) is_null(a *ast) {

	if a == nil {
		return
	}
	p2.is_null(a.left)
	p2.is_null(a.right)

	l := a.left
	switch a.yy_tok {
	case IS_NULL:
		switch {
		case l.is_string():
			a.yy_tok = IS_NULL_STRING
		case l.is_uint64():
			a.yy_tok = IS_NULL_UINT64
		case l.is_bool():
			a.yy_tok = IS_NULL_BOOL
		}
	case IS_NOT_NULL:
		switch {
		case l.is_string():
			a.yy_tok = IS_NOT_NULL_STRING
		case l.is_uint64():
			a.yy_tok = IS_NOT_NULL_UINT64
		case l.is_bool():
			a.yy_tok = IS_NOT_NULL_BOOL
		}
	}
	p2.is_null(a.next)
}

//  frisk&optimize abstract syntax tree compiled by pass1 (yacc grammar)

func xpass2(root *ast) error {

	if root == nil {
		return errors.New("root is nil")
	}

	_c := func(format string, args...interface{}) {
		root.corrupt("xpass2: " + format, args...)
	}

	if root.yy_tok != FLOW {
		_c("root not yy FLOW")
	}
	if root.parent != nil {
		_c("parent of root not nil: %s", root.parent)
	}
	if root.left == nil {
		return nil
	}
	if root.left.parent != root {
		_c("left: parent not root: %s", root.left)
	}
	if root.left.yy_tok != STMT_LIST {
		_c("left: not STMT_LIST: %s", root.left)
	}

	if root.right != nil {
		_c("root.right not nil: %s", root.right)
	}

	p2 := &pass2{
		root:		root,
		run:		make(map[string]*ast),
		depends:	make(map[string]string),
		run_call:	make(map[*command]*ast),
		run_sysatt:	make(map[*command][]*ast),
		run_att:	make(map[*command][]*ast),
	}

	p2.plumb(root.left)
	p2.plumb(root.right)

	p2.map_run()		//  build a map of "run <command" nodes


	//  verify "run command only occurs once"
	if err := p2.xrun(root);  err != nil {
		return err
	}

	p2.is_null(root)	// rewrite "IS NULL" ops

	if err := p2.cycle();  err != nil {	//  find cyclic dependencies
		return err
	}

	//  resolve file system paths to executables in COMMAND_REF nodes

	if err := p2.look_path(root);  err != nil {
		return err
	}

	//  verify <command>$sysatt expressions after "run <command>"
	//  statements.
	if err := p2.xrun_sysatt(root);  err != nil {
		return err
	}

	//  verify <command>.attribute expressions after "run <command>"
	//  statements.
	if err := p2.xrun_att(root);  err != nil {
		return err
	}

	if err := p2.cast(root);  err != nil {
		return err
	}

	//  all arguments to argv must be a string
	if err := p2.argv_is_string(root);  err != nil {
		return err
	}

	p2.run_parent_argv(root)

	/*
	 *  second pass check of tree plumbing
	 *
	 *  Note:
	 *	would be nice to specify error occured in after rewiring
	 */

	p2.plumb(root.left)
	p2.plumb(root.right)
	return nil
}
