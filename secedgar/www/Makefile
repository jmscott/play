#
#  Synopsis:
#	Makefile for www gui on SECEdgar.
#
  
include ../../local.mk

ifndef JMSCOTT_ROOT
$(error env JMSCOTT_ROOT is not set)
endif

DIST := secedgar-www.dist
HTDOCs :=	$(shell (. ./$(DIST) && echo $$HTDOCs))
LIBs :=		$(shell (. ./$(DIST) && echo $$LIBs))
CGIBINs :=	$(shell (. ./$(DIST) && echo $$CGIBINs))
COMPILEs :=	$(shell (. ./$(DIST) && echo $$COMPILEs))

WWW_USER=$(INSTALL_USER)
WWW_GROUP=$(INSTALL_GROUP)
WWW_ROOT=$(INSTALL_PREFIX)/www

all: $(COMPILEs)

clean:
	rm -f $(COMPILEs)
world:
	$(MAKE) $(MFLAGS) clean
	$(MAKE) $(MFLAGS) all
	$(MAKE) $(MFLAGS) distclean
	$(MAKE) $(MFLAGS) install
distclean:
	rm -rf $(WWW_ROOT)/htdocs $(WWW_ROOT)/lib $(WWW_ROOT)/cgi-bin

install: all
	install								\
		-g $(WWW_GROUP)						\
		-o $(WWW_USER)						\
		-m u=rwx,g=x,o=						\
		-d $(WWW_ROOT)

	install								\
		-g $(WWW_GROUP)						\
		-o $(WWW_USER)						\
		-m u=rwx,g=x,o=						\
		-d $(WWW_ROOT)/htdocs
	install								\
		-g $(WWW_GROUP)						\
		-o $(WWW_USER)						\
		-m u=rw,g=r,o=						\
		-C							\
		$(HTDOCs)						\
		$(WWW_ROOT)/htdocs

	install								\
		-g $(WWW_GROUP)						\
		-o $(WWW_USER)						\
		-m u=xrw,g=x,o=						\
		-d $(WWW_ROOT)/cgi-bin

	test -e $(WWW_ROOT)/cgi-bin/jmscott ||			\
		ln -s $(JMSCOTT_ROOT)/www/cgi-bin 			\
			$(WWW_ROOT)/cgi-bin/jmscott

	echo WTF
	install -g $(INSTALL_GROUP) -o $(INSTALL_USER) -C -m ug=rx,o=         \
                $(CGIBINs)                                              \
                $(WWW_ROOT)/cgi-bin
	install								\
		-g $(WWW_GROUP)						\
		-o $(WWW_USER)						\
		-m u=rwx,g=rx,o=					\
		-d $(WWW_ROOT)/lib
	test -e $(WWW_ROOT)/lib/jmscott ||				\
		ln -s $(JMSCOTT_ROOT)/www/lib $(WWW_ROOT)/lib/jmscott
	rsync								\
		--relative						\
		--no-perms						\
		--chmod=u=rw,Du+x,g=r,Dg+x,o=				\
		$(LIBs)							\
		$(WWW_ROOT)/lib

dist: all
	make-dist $(DIST)
daily: daily.cgi
	cgi2perl5 --source-path daily.cgi
	chmod +x daily
