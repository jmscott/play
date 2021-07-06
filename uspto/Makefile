GODIST=/usr/local/go

all: scan-xml-preamble

clean:
	rm -f scan-xml-preamble

scan-xml-preamble: scan-xml-preamble.go
	$(GODIST)/bin/go build scan-xml-preamble.go
