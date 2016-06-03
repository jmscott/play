GO:=/usr/local/go/bin/go

all: raqd

clean:
	rm -f raqd

raqd: raqd.go preamble.go
	$(GO) build raqd.go preamble.go
