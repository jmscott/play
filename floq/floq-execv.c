/*
 *  Synopsis:
 *	Execute command vectors read from stdin and write summaries on stdout
 *  Description:
 *	For command line testing, do
 *
 *		echo /usr/bin/true | floq-execv
 *		echo /usr/bin/false | floq-execv
 *		echo /usr/bin/date | floq-execv
 *
 *	A single new-line tells floq to exit cleanly, eliminating the confusing
 *	"ERROR: unexpected read(request) of 0" in the examples above.
 *
 *		#  exit cleanly with no ERROR
 *
 *		(echo /usr/bin/date;  echo) | floq-execv
 *	
 *	A summary of the exit status of each execed process is written to
 *	standard out, tab separated:
 *
 *		<xclass>\t<xstatus>\t<pid>\t<usecs>\t<ssecs>\t<<merge-out-len>\n
 *		...<merge-out-len> bytes
 *
 *	where <merge-out-len> is number of bytes of stdout and stderr merged
 *	together.  The exit class of the process is EXIT, SIG, STOP, FAULT
 *	
 *		# normal process exit: 0 <= 255
 *		EXIT\t<exit-code>\t<user-seconds>\t<system-seconds>\tout-count\n
 *		merged std{out,err} bytes ...
 *
 *		# process was interupted by a signal
 *		SIG\t<signal>\t<user-seconds>\t<system-seconds>
 *
 *		# process got the KILLSTOP signal and was killed
 *		STOP\t<stop-signal>\t<user-seconds>\t<system-seconds>
 *
 * 		A fatal error occured when execv'ing the process
 *		FAULT\t<error description>
 *
 *	followed by exit status, process id, user sec, sys secs, merged output
 *	length, finally followed by exactly <merged-length> bytes.
 *
 *	Fatal errors for floq-execv itself are written to standard err and then
 *	floq-execv exits with value 1.
 *
 *		"unexpected read(request) of 0" means
 *
 *	This program exists because older golang have following issues.
 *
 *		1.  no support RUSAGE_CHILDREN
 *		2.  race detector does not play well with fast execs.
 *		3.  a global lock once existed, inhibiting fasy execs.
 *
 *  Exit Status:
 *  	0	exit ok
 *  	1	exit error (written to standard error)
 *  Note:
 *	Segregate stdout from stderr by having a count for each instead of a
 *	merged count.
 *	
 *	Should wall_duration be added to the process exit status?
 *	For daemon mode this makes sense.
 *
 *	No reason to limit output to 4096 bytes, since no need for atomic
 *	write.  The output limit should be either passed as command line
 *	option to floq-execv or in the request record.
 *
 *	What reaps child processes when parent panics?
 *
 *	Should the process be killed upon receiving a STOP signal?
 */
#include <sys/times.h>
#include <sys/types.h>
#include <sys/resource.h>
#include <sys/wait.h>

#include <unistd.h>
#include <string.h>
#include <errno.h>
#include <ctype.h>
#include <stdio.h>

#include <stdlib.h>		//  zap me when done debugging

#include "jmscott/libjmscott.h"

char *jmscott_progname = "floq-execv";
static char *usage = "floq-execv";

//#define MAX_MSG	JMSCOTT_ATOMIC_WRITE_SIZE
#define MAX_MSG	4096

/*  maximum number arg vector sent to floq-execv from floq */

#define MAX_X_ARG	256	//  max byte length of a single string argv[]
#define MAX_X_ARGC	64	//  max elements in argv[]

static int	x_argc;

//  the argv for the execv() (not main())
static char	*x_argv[MAX_X_ARGC + 1];

static char	args[MAX_X_ARGC * (MAX_X_ARG + 1)];

static void
die(char *msg)
{
	write(2, "\nERROR: ", 8);
	write(2, msg, strlen(msg));
	write(2, "\n", 1);
	exit(1);
}

static void
die2(char *msg1, char *msg2)
{
	jmscott_die2(1, msg1, msg2);
}

static void
die3(char *msg1, char *msg2, char *msg3)
{
	jmscott_die3(1, msg1, msg2, msg3);
}

/*  read a request from floq to execute a command */
static void
_read_request(char *buf)
{
	ssize_t nr, nread = 0;
	char *p = buf;
	
	*p = 0;

AGAIN:
	//  room for null (not written), since sizeof buf == MAX_MSG + 1
	nr = jmscott_read(0, p, MAX_MSG - nread);
	if (nr < 0)
		die2("read(request) failed", strerror(errno));
	if (nr == 0)
		die("unexpected read(request<stdin) of 0 bytes");
	nread += nr;
	p += nr;
	if (buf[nread - 1] == '\n') {
		buf[nread] = 0;
		return;
	}
	goto AGAIN;
}

/*  blocking read of output from a child process */

static int
_read_child(int fd, char *buf)
{
	ssize_t nb, nread = 0;
	char *p = buf;
	*buf = 0;

AGAIN:
	nb = jmscott_read(fd, p, MAX_MSG - nread);
	if (nb == 0) {
		buf[nread] = 0;		// sizeof buf is MAX_MSG + 1
		return nread;
	}
	if (nb > 0) {
		nread += nb;
		p += nb;
		goto AGAIN;
	}
	die2("read(child) failed", strerror(errno));

	/*NOTREACHED*/
	return -1;
}

