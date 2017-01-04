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

package external

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"syscall"

	"holocm.org/cmd/holo/output"
	"holocm.org/lib/holo"
)

//Entity represents an entity known to some Holo plugin.
type Entity struct {
	plugin       *Plugin
	id           string
	actionVerb   string
	actionReason string
	sourceFiles  []string
	infoLines    []holo.KV
}

var _ holo.Entity = &Entity{}

//EntityID returns a string that uniquely identifies the entity.
func (e *Entity) EntityID() string { return e.id }

func (e *Entity) EntitySource() []string { return e.sourceFiles }

func (e *Entity) EntityUserInfo() []holo.KV { return e.infoLines }

func (e *Entity) EntityAction() string {
	if e.actionReason == "" {
		return e.actionVerb
	}
	return fmt.Sprintf("%s (%s)", e.actionVerb, e.actionReason)
}

//MatchesSelector checks whether the given string is either the entity ID or a
//source file of this entity.
func (e *Entity) MatchesSelector(value string) bool {
	if e.id == value {
		return true
	}
	for _, file := range e.sourceFiles {
		if file == value {
			return true
		}
	}
	return false
}

//PrintReport prints the scan report describing this Entity.
func (e *Entity) PrintReport(withAction bool) {
	//print initial line with action and entity ID
	//(note that output.Stdout != os.Stdout)
	var lineFormat string
	if e.actionVerb == "" || !withAction {
		lineFormat = "%12s %s\n"
		fmt.Fprintf(output.Stdout, "\x1b[1m%s\x1b[0m", e.EntityID())
	} else {
		lineFormat = fmt.Sprintf("%%%ds %%s\n", len(e.actionVerb))
		fmt.Fprintf(output.Stdout, "%s \x1b[1m%s\x1b[0m", e.actionVerb, e.id)
	}
	if e.actionReason == "" {
		output.Stdout.Write([]byte{'\n'})
	} else {
		fmt.Fprintf(output.Stdout, " (%s)\n", e.actionReason)
	}

	//print info lines
	for _, line := range e.EntityUserInfo() {
		fmt.Fprintf(output.Stdout, lineFormat, line.Key, line.Val)
	}
	output.Stdout.EndParagraph()
	os.Stdout.Sync()
}

//PrintScanReport reproduces the original scan report for this Entity.
func (e *Entity) PrintScanReport() {
	fmt.Fprintf(output.Stdout, "ENTITY: %s\n", e.EntityID())
	fmt.Fprintf(output.Stdout, "ACTION: %s\n", e.EntityAction())

	for _, sourceFile := range e.EntitySource() {
		fmt.Fprintf(output.Stdout, "SOURCE: %s\n", sourceFile)
	}
	for _, infoLine := range e.EntityUserInfo() {
		fmt.Fprintf(output.Stdout, "%s: %s\n", infoLine.Key, infoLine.Val)
	}

	output.Stdout.EndParagraph()
}

//Apply performs the complete application algorithm for the given Entity.
func (e *Entity) Apply(withForce bool) {
	// track whether the report was already printed
	tracker := &output.PrologueTracker{Printer: func() { e.PrintReport(true) }}
	stdout := &output.PrologueWriter{Tracker: tracker, Writer: output.Stdout}
	stderr := &output.PrologueWriter{Tracker: tracker, Writer: output.Stderr}

	result := e.plugin.HoloApply(e.id, withForce, stdout, stderr)

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
		showDiff = false
	case holo.ApplyExternallyDeleted:
		output.Errorf(stderr, "Entity has been deleted by user (use --force to restore)")
		showReport = true
		showDiff = false
	default: // assume holo.ApplyErr
		return
	}

	if showReport {
		tracker.Exec()
	}
	if showDiff {
		diff, err := e.RenderDiff()
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

//RenderDiff creates a unified diff of a target file and its last provisioned
//version, similar to `diff /var/lib/holo/files/provisioned/$FILE $FILE`, but it also
//handles symlinks and missing files gracefully. The output is always a patch
//that can be applied to last provisioned version into the current version.
func (e *Entity) RenderDiff() ([]byte, error) {
	new, cur := e.plugin.HoloDiff(e.EntityID())
	if new == "" && cur == "" {
		return nil, nil
	}
	return renderFileDiff(new, cur)
}

func renderFileDiff(fromPath, toPath string) ([]byte, error) {
	fromPathToUse, err := checkFile(fromPath)
	if err != nil {
		return nil, err
	}
	toPathToUse, err := checkFile(toPath)
	if err != nil {
		return nil, err
	}

	//run git-diff to obtain the diff
	var buffer bytes.Buffer
	cmd := exec.Command("git", "diff", "--no-index", "--", fromPathToUse, toPathToUse)
	cmd.Stdout = &buffer
	cmd.Stderr = output.Stderr

	//error "exit code 1" is normal for different files, only exit code > 2 means trouble
	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 1 {
					err = nil
				}
			}
		}
	}
	//did a relevant error occur?
	if err != nil {
		return nil, err
	}

	//remove "index <SHA1>..<SHA1> <mode>" lines
	result := buffer.Bytes()
	rx := regexp.MustCompile(`(?m:^index .*$)\n`)
	result = rx.ReplaceAll(result, nil)

	//fix paths in headers, especially remove the unnecessary "a/" and "b/"
	//path prefixes
	rx = regexp.MustCompile(`(?m:^diff --git .*$)`)
	result = rx.ReplaceAll(result, []byte(fmt.Sprintf("diff --holo %s %s", fromPath, toPath)))
	rx = regexp.MustCompile(`(?m:^--- a/.*$)`)
	result = rx.ReplaceAll(result, []byte("--- "+fromPath))
	rx = regexp.MustCompile(`(?m:^\+\+\+ b/.*$)`)
	result = rx.ReplaceAll(result, []byte("+++ "+toPath))

	//colorize diff
	rules := []output.LineColorizingRule{
		{[]byte("diff "), []byte("\x1B[1m")},
		{[]byte("new "), []byte("\x1B[1m")},
		{[]byte("deleted "), []byte("\x1B[1m")},
		{[]byte("--- "), []byte("\x1B[1m")},
		{[]byte("+++ "), []byte("\x1B[1m")},
		{[]byte("@@ "), []byte("\x1B[36m")},
		{[]byte("-"), []byte("\x1B[31m")},
		{[]byte("+"), []byte("\x1B[32m")},
	}

	return output.ColorizeLines(result, rules), nil
}

func checkFile(path string) (pathToUse string, returnError error) {
	if path == "/dev/null" {
		return path, nil
	}

	//check that files are either non-existent (in which case git-diff needs to
	//be given /dev/null instead or manageable (e.g. we can't diff directories
	//or device files)
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "/dev/null", nil
		}
		return path, err
	}

	//can only diff regular files and symlinks
	switch {
	case info.Mode().IsRegular():
		return path, nil //regular file is ok
	case (info.Mode() & os.ModeType) == os.ModeSymlink:
		return path, nil //symlink is ok
	default:
		return path, fmt.Errorf("file %s has wrong file type", path)
	}

}
