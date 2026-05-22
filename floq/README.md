# What is `floq`

`floq` stands for "Flow Query".

The program `floq` is a simple interpreter for coordinating the execution of
unix processes by qualifying on the standard output, error and exit status
of those processes.  The output of the processes - UTF8 only - can be
`tuplified` and qualfied in a pidgin query language.

# Hello, World

``` 
	#  tab separated with two fields
	cat >hellow.flow <<END
	hello,	world
	good bye,	cruel world
	END

	cat >hello.floq <<END
	#!/usr/bin/env floq server

	define tuple example_tup as {
		attributes: {
			salutation: {
				matches: "^..{0,31}$",
				tsv_field: 1
			},
			who: {
				matches: "^..{0,31}$",
				tsv_field: 2
			}
		}
	};
	define command tail.example_tup as {
		path: "/usr/bin/tail"
		args: ["-F"]
	};
	define hello command as {
		path: "/bin/echo"
	};

	flow tail("hello.floq");
```
