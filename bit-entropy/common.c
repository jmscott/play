/*
 *  Synopsis:
 *	Common routies used clang programs in play/bit-entropy.
 *  Usage:
 *	#include "./common.h"
 */

#define EXIT_OK		0
#define EXIT_FAIL	2
#define PIPE_MAX	4096

/*
 * Synopsis:
 *      Fast, safe and simple string concatenator
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
	exit(EXIT_FAIL);
}

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
