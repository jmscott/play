include ../local.mk
include ../play.mk

DIST=./jmscott-play-secedgar.dist

SBINs := $(shell (. $(DIST) && echo $$SBINs))
LIBs := $(shell (. $(DIST) && echo $$LIBs))
COMPILED := $(shell (. $(DIST) && echo $$COMPILED))

clean:
	rm -f $(COMPILED)

install:
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=rx	\
		-d $(PLAY_PREFIX)
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=	\
		-d $(PLAY_PREFIX)/sbin
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=	\
		-d $(PLAY_PREFIX)/spool
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=rx	\
		-d $(PLAY_PREFIX)/lib
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=	\
		-d $(PLAY_PREFIX)/tmp

	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m ugo=xr	\
		$(SBINs)						\
		$(PLAY_PREFIX)/sbin
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m ugo=r		\
		$(LIBs)							\
		$(PLAY_PREFIX)/lib

dist:
	make-dist jmscott-play-secedgar.dist

distclean:
	rm -rf $(PLAY_PREFIX)/sbin
	rm -rf $(PLAY_PREFIX)/lib

world:
	$(MAKE) $(MFLAGS) clean
	$(MAKE) $(MFLAGS) distclean
	$(MAKE) $(MFLAGS) install
dev-links:
	test -e data || ln -s . data
	test -e run || ln -s . run
	test -e etc || ln -s . etc
	test -e spool || ln -s . spool
	test -e profile || cp profile.example profile
