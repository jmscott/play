package main

/*
 *  Note:
 *	Pardon the lack of single function to recurse through the ast.
 *	methods are not allowed as function pointers.
 *	doing the trick like in relop_string[]@string.go is more verbose
 *	than eexplicitly doing the node walk.
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
	run_sysatt	map[*command][]*ast

	run_call	map[*command]bool
}


//  depth first check of node pointers

func (p2 *pass2) plumb(a *ast) error {

	if a == nil {
		return nil
	}
	if a.parent == nil {
		return a.error("parent is nil")
	}

	plumb_kid := func(what string, kid *ast) error {
		if kid == nil {
			return nil
		}
		what = fmt.Sprintf("%s: %s", kid)
		if kid.prev != nil {
			return kid.error("%s: prev %s not nil", what, kid.prev)
		}
		if kid.parent != a {
			return kid.error("%s:  parent not %s", what, a)
		}
		if err := p2.plumb(kid);  err != nil {
			return err
		}
		return nil
	}

	if err := plumb_kid("left", a.left);  err != nil {
		return err
	}
	if err := plumb_kid("right", a.right);  err != nil {
		return err
	}

	if a.prev != nil {	//  in middle of sibling list
		return nil
	}

	//  plumb each sibling

	for sib, prev := a.next, (*ast)(nil);  sib != nil;  sib = sib.next {
		if prev != nil {
			if sib.prev == nil {
				return sib.error("prev sibling is nil")
			}
			if sib.prev != prev {
				return sib.error("prev sibling not %s", prev)
			}
		}
		if err := p2.plumb(sib);  err != nil {
			return err
		}
		prev = sib
	}
	return nil
}

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

//  find all sysatt references to command in "run <command>" statement

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
	
	_die := func(format string, args...interface{}) {
		a.corrupt("pass2: xrun_sysatt: " + format, args...)
	}

	switch a.yy_tok {

	//  tag the active run branch

	case RUN:
		cmd := a.command_ref

		if cmd == nil {
			_die("command_ref is nil")
		}
		if p2.run_call[cmd] {
			return a.error("run more than once: %s", cmd.name)
		}
		p2.run_call[cmd] = true

	//  add sysatt to list of what references this active run

	case COMMAND_SYSATT, COMMAND_SYSATT_EXIT_CODE:
		sa := a.sysatt_ref
		if sa == nil {
			_die("sysatt_ref is nil")
		}
		cmd := sa.command_ref
		if cmd == nil {
			_die("sysatt_ref.command_ref is nil")
		}
		if len(p2.run_sysatt[cmd]) == 255 {
			return a.error("too many ref to coammnd: %s", cmd.name)
		}
		p2.run_sysatt[cmd] = append(p2.run_sysatt[cmd], a)
	}
	return p2.xrun_sysatt(a.next)
}

func (p2 *pass2) walk_depends(a *ast) error {

	if a == nil {
		return nil
	}

	_err := func(format string, args...interface{}) error {
		return a.error(format, args...)
	}

	switch a.yy_tok {
	case RUN:
		rn := p2.run[a.name]
		if rn == nil {
			return _err("node is nil in map p2.run")
		}
		if rn != a {
			return _err("node in p2.run unexpected: %s", rn)
		}
	case COMMAND_SYSATT:
		sa := a.sysatt_ref
		if sa == nil {
			return _err("sysatt_ref is nil")
		}
		cmd := sa.command_ref
		if cmd == nil {
			return _err("sysatt_ref.command_ref is nil")
		}

		run := a.yy_ancestor(RUN)
		if run == nil {
			a.corrupt("no ancestor RUN node")
		}

		if p2.run[cmd.name] == nil {
			return _err("command never run: %s", cmd.name)
		}
		p2.depends[run.name] = cmd.name
	}
	if err := p2.walk_depends(a.left);  err != nil {
		return err
	}
	if err := p2.walk_depends(a.right);  err != nil {
		return err
	}
	if a.prev != nil {	//  in middle of sibling list
		return nil
	}

	for sib := a.next;  sib != nil;  sib = sib.next {
		if err := p2.walk_depends(sib);  err != nil {
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

		if err := p2.walk_depends(stmt.left);  err != nil {
			return err
		}

		if err := p2.walk_depends(stmt.right);  err != nil {
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
	return
	p2.run_parent_argv(a.left)
	p2.run_parent_argv(a.right)
	if a.yy_tok == ARGV {
		p := a.parent
		if p.yy_tok != RUN {
			a.corrupt("parent not RUN: %s", a.parent)
		}
		cmd := p.command_ref
		if cmd == nil {
			p.corrupt("command_ref is nil")
		}
		cnt := a.count
		pcnt := uint32(len(cmd.args))
		if pcnt != cnt + 1 {
			a.corrupt("parent RUN pcnt=%d != cnt=%d+1", pcnt, cnt)
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
				return a.error("arg #%d not string", arg.order)
			}
		}
	}
	return p2.argv_is_string(a.next)
}

func xpass2(root *ast) error {

	if root == nil {
		return errors.New("root is nil")
	}

	_err := func (format string, args...interface{}) error {
		return root.error("root: " + format, args...)
	}

	if root.yy_tok != FLOW {
		return _err("not yy FLOW: %s", root)
	}
	if root.parent != nil {
		return _err("parent of root not nil: %s", root.parent)
	}
	if root.left == nil {
		return nil
	}
	if root.left.parent != root {
		return _err("left: parent not root: %s", root.left)
	}
	if root.left.yy_tok != STMT_LIST {
		return _err("left: not STMT_LIST: %s", root.left)
	}

	if root.right != nil {
		return _err("root.right not nil: %s", root.right)
	}

	p2 := &pass2{
		root:		root,
		run:		make(map[string]*ast),
		depends:	make(map[string]string),
		run_sysatt:	make(map[*command][]*ast),
		run_call:	make(map[*command]bool),
	}

	if err := p2.plumb(root.left);  err != nil {
		return err
	}
	if err := p2.plumb(root.right);  err != nil {
		return err
	}

	p2.map_run()

	if err := p2.cycle();  err != nil {
		return err
	}

	//  resolve paths to executables in COMMAND_REF nodes
	if err := p2.look_path(root);  err != nil {
		return err
	}

	//  check all references to RUN
	if err := p2.xrun_sysatt(root);  err != nil {
		return err
	}

	//  all arguments to argv must be a string
	if err := p2.argv_is_string(root);  err != nil {
		return err
	}

	p2.run_parent_argv(root)
	return nil
}
