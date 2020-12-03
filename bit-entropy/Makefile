
DIST=./play-bit-entropy.dist
BINs := $(shell (. $(DIST) && echo $$BINs))
SRCs := $(shell (. $(DIST) && echo $$SRCs))
LIBs := $(shell (. $(DIST) && echo $$LIBs))
COMPILED := $(shell (. $(DIST) && echo $$COMPILED))

all: $(COMPILED)

clean:
	rm -f $(COMPILED)

shift-xor-pop-count-v1: shift-xor-pop-count-v1.c common.c
	cc -Wall -Wextra -o shift-xor-pop-count-v1 shift-xor-pop-count-v1.c
