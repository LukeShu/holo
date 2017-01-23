/*******************************************************************************
*
* Copyright 2017 Luke Shumaker <lukeshu@parabola.nu>
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

package holo

import (
	"fmt"
	"io"
	"os"
)

// An ApplyResult is an outcome from HoloApply.
type ApplyResult interface {
	isApplyResult()
	ExitCode() int
}

var (
	// ApplyApplied indicates that the entity has been
	// successfully modified to be in the desired state.
	ApplyApplied = applyApplied{}

	// ApplyAlreadyApplied indicates that the entity is already in
	// the desired state, so no changes have been made. Holo will
	// format its output accordingly (at the time of this writing,
	// by omitting the entity from the output).
	ApplyAlreadyApplied = ApplyMessage{"not changed\n"}

	// ApplyExternallyChanged indicates that the entity was
	// provisioned by this plugin, but has been changed by a user
	// or external application since then.  Holo will output an
	// error message indicating that "--force" is needed to
	// overwrite these manual changes.
	ApplyExternallyChanged = ApplyMessage{"requires --force to overwrite\n"}

	// ApplyExternallyDeleted indicates that the entity was
	// provisioned by this plugin, but has been deleted by the
	// user or external application since then.  Holo will output
	// an error message indicating that "--force" is needed to
	// overwrite these manual changes.
	ApplyExternallyDeleted = ApplyMessage{"requires --force to restore\n"}

	// ApplyError (N) indicates that this plugin attempty to
	// provision the entity, but failed.
	ApplyError = func(exitCode int) ApplyResult { return applyError{exitCode} }
)

////////////////////////////////////////////////////////////////////////////////

// ApplyMessage is the implementing type of ApplyAlreadyApplied,
// ApplyExternallyChanged, and ApplyExternally deleted.  You should
// generally not use it directly; but the type needs to be public for
// the plugin-side runner.
type ApplyMessage struct {
	msg string
}

func (a ApplyMessage) isApplyResult() {}

// ExitCode returns 0; as a the ApplyMessage values types do not
// indicate operational arrors.
func (a ApplyMessage) ExitCode() int {
	return 0
}

// Send sends the message to the controlling `holo` process over file
// descriptor 3.
func (a ApplyMessage) Send() {
	file := os.NewFile(3, "/dev/fd/3")
	_, err := io.WriteString(file, a.msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
	}
}

////////////////////////////////////////////////////////////////////////////////

type applyError struct {
	exitCode int
}

func (e applyError) isApplyResult() {}

func (e applyError) ExitCode() int {
	return e.exitCode
}

////////////////////////////////////////////////////////////////////////////////

type applyApplied struct{}

func (e applyApplied) isApplyResult() {}
func (e applyApplied) ExitCode() int  { return 0 }
