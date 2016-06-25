//Parse a C style comment written jmscott preamble syntax

package main

import (
	"bufio"
	"bytes"
	"errors"
	"regexp"
	"strings"
)

var section_re = regexp.MustCompile(`^ [*]  ([A-Z][^:]*):(.*)`)

type CcommentPreamble map[string]string

func (pre CcommentPreamble) parse(
	in *bufio.Reader,
) (
	line_count int,
	err error,
) {

	var name string
	var value bytes.Buffer

	for {
		var line string

		line, err = in.ReadString('\n')
		if err != nil {

			//  EOF is an error until end of preamble seen

			return
		}
		line_count++

		//  end of preamble

		if line == " */\n" {

			//  close final session
			if name != "" {
				pre[name] = value.String()
			}
			return
		}

		if !strings.HasPrefix(line, " *") {
			err = errors.New("line must start with \" *\"")
			return
		}

		//  new section

		matches := section_re.FindStringSubmatch(line)
		if len(matches) > 1 {

			//  close the previous section

			if name != "" {
				pre[name] = value.String()
				value.Truncate(0)
			}

			// new section

			name = matches[1]
			if pre[name] != "" {
				err = errors.New("section redefined: " + name)
				return
			}
			value.WriteString(matches[2])
		} else if name != "" {
			value.WriteString(line[2:])
		}
	}
}
