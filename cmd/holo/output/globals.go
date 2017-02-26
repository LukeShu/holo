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

package output

import (
	"fmt"
	"io"
	"os"
	"strings"
)

//Errorf formats and prints an error message on stderr.
func Errorf(writer io.Writer, text string, args ...interface{}) {
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}
	fmt.Fprintf(writer, "\x1b[1;31m!! %s\x1b[0m\n", strings.TrimSuffix(text, "\n"))
}

//Warnf formats and prints an warning message on stderr.
func Warnf(writer io.Writer, text string, args ...interface{}) {
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}
	fmt.Fprintf(writer, "\x1b[1;33m>> %s\x1b[0m\n", strings.TrimSuffix(text, "\n"))
}

var stdTracker = &ParagraphTracker{PrimaryWriter: os.Stdout}

//Stdout wraps os.Stdout into a ParagraphWriter.
var Stdout = &ParagraphWriter{Writer: os.Stdout, Tracker: stdTracker}

//Stderr wraps os.Stderr into a ParagraphWriter.
var Stderr = &ParagraphWriter{Writer: os.Stderr, Tracker: stdTracker}
