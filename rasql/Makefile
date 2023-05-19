#
#  Synopsis:
#	Simple development Makefile file
#  Usage:
#	Convert to make-make
#	rasqld <schema>.rasql >rasqld-<schema>.log 2>&1 &
#
GO:=/usr/local/go/bin/go
export GOPATH := /usr/local

RASQLD_SRC=rasqld.go							\
	config.go							\
	http.go								\
	log.go								\
	preamble.go							\
	sql.go

all: rasqld

install: rasqld
	cp rasqld /usr/local/bin

clean:
	rm -f rasqld

rasqld: $(RASQLD_SRC)
	$(GO) build $(RASQLD_SRC)
fmt:
	$(GO) fmt $(RASQLD_SRC)
vet:
	#  vet wants formating functions to end in 'f'.  jmscott doesn't want.
	#  -printf=0 eliminates "possible formatting directive in log call"
	#  gripes
	$(GO) tool vet -printf=0 $(RASQLD_SRC)
pem:
	#
	#  Generate test pem files for public certificate and private keys.
	# 
	#  See json fields "tls-cert-path" and "tls-key-path" in file
	#  pg_catalog.rasql.example.
	#
	openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 \
			-keyout self-signed.key -out self-signed.cert

dev-links:
	test -e log || ln -s . log
