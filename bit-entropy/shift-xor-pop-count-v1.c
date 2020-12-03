/*
 *  Synopsis:
 *	Count bits in blob XOR'ed with itself shifted one bit to the left, V1.
 *  Usage:
 *	shift-xor-pop-count-v1 <blob-size> <ignore-bits>
 */

#define COMMON_NEED_SIZE64
#define COMMON_NEED_READ

#include "./common.c"

char	*prog = "shift-xor-pop-count-v1";

int
main(int argc, char **argv) {

	uint64_t blob_size, ignore_bits;

	(void)argv;
	if (argc != 3)
		die("wrong argument count: expected 2 args");

	blob_size = parse_size64("blob-size", argv[1]);

	ignore_bits = parse_size64("ignore-bits", argv[2]);
	if (ignore_bits == 0)
		die("ignore bits can not == 0");
	if (ignore_bits >= blob_size)
		die("ignore bits can not >= blob size");

	void *buf = malloc(blob_size);
	if (buf == NULL)
		die2("malloc(blob) failed", argv[1]);

	//  Slurp the entire blob into ram memory.
	size_t nr;
	void *p = buf;
	while ((nr = _read(p, PIPE_MAX)) > 0) {
		p += nr;
		if ((uint64_t)(p - buf) > blob_size)
			die("blob too big");
	}
	if ((uint64_t)(p - buf) != blob_size)
		die("blob too small");

	_exit(EXIT_OK);
}
