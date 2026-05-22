# What is `floq`?

`floq` abbreviates "Flow Query".

The program `floq` is a simple interpreter that coordinates the execution of
unix processes by qualifying on the output and exit status
of those processes.  The output of the processes - text only - can be
[tuplified](https://en.wikipedia.org/wiki/Tuple), so to speak, and then
qualfied in a pidgin query language dictates the order of execution of the
procees

# Hello, World

```floq

#  create a source text file where each line has two, tab separated fields.
$ cat >phrases.txt <<END
hello,	world
good bye,	cruel world
END

#  create floq script to echo phrases to terminal (not stdout)
$ cat >echo.floq <<END

#  define a tuple `phrase` with two attributes: `salutation` and `who`,
#  the tuple will be built from two, tab separated text fields.

define tuple phrase as {
	attributes: {
		salutation: {
			#  regexp to match value of `salutation`
			matches: "^[a-z][a-z0-9, ]{0,31}$",

			#  first field qualfied as `<command>.salutation`
			tsv_field: 1
		},
		who: {
			#  regexp to match value of `who`
			matches: "^[a-z][a-z0-9, ]{0,31}$",

			#  second field qualified as `<command>.who`
			tsv_field: 2
		}
	}
};

#  write to /dev/tty instead of stdout

define command say as {
	path:	"/bin/sh",
	args:	["-c", "echo $0 $1 >/dev/tty"]
};

#  tail generates tuples by following output of file
define command tail.phrase as {
	path: "/usr/bin/tail",
	args: ["-F", "phrases.txt"]
};

#  start the flow by tailing file `phrases.txt` 
flow tail();

#  for each tuple
run say(tail.salutation, tail.who)
  when
  	tail.who != ""
	and
	tail.salutation != ""
;
END

#  now run the script
$ floq server echo.floq
```
