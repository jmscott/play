include local.mk
include play.mk

DIST=./jmscott-play.dist

BINs := $(shell (. $(DIST) && echo $$BINs))
SRCs := $(shell (. $(DIST) && echo $$SRCs))
LIBs := $(shell (. $(DIST) && echo $$LIBs))
COMPILED := $(shell (. $(DIST) && echo $$COMPILED))

all: $(COMPILED)

clean:
	rm -f $(COMPILED)

install: all
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=rx	\
		-d $(PLAY_PREFIX)
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=rx	\
		-d $(PLAY_PREFIX)/bin
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=rx	\
		-d $(PLAY_PREFIX)/src
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=rx	\
		-d $(PLAY_PREFIX)/lib

	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m ugo=xr	\
		$(COMPILED)						\
		$(PLAY_PREFIX)/bin
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=rx	\
		$(SRCs)							\
		$(PLAY_PREFIX)/src
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m ugo=r		\
		$(LIBs)							\
		$(PLAY_PREFIX)/lib

utf8-frisk: utf8-frisk.c
	cc -Wall -Wextra -o utf8-frisk utf8-frisk.c

dist: all
	make-dist $(DIST)

distclean:
	rm -rf $(PLAY_PREFIX)/bin
	rm -rf $(PLAY_PREFIX)/lib
	rm -rf $(PLAY_PREFIX)/src

world:
	$(MAKE) $(MFLAGS) clean
	$(MAKE) $(MFLAGS) all
	$(MAKE) $(MFLAGS) distclean
	$(MAKE) $(MFLAGS) install
