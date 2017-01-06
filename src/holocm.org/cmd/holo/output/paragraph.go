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
	"io"
)

//ParagraphTracker is used in conjunction with ParagraphWriter. See explanation
//over there.
type ParagraphTracker struct {
	PrimaryWriter        io.Writer
	hadOutput            bool
	trailingNewlineCount int
}

//ParagraphWriter is an io.Writer that forwards to another io.Writer, but
//ensures that input is written in paragraphs, with newlines in between.
//
//Since, in this usecase, both stdout and stderr need to be PrologueWriter
//instances, the logic that prints the additional newlines must be shared by
//both. Thus the newlines are tracked with a ParagraphTracker instance.
type ParagraphWriter struct {
	Writer  io.Writer
	Tracker *ParagraphTracker
}

func (t *ParagraphTracker) observeOutput(p []byte) {
	//print the initial newline before any other output
	if !t.hadOutput {
		t.PrimaryWriter.Write([]byte{'\n'})
		t.hadOutput = true
	}

	//count trailing newlines on the output that was seen
	cnt := 0
	for cnt < len(p) && p[len(p)-1-cnt] == '\n' {
		cnt++
	}
	if cnt == len(p) {
		t.trailingNewlineCount += cnt
	} else {
		t.trailingNewlineCount = cnt
	}
}

//Write implements the io.Writer interface.
func (w *ParagraphWriter) Write(p []byte) (n int, e error) {
	w.Tracker.observeOutput(p)
	return w.Writer.Write(p)
}

//EndParagraph inserts newlines to start the next paragraph of output.
func (w *ParagraphWriter) EndParagraph() {
	if !w.Tracker.hadOutput {
		return
	}
	for w.Tracker.trailingNewlineCount < 2 {
		w.Write([]byte{'\n'})
	}
}
