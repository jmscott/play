/*
 *  Synopsis:
 *	Count bits in blob XOR'ed with itself shifted one bit to the left.
 *  Usage:
 *	shift-xor-pop-count-v1 <bytes-count> <mask-bits>
 */

#include "./common.c"

char	*prog = "shift-xor-pop-count-v1";

int
main(int argc, char **argv) {

	(void)argv;
	if (argc != 3)
		die("wrong argument count: expected 2 args");

	_exit(EXIT_OK);
}
