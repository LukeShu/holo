/*******************************************************************************
*
* Copyright 2016 Stefan Majewsky <majewsky@gmx.net>
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
	"fmt"
	"os"

	"holocm.org/cmd/holo/output"
)

type PidFile struct {
	file *os.File
}

// AcquirePidFile will create a pid file to ensure that only one
// instance of Holo is running at the same time.  Returns whether the
// lock was successfully aquired.
func AcquirePidFile(pidPath string) *PidFile {
	file, err := os.OpenFile(pidPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		output.Errorf(output.Stderr, "Cannot create pid file %s: %s", pidPath, err.Error())
		if os.IsExist(err) {
			fmt.Fprintln(output.Stderr, "This usually means that another instance of Holo is currently running.")
			fmt.Fprintln(output.Stderr, "If not, you can try to delete the pid file manually.")
		}
		return nil
	}
	fmt.Fprintf(file, "%d\n", os.Getpid())
	file.Sync()
	return &PidFile{file}
}

// Release removes the pid file created by AcquirePidFile.
func (pidFile *PidFile) Release() {
	pidPath := pidFile.file.Name()
	err := pidFile.file.Close()
	if err != nil {
		output.Errorf(output.Stderr, err.Error())
	}
	err = os.Remove(pidPath)
	if err != nil {
		output.Errorf(output.Stderr, err.Error())
	}
}
