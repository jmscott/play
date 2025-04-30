/*
 *  Synopsis:
 *	Deduplicate lines from stdin to stdout, in scan order
 *  Usage:
 *	dedup <inet-addr.txt | wc -l
 *  Exit Status:
 *	0	ok
 *	1	unexpected error
 *  Note:
 *	About 10 times quicker than "sort -u".  However, RAM is probably
 *	better managedby "sort -u".  The clang version is about %30 faster
 *	than the golang version.
 */

#include <sys/errno.h>
#include <unistd.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>

extern int errno;

//  the 560000th prime number ~ 63megs of memory
#define HASH_TABLE_SIZE		8322241

static void
die(char *msg)
{
	write(2, "ERROR: ", 6);
	write(2, msg, strlen(msg));
	write(2, "\n", 1);
	_exit(1);
}

/*
 *  Attributed to Dan Berstein from a usenet comp.lang.c newsgroup.
 */
static unsigned int
djb(unsigned char *buf, int nbytes)
{
	unsigned int hash = 5381;
	unsigned char *p;

	p = buf;
	while (--nbytes >= 0)
		 hash = ((hash << 5) + hash) + *p++;
	return hash;
}

struct hash_set_element
{
	unsigned char		*value;
	int			size;		

	struct hash_set_element	*next;
};

struct hash_set
{
	int			size;
	struct hash_set_element	**table;
};

void
blob_set_alloc(void **set)
{
	struct hash_set *s;

	s = malloc(sizeof *s);
	if (s == NULL)
		die("malloc(hash_set) failed");
	s->size = HASH_TABLE_SIZE;
	s->table = calloc(s->size * sizeof *s->table, sizeof *s->table);
	if (s->table == NULL)
		die("calloc(table) failed");
	*set = (void *)s;
}

/*
 *  Does an element exist in a set?
 */
int
blob_set_exists(void *set, unsigned char *element, int size)
{
	struct hash_set *s = (struct hash_set *)set;
	struct hash_set_element *e;
	unsigned int hash;

	hash = djb(element, size) % s->size;
	e = s->table[hash];
	if (e) do {
		if (size == e->size && memcmp(e->value, element, size) == 0)
			return 1;
	} while ((e = e->next));

	return 0;
}

/*
 *  Synopsis:
 *	Add an element to the hash table.
 *  Returns:
 *	0	added, did not exist
 *	1	already exists
 */
int
blob_set_put(void *set, unsigned char *element, int size)
{
	unsigned int hash;
	struct hash_set *s;
	struct hash_set_element *e;

	s = (struct hash_set *)set;
	if (blob_set_exists(set, element, size))
		return 1;

	/*
	 *  Allocate new hash entry.
	 */
	e = malloc(sizeof *e + size);
	if (e == NULL)
		die("malloc(entry+value) failed");
	e->value = (unsigned char *)e + sizeof *e;
	memcpy(e->value, element, size);
	e->size = size;
	hash = djb(element, size) % s->size;
	e->next = s->table[hash];
	s->table[hash] = e;

	return 0;
}

int
main()
{
	unsigned char buf[4096];

	struct hash_set *hs;
	blob_set_alloc((void **)&hs);

	char *p;
	errno = 0;
	while (fgets((char *)buf, sizeof buf, stdin)) {
		int len = strlen((char *)buf);
		if (len == 0)
			die("impossible 0 len line");

		if (blob_set_put(hs, (unsigned char *)buf, len) == 0)
			if (write(1, buf, len) < 0)
				die(strerror(errno));
	}
	if (errno > 0)
		die(strerror(errno));
	_exit(0);
}
