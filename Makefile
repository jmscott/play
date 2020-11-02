include local.mk

PLAY_PREFIX=$(INSTALL_PREFIX)/play

BINs := $(shell (. ./play.dist && echo $$BINs))
SRCs := $(shell (. ./play.dist && echo $$SRCs))
LIBs := $(shell (. ./play.dist && echo $$LIBs))
COMPILED := $(shell (. ./play.dist && echo $$COMPILED))

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

bit-shift-left-one: bit-shift-left-one.c
	cc -Wall -Wextra -o bit-shift-left-one bit-shift-left-one.c

dist: all
	make-dist play.dist

distclean:
	rm -rf $(PLAY_PREFIX)/bin
	rm -rf $(PLAY_PREFIX)/lib
	rm -rf $(PLAY_PREFIX)/src

world:
	$(MAKE) $(MFLAGS) clean
	$(MAKE) $(MFLAGS) all
	$(MAKE) $(MFLAGS) distclean
	$(MAKE) $(MFLAGS) install
