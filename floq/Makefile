#
#  Synopsis:
#  	Makefile for building "floq" command
#  Usage:
#	make world
#  Yacc:
#	Compiled with yacc at
#
#		export GO111MODULE=off
#		/usr/local/go/bin/go get -v -u golang.org/x/tools/cmd/goyacc
#		/usr/local/go/bin/go get -v -u github.com/lib/pq
#
#	which installs goyacc in $GOPATH/bin.
#
#	ubunutu22 packages goyacc in
#
#		apt install golang-golang-x-tools
#  Note:
#	Need to eliminate redundant create of dir such as /bin, /sbin.
#
#	Consider always compiling in pprof.
#	Apparently the execution overhead of pprof is not much.
#
include ../local.mk
include ../play.mk

_MAKE=$(MAKE) $(MFLAGS)

MKMK=play-floq.mkmk

SRCs := $(shell  (. ./$(MKMK) && echo $$SRCs))
BINs := $(shell  (. ./$(MKMK) && echo $$BINs))
SBINs := $(shell  (. ./$(MKMK) && echo $$SBINs))
LIBs := $(shell  (. ./$(MKMK) && echo $$LIBs))
COMPILEs := $(shell  (. ./$(MKMK) && echo $$COMPILEs))
GOSRCs := $(shell  (. ./$(MKMK) && echo $$GOSRCs))

export GO111MODULE := off

#
#  To enable race detection, uncomment GO_BUILD_RACE
#
#GO_BUILD_RACE=-race

#  Do 'go tool compile' for list of compile time flags
#
#  Disable opimixations like function inline
#	-B	disable bounds checking
#	-l	disable inlining
#	-N	disable optimizations
#  
#GCFLAGS=-gcflags "-N -l"
#GCFLAGS=-gcflags "-l"

all: floq floq-execv

clean:
	rm --force $(COMPILEs) y.output

parser.go: parser.y
	$(GOYACC) -o parser.go parser.y || (rm -f parser.go; exit 1)
	$(GOEXE) fmt parser.go

floq: $(GOSRCs)
	$(GOEXE) build $(GOSRCs)

floq-execv: floq-execv.c
	cc $(CFLAGS)							\
		-I$(JMSCOTT_ROOT)/include				\
		floq-execv.c						\
		-o floq-execv						\
		-L $(JMSCOTT_ROOT)/lib					\
		-ljmscott

tar:
	make-make tar $(MKMK)
frisk:
	make-make tar $(MKMK)
	@rm --force --verbose play-floq-20*.tar.bz2
	@echo;  echo;  echo OK frisk passed
world:
	$(MAKE) $(MFLAGS) clean
	$(MAKE) $(MFLAGS) all
