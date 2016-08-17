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
	#  vet wants formating functions to end in 'f'.  jmscott doesn't want.
	#  -printf=0 eliminates "possible formatting directive in log call"
	#  gripes
	$(GO) tool vet -printf=0 $(GOSRC)
pem:
	#
	#  Generate test pem files for public certificate and private keys.
	# 
	#  See json fields "tls-cert-path" and "tls-key-path" in file
	#  pg_catalog.rasql.example.
	#
	openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 \
			-keyout key.pem -out cert.pem
