GO:=/usr/local/go/bin/go
export GOPATH := /usr/local

GOSRC=rasqld.go								\
	config.go							\
	http.go								\
	preamble.go							\
	sql.go

all: rasqld

install: rasqld
	cp rasqld /usr/local/bin

clean:
	rm -f rasqld

rasqld: $(GOSRC)
	$(GO) build $(GOSRC)
fmt:
	$(GO) fmt $(GOSRC)
vet:
	#  vet wants formating functions to end in 'f'.  jmscott doesn't.
	#  -printf=0 eliminates "possible formatting directive in log call"
	#  gripes
	$(GO) tool vet -printf=0 $(GOSRC)
