/*******************************************************************************
*
* Copyright 2015 Stefan Majewsky <majewsky@gmx.net>
*
* This file is part of Holo.
*
* Holo is free software: you can redistribute it and/or modify it under the
* terms of the GNU General Public License as published by the Free Software
* Foundation, either version 3 of the License, or (at your option) any later
* version.
*
* Holo is distributed in the hope that it will be useful, but WITHOUT ANY
* WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
* A PARTICULAR PURPOSE. See the GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License along with
* Holo. If not, see <http://www.gnu.org/licenses/>.
*
*******************************************************************************/

package color

import (
	"bytes"
	"io"
)

//LineColorizingRule is a rule for the LineColorizingWriter (see there).
type LineColorizingRule struct {
	Prefix []byte
	Color  []byte
}

//ColorizeLine adds color to the given line according to the first of the given
//`rules` that matches.
func ColorizeLine(line []byte, rules []LineColorizingRule) []byte {
	for _, rule := range rules {
		if bytes.HasPrefix(line, rule.Prefix) {
			return bytes.Join([][]byte{rule.Color, line, []byte("\x1b[0m")}, nil)
		}
	}
	return line
}

//ColorizeLines is like ColorizeLine, but acts on multiple lines.
func ColorizeLines(lines []byte, rules []LineColorizingRule) []byte {
	sep := []byte{'\n'}
	in := bytes.Split(lines, sep)
	out := make([][]byte, 0, len(in))
	for _, line := range in {
		out = append(out, ColorizeLine(line, rules))
	}
	return bytes.Join(out, sep)
}

//LineColorizingWriter is an io.Writer that adds ANSI colors to lines of text
//written into it. It then passes the colorized lines to another writer.
//Coloring is based on prefixes. For example, to turn all lines with a "!!"
//prefix red, use
//
//    colorizer = &LineColorizingWriter {
//        Writer: otherWriter,
//        Rules: []LineColorizingRule {
//            LineColorizingRule { []byte("!!"), []byte("\x1B[1;31m") },
//        },
//    }
//
type LineColorizingWriter struct {
	Writer io.Writer
	Rules  []LineColorizingRule
	buffer []byte
}

//Write implements the io.Writer interface.
func (w *LineColorizingWriter) Write(p []byte) (n int, err error) {
	//append `p` to buffer and report everything as written
	w.buffer = append(w.buffer, p...)
	n = len(p)

	for {
		//check if we have a full line in the buffer
		idx := bytes.IndexByte(w.buffer, '\n')
		if idx == -1 {
			return n, nil
		}

		//extract line from buffer
		line := append(ColorizeLine(w.buffer[0:idx], w.Rules), '\n')
		w.buffer = w.buffer[idx+1:]

		//check if a colorizing rule matches
		_, err := w.Writer.Write(line)
		if err != nil {
			return n, err
		}
	}
}
