GO=/usr/local/go/bin/go
GOSRC=hoq.go parser.go ast.go command.go logic.go opcode.go

export GOPATH := /usr/local/go/bin

all: hoq

clean:
	rm -f parser.go hoq

hoq: $(GOSRC)
	$(GO) build $(GOSRC)

parser.go: parser.y
	$(GO) tool yacc -o parser.go parser.y || (rm -f parser.go; exit 1)
	$(GO) fmt parser.go
