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
	"os"
	"os/exec"
	"strings"

	"github.com/holocm/holo/cmd/holo-files/filesplugin/fileutil"
)

func (repoFile RepoFile) ApplyTo(buf fileutil.FileBuffer, stdout, stderr io.Writer) (out fileutil.FileBuffer, err error) {
	if strings.HasSuffix(repoFile.Path, ".holoscript") {
		// script //////////////////////////////////////////////

		// this application strategy requires file contents
		buf, err = buf.ResolveSymlink()
		if err != nil {
			return fileutil.FileBuffer{}, err
		}

		// run command, fetch result file into out (not into
		// the targetPath directly, in order not to corrupt
		// the file there if the script run fails)
		var stdout bytes.Buffer
		cmd := exec.Command(repoFile.Path)
		cmd.Stdin = strings.NewReader(buf.Contents)
		cmd.Stdout = &stdout
		cmd.Stderr = stderr
		err = cmd.Run()
		if err != nil {
			return fileutil.FileBuffer{}, fmt.Errorf("execution of %s failed: %s", repoFile.Path, err.Error())
		}

		// result is the stdout of the script
		buf.Mode &^= os.ModeType
		buf.Contents = stdout.String()
		return buf, nil
	} else {
		// file ////////////////////////////////////////////////

		// Replace the FileBuffer's filetype and contents with
		// that of the repofile.
		repobuf, err := fileutil.NewFileBuffer(repoFile.Path, false)
		if err != nil {
			return fileutil.FileBuffer{}, err
		}
		buf.Mode = (buf.Mode &^ os.ModeType) | (repobuf.Mode & os.ModeType)
		buf.Contents = repobuf.Contents
		return buf, nil
	}
}
