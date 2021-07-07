GODIST=/usr/local/go

all: scan-xml-preamble splice-bytes

clean:
	rm -f scan-xml-preamble split-bytes

scan-xml-preamble: scan-xml-preamble.go
	$(GODIST)/bin/go build scan-xml-preamble.go

split-bytes: splice-bytes.c
	cc -Wall -Wextra -o splice-bytes splice-bytes.c
