GO:=/usr/local/go/bin/go

all: raqd

clean:
	rm -f raqd

raqd: raqd.go
	$(GO) build raqd.go
