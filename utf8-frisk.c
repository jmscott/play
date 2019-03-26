/*
 *  Synopsis:
 *	Write lines of valid utf8 onto standard output, ignoring invalid lines.
 *  Usage:
 *  	utf8-frisk <BLOB.csv >BLOB.utf8
 *  Exit Status:
 *	0	if all bytes match utf8 rfc3629
 *	1	some lines are not well formed utf8.
 *	2	standard input is empty
 *	3	bad i/o error (getline, write)
 *	4	premature end of stream on stdin
 *	9	invocation error
 *  Blame:
 *  	jmscott@setspace.com
 *  Note:
 *	Perhaps an "--number" option to list the line numbers of broken lines.
 *
 *	The well known iconv program should have a --skip-line option to
 *	eliminate  fine malformed lines.  This program only exists due to this
 *	limitation.  Also, python and perl both probably have methods/regexps
 *	to strip lines, but the invocations i found were way too complex
 *	and package dependent.
 *
 *	Some interesting "pure", table driven state machine algorithms exist.
 *	For example, these links describe utf8 state recognizers that expect
 *	entire strings as input before processing:
 *
 *	      http://lists.w3.org/Archives/Public/www-archive/2009Apr/0001.html
 *	      http://bjoern.hoehrmann.de/utf-8/decoder/dfa/
 *
 *	Since this program is byte/stream oriented, the above algorithms would
 *	need to be modified if used here.
 *
 *	A vectorized assembly version of validation has been written by Daniel
 *	Lemire:
 *
 *	      https://github.com/lemire/fastvalidate-utf-8
 *	      https://lemire.me/blog/2018/05/09/how-quickly-can-you-check-that-a-string-is-valid-unicode-utf-8/
 *
 *	The vectorized version appears to be about 8 time quicker than the
 *	fastest C state machine.
 *
 *	And, while you are exploring, be sure to read perl's mighty
 *	unicode/set spanner packages that built the decoders described above:
 *
 *		http://search.cpan.org/dist/Unicode-SetAutomaton/
 *		http://search.cpan.org/dist/Set-IntSpan-Partition/
 */

#include <sys/errno.h>

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>

static char progname[] = "utf8-frisk";

#define EXIT_OK			0
#define EXIT_BAD_UTF8		1
#define EXIT_EMPTY		2
#define EXIT_MAX_LINE		3
#define EXIT_BAD_IO		4
#define EXIT_BAD_INVO		9

/*
 *  States of utf8 scanner.
 */
#define STATE_START	0	/* goto new code sequence */

#define STATE_2BYTE2	1	/* goto second byte of 2 byte sequence */

#define STATE_3BYTE2	2	/* goto second byte of 3 byte sequence */
#define STATE_3BYTE3	3	/* goto third byte of 3 byte sequence */

#define STATE_4BYTE2	4	/* goto second byte of 4 byte sequence */
#define STATE_4BYTE3	5	/* goto third byte of 4 byte sequence */
#define STATE_4BYTE4	6	/* goto fourth byte of 4 byte sequence */

/*
 *  Bit masks up to 4 bytes per character
 */
#define B00000000	0x0
#define B10000000	0x80
#define B11000000	0xC0
#define B11100000	0xE0
#define B11110000	0xF0
#define B11111000	0xF8

/*
 * Synopsis:
 *  	Fast, safe and simple string concatenator
 *  Usage:
 *  	buf[0] = 0
 *  	_strcat(buf, sizeof buf, "hello, world");
 *  	_strcat(buf, sizeof buf, ": ");
 *  	_strcat(buf, sizeof buf, "good bye, cruel world");
 */
static void
_strcat(char *tgt, int tgtsize, char *src)
{
	//  find null terminated end of target buffer
	while (*tgt++)
		--tgtsize;
	--tgt;

	//  copy non-null src bytes, leaving room for trailing null
	while (--tgtsize > 0 && *src)
		*tgt++ = *src++;

	// target always null terminated
	*tgt = 0;
}

/*
 *  Write error message to standard error and exit process with status code.
 */
static void
die(int status, char *msg1)
{
	char msg[4096];
	static char ERROR[] = "ERROR: ";
	static char	colon[] = ": ";
	static char nl[] = "\n";

	msg[0] = 0;
	_strcat(msg, sizeof msg, progname);
	_strcat(msg, sizeof msg, colon);
	_strcat(msg, sizeof msg, ERROR);
	_strcat(msg, sizeof msg, msg1);
	_strcat(msg, sizeof msg, nl);

	write(2, msg, strlen(msg));

	_exit(status);
}

static void
die2(int status, char *msg1, char *msg2)
{
	static char colon[] = ": ";
	char msg[4096];

	msg[0] = 0;
	_strcat(msg, sizeof msg, msg1);
	_strcat(msg, sizeof msg, colon);
	_strcat(msg, sizeof msg, msg2);

	die(status, msg);
}

/*
 *  Are the characters in the line well formed UTF8 sequences?
 */
