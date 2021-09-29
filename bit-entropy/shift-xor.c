/*
 *  Synopsis:
 *	XOR stdin with itself shifted left one bit and write to stdout.
 *  Description:
 *	Write a new blob which is the XOR of the standard input shifted
 *	one bit to the left.  The output is two lines followed by the
 *	raw bytes
 *
 *		blob-size
 *		ignore-tail-bits
 *		bytes of shifted xor bytes
 *  Usage:
 *	shift-xor <in-blob-size> <ignore-tail-bits>
 */

#define COMMON_NEED_SIZE64
#define COMMON_NEED_READ
#define COMMON_NEED_WRITE

#include "./common.c"

char	*prog = "shift-xor";

int
main(int argc, char **argv) {

	uint64_t in_size, ignore_tail_bits;

	if (argc != 3)
		die("wrong argument count: expected 2 args");

	in_size = parse_size64("in-blob-size", argv[1]);
	if (in_size == 0) {
		write(1, "0\n0\n", 4);
		_exit(0);
	}

	ignore_tail_bits = parse_size64("ignore-tail-bits", argv[2]);
	if (ignore_tail_bits >= 8)
		die("ignore bits can not >= 8");

	unsigned char *in = malloc(in_size * 2);
	if (in == NULL)
		die2("malloc(stdin+stdout) failed", argv[1]);

	//  Slurp the entire blob into ram memory.
	size_t nr;
	unsigned char *p = in;
	while ((nr = _read(p, PIPE_MAX)) > 0) {
		p += nr;
		if ((uint64_t)(p - in) > in_size)
			die("blob too big");
	}
	if ((uint64_t)(p - in) != in_size)
		die("blob too small");

	p = in;
	unsigned char *p_limit = in + in_size - 1;
	unsigned char *x = in + in_size;

	/*
	 *  xor the in blob with the shifted version of itself.
	 */
	unsigned char c;
	while (p < p_limit) {
		c = *p++;
		*x++ = (c ^ (c << 1)) | ((c & 0x1) ^ ((*p >> 7) &0x1));
	}
	c = *p;
	// *x = (c ^ (c << 1)) | ((c & 0x1) ^ ((*b >> 7) & 0x1));
	_write(in + in_size, in_size);

	_exit(EXIT_OK);
}
