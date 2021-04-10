#
#  Synopsis:
#	Makefile for www gui on SECEdgar.
#
  
include local.mk

DIST := secedgar-www.dist
HTDOCs :=	$(shell (. ./$(DIST) && echo $$HTDOCs))
LIBs :=		$(shell (. ./$(DIST) && echo $$LIBs))
CGIBINs :=	$(shell (. ./$(DIST) && echo $$CGIBINs))
COMPILEs :=	$(shell (. ./$(DIST) && echo $$COMPILEs))

WWW_USER=$(DIST_USER)
WWW_GROUP=$(DIST_GROUP)
WWW_PREFIX=$(DIST_ROOT)/www

all: $(COMPILEs)

clean:
	rm -f $(COMPILEs)
world:
	$(MAKE) $(MFLAGS) clean
	$(MAKE) $(MFLAGS) all
	$(MAKE) $(MFLAGS) distclean
	$(MAKE) $(MFLAGS) install
distclean:
	rm -rf $(WWW_PREFIX)/htdocs $(WWW_PREFIX)/lib $(WWW_PREFIX)/cgi-bin

install: all
	install								\
		-g $(WWW_GROUP)						\
		-o $(WWW_USER)						\
		-m u=rwx,g=x,o=						\
		-d $(WWW_PREFIX)

	install								\
		-g $(WWW_GROUP)						\
		-o $(WWW_USER)						\
		-m u=rwx,g=x,o=						\
		-d $(WWW_PREFIX)/htdocs
	install								\
		-g $(WWW_GROUP)						\
		-o $(WWW_USER)						\
		-m u=rw,g=r,o=						\
		-C							\
		$(HTDOCs)						\
		$(WWW_PREFIX)/htdocs

	install								\
		-g $(WWW_GROUP)						\
		-o $(WWW_USER)						\
		-m u=xrw,g=x,o=						\
		-d $(WWW_PREFIX)/cgi-bin

	test -e $(WWW_PREFIX)/cgi-bin/jmscott ||			\
		ln -s $(JMSCOTT_ROOT)/www/cgi-bin $(WWW_PREFIX)/cgi-bin/jmscott

	install -g $(DIST_GROUP) -o $(DIST_USER) -C -m ug=rx,o=         \
                $(CGIBINs)                                              \
                $(WWW_PREFIX)/cgi-bin
	install								\
		-g $(WWW_GROUP)						\
		-o $(WWW_USER)						\
		-m u=rwx,g=rx,o=					\
		-d $(WWW_PREFIX)/lib
	ln -s $(JMSCOTT_ROOT)/www/lib $(WWW_PREFIX)/lib/jmscott
	rsync								\
		--relative						\
		--no-perms						\
		--chmod=u=rw,Du+x,g=r,Dg+x,o=				\
		$(LIBs)							\
		$(WWW_PREFIX)/lib

dist: all
	make-dist $(DIST)
daily: daily.cgi
	cgi2perl5 --source-path daily.cgi
	chmod +x daily
