/*
 *  Synopsis:
 *	Common routies used clang programs in play/bit-entropy.
 *  Usage:
 *	#include "./common.c"
 */

#include <sys/errno.h>
#include <unistd.h>
#include <string.h>
#include <stdlib.h>
#include <limits.h>

extern char *prog;			//  defined in main() file.

#define EXIT_OK		0
#define EXIT_FAIL	2
#ifndef PIPE_MAX
#define PIPE_MAX	4096
#endif

#ifdef COMMON_NEED_SIZE64
#define COMMON_NEED_PARSE_UINT64
#define COMMON_NEED_DIE2
#endif

#ifdef COMMON_NEED_PARSE_UINT64
#define COMMON_NEED_DIE2
#endif

#ifdef COMMON_NEED_DIE2
#define COMMON_NEED_DIE
#define COMMON_NEED_STRCAT
#endif


#ifdef COMMON_NEED_STRCAT
/*
 * Synopsis:
 *      Fast, safe and ergonomically simple string concatenator
 *  Usage:
 *      buf[0] = 0
 *      _strcat(buf, sizeof buf, "hello, world");
 *      _strcat(buf, sizeof buf, ": ");
 *      _strcat(buf, sizeof buf, "good bye, cruel world");
 */
static void
_strcat(char *restrict tgt, int tgtsize, const char *restrict src)
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
#endif

static void
die(char *msg)
{
	char buf[1024];

	strcpy(buf, prog);
	strcat(buf, ": ERROR: ");
	strncat(buf, msg, 1024 - (strlen(buf) + 2));
	strncat(buf, "\n", 1024 - (strlen(buf) + 2));

	buf[sizeof buf - 2] = '\n';
	buf[sizeof buf - 1] = 0;
	write(2, buf, strlen(buf)); 
	_exit(EXIT_FAIL);
}

#ifdef COMMON_NEED_DIE2

static void
die2(char *msg1, char *msg2)
{
        static char colon[] = ": ";
        char msg[PIPE_MAX];

        msg[0] = 0;
        _strcat(msg, sizeof msg, msg1);
        _strcat(msg, sizeof msg, colon);
        _strcat(msg, sizeof msg, msg2);

        die(msg);
}
#endif

#ifdef COMMON_NEED_READ
/*
 *  read() bytes from stdin, restarting on interrupt and dying on error.
 */
static ssize_t
_read(void *p, ssize_t nbytes)
{
        ssize_t nb;

again:
        nb = read(0, p, nbytes);
        if (nb >= 0)
                return nb;
        if (errno == EINTR)             //  try read()
                goto again;

        die2("read(stdin) failed", strerror(errno));

        /*NOTREACHED*/
        return -1;
}
#endif

#ifdef COMMON_NEED_WRITE
/*
 *  write() exactly nbytes bytes to stdout,
 *  restarting on interrupt and dying on error.
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
                die2("write(stdout) failed", strerror(errno));
        }
        nbytes -= nb;
        if (nbytes > 0)
                goto again;
}
#endif

#ifdef COMMON_NEED_PARSE_UINT64

/*
 *  Parse an unsigned 64 bit integer or croak.
 */
static uint64_t
parse_uint64(char *what, char *src)
{
	size_t len;

	len = strlen(src);
	if (len == 0)
		die2(what, "zero length");
	if (len > 20)
		die2(what, "length greater than 20 chars");
	
	//  only digits 0-9 allowed
	for (size_t i = 0;  i < len;  i++)
		if (src[i] < '0' || src[i] > '9')
			die2(what, "non-digit not valid");

	errno = 0;
	unsigned long long ui64 = strtoull(src, (char **)0, 10);
	if (ui64 == ULLONG_MAX && errno != 0)
		die2(what, strerror(errno));

	return (uint64_t)ui64;
}
#endif

#ifdef COMMON_NEED_SIZE64
/*
 *  Parse unsigned 64 bit <= max of a signed 64 bit int.
 */
static uint64_t
parse_size64(char *what, char *src)
{
	uint64_t size64 = parse_uint64(what, src);
	if (size64 > LLONG_MAX)
		die2(what, "size > max int64");
	return size64;
}

#endif
