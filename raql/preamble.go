//Parse a C style comment written jmscott preamble syntax

package main

import (
	"bufio"
	"bytes"
	"errors"
	"regexp"
	"strings"
)

func parse_Ccomment_preamble(
	in *bufio.Reader,
) (
	pre map[string]string,
	line_count int,
	err error,
) {

	var name string
	var value bytes.Buffer

	pre = make(map[string]string)

	//  Note: Section prefix needs to match Unicode Graphic

	section_re := regexp.MustCompile(`^ [*]  ([A-Z][^:]*):\s*(.*)$`)

	for {
		var line string

		line, err = in.ReadString('\n')
		if err != nil {

			//  EOF is an error until end of preamble seen

			return nil, line_count, err
		}
		line_count++

		if line == " */" {
			return
		}

		if !strings.HasPrefix(line, " *") {
			err = errors.New("line must start with \" *\"")
			return nil, line_count, err
		}

		//  new section

		matches := section_re.FindStringSubmatch(line)
		if len(matches) > 1 {

			//  update section value
			if name != "" {
				pre[name] = value.String()
				value.Truncate(0)
			}

			name = matches[1]
			if len(matches) == 3 {
				value.WriteString(matches[1])
			}
		} else {
			value.WriteString(line)
		}
	}
}
