
DIST=./play-bit-entropy.dist
BINs := $(shell (. $(DIST) && echo $$BINs))
SRCs := $(shell (. $(DIST) && echo $$SRCs))
LIBs := $(shell (. $(DIST) && echo $$LIBs))
COMPILED := $(shell (. $(DIST) && echo $$COMPILED))

all: $(COMPILED)

clean:
	rm -f $(COMPILED)

shift-xor: shift-xor.c common.c
	cc -Wall -Wextra -o shift-xor shift-xor.c

bit-pop-count: bit-pop-count.c common.c
	cc -Wall -Wextra -o bit-pop-count bit-pop-count.c