static void
_write(void *p, ssize_t nbytes)
{
	int nb = 0;

AGAIN:
	nb = jmscott_write(1, p + nb, nbytes);
	if (nb < 0)
		die2("write(1) failed", strerror(errno));
	if (nb == 0)
		die("write(1) wrote 0 bytes");
	nbytes -= nb;
	if (nbytes == 0)
		return;
	goto AGAIN;
}

static void
_wait4(pid_t pid, int *statp, struct rusage *ru)
{
AGAIN:
	if (wait4(pid, statp, 0, ru) == pid)
		return;
	if (errno == EINTR)
		goto AGAIN;
	die2("wait4() failed", strerror(errno));
}

static void
_close(int fd)
{
	if (jmscott_close(fd) < 0)
		die2("close() failed", strerror(errno));
}

static void
_dup2(int old, int new)
{
AGAIN:
	if (dup2(old, new) < 0) {
		if (errno == EINTR)
			goto AGAIN;
		die2("dup2() failed", strerror(errno));
	}
}

static void
fork_wait() {

	pid_t pid;
	int status;
	char reply[MAX_MSG + 1], child_out[MAX_MSG + 1];
	struct rusage ru;
	char *xclass = 0;
	int xstatus = 0;
	ssize_t olen = 0;
	int merge[2];

	if (pipe(merge) < 0)
		die2("pipe(merge) failed", strerror(errno));
	pid = fork();
	if (pid < 0)
		die2("fork(request) failed", strerror(errno));

	//  in the child process

	if (pid == 0) {
		_close(0);
		_close(merge[0]);

		//  dup stdout and stderr onto merge[1]
		_dup2(merge[1], 1);
		_dup2(merge[1], 2);
		_close(merge[1]);
		execv(x_argv[0], x_argv);
		die3("execv(request) failed", strerror(errno), x_argv[0]);
	}

	//  in parent, so wait for output from the child,
	//  and reply with an execution description record,
	//  followed by the child's output.

	_close(merge[1]);
	olen = _read_child(merge[0], child_out);
	_close(merge[0]);

	//  reap the dead

	_wait4(pid, &status, &ru);

	//  determine process exit class, per xdr records

	if (WIFEXITED(status)) {
		xclass = "EXIT";
		xstatus = WEXITSTATUS(status);
	} else if (WIFSIGNALED(status)) {
		xclass = "SIG";
		xstatus = WTERMSIG(status);
	} else if (WIFSTOPPED(status)) {
		xclass = "STOP";
		xstatus = WSTOPSIG(status);
	} else {
		char buf[22];

		snprintf(buf, sizeof buf, "0x%x", status);
		die2("wait(request) impossible status", buf);
	}

	//  write the execution description record (xdr) back to floq

	snprintf(
		reply,
		sizeof reply, "%s\t%d\t%ld\t%ld.%06ld\t%ld.%06ld\t%ld\n",
		xclass,
		xstatus,
		(long)pid,
		ru.ru_utime.tv_sec,
		(long)ru.ru_utime.tv_usec,
		ru.ru_stime.tv_sec,
		(long)ru.ru_stime.tv_usec,
		olen
	);
	_write(reply, strlen(reply));
	if (olen > 0)
		_write(child_out, olen);
}

int
main(int argc, char **argv)
{
	char *arg, c = 0;
	char buf[MAX_MSG + 1];
	int i;

	if (argc != 1)
		jmscott_die_argc(1, argc, 1, usage);
	(void)argv;

	//  initialize static memory for x_args[] vector read from floq

	for (i = 0;  i < MAX_X_ARGC;  i++)
		x_argv[i] = &args[i * (MAX_X_ARG + 1)];

	x_argc = 0;
	arg = x_argv[0];

READ_REQUEST:
	_read_request(buf);
	char *p;

	p = buf;

	while ((c = *p++)) {
		switch (c) {

		//  finished parsing an element of string vector
		case '\t':
			*arg = 0;
			x_argc++;
			if (x_argc > MAX_X_ARGC)
				die("argc too big");
			arg = x_argv[x_argc];
			continue;

		//  parsed final element of string vector, so exec()
		case '\n':
			*arg = 0;
			x_argc++;
			break;

		//  partial parse of an element in the string vector
		default:
			if (!isascii(c))
				die("non-ascii input");
			if (arg - x_argv[x_argc] > MAX_X_ARG)
				die("arg too big");
			*arg++ = c;
			continue;
		}
		arg = x_argv[x_argc];
		x_argv[x_argc] = 0;	//  null-terminate vector

		/*
		 *  A request of zero length string is a request from floq
		 *  for floq-execv process to exit cleanly.
		 */  

		if (x_argc == 1 && strcmp(x_argv[0], "") == 0)
			exit(0);

		fork_wait();

		x_argv[x_argc] = arg;	//  reset null to arg buffer
		arg = x_argv[0];
		x_argc = 0;
	}
	goto READ_REQUEST;
}
