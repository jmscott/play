GO:=/usr/local/go/bin/go
GOSRC=rasqld.go preamble.go

all: rasqld

clean:
	rm -f rasqld

rasqld: rasqld.go preamble.go
	$(GO) build $(GOSRC)
fmt:
	$(GO) fmt $(GOSRC)
