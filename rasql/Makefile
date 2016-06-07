GO:=/usr/local/go/bin/go
GOSRC=rasqld.go preamble.go sql.go

all: rasqld

clean:
	rm -f rasqld

rasqld: rasqld.go preamble.go
	$(GO) build $(GOSRC)
fmt:
	$(GO) fmt $(GOSRC)
vet:
	#  vet wants formating functions to end in 'f'.  jmscott doesn't.
	#  -printf=0 eliminates "possible formatting directive in log call"
	#  gripes
	$(GO) tool vet -printf=0 $(GOSRC)
