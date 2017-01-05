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

package impl

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"github.com/holocm/holo/cmd/holo/internal/output"
	"github.com/holocm/holo/lib/holo"
)

type EntityHandle struct {
	PluginHandle *PluginHandle
	Entity       holo.Entity
}

// MatchesSelector checks whether the given string is either the
// entity ID or a source file of this entity.
func (ehandle *EntityHandle) MatchesSelector(value string) bool {
	if ehandle.Entity.EntityID() == value {
		return true
	}
	for _, file := range ehandle.Entity.EntitySource() {
		if file == value {
			return true
		}
	}
	return false
}

func (ehandle *EntityHandle) Apply(withForce bool) {
	// track whether the report was already printed
	tracker := &output.PrologueTracker{Printer: func() { ehandle.PrintReport(true) }}
	stdout := &output.PrologueWriter{Tracker: tracker, Writer: output.Stdout}
	stderr := &output.PrologueWriter{Tracker: tracker, Writer: output.Stderr}

	result := ehandle.PluginHandle.Plugin.HoloApply(ehandle.Entity.EntityID(), withForce, stdout, stderr)

	var showReport bool
	var showDiff bool
	switch result {
	case holo.ApplyApplied:
		showReport = true
		showDiff = false
	case holo.ApplyAlreadyApplied:
		showReport = false
		showDiff = false
	case holo.ApplyExternallyChanged:
		output.Errorf(stderr, "Entity has been modified by user (use --force to overwrite)")
		showReport = false
		showDiff = true
	case holo.ApplyExternallyDeleted:
		output.Errorf(stderr, "Entity has been deleted by user (use --force to restore)")
		showReport = true
		showDiff = false
	default: // holo.ApplyError
		showReport = false
		showDiff = false
	}

	if showReport {
		tracker.Exec()
	}
	if showDiff {
		diff, err := ehandle.RenderDiff()
		if err != nil {
			output.Errorf(stderr, err.Error())
			return
		}
		// indent diff
		indent := []byte("    ")
		diff = regexp.MustCompile("(?m:^)").ReplaceAll(diff, indent)
		diff = bytes.TrimSuffix(diff, indent)

		tracker.Exec()
		output.Stdout.EndParagraph()
		output.Stdout.Write(diff)
	}
}

// PrintReport prints the scan report describing this Entity.
//
// The output should look like
//
//               lined up
//               V
//     ACTION_VERB ENTITY (ACTION_REASON)
//            KEY1 VAL1
//            KEY2 VAL2
//
// or
//
//                12
//                V
//     123456789012345678901234567890
//                V
//     ENTITY (ACTION_REASON)
//             KEY1 VAL1
//             KEY2 VAL2
//
// "ENTITY" is colored with ASNI escape codes.  " (ACTION_REASON)" is
// omitted if the action doesn't have a reason specified.
func (ehandle *EntityHandle) PrintReport(withAction bool) {
	// Initial header line
	align := 12
	verb, reason := ehandle.Entity.EntityAction()
	if withAction && verb != "" {
		align = len(verb)
		fmt.Fprintf(output.Stdout, "%s ", verb)
	}
	fmt.Fprintf(output.Stdout, "\x1b[1m%s\x1b[0m", ehandle.Entity.EntityID())
	if reason != "" {
		fmt.Fprintf(output.Stdout, " (%s)", reason)
	}
	output.Stdout.Write([]byte{'\n'})

	// Remaining info lines
	for _, line := range ehandle.Entity.EntityUserInfo() {
		fmt.Fprintf(output.Stdout, "%*s %s\n", align, line.Key, line.Val)
	}
	output.Stdout.EndParagraph()
	os.Stdout.Sync()
}

// PrintScanReport reproduces the original scan report for an Entity.
func (ehandle *EntityHandle) PrintScanReport() {
	fmt.Fprintf(output.Stdout, "ENTITY: %s\n", ehandle.Entity.EntityID())
	if verb, reason := ehandle.Entity.EntityAction(); reason == "" {
		fmt.Fprintf(output.Stdout, "ACTION: %s\n", verb)
	} else {
		fmt.Fprintf(output.Stdout, "ACTION: %s (%s)\n", verb, reason)
	}

	for _, sourceFile := range ehandle.Entity.EntitySource() {
		fmt.Fprintf(output.Stdout, "SOURCE: %s\n", sourceFile)
	}
	for _, infoLine := range ehandle.Entity.EntityUserInfo() {
		fmt.Fprintf(output.Stdout, "%s: %s\n", infoLine.Key, infoLine.Val)
	}

	output.Stdout.EndParagraph()
}

// RenderDiff creates a unified diff of a target file and its last
// provisioned version, similar to `diff
// /var/lib/holo/files/provisioned/$FILE $FILE`, but it also handles
// symlinks and missing files gracefully. The output is always a patch
// that can be applied to last provisioned version into the current
// version.
func (ehandle *EntityHandle) RenderDiff() ([]byte, error) {
	new, cur := ehandle.PluginHandle.Plugin.HoloDiff(ehandle.Entity.EntityID(), output.Stderr)
	if new == "" && cur == "" {
		return nil, nil
	}
	return renderFileDiff(new, cur)
}
