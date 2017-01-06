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

package filesplugin

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"

	"holocm.org/plugins/filesplugin/fileutil"
)

func (repoFile RepoFile) ApplyTo(in *fileutil.FileBuffer, stdout, stderr io.Writer) (out *fileutil.FileBuffer, err error) {
	if repoFile.ApplicationStrategy() == "passthru" {
		// script //////////////////////////////////////////////

		// this application strategy requires file contents
		in, err = in.ResolveSymlink()
		if err != nil {
			return nil, err
		}

		// run command, fetch result file into out (not into
		// the targetPath directly, in order not to corrupt
		// the file there if the script run fails)
		var stdout bytes.Buffer
		cmd := exec.Command(repoFile.Path())
		cmd.Stdin = bytes.NewBuffer(in.Contents)
		cmd.Stdout = &stdout
		cmd.Stderr = stderr
		err = cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("execution of %s failed: %s", repoFile.Path(), err.Error())
		}

		// result is the stdout of the script
		return fileutil.NewFileBufferFromContents(stdout.Bytes(), in.BasePath), nil
	} else {
		// file ////////////////////////////////////////////////

		// if the repo contains a plain file (or symlink), the file
		// buffer is replaced by it, thus ignoring the target base (or
		// any previous application steps)
		return fileutil.NewFileBuffer(repoFile.Path(), in.BasePath)
	}
}
