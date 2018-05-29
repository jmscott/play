#
#  Synopsis:
#	Simple makefile to build hoq
#
GO=/usr/local/go/bin/go
GOYACC=goyacc

GOSRC=hoq.go 								\
	ast.go								\
	command.go							\
	compile.go							\
	opcode.go							\
	parser.go							\
	rummy.go							\
	tsort.go							\

all: hoq

clean:
	rm -f parser.go hoq

hoq: $(GOSRC)
	$(GO) build $(GOSRC)

parser.go: parser.y
	$(GOYACC) -o parser.go parser.y || (rm -f parser.go; exit 1)
	$(GO) fmt parser.go
fmt:
	$(GO) fmt $(GOSRC)