static int
is_utf8wf(unsigned char *p)
{
	unsigned int code_point = 0;
	int state = STATE_START;
	unsigned char c;

again:
	c = *p++;
	if (c == '\n' || c == 0)
		return state == STATE_START ? 1 : 0;

	switch (state) {
	case STATE_START:
		/*
		 *  Single byte/7 bit ascii?
		 *  Remain in START_START.
		 */
		if ((c & B10000000) == B00000000)
			goto again;

		/*
		 *  Mutibyte code point.
		 */
		code_point = 0;
		if ((c & B11100000) == B11000000) {
			/*
			 *  Start of 2 byte/11 bit sequence, so shift
			 *  the lower 5 bits of the first byte left
			 *  6 bits.
			 */
			code_point = (c & ~B11100000) << 6;
			state = STATE_2BYTE2;
		} else if ((c & B11110000) == B11100000) {
			/*
			 *  Start of 3 byte/16 bit sequence, so shift
			 *  the lower 4 bits of the first byte left 12
			 *  bits.
			 */
			code_point = (c & ~B11110000) << 12;
			state = STATE_3BYTE2;
		} else if ((c & B11111000) == B11110000) {
			/*
			 *  Start of 4 byte/21 bit sequence, so shift
			 *  the lower 3 bits of the first byte left 18
			 *  bits.
			 */
			code_point = (c & ~B11111000) << 18;
			state = STATE_4BYTE2;
		} else
			return 0;
		goto again;
	/*
	 *  Expect the second and final byte of two byte/11 bit
	 *  code point.
	 */
	case STATE_2BYTE2:
		/*
		 *  No continuation byte implies malformed sequence.
		 */
		if ((c & B11000000) != B10000000)
			return 0;
		/*
		 *  Or in the lower 6 bits of the second & final byte.
		 */
		code_point |= (c & ~B11000000);

		/*
		 *  Is an overlong representation.  Any value less than
		 *  128 must be represented with a single byte/7 bits.
		 */
		if (code_point < 128)
			return 0;

		state = STATE_START;
		goto again;
	/*
	 *  Expect the second byte of a three byte sequence.
	 */
	case STATE_3BYTE2:
		/*
		 *  No continuation byte implies malformed sequence.
		 */
		if ((c & B11000000) != B10000000)
			return 0;
		/*
		 *  Or in the lower 6 bits of the second byte into
		 *  bits 12 through 7 of the code point.
		 */
		code_point |= (c & ~B11000000) << 6;

		state = STATE_3BYTE3;
		goto again;

	/*
	 *  Third byte of three byte/16 bit sequence.
	 */
	case STATE_3BYTE3:
		/*
		 *  No continuation byte implies malformed sequence.
		 */
		if ((c & B11000000) != B10000000)
			return 0;
		/*
		 *  Or in the lower 6 bits of the third & final byte
		 *  into bits 6 through 1 of the code point.
		 */
		code_point |= c & ~B11000000;

		/*
		 *  Is an overlong representation?  Any value less than
		 *  2048 must be represented with either a
		 *  one byte/7 bit or two byte/11 bit sequence.
		 *
		 *  Second test is for UTF-16 surrogate pairs.
		 */
		if (code_point < 2048 ||
		    (0xD800 <= code_point&&code_point <= 0xDFFF))
			return 0;
		state = STATE_START;
		goto again;
	/*
	 *  Expect the second byte of four byte/21 bit sequence
	 */
	case STATE_4BYTE2:
		/*
		 *  No continuation byte implies malformed sequence.
		 */
		if ((c & B11000000) != B10000000)
			return 0;
		/*
		 *  Or in the lower 6 bits of the second byte into
		 *  bits 18 through 13 of the code point.
		 */
		code_point |= (c & ~B11000000) << 12;
		state = STATE_4BYTE3;
		goto again;
	/*
	 *  Expect the third byte of four byte/21 bit sequence
	 */
	case STATE_4BYTE3:
		/*
		 *  No continuation byte implies malformed sequence.
		 */
		if ((c & B11000000) != B10000000)
			return 0;
		/*
		 *  Or in the lower 6 bits of the third byte into
		 *  bits 12 through 7 of the code point.
		 */
		code_point |= (c & ~B11000000) << 6;
		state = STATE_4BYTE4;
		goto again;
	/*
	 *  Expect the fourth byte of four byte/21 bit sequence
	 */
	case STATE_4BYTE4:
		/*
		 *  No continuation byte implies malformed sequence.
		 */
		if ((c & B11000000) != B10000000)
			return 0;
		/*
		 *  Or in the lower 6 bits of the fourth byte into
		 *  bits 6 through 1 of the code point.
		 */
		code_point |= c & ~B11000000;
		/*
		 *  Is an overlong representation.  Any value less than
		 *  65536 must be represented with either a
		 *  one byte/7 bit, two byte/11 bit or three byte/16
		 *  sequence.
		 */
		if (code_point < 65536)
			return 0;
		state = STATE_START;
		goto again;
	}

	/* NOT REACHED */
	return 0;
}

/*
 *  write() exactly nbytes bytes, restarting on interrupt and dieing on error.
 */
static void
_write(void *p, ssize_t nbytes)
{
	int nb = 0;

	again:

	nb = write(1, p + nb, nbytes);
	if (nb < 0) {
		if (errno == EINTR)
			goto again;
		die2(EXIT_BAD_IO, "write(1) failed", strerror(errno));
	}
	nbytes -= nb;
	if (nbytes > 0)
		goto again;
}

int
main(int argc, char **argv)
{
	int seen_wf = 0, seen_bad = 0;

	if (argc != 1)
		die(EXIT_BAD_INVO, "wrong number of arguments");
	(void)argv;

	unsigned char *line = NULL;
	size_t cap = 0;
	ssize_t len;
	while ((len = getline((char **)&line, &cap, stdin)) > 0)
		if (is_utf8wf(line)) {
			_write(line, len);
			seen_wf = 1;
		} else
			seen_bad = 1;
fprintf(stderr, "WTF: len=%ld\n", len);
	if (len < 0) {
		if (errno > 0)
			die2(
				EXIT_BAD_IO,
				"getline(stdin) failed: %s",
				strerror(errno)
			);
	}
	if (seen_bad)
		exit(EXIT_BAD_UTF8);
	if (seen_wf)
		exit(EXIT_OK);
	exit(EXIT_EMPTY);
}
