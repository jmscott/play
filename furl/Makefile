dev-links:
	test -e state || ln -s . state
	test -e spool || ln -s . spool
	test -e log || ln -s . log
	test -e etc || ln -s . etc
	test -e run || ln -s . run
	test -e tmp || ln -s . tmp
clean-dev-links:
	test ! -e state || rm state
	test ! -e spool || rm spool
	test ! -e log || rm log
	test ! -e etc || rm etc
	test ! -e run || rm run
	test ! -e tmp || rm tmp
