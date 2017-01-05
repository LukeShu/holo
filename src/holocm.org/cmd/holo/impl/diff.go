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
	"os/exec"
	"regexp"
	"syscall"

	"holocm.org/cmd/holo/output"
)

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
