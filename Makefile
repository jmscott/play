include local.mk
include play.mk

MKMK=./jmscott-play.mkmk

BINs := $(shell (. $(MKMK) && echo $$BINs))
SRCs := $(shell (. $(MKMK) && echo $$SRCs))
COMPILEs := $(shell (. $(MKMK) && echo $$COMPILEs))

all: $(COMPILEs)

clean:
	rm -f $(COMPILEs)

install-dirs:
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=rx	\
		-d $(PLAY_PREFIX)
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=rx	\
		-d $(PLAY_PREFIX)/bin
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=rx	\
		-d $(PLAY_PREFIX)/src

install: install-dirs all
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m ugo=xr	\
		$(COMPILEs)						\
		$(PLAY_PREFIX)/bin
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -m u=rwx,go=rx	\
		$(SRCs)							\
		$(PLAY_PREFIX)/src

utf8-frisk: utf8-frisk.c
	cc -Wall -Wextra -o utf8-frisk utf8-frisk.c

tar: all
	make-make tar $(MKMK)

distclean:
	rm -rf $(PLAY_PREFIX)/bin
	rm -rf $(PLAY_PREFIX)/src

world: clean all distclean install
