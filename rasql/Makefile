GO:=/usr/local/go/bin/go
GOSRC=raqd.go preamble.go

all: raqd

clean:
	rm -f raqd

raqd: raqd.go preamble.go
	$(GO) build $(GOSRC)
fmt:
	$(GO) fmt $(GOSRC)
