/*
 *  Synopsis:
 *	Count bits in blob XOR'ed with itself shifted one bit to the left, V1.
 *  Usage:
 *	shift-xor-pop-count-v1 <blob-size> <ignore-bits>
 *  Note:
 *	Incorrect algo for blob-size == 1
 */

#define COMMON_NEED_SIZE64
#define COMMON_NEED_READ
#define COMMON_NEED_WRITE

#include "./common.c"

char	*prog = "shift-xor";

int
main(int argc, char **argv) {

	uint64_t blob_size, ignore_bits;

	if (argc != 3)
		die("wrong argument count: expected 2 args");

	blob_size = parse_size64("blob-size", argv[1]);

	ignore_bits = parse_size64("ignore-bits", argv[2]);
	if (ignore_bits == 0)
		die("ignore bits can not == 0");
	if (ignore_bits >= blob_size)
		die("ignore bits can not >= blob size");

	void *blob = malloc(blob_size * 2);
	if (blob == NULL)
		die2("malloc(blob+xor) failed", argv[1]);
	void *xor = blob + blob_size;

	//  Slurp the entire blob into ram memory.
	size_t nr;
	void *p = blob;
	while ((nr = _read(p, PIPE_MAX)) > 0) {
		p += nr;
		if ((uint64_t)(p - blob) > blob_size)
			die("blob too big");
	}
	if ((uint64_t)(p - blob) != blob_size)
		die("blob too small");

	unsigned char *b = (unsigned char *)blob;
	unsigned char *b_limit = b + blob_size - 1;
	unsigned char *x = (unsigned char *)xor;

	unsigned char c;
	while (b < b_limit) {
		c = *b++;
		*x++ = (c ^ (c << 1)) | ((c & 0x1) ^ ((*b >> 7) & 0x1));
	}
	c = *b;
	*x = (c ^ (c << 1)) | ((c & 0x1) ^ ((*b >> 7) & 0x1));
	_write(xor, blob_size);

	_exit(EXIT_OK);
}
