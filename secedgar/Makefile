#
#  Synopsis:
#	Makefile file for secedgar system.
#  Note:
#	No dependency on jmscott/play exists, in theory.
#	However, dependency on jmscott/work exists. 
#
include ../local.mk
include ../play.mk
include secedgar.mk

DIST=./secedgar.dist

SBINs := $(shell (. $(DIST) && echo $$SBINs))
LIBs := $(shell (. $(DIST) && echo $$LIBs))
COMPILED := $(shell (. $(DIST) && echo $$COMPILED))

clean:
	rm -f $(COMPILED)

install:
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=rx	\
		-d $(SECEDGAR_PREFIX)
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=	\
		-d $(SECEDGAR_PREFIX)/sbin
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=	\
		-d $(SECEDGAR_PREFIX)/spool
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=	\
		-d $(SECEDGAR_PREFIX)/run
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=	\
		-d $(SECEDGAR_PREFIX)/data
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=	\
		-d $(SECEDGAR_PREFIX)/attic
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=	\
		-d $(SECEDGAR_PREFIX)/etc
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=rx	\
		-d $(SECEDGAR_PREFIX)/lib
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=	\
		-d $(SECEDGAR_PREFIX)/tmp

	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m ugo=xr	\
		$(SBINs)						\
		$(SECEDGAR_PREFIX)/sbin
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m ugo=r		\
		$(LIBs)							\
		$(SECEDGAR_PREFIX)/lib

dist:
	make-dist secedgar.dist

distclean:
	rm -rf $(SECEDGAR_PREFIX)/sbin
	rm -rf $(SECEDGAR_PREFIX)/lib

world:
	$(MAKE) $(MFLAGS) clean
	$(MAKE) $(MFLAGS) distclean
	$(MAKE) $(MFLAGS) install
dev-links:
	test -e attiattic || ln -s . attic
	test -e data || ln -s . data
	test -e etc || ln -s . etc
	test -e run || ln -s . run
	test -e sbin || ln -s . sbin
	test -e spool || ln -s . spool
	test -e profile || cp profile.example profile
