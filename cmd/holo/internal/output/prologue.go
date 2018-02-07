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

//PrologueTracker is used in conjunction with PrologueWriter. See explanation
//over there.
type PrologueTracker struct {
	Printer func()
}

//Exec prints the prologue if it has not been printed before.
func (t *PrologueTracker) Exec() {
	//print prologue exactly once
	if t.Printer != nil {
		t.Printer()
		t.Printer = nil
	}
}

//PrologueWriter is an io.Writer that ensures that a prologue is printed before
//any writes to the underlying io.Writer occur. This is used by entity.Apply()
//to print the scan report before any other output, but only if there is output.
//
//Since, in this usecase, both stdout and stderr need to be PrologueWriter
//instances, the function that prints the prologue must be shared by both, and
//it needs to be made sure that the prologue is only printed once. Thus the
//prologue is tracked with a PrologueTracker instance.
type PrologueWriter struct {
	Writer  io.Writer
	Tracker *PrologueTracker
}

//Write implements the io.Writer interface.
func (w *PrologueWriter) Write(p []byte) (n int, e error) {
	//skip empty writes
	if len(p) == 0 {
		return 0, nil
	}

	//ensure that prologue is printed
	w.Tracker.Exec()
	return w.Writer.Write(p)
}
