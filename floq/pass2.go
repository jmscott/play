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
	run_proj	map[*projection][]*ast

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

	_e := func(format string, args...interface{}) error {
		return a.error("xrun_sysatt: " + format, args...)
	}

	switch a.yy_tok {

	//  add projection to list of what references this "run".
	case PROJECT_OSX_EXIT_CODE,
	     PROJECT_OSX_PID,
	     PROJECT_OSX_START_TIME,
	     PROJECT_OSX_WALL_DURATION,
	     PROJECT_OSX_USER_SEC,  PROJECT_OSX_USER_USEC,
	     PROJECT_OSX_SYS_SEC,  PROJECT_OSX_SYS_USEC,
	     PROJECT_OSX_STDOUT,
	     PROJECT_OSX_STDERR:
		proj := a.proj_ref

		cmd := proj.sysatt_ref.command_ref
		ar := p2.run_call[cmd]
		if ar == nil {
			return _e("command for sysatt never run: %s", cmd.name)
		}
		if ar.line_no >= a.line_no {
			return _e("run call after sysatt: %s", )
		}

		if len(p2.run_proj[proj]) == 255 {
			return _e(
				"too many sysatt ref to command: %s",
				cmd.name,
			)
		}

		//  append PROJECT_OSX... ast node to array of sysatt
		//  references.
		p2.run_proj[proj] = append(p2.run_proj[proj], a)
		a.proj_ref.call_order = uint8(len(p2.run_proj[proj]))
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

	if a.yy_tok == PROJECT_OSX_TUPLE_TSV {

		/*
		proj := a.proj_ref
		cmd := a.proj_ref.att_ref.command_ref

		ar := p2.run_call[cmd]
		if ar == nil {
			return _e("command for att never run: %s", cmd.name)
		}
		if ar.line_no >= a.line_no {
			return _e("run call after att: %s", proj.name)
		}

		if len(p2.run_proj[proj]) == 255 {
			return _e(
				"too many att ref to command: %s",
				cmd.name,
			)
		}

		//  append PROJECT_OSX... ast node to array of sysatt
		//  references.
		p2.run_proj[proj] = append(p2.run_proj[proj], a)
		proj.call_order = uint8(len(p2.run_proj[proj]))
		*/
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
		a.corrupt("run_depends: " + format, args...)
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
		proj := a.proj_ref
		if proj == nil {
			_c("proj_ref is nil")
		}
		_e("%#v", proj)

	case PROJECT_OSX_TUPLE_TSV:
	/*
		proj := a.proj_ref
		if proj == nil {
			_c("proj_ref is nil")
		}

		//p2.depends[run.name] = cmd.name
	*/
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

	return fmt.Errorf("pass2: " + format, args...)
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

func (p2 *pass2) parse_set(a *ast) error {

	if a == nil {
		return nil
	}
	if a.parent == nil {
		a.corrupt("a.parent is nil")
	}

	//  cheap sanity checks

	if a.parent.yy_tok != DEFINE && a.parent.yy_tok != yy_SET {
		a.corrupt("a.parent not DEFINE nor yy_SET")
	}
	s := a.set_ref
	if s == nil {
		a.corrupt("a.set_ref is nil")
	}
	for ele := a.left;  ele != nil;  ele = ele.next {

		switch ele.yy_tok {
		case yy_TRUE, yy_FALSE:
			if err := s.add_bare_bool(ele.bool);  err != nil {
				return err
			}
		case UINT64:
			if err := s.add_bare_uint64(ele.uint64);  err != nil {
				return err
			}
		case STRING:
			if err := s.add_bare_string(ele.string);  err != nil {
				return  err
			}
		case yy_SET:
			//  parse elements of set before this set
			if err := p2.parse_set(ele);  err != nil {
				return err
			}
			if err := s.add_bare_set(ele.set_ref);  err != nil {
				return  err
			}
		default:
			ele.corrupt("unexpected element: %s", ele)
		}
	}

	return nil
}

func (p2 *pass2) parse_sets(root *ast) error {

	_c := func(format string, args...interface{}) {
		root.corrupt("parse_sets(): " + format, args...)
	}

	//  cheap sanity tests of tree
	if root.left == nil {
		_c("root.left is nil")
	}

	if root.left.yy_tok != STMT_LIST {
		_c("root.left != STMT_LIST: %s", yy_name(root.left.yy_tok))
	}
	if root.left.left == nil {
		_c("root.left.left is nil")
	}
		
	for stmt := root.left.left;  stmt != nil;  stmt = stmt.next {
		if stmt.yy_tok != DEFINE {
			continue
		}

		//  parse the outer mst set of "define set <name> as {...}"

		aset := stmt.left
		if aset == nil {
			_c("stmt.left is nil")
		}

		if aset.yy_tok != yy_SET {
			continue
		}
		if aset.set_ref == nil {
			_c("stmt.set_reft is nil: %s", aset)
		}

		if err := p2.parse_set(aset);  err != nil {
			return err
		}
	}
	return nil;
}

//  frisk&optimize abstract syntax tree compiled by pass1 (yacc grammar)

func xpass2(root *ast) error {

	if root == nil {
		return errors.New("root is nil")
	}

	if root.yy_tok != FLOW {
		root.corrupt("root not yy FLOW")
	}
	if root.parent != nil {
		root.corrupt("parent of root not nil: %s", root.parent)
	}
	if root.left == nil {
		return nil
	}
	if root.left.parent != root {
		root.corrupt("left: parent not root: %s", root.left)
	}
	if root.left.yy_tok != STMT_LIST {
		root.corrupt("left: not STMT_LIST: %s", root.left)
	}

	if root.right != nil {
		root.corrupt("root.right not nil: %s", root.right)
	}

	p2 := &pass2{
		root:		root,
		run:		make(map[string]*ast),
		depends:	make(map[string]string),
		run_call:	make(map[*command]*ast),
		run_proj:	make(map[*projection][]*ast),
	}

	p2.plumb(root.left)
	p2.plumb(root.right)

	if err := p2.parse_sets(root);  err != nil {
		return err
	}

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
