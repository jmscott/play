package main

import (
	"errors"
	"fmt"
)

type pass2 struct {

	root	*ast

	run		map[string]*ast
	depends		map[string]string

	active_run	*ast
}

/*
func (rw *rewire) command_sysatt(a *ast) {

	if a == nil {
		return
	}

	if a.yy_tok != COMMAND_SYSATT {
		a.left.rewire_command_sysatt()
		a.right.rewire_command_sysatt()
		return
	}

	sa := a.sysatt_ref
	if sa.name != "exit_status" {
		a.corrupt("rewire_command_sysatt: impossible att: %s", sa)
	}
	a.yy_tok = PROJECT_OSX_EXIT_STATUS
}
*/

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
		err := p2.plumb(kid)
		if err != nil {
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

	if a.prev != nil {	//  avoid redundant plumbs of siblings
		return nil
	}

	//  plumb each sibling

	var prev *ast
	for sib := a.next;  sib != nil;  sib = sib.next {
		if prev != nil {
			if sib.prev == nil {
				return sib.error("prev sibling is nil")
			}
			if sib.prev != prev {
				return sib.error("prev sibling not %s", prev)
			}
		}
		err := p2.plumb(sib)
		if err != nil {
			return err
		}
		prev = sib
	}
	return nil
}

func (p2 *pass2) find_run() {
	if p2.root.left == nil {
		return
	}

	for stmt := p2.root.left.left;  stmt != nil;  stmt = stmt.next {
		if stmt.yy_tok == RUN {
			p2.run[stmt.command_ref.name] = stmt
		}
	}
}

func parse2(root *ast) error {

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
		return _err("parent of not nil: %s", root.parent)
	}
	if root.left != nil && root.left.parent != root {
		return _err("left: parent not root: %s", root.left)
	}
	if root.right != nil && root.right.parent != root {
		return _err("right: parent not root: %s", root.right)
	}

	p2 := &pass2{
		root:	root,
		run:	make(map[string]*ast),
	}

	if err := p2.plumb(root.left);  err != nil {
		return err
	}
	if err := p2.plumb(root.right);  err != nil {
		return err
	}

	p2.find_run()
	return nil
}
