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
)

//ApplyTo applies this RepoFile to a file buffer, as part of the `holo apply`
//algorithm. Regular repofiles will replace the file buffer, while a holoscript
//will be executed on the file buffer to obtain the new buffer.
func (file RepoFile) ApplyTo(buffer *FileBuffer) (*FileBuffer, error) {
	if file.ApplicationStrategy() == "apply" {
		return NewFileBuffer(file.Path(), buffer.BasePath)
	}

	//application of a holoscript requires file contents
	buffer, err := buffer.ResolveSymlink()
	if err != nil {
		return nil, err
	}

	//run command, fetch result file into buffer (not into the targetPath
	//directly, in order not to corrupt the file there if the script run fails)
	var stdout bytes.Buffer
	cmd := exec.Command(file.Path())
	cmd.Stdin = bytes.NewBuffer(buffer.Contents)
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("execution of %s failed: %s", file.Path(), err.Error())
	}

	//result is the stdout of the script
	return NewFileBufferFromContents(stdout.Bytes(), buffer.BasePath), nil
}
